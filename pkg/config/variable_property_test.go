// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"testing/quick"

	"gopkg.in/yaml.v3"
)

func generatedVariables(text string, number int32, enabled bool, items []uint8) *Variables {
	variables := NewVariables()
	variables.Set("text", &StringVariable{Value: text})
	variables.Set("number", &NumberVariable{Value: float64(number)})
	variables.Set("enabled", &BoolVariable{Value: enabled})

	metadata := NewMapVariable()
	metadata.Values["label"] = &StringVariable{Value: text}
	metadata.Values["count"] = &NumberVariable{Value: float64(number)}
	variables.Set("metadata", metadata)

	entries := NewSliceVariable()
	numbers := NewSliceVariable()
	for _, item := range items {
		entry := NewMapVariable()
		entry.Values["value"] = &NumberVariable{Value: float64(item)}
		entries.Values = append(entries.Values, entry)
		numbers.Values = append(numbers.Values, &NumberVariable{Value: float64(item)})
	}
	variables.Set("entries", entries)
	variables.Set("numbers", numbers)

	return variables
}

func sameLeafTypes(left, right *Variables) bool {
	for path := range left.Flatten() {
		leftValue, leftExists := left.GetByPath(path)
		rightValue, rightExists := right.GetByPath(path)
		if !leftExists || !rightExists || leftValue.Type() != rightValue.Type() {
			return false
		}
	}
	return true
}

func TestPropertyFlattenedPathsResolveToTheirLeaves(t *testing.T) {
	property := func(text string, number int32, enabled bool, items []uint8) bool {
		variables := generatedVariables(text, number, enabled, items)
		flattened := variables.Flatten()
		if len(flattened) != 5+(2*len(items)) {
			return false
		}

		for path, want := range flattened {
			got, exists := variables.GetString(path)
			if !exists || got != want {
				return false
			}
		}
		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 1_000}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyJSONRoundTripPreservesVariableMeaning(t *testing.T) {
	property := func(text string, number int32, enabled bool, items []uint8) bool {
		variables := generatedVariables(text, number, enabled, items)
		encoded, err := json.Marshal(variables)
		if err != nil {
			return false
		}

		var decoded Variables
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			return false
		}
		return reflect.DeepEqual(decoded.Flatten(), variables.Flatten()) && sameLeafTypes(variables, &decoded)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 500}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyYAMLRoundTripPreservesVariableMeaning(t *testing.T) {
	property := func(text string, number int32, enabled bool, items []uint8) bool {
		variables := generatedVariables(text, number, enabled, items)
		encoded, err := yaml.Marshal(variables)
		if err != nil {
			return false
		}

		var decoded Variables
		if err := yaml.Unmarshal(encoded, &decoded); err != nil {
			return false
		}
		return reflect.DeepEqual(decoded.Flatten(), variables.Flatten()) && sameLeafTypes(variables, &decoded)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 500}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyOutOfRangeIndexesNeverResolve(t *testing.T) {
	property := func(items []uint8, extra uint8) bool {
		variables := generatedVariables("value", 1, true, items)
		indexes := []int{-1, len(items), len(items) + int(extra) + 1}
		for _, index := range indexes {
			if _, exists := variables.GetString(fmt.Sprintf("numbers[%d]", index)); exists {
				return false
			}
		}
		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 1_000}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertySliceLookupObeysIndexDomain(t *testing.T) {
	property := func(items []uint8) bool {
		values := NewSliceVariable()
		for _, item := range items {
			values.Values = append(values.Values, &NumberVariable{Value: float64(item)})
		}

		self, exists := values.Get("")
		if !exists || self.Type() != VariableTypeSlice || self.Len() != len(items) {
			return false
		}
		if _, exists := values.Get("."); exists {
			return false
		}
		for index, want := range items {
			got, exists := values.Get(strconv.Itoa(index))
			if !exists || got.String() != strconv.Itoa(int(want)) {
				return false
			}
		}
		_, exists = values.Get(strconv.Itoa(len(items)))
		return !exists
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 1_000}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyNumberStringUsesLosslessDecimalForm(t *testing.T) {
	property := func(value int64) bool {
		number := float64(value) / 10
		got := (&NumberVariable{Value: number}).String()
		want := strconv.FormatFloat(number, 'f', -1, 64)
		return got == want
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 1_000}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyConfigPassesAreClampedToDomainRange(t *testing.T) {
	property := func(input int16) bool {
		config, err := NewConfigFromYAML([]byte(fmt.Sprintf("passes: %d", input)))
		if err != nil {
			return false
		}

		want := int(input)
		if want < 1 {
			want = 1
		}
		if want > 10 {
			want = 10
		}
		return config.Passes == want
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 1_000}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyMalformedYAMLMappingDoesNotPanic(t *testing.T) {
	property := func(values []string) bool {
		content := make([]*yaml.Node, 0, (2*len(values))+1)
		for index, value := range values {
			content = append(content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("key%d", index)},
				&yaml.Node{Kind: yaml.ScalarNode, Value: value},
			)
		}
		content = append(content, &yaml.Node{Kind: yaml.ScalarNode, Value: "orphan"})

		variables := NewVariables()
		node := &yaml.Node{Kind: yaml.MappingNode, Content: content}
		converted := variables.convertYAMLNodeToVariable(node)
		mapping, ok := converted.(*MapVariable)
		return ok && mapping.Len() == len(values)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 500}); err != nil {
		t.Fatal(err)
	}
}
