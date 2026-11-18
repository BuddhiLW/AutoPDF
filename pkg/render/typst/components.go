package typst

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/escape"
)

// Target is the dispatch target these definitions register under and the
// backend name the projector answers to.
const Target = "typst"

// builtinSchema documents the props every builtin kind reads.
const builtinSchema = `{"type":"object","properties":{"content":{"type":"string"},"title":{"type":"string"}}}`

// Definitions returns the Typst projections of the builtin kinds of
// pkg/component: text (flow), section (section) and artifact (artifact).
// Text props are escaped as Typst markup; they are never interpreted.
func Definitions() ([]component.Definition, error) {
	specs := []struct {
		kind   string
		mode   document.CompositionMode
		render component.Renderer
	}{
		{kind: "text", mode: document.Flow, render: renderText},
		{kind: "section", mode: document.Section, render: renderSection},
		{kind: "artifact", mode: document.Artifact, render: renderText},
	}
	definitions := make([]component.Definition, 0, len(specs))
	for _, spec := range specs {
		definition, err := component.NewDefinition(
			component.Key{Kind: spec.kind, Variant: "default", Target: Target},
			spec.mode,
			json.RawMessage(builtinSchema),
			document.Props{"content": ""},
			validateStringProps,
			spec.render,
		)
		if err != nil {
			return nil, fmt.Errorf("define typst component %q: %w", spec.kind, err)
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

// DefaultKeys are the dispatch keys the projector accepts by default: exactly
// the keys of Definitions.
func DefaultKeys() []component.Key {
	definitions, err := Definitions()
	if err != nil {
		return nil
	}
	keys := make([]component.Key, 0, len(definitions))
	for _, definition := range definitions {
		keys = append(keys, definition.Key())
	}
	return keys
}

// Register installs every Typst definition into a catalog builder.
func Register(builder *component.Builder) error {
	definitions, err := Definitions()
	if err != nil {
		return err
	}
	for _, definition := range definitions {
		if err := builder.Register(definition); err != nil {
			return err
		}
	}
	return nil
}

// Catalog freezes a catalog holding exactly the Typst definitions.
func Catalog() (*component.Catalog, error) {
	builder := component.NewBuilder()
	if err := Register(builder); err != nil {
		return nil, err
	}
	return builder.Freeze()
}

func validateStringProps(node document.Component) document.Problems {
	var problems document.Problems
	for _, key := range []string{"content", "title"} {
		value, exists := node.Props[key]
		if !exists {
			continue
		}
		if _, ok := value.(string); !ok {
			problems = append(problems, document.Problem{
				ComponentID: node.ID,
				Path:        "props." + key,
				Code:        "component." + key + ".invalid",
				Message:     key + " must be a string",
			})
		}
	}
	return problems
}

func propString(node document.Component, key string) string {
	value, _ := node.Props[key].(string)
	return value
}

func paragraph(text string) string {
	if text == "" {
		return ""
	}
	return escape.Typst(text) + "\n\n"
}

func renderText(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	return component.Output{Content: []byte(paragraph(propString(node, "content"))), Assets: node.Assets}, nil
}

func renderSection(_ context.Context, _ component.RenderContext, node document.Component) (component.Output, error) {
	var content string
	if title := propString(node, "title"); title != "" {
		content = "= " + escape.Typst(title) + "\n"
	}
	content += paragraph(propString(node, "content"))
	return component.Output{Content: []byte(content), Assets: node.Assets}, nil
}
