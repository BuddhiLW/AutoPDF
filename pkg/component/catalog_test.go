package component_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/component"
	"github.com/BuddhiLW/AutoPDF/pkg/document"
)

func definition(t *testing.T, key component.Key, mode document.CompositionMode, content string) component.Definition {
	t.Helper()
	definition, err := component.NewDefinition(
		key,
		mode,
		json.RawMessage(`{"type":"object"}`),
		document.Props{"content": content},
		nil,
		func(context.Context, component.RenderContext, document.Component) (component.Output, error) {
			return component.Output{Content: []byte(content)}, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return definition
}

func TestCatalogExactDispatchAndFreeze(t *testing.T) {
	key := component.Key{Kind: "table", Variant: "long", Target: "latex"}
	def := definition(t, key, document.Section, "long-table")
	builder := component.NewBuilder()
	if err := builder.Register(def); err != nil {
		t.Fatal(err)
	}
	if err := builder.Register(def); !errors.Is(err, component.ErrDuplicate) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	catalog, err := builder.Freeze()
	if err != nil {
		t.Fatal(err)
	}
	if err := builder.Register(definition(t, component.Key{Kind: "x", Variant: "y", Target: "z"}, document.Flow, "x")); !errors.Is(err, component.ErrCatalogFrozen) {
		t.Fatalf("expected frozen error, got %v", err)
	}

	got, err := catalog.Lookup(key)
	if err != nil {
		t.Fatal(err)
	}
	output, err := got.Render(context.Background(), component.RenderContext{}, document.Component{ID: "table"})
	if err != nil || string(output.Content) != "long-table" {
		t.Fatalf("unexpected dispatch output %q, %v", output.Content, err)
	}
	if _, err := catalog.Lookup(component.Key{Kind: "table", Variant: "simple", Target: "latex"}); !errors.Is(err, component.ErrDefinitionAbsent) {
		t.Fatalf("expected absent definition, got %v", err)
	}
}

func TestCatalogMetadataIsStableAndOwned(t *testing.T) {
	a := definition(t, component.Key{Kind: "z", Variant: "default", Target: "latex"}, document.Flow, "z")
	b := definition(t, component.Key{Kind: "a", Variant: "default", Target: "latex"}, document.Section, "a")
	catalog, err := component.NewCatalog(a, b)
	if err != nil {
		t.Fatal(err)
	}
	metadata := catalog.Metadata()
	if metadata[0].Key.Kind != "a" || metadata[1].Key.Kind != "z" {
		t.Fatalf("metadata not sorted: %#v", metadata)
	}
	metadata[0].Defaults["content"] = "mutated"
	if catalog.Metadata()[0].Defaults["content"] != "a" {
		t.Fatal("metadata borrowed definition defaults")
	}
}

func TestCatalogConcurrentLookup(t *testing.T) {
	defs, err := component.Builtins("latex")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := component.NewCatalog(defs...)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for index := 0; index < 100; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			definition, lookupErr := catalog.Lookup(component.Key{Kind: "text", Variant: "default", Target: "latex"})
			if lookupErr != nil {
				t.Errorf("lookup: %v", lookupErr)
				return
			}
			if definition.Mode() != document.Flow {
				t.Errorf("unexpected mode %s", definition.Mode())
			}
		}()
	}
	wg.Wait()
}

func TestBuiltinsCoverEveryCompositionMode(t *testing.T) {
	definitions, err := component.Builtins("latex")
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[document.CompositionMode]bool)
	for _, definition := range definitions {
		seen[definition.Mode()] = true
	}
	for _, mode := range []document.CompositionMode{document.Flow, document.Section, document.Artifact} {
		if !seen[mode] {
			t.Fatalf("missing builtin mode %s", mode)
		}
	}
}
