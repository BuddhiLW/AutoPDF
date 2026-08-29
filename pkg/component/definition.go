package component

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/BuddhiLW/AutoPDF/pkg/document"
)

var (
	ErrInvalidKey       = errors.New("component key requires kind, variant, and target")
	ErrInvalidMode      = errors.New("component definition has invalid composition mode")
	ErrNilRenderer      = errors.New("component definition renderer must not be nil")
	ErrDuplicate        = errors.New("component definition already registered")
	ErrCatalogFrozen    = errors.New("component catalog builder is frozen")
	ErrDefinitionAbsent = errors.New("component definition not found")
)

// Key selects one component implementation. Dispatch is exact and never falls
// back silently to another variant or rendering target.
type Key struct {
	Kind    string `json:"kind"`
	Variant string `json:"variant"`
	Target  string `json:"target"`
}

// Valid reports whether all dispatch dimensions are present.
func (key Key) Valid() bool {
	return strings.TrimSpace(key.Kind) != "" &&
		strings.TrimSpace(key.Variant) != "" &&
		strings.TrimSpace(key.Target) != ""
}

func (key Key) String() string {
	return key.Kind + "/" + key.Variant + "@" + key.Target
}

// RenderContext contains inert values available to pure component renderers.
type RenderContext struct {
	Theme  string
	Values map[string]any
}

// Output is a component definition's target-specific result before the
// composition pipeline hashes and assembles it.
type Output struct {
	Content      []byte
	Assets       []document.AssetRef
	Dependencies []string
}

// Renderer projects one component. Implementations should be calculations;
// context exists for prompt cancellation, not hidden I/O.
type Renderer func(context.Context, RenderContext, document.Component) (Output, error)

// Validator returns field-addressable domain problems for definition-specific
// props and style rules.
type Validator func(document.Component) document.Problems

// Definition supplies metadata and behavior for one exact dispatch key.
type Definition interface {
	Key() Key
	Version() string
	Mode() document.CompositionMode
	Schema() json.RawMessage
	Defaults() document.Props
	Validate(document.Component) document.Problems
	Render(context.Context, RenderContext, document.Component) (Output, error)
}

// FuncDefinition is an immutable Definition backed by supplied functions.
type FuncDefinition struct {
	key       Key
	version   string
	mode      document.CompositionMode
	schema    json.RawMessage
	defaults  document.Props
	validator Validator
	renderer  Renderer
}

// NewDefinition validates and defensively owns definition metadata.
func NewDefinition(
	key Key,
	mode document.CompositionMode,
	schema json.RawMessage,
	defaults document.Props,
	validator Validator,
	renderer Renderer,
) (*FuncDefinition, error) {
	return NewVersionedDefinition(key, "1", mode, schema, defaults, validator, renderer)
}

// NewVersionedDefinition includes an explicit renderer/template revision in
// composition cache keys. Change version whenever rendering semantics change.
func NewVersionedDefinition(
	key Key,
	version string,
	mode document.CompositionMode,
	schema json.RawMessage,
	defaults document.Props,
	validator Validator,
	renderer Renderer,
) (*FuncDefinition, error) {
	if !key.Valid() {
		return nil, ErrInvalidKey
	}
	if strings.TrimSpace(version) == "" {
		return nil, fmt.Errorf("component definition version is required")
	}
	if !mode.Valid() {
		return nil, ErrInvalidMode
	}
	if renderer == nil {
		return nil, ErrNilRenderer
	}
	ownedDefaults, err := cloneProps(defaults)
	if err != nil {
		return nil, fmt.Errorf("clone component defaults: %w", err)
	}
	return &FuncDefinition{
		key:       key,
		version:   version,
		mode:      mode,
		schema:    append(json.RawMessage(nil), schema...),
		defaults:  ownedDefaults,
		validator: validator,
		renderer:  renderer,
	}, nil
}

func (definition *FuncDefinition) Key() Key { return definition.key }

func (definition *FuncDefinition) Version() string { return definition.version }

func (definition *FuncDefinition) Mode() document.CompositionMode { return definition.mode }

func (definition *FuncDefinition) Schema() json.RawMessage {
	return append(json.RawMessage(nil), definition.schema...)
}

func (definition *FuncDefinition) Defaults() document.Props {
	defaults, _ := cloneProps(definition.defaults)
	return defaults
}

func (definition *FuncDefinition) Validate(node document.Component) document.Problems {
	if definition.validator == nil {
		return nil
	}
	return append(document.Problems(nil), definition.validator(node)...)
}

func (definition *FuncDefinition) Render(ctx context.Context, renderContext RenderContext, node document.Component) (Output, error) {
	if err := ctx.Err(); err != nil {
		return Output{}, err
	}
	output, err := definition.renderer(ctx, renderContext, node)
	if err != nil {
		return Output{}, err
	}
	output.Content = append([]byte(nil), output.Content...)
	output.Assets = append([]document.AssetRef(nil), output.Assets...)
	output.Dependencies = append([]string(nil), output.Dependencies...)
	return output, nil
}

func cloneProps(props document.Props) (document.Props, error) {
	if props == nil {
		return nil, nil
	}
	data, err := json.Marshal(props)
	if err != nil {
		return nil, err
	}
	var clone document.Props
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err := decoder.Decode(&clone); err != nil {
		return nil, err
	}
	return clone, nil
}
