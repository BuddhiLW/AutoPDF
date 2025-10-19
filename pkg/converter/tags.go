// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"strings"
)

// FieldTag represents parsed autopdf struct tag
type FieldTag struct {
	Name      string   // Field name in template
	Omit      bool     // Skip this field (for "-" tag)
	OmitEmpty bool     // Omit if empty value
	Flatten   bool     // Flatten nested structures
	Inline    bool     // Inline nested struct fields
	Options   []string // Additional options
}

// ParseTag parses "autopdf" struct tag
// Supports formats:
//
//	autopdf:"field_name"
//	autopdf:"field_name,omitempty"
//	autopdf:"field_name,flatten"
//	autopdf:"field_name,inline"
//	autopdf:"-" (skip field)
func ParseTag(tag string) FieldTag {
	if tag == "" {
		return FieldTag{
			Options: []string{},
		}
	}

	// Handle skip field
	if tag == "-" {
		return FieldTag{
			Omit: true,
		}
	}

	// Split by comma to get name and options
	parts := strings.Split(tag, ",")

	fieldTag := FieldTag{
		Name:    strings.TrimSpace(parts[0]),
		Options: []string{},
	}

	// Parse options
	for i := 1; i < len(parts); i++ {
		option := strings.TrimSpace(parts[i])
		fieldTag.Options = append(fieldTag.Options, option)

		switch option {
		case "omitempty":
			fieldTag.OmitEmpty = true
		case "flatten":
			fieldTag.Flatten = true
		case "inline":
			fieldTag.Inline = true
		}
	}

	return fieldTag
}

// String returns the string representation of the tag
func (ft FieldTag) String() string {
	if ft.Omit && ft.Name == "" {
		return "-"
	}

	var parts []string
	if ft.Name != "" {
		parts = append(parts, ft.Name)
	}

	parts = append(parts, ft.Options...)

	return strings.Join(parts, ",")
}

// HasOption checks if the tag has a specific option
func (ft FieldTag) HasOption(option string) bool {
	for _, opt := range ft.Options {
		if opt == option {
			return true
		}
	}
	return false
}

// IsEmpty returns true if the tag is empty (no name and no options)
func (ft FieldTag) IsEmpty() bool {
	return ft.Name == "" && len(ft.Options) == 0 && !ft.Omit
}

// IsValid returns true if the tag is valid
func (ft FieldTag) IsValid() bool {
	// Tag is valid if it has a name or is marked to omit
	return ft.Name != "" || ft.Omit
}
