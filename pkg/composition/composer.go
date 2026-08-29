package composition

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/BuddhiLW/AutoPDF/pkg/component"
	"github.com/BuddhiLW/AutoPDF/pkg/document"
)

var (
	ErrCatalogRequired = errors.New("composition catalog is required")
	ErrTargetRequired  = errors.New("composition target is required")
)

// Options controls bounded execution and reuse.
type Options struct {
	MaxWorkers int
	Cache      Cache
	Values     map[string]any
}

// Composer promotes and renders immutable document values.
type Composer struct {
	catalog    *component.Catalog
	target     string
	maxWorkers int
	cache      Cache
	values     map[string]any
}

func New(catalog *component.Catalog, target string, options Options) (*Composer, error) {
	if catalog == nil {
		return nil, ErrCatalogRequired
	}
	if strings.TrimSpace(target) == "" {
		return nil, ErrTargetRequired
	}
	workers := options.MaxWorkers
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}
	if workers < 1 {
		workers = 1
	}
	cache := options.Cache
	if cache == nil {
		cache = NewMemoryCache()
	}
	values, err := cloneValues(options.Values)
	if err != nil {
		return nil, fmt.Errorf("clone composition values: %w", err)
	}
	return &Composer{catalog: catalog, target: target, maxWorkers: workers, cache: cache, values: values}, nil
}

type workItem struct {
	index        int
	node         document.Component
	dependencies []string
}

type workResult struct {
	fragment Fragment
	cacheHit bool
	err      error
}

// Compose performs validation, promotion, bounded rendering, and deterministic
// manifest assembly. Effects belong after this method.
func (composer *Composer) Compose(ctx context.Context, spec document.DocumentSpec) (RenderManifest, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return RenderManifest{}, err
	}
	if problems := spec.Problems(); len(problems) > 0 {
		return RenderManifest{}, problems
	}

	items := flatten(spec.Blocks)
	results := make([]workResult, len(items))
	jobs := make(chan workItem)
	var workers sync.WaitGroup
	workerCount := composer.maxWorkers
	if workerCount > len(items) {
		workerCount = len(items)
	}
	for worker := 0; worker < workerCount; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for item := range jobs {
				results[item.index] = composer.renderOne(ctx, spec.Theme, item)
			}
		}()
	}

	dispatchDone := false
	for _, item := range items {
		select {
		case <-ctx.Done():
			dispatchDone = true
		case jobs <- item:
		}
		if dispatchDone {
			break
		}
	}
	close(jobs)
	workers.Wait()
	if err := ctx.Err(); err != nil {
		return RenderManifest{}, err
	}

	manifest := RenderManifest{
		DocumentID:   spec.ID,
		Theme:        spec.Theme,
		Target:       composer.target,
		RootOrder:    make([]string, len(spec.Blocks)),
		Fragments:    make([]Fragment, 0, len(results)),
		SourceMap:    make(map[string]SourceLocation, len(results)),
		Dependencies: make(map[string][]string, len(results)),
		Sources:      make(map[string][]document.AssetRef, len(results)),
	}
	for index, block := range spec.Blocks {
		manifest.RootOrder[index] = block.ID
	}
	for index, result := range results {
		if result.err != nil {
			return RenderManifest{}, result.err
		}
		manifest.Fragments = append(manifest.Fragments, result.fragment)
		manifest.SourceMap[result.fragment.ComponentID] = SourceLocation{FragmentHash: result.fragment.Hash, Order: index}
		manifest.Dependencies[result.fragment.ComponentID] = append([]string(nil), result.fragment.Dependencies...)
		manifest.Sources[result.fragment.ComponentID] = append([]document.AssetRef(nil), result.fragment.Assets...)
		if result.cacheHit {
			manifest.Stats.CacheHits++
		} else {
			manifest.Stats.Rendered++
		}
	}
	manifest.Hash = hashManifest(manifest)
	return manifest, nil
}

