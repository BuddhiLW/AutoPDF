package composition

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
)

// Fragment is one content-addressed projection of a component.
type Fragment struct {
	ComponentID  string                   `json:"componentId"`
	Key          component.Key            `json:"key"`
	Mode         document.CompositionMode `json:"mode"`
	Content      []byte                   `json:"content,omitempty"`
	Assets       []document.AssetRef      `json:"assets,omitempty"`
	Children     []string                 `json:"children,omitempty"`
	Dependencies []string                 `json:"dependencies,omitempty"`
	Hash         string                   `json:"hash"`
}

// SourceLocation maps user-facing component identity to generated material.
type SourceLocation struct {
	FragmentHash string `json:"fragmentHash"`
	Order        int    `json:"order"`
}

// Stats reports work avoidance without changing the manifest's semantics.
type Stats struct {
	Rendered  int `json:"rendered"`
	CacheHits int `json:"cacheHits"`
}

// RenderManifest is the stable boundary value produced by composition.
type RenderManifest struct {
	DocumentID   string                         `json:"documentId,omitempty"`
	Theme        string                         `json:"theme,omitempty"`
	Target       string                         `json:"target"`
	RootOrder    []string                       `json:"rootOrder"`
	Fragments    []Fragment                     `json:"fragments"`
	SourceMap    map[string]SourceLocation      `json:"sourceMap"`
	Dependencies map[string][]string            `json:"dependencies"`
	Sources      map[string][]document.AssetRef `json:"sources"`
	Hash         string                         `json:"hash"`
	Stats        Stats                          `json:"stats"`
}

func hashBytes(parts ...[]byte) string {
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write(part)
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}
