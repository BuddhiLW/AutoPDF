package domain

import (
	"errors"
	"path/filepath"
	"strings"
)

// OutputPath represents an output file path as a value object
type OutputPath struct {
	value string
}

// OutputPath errors
var (
	ErrEmptyOutputPath        = errors.New("output path cannot be empty")
	ErrInvalidOutputExtension = errors.New("output path must have .pdf extension")
	ErrInvalidOutputPath      = errors.New("output path contains invalid characters")
	ErrOutputPathTooLong      = errors.New("output path cannot exceed 255 characters")
)

// NewOutputPath creates a new OutputPath value object with validation
func NewOutputPath(path string) (OutputPath, error) {
	trimmed := strings.TrimSpace(path)

	if trimmed == "" {
		return OutputPath{}, ErrEmptyOutputPath
	}

	if len(trimmed) > 255 {
		return OutputPath{}, ErrOutputPathTooLong
	}

	// Check for invalid characters
	if strings.ContainsAny(trimmed, "<>|*?\"") {
		return OutputPath{}, ErrInvalidOutputPath
	}

	// Check for valid extension
	ext := strings.ToLower(filepath.Ext(trimmed))
	if ext != ".pdf" {
		return OutputPath{}, ErrInvalidOutputExtension
	}

	// Normalize the path
	normalized := filepath.Clean(trimmed)

	return OutputPath{value: normalized}, nil
}

// String returns the string representation of the output path
func (op OutputPath) String() string {
	return op.value
}

// Value returns the underlying string value
func (op OutputPath) Value() string {
	return op.value
}

// IsEmpty returns true if the output path is empty
func (op OutputPath) IsEmpty() bool {
	return op.value == ""
}

// Extension returns the file extension
func (op OutputPath) Extension() string {
	return filepath.Ext(op.value)
}

// BaseName returns the base name of the file
func (op OutputPath) BaseName() string {
	return filepath.Base(op.value)
}

// Dir returns the directory part of the path
func (op OutputPath) Dir() string {
	return filepath.Dir(op.value)
}

// Equals compares two OutputPath values for equality
func (op OutputPath) Equals(other OutputPath) bool {
	return op.value == other.value
}

// IsAbsolute returns true if the path is absolute
func (op OutputPath) IsAbsolute() bool {
	return filepath.IsAbs(op.value)
}

// Join joins the output path with additional path elements
func (op OutputPath) Join(elem ...string) (OutputPath, error) {
	joined := filepath.Join(op.value, filepath.Join(elem...))
	return NewOutputPath(joined)
}

// WithExtension creates a new OutputPath with a different extension
func (op OutputPath) WithExtension(ext string) (OutputPath, error) {
	base := strings.TrimSuffix(op.value, filepath.Ext(op.value))
	newPath := base + ext
	return NewOutputPath(newPath)
}

// MarshalJSON implements json.Marshaler
func (op OutputPath) MarshalJSON() ([]byte, error) {
	return []byte(`"` + op.value + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (op *OutputPath) UnmarshalJSON(data []byte) error {
	// Remove quotes and create new OutputPath
	value := string(data)
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	path, err := NewOutputPath(value)
	if err != nil {
		return err
	}

	*op = path
	return nil
}