func (composer *Composer) renderOne(ctx context.Context, theme string, item workItem) workResult {
	key := component.Key{Kind: item.node.Kind, Variant: item.node.Variant, Target: composer.target}
	definition, err := composer.catalog.Lookup(key)
	if err != nil {
		return workResult{err: fmt.Errorf("dispatch component %s: %w", item.node.ID, err)}
	}
	promoted, err := applyDefaults(item.node, definition.Defaults())
	if err != nil {
		return workResult{err: fmt.Errorf("promote component %s: %w", item.node.ID, err)}
	}
	if problems := definition.Validate(promoted); len(problems) > 0 {
		return workResult{err: problems}
	}
	inputHash, err := hashInput(theme, composer.values, promoted, key, definition.Version(), definition.Mode())
	if err != nil {
		return workResult{err: fmt.Errorf("hash component %s: %w", item.node.ID, err)}
	}
	if fragment, found := composer.cache.Get(inputHash); found {
		return workResult{fragment: fragment, cacheHit: true}
	}
	output, err := definition.Render(ctx, component.RenderContext{Theme: theme, Values: composer.values}, promoted)
	if err != nil {
		return workResult{err: fmt.Errorf("render component %s: %w", item.node.ID, err)}
	}
	dependencies := append([]string(nil), item.dependencies...)
	dependencies = append(dependencies, output.Dependencies...)
	dependencies = stableUnique(dependencies)
	fragment := Fragment{
		ComponentID:  promoted.ID,
		Key:          key,
		Mode:         definition.Mode(),
		Content:      output.Content,
		Assets:       output.Assets,
		Children:     append([]string(nil), item.dependencies...),
		Dependencies: dependencies,
	}
	fragment.Hash = hashFragment(fragment)
	composer.cache.Put(inputHash, fragment)
	return workResult{fragment: fragment}
}

func flatten(roots []document.Component) []workItem {
	items := make([]workItem, 0)
	var visit func(document.Component)
	visit = func(node document.Component) {
		dependencies := make([]string, len(node.Children))
		for index, child := range node.Children {
			dependencies[index] = child.ID
		}
		items = append(items, workItem{index: len(items), node: node, dependencies: dependencies})
		for _, child := range node.Children {
			visit(child)
		}
	}
	for _, root := range roots {
		visit(root)
	}
	return items
}

func applyDefaults(node document.Component, defaults document.Props) (document.Component, error) {
	cloneSpec, err := (document.DocumentSpec{SchemaVersion: document.CurrentSchemaVersion, Blocks: []document.Component{node}}).Clone()
	if err != nil {
		return document.Component{}, err
	}
	promoted := cloneSpec.Blocks[0]
	if promoted.Props == nil {
		promoted.Props = make(document.Props)
	}
	for key, value := range defaults {
		if _, exists := promoted.Props[key]; !exists {
			promoted.Props[key] = value
		}
	}
	return promoted, nil
}

func cloneValues(values map[string]any) (map[string]any, error) {
	if values == nil {
		return nil, nil
	}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	var clone map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err := decoder.Decode(&clone); err != nil {
		return nil, err
	}
	return clone, nil
}

func hashInput(theme string, values map[string]any, node document.Component, key component.Key, definitionVersion string, mode document.CompositionMode) (string, error) {
	data, err := json.Marshal(struct {
		Theme   string
		Values  map[string]any
		Node    document.Component
		Key     component.Key
		Version string
		Mode    document.CompositionMode
	}{theme, values, node, key, definitionVersion, mode})
	if err != nil {
		return "", err
	}
	return hashBytes(data), nil
}

func hashFragment(fragment Fragment) string {
	data, _ := json.Marshal(struct {
		ID           string
		Key          component.Key
		Mode         document.CompositionMode
		Content      []byte
		Assets       []document.AssetRef
		Children     []string
		Dependencies []string
	}{fragment.ComponentID, fragment.Key, fragment.Mode, fragment.Content, fragment.Assets, fragment.Children, fragment.Dependencies})
	return hashBytes(data)
}

func hashManifest(manifest RenderManifest) string {
	parts := make([][]byte, 0, len(manifest.Fragments)+3)
	parts = append(parts, []byte(manifest.DocumentID), []byte(manifest.Theme), []byte(manifest.Target))
	for _, fragment := range manifest.Fragments {
		parts = append(parts, []byte(fragment.Hash))
	}
	return hashBytes(parts...)
}

func stableUnique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
