package composition_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/component"
	"github.com/BuddhiLW/AutoPDF/pkg/composition"
	"github.com/BuddhiLW/AutoPDF/pkg/document"
)

func makeDefinition(t *testing.T, key component.Key, mode document.CompositionMode, renderer component.Renderer) component.Definition {
	t.Helper()
	definition, err := component.NewDefinition(key, mode, json.RawMessage(`{"type":"object"}`), document.Props{"content": "default"}, nil, renderer)
	if err != nil {
		t.Fatal(err)
	}
	return definition
}

func makeComposer(t *testing.T, maxWorkers int, renderer component.Renderer) *composition.Composer {
	t.Helper()
	definitions := []component.Definition{
		makeDefinition(t, component.Key{Kind: "node", Variant: "flow", Target: "latex"}, document.Flow, renderer),
		makeDefinition(t, component.Key{Kind: "node", Variant: "section", Target: "latex"}, document.Section, renderer),
	}
	catalog, err := component.NewCatalog(definitions...)
	if err != nil {
		t.Fatal(err)
	}
	composer, err := composition.New(catalog, "latex", composition.Options{MaxWorkers: maxWorkers, Cache: composition.NewMemoryCache()})
	if err != nil {
		t.Fatal(err)
	}
	return composer
}

func specWithTree(value string) document.DocumentSpec {
	return document.DocumentSpec{
		SchemaVersion: document.CurrentSchemaVersion,
		ID:            "doc",
		Theme:         "theme@1",
		Blocks: []document.Component{{
			ID: "parent", Kind: "node", Variant: "section", Props: document.Props{"content": "parent"},
			Children: []document.Component{
				{ID: "first", Kind: "node", Variant: "flow", Props: document.Props{"content": value}},
				{ID: "second", Kind: "node", Variant: "flow", Props: document.Props{"content": "second"}},
			},
		}},
	}
}

func TestComposeDeterministicOrderCacheAndDirtyPropagation(t *testing.T) {
	var renders atomic.Int64
	renderer := func(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
		renders.Add(1)
		time.Sleep(time.Duration(len(node.ID)%3) * time.Millisecond)
		return component.Output{Content: []byte(fmt.Sprint(node.Props["content"]))}, nil
	}
	composer := makeComposer(t, 3, renderer)
	first, err := composer.Compose(context.Background(), specWithTree("one"))
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{first.Fragments[0].ComponentID, first.Fragments[1].ComponentID, first.Fragments[2].ComponentID}; !reflect.DeepEqual(got, []string{"parent", "first", "second"}) {
		t.Fatalf("unstable order: %v", got)
	}
	second, err := composer.Compose(context.Background(), specWithTree("one"))
	if err != nil {
		t.Fatal(err)
	}
	if second.Stats.CacheHits != 3 || renders.Load() != 3 || first.Hash != second.Hash {
		t.Fatalf("cache mismatch: stats=%+v renders=%d", second.Stats, renders.Load())
	}
	changed, err := composer.Compose(context.Background(), specWithTree("changed"))
	if err != nil {
		t.Fatal(err)
	}
	if dirty := composition.DirtySet(second, changed); !reflect.DeepEqual(dirty, []string{"first", "parent"}) {
		t.Fatalf("dirty propagation mismatch: %v", dirty)
	}
}

func TestComposeHonorsWorkerBound(t *testing.T) {
	var active atomic.Int64
	var peak atomic.Int64
	renderer := func(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
		current := active.Add(1)
		defer active.Add(-1)
		for {
			old := peak.Load()
			if current <= old || peak.CompareAndSwap(old, current) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		return component.Output{Content: []byte(node.ID)}, nil
	}
	composer := makeComposer(t, 2, renderer)
	spec := document.DocumentSpec{SchemaVersion: 1}
	for index := 0; index < 12; index++ {
		spec.Blocks = append(spec.Blocks, document.Component{ID: fmt.Sprintf("n-%d", index), Kind: "node", Variant: "flow"})
	}
	if _, err := composer.Compose(context.Background(), spec); err != nil {
		t.Fatal(err)
	}
	if peak.Load() > 2 || peak.Load() < 2 {
		t.Fatalf("expected peak 2, got %d", peak.Load())
	}
}

func TestComposeCancellationAndUnknownDefinition(t *testing.T) {
	started := make(chan struct{})
	renderer := func(ctx context.Context, _ component.RenderContext, _ document.Component) (component.Output, error) {
		close(started)
		<-ctx.Done()
		return component.Output{}, ctx.Err()
	}
	composer := makeComposer(t, 1, renderer)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := composer.Compose(ctx, document.DocumentSpec{SchemaVersion: 1, Blocks: []document.Component{{ID: "x", Kind: "node", Variant: "flow"}}})
		done <- err
	}()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}

	composer = makeComposer(t, 1, func(context.Context, component.RenderContext, document.Component) (component.Output, error) {
		return component.Output{}, nil
	})
	_, err := composer.Compose(context.Background(), document.DocumentSpec{SchemaVersion: 1, Blocks: []document.Component{{ID: "x", Kind: "missing", Variant: "default"}}})
	if !errors.Is(err, component.ErrDefinitionAbsent) {
		t.Fatalf("expected exact dispatch failure, got %v", err)
	}
}
