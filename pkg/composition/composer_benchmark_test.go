package composition_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/composition"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
)

type discardCache struct{}

func (discardCache) Get(string) (composition.Fragment, bool) { return composition.Fragment{}, false }
func (discardCache) Put(string, composition.Fragment)        {}

func benchmarkDocument(size int) document.DocumentSpec {
	spec := document.DocumentSpec{SchemaVersion: document.CurrentSchemaVersion, ID: "benchmark"}
	for index := 0; index < size; index++ {
		spec.Blocks = append(spec.Blocks, document.Component{
			ID: fmt.Sprintf("component-%03d", index), Kind: "node", Variant: "flow",
			Props: document.Props{"content": "benchmark content"},
		})
	}
	return spec
}

func benchmarkComposer(b *testing.B, cache composition.Cache) *composition.Composer {
	b.Helper()
	definition, err := component.NewDefinition(component.Key{Kind: "node", Variant: "flow", Target: "latex"}, document.Flow, json.RawMessage(`{"type":"object"}`), nil, nil,
		func(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
			return component.Output{Content: []byte(node.Props["content"].(string))}, nil
		})
	if err != nil {
		b.Fatal(err)
	}
	catalog, err := component.NewCatalog(definition)
	if err != nil {
		b.Fatal(err)
	}
	composer, err := composition.New(catalog, "latex", composition.Options{MaxWorkers: 8, Cache: cache})
	if err != nil {
		b.Fatal(err)
	}
	return composer
}

func BenchmarkComposeCold100(b *testing.B) {
	composer := benchmarkComposer(b, discardCache{})
	spec := benchmarkDocument(100)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := composer.Compose(context.Background(), spec); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkComposeWarm100(b *testing.B) {
	composer := benchmarkComposer(b, composition.NewMemoryCache())
	spec := benchmarkDocument(100)
	if _, err := composer.Compose(context.Background(), spec); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := composer.Compose(context.Background(), spec); err != nil {
			b.Fatal(err)
		}
	}
}
