// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package strict

import "strings"

// Reference is one site in a template that reads a variable off the root data.
//
// Line and Column are 1-based positions in the template source.
type Reference struct {
	// Key is the dotted path read at this site, e.g. "section.range.start".
	Key string
	// Line is the 1-based source line of the site.
	Line int
	// Column is the 1-based source column of the site.
	Column int
	// Waived reports that the site declared a blank here to be deliberate.
	Waived bool
	// Reason is the justification the site stated for the waiver.
	Reason string
	// Malformed reports that the site used the waiver builtin in a shape the
	// engine cannot honour.
	Malformed bool
}

// Bindings is the data a template would be executed against, addressed by
// dotted path.
type Bindings map[string]any

// Lookup resolves a dotted key. found is false when any segment of the path is
// missing, or when a non-final segment holds something that cannot be indexed
// by name.
func (b Bindings) Lookup(key string) (value any, found bool) {
	if b == nil {
		return nil, false
	}
	segments := strings.Split(key, ".")
	var current any = map[string]any(b)
	for _, segment := range segments {
		next, ok := index(current, segment)
		if !ok {
			return nil, false
		}
		current = next
	}
	return current, true
}

// index reads one named field off a container.
func index(container any, name string) (any, bool) {
	switch typed := container.(type) {
	case map[string]any:
		value, ok := typed[name]
		return value, ok
	case Bindings:
		value, ok := typed[name]
		return value, ok
	case map[string]string:
		value, ok := typed[name]
		if !ok {
			return nil, false
		}
		return value, true
	default:
		return nil, false
	}
}

// Set binds a dotted key, creating intermediate maps as needed. It reports
// false when a non-final segment already holds something that is not a map.
func (b Bindings) Set(key string, value any) bool {
	if b == nil {
		return false
	}
	segments := strings.Split(key, ".")
	current := map[string]any(b)
	for _, segment := range segments[:len(segments)-1] {
		existing, ok := current[segment]
		if !ok {
			created := make(map[string]any)
			current[segment] = created
			current = created
			continue
		}
		nested, isMap := existing.(map[string]any)
		if !isMap {
			return false
		}
		current = nested
	}
	current[segments[len(segments)-1]] = value
	return true
}
