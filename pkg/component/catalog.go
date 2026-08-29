package component

import (
	"fmt"
	"sort"
	"sync"
)

// Metadata is the serializable catalog surface suitable for editor discovery.
type Metadata struct {
	Key      Key            `json:"key"`
	Version  string         `json:"version"`
	Mode     string         `json:"mode"`
	Schema   []byte         `json:"schema,omitempty"`
	Defaults map[string]any `json:"defaults,omitempty"`
}

// Builder owns mutable registration. Freeze creates a read-only Catalog and
// permanently closes the builder against later mutation.
type Builder struct {
	mu     sync.Mutex
	defs   map[Key]Definition
	frozen bool
}

func NewBuilder() *Builder {
	return &Builder{defs: make(map[Key]Definition)}
}

func (builder *Builder) Register(definition Definition) error {
	if definition == nil {
		return fmt.Errorf("register component: nil definition")
	}
	key := definition.Key()
	if !key.Valid() {
		return ErrInvalidKey
	}
	if !definition.Mode().Valid() {
		return ErrInvalidMode
	}

	builder.mu.Lock()
	defer builder.mu.Unlock()
	if builder.frozen {
		return ErrCatalogFrozen
	}
	if _, exists := builder.defs[key]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicate, key)
	}
	builder.defs[key] = definition
	return nil
}

func (builder *Builder) Freeze() (*Catalog, error) {
	builder.mu.Lock()
	defer builder.mu.Unlock()
	if builder.frozen {
		return nil, ErrCatalogFrozen
	}
	builder.frozen = true
	definitions := make(map[Key]Definition, len(builder.defs))
	for key, definition := range builder.defs {
		definitions[key] = definition
	}
	return &Catalog{definitions: definitions}, nil
}

// Catalog is immutable after construction and safe for concurrent readers.
type Catalog struct {
	definitions map[Key]Definition
}

// NewCatalog registers and freezes definitions in one operation.
func NewCatalog(definitions ...Definition) (*Catalog, error) {
	builder := NewBuilder()
	for _, definition := range definitions {
		if err := builder.Register(definition); err != nil {
			return nil, err
		}
	}
	return builder.Freeze()
}

// Lookup performs exact multi-dispatch.
func (catalog *Catalog) Lookup(key Key) (Definition, error) {
	if catalog == nil {
		return nil, fmt.Errorf("%w: %s", ErrDefinitionAbsent, key)
	}
	definition, exists := catalog.definitions[key]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrDefinitionAbsent, key)
	}
	return definition, nil
}

// Metadata returns a stable defensive snapshot for web editor discovery.
func (catalog *Catalog) Metadata() []Metadata {
	if catalog == nil {
		return nil
	}
	keys := make([]Key, 0, len(catalog.definitions))
	for key := range catalog.definitions {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })

	metadata := make([]Metadata, 0, len(keys))
	for _, key := range keys {
		definition := catalog.definitions[key]
		metadata = append(metadata, Metadata{
			Key:      key,
			Version:  definition.Version(),
			Mode:     string(definition.Mode()),
			Schema:   append([]byte(nil), definition.Schema()...),
			Defaults: map[string]any(definition.Defaults()),
		})
	}
	return metadata
}
