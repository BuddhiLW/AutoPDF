package component

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/pkg/document"
)

// Builtins returns minimal target-neutral definitions covering each layout
// coupling class. Applications can replace or extend them before freezing a
// catalog.
func Builtins(target string) ([]Definition, error) {
	types := []struct {
		key  Key
		mode document.CompositionMode
	}{
		{key: Key{Kind: "text", Variant: "default", Target: target}, mode: document.Flow},
		{key: Key{Kind: "section", Variant: "default", Target: target}, mode: document.Section},
		{key: Key{Kind: "artifact", Variant: "default", Target: target}, mode: document.Artifact},
	}
	definitions := make([]Definition, 0, len(types))
	for _, item := range types {
		definition, err := NewDefinition(
			item.key,
			item.mode,
			json.RawMessage(`{"type":"object","properties":{"content":{"type":"string"}}}`),
			document.Props{"content": ""},
			validateBuiltinContent,
			renderBuiltinContent,
		)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func validateBuiltinContent(node document.Component) document.Problems {
	if value, exists := node.Props["content"]; exists {
		if _, ok := value.(string); !ok {
			return document.Problems{{
				ComponentID: node.ID,
				Path:        "props.content",
				Code:        "component.content.invalid",
				Message:     "content must be a string",
			}}
		}
	}
	return nil
}

func renderBuiltinContent(_ context.Context, _ RenderContext, node document.Component) (Output, error) {
	value, exists := node.Props["content"]
	if !exists {
		return Output{Assets: node.Assets}, nil
	}
	content, ok := value.(string)
	if !ok {
		return Output{}, fmt.Errorf("component %s content must be a string", node.ID)
	}
	return Output{Content: []byte(content), Assets: node.Assets}, nil
}
