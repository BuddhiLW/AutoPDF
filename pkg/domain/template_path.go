package domain

import (
	"errors"
	"path/filepath"
	"strings"
)

// TemplatePath represents a template file path as a value object
type TemplatePath struct {
	value string
}

// TemplatePath errors
var (
	ErrEmptyPath                = errors.New("template path cannot be empty")
	ErrInvalidTemplateExtension = errors.New("template path must have .tex extension")
	ErrInvalidPath              = errors.New("template path contains invalid characters")
	ErrPathTooLong              = errors.New("template path cannot exceed 255 characters")
	ErrRelativePath             = errors.New("template path must be absolute or relative to current directory")
)

// NewTemplatePath creates a new TemplatePath value object with validation
func NewTemplatePath(path string) (TemplatePath, error) {
	trimmed := strings.TrimSpace(path)

	if trimmed == "" {
		return TemplatePath{}, ErrEmptyPath
	}

	if len(trimmed) > 255 {
		return TemplatePath{}, ErrPathTooLong
	}

	// Check for invalid characters
	if strings.ContainsAny(trimmed, "<>|*?\"") {
		return TemplatePath{}, ErrInvalidPath
	}

	// Check for valid extension
	ext := strings.ToLower(filepath.Ext(trimmed))
	if ext != ".tex" {
		return TemplatePath{}, ErrInvalidTemplateExtension
	}

	// Normalize the path
	normalized := filepath.Clean(trimmed)

	return TemplatePath{value: normalized}, nil
}

// String returns the string representation of the template path
func (tp TemplatePath) String() string {
	return tp.value
}

// Value returns the underlying string value
func (tp TemplatePath) Value() string {
	return tp.value
}

// IsEmpty returns true if the template path is empty
func (tp TemplatePath) IsEmpty() bool {
	return tp.value == ""
}

// Extension returns the file extension
func (tp TemplatePath) Extension() string {
	return filepath.Ext(tp.value)
}

// BaseName returns the base name of the file
func (tp TemplatePath) BaseName() string {
	return filepath.Base(tp.value)
}

// Dir returns the directory part of the path
func (tp TemplatePath) Dir() string {
	return filepath.Dir(tp.value)
}

// Equals compares two TemplatePath values for equality
func (tp TemplatePath) Equals(other TemplatePath) bool {
	return tp.value == other.value
}

// IsAbsolute returns true if the path is absolute
func (tp TemplatePath) IsAbsolute() bool {
	return filepath.IsAbs(tp.value)
}

// Join joins the template path with additional path elements
func (tp TemplatePath) Join(elem ...string) (TemplatePath, error) {
	joined := filepath.Join(tp.value, filepath.Join(elem...))
	return NewTemplatePath(joined)
}

// MarshalJSON implements json.Marshaler
func (tp TemplatePath) MarshalJSON() ([]byte, error) {
	return []byte(`"` + tp.value + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (tp *TemplatePath) UnmarshalJSON(data []byte) error {
	// Remove quotes and create new TemplatePath
	value := string(data)
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	path, err := NewTemplatePath(value)
	if err != nil {
		return err
	}

	*tp = path
	return nil
}
