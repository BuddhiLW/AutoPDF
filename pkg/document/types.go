package document

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// CurrentSchemaVersion is the latest DocumentSpec wire format understood by
// this package.
const CurrentSchemaVersion = 1

// CompositionMode describes layout coupling. Component definitions choose the
// mode; users only choose semantic kinds and variants.
type CompositionMode string

const (
	// Flow participates in surrounding pagination and is normally projected as
	// a LaTeX \input fragment.
	Flow CompositionMode = "flow"
	// Section has explicit page boundaries and may be projected as a LaTeX
	// \include unit for focused previews.
	Section CompositionMode = "section"
	// Artifact is independently renderable and may be cached or embedded.
	Artifact CompositionMode = "artifact"
)

// Valid reports whether mode is a supported composition mode.
func (mode CompositionMode) Valid() bool {
	switch mode {
	case Flow, Section, Artifact:
		return true
	default:
		return false
	}
}

// Props contains JSON-shaped semantic component input.
type Props map[string]any

// Style contains JSON-shaped visual choices. Renderers interpret these values;
// the document domain does not contain target-specific commands.
type Style map[string]any

// AssetRef identifies an asset without coupling the domain to a filesystem or
// transport. A collector resolves URI at the effect boundary.
type AssetRef struct {
	ID        string `json:"id"`
	URI       string `json:"uri,omitempty"`
	MediaType string `json:"mediaType,omitempty"`
	Digest    string `json:"digest,omitempty"`
}

// Component is one ordered node in a document tree.
type Component struct {
	ID       string      `json:"id"`
	Kind     string      `json:"kind"`
	Variant  string      `json:"variant"`
	Props    Props       `json:"props,omitempty"`
	Style    Style       `json:"style,omitempty"`
	Assets   []AssetRef  `json:"assets,omitempty"`
	Children []Component `json:"children,omitempty"`
}

// DocumentSpec is the versioned, renderer-independent source document.
type DocumentSpec struct {
	SchemaVersion int         `json:"schemaVersion"`
	ID            string      `json:"id,omitempty"`
	Theme         string      `json:"theme,omitempty"`
	Blocks        []Component `json:"blocks"`
}

// Decode parses one strict JSON document and validates its domain invariants.
func Decode(data []byte) (DocumentSpec, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()

	var spec DocumentSpec
	if err := decoder.Decode(&spec); err != nil {
		return DocumentSpec{}, fmt.Errorf("decode document spec: %w", err)
	}
	if err := ensureEOF(decoder); err != nil {
		return DocumentSpec{}, err
	}
	if problems := spec.Problems(); len(problems) > 0 {
		return DocumentSpec{}, problems
	}
	return spec, nil
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return fmt.Errorf("decode trailing document data: %w", err)
	}
	return fmt.Errorf("decode document spec: multiple JSON values")
}

// Clone returns a deeply owned copy through the canonical JSON representation.
func (spec DocumentSpec) Clone() (DocumentSpec, error) {
	data, err := json.Marshal(spec)
	if err != nil {
		return DocumentSpec{}, fmt.Errorf("clone document spec: %w", err)
	}
	var clone DocumentSpec
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&clone); err != nil {
		return DocumentSpec{}, fmt.Errorf("clone document spec: %w", err)
	}
	return clone, nil
}
