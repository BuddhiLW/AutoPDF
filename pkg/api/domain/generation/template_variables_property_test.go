// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package generation_test

import (
	"reflect"
	"testing"
	"testing/quick"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

func nestedTemplateVariables(value string) *generation.TemplateVariables {
	variables, err := generation.NewTemplateVariablesFromMap(map[string]interface{}{
		"metadata": map[string]interface{}{"value": value},
		"items":    []interface{}{map[string]interface{}{"value": value}},
	})
	if err != nil {
		panic(err)
	}
	return variables
}

func replaceNestedValues(variables *generation.TemplateVariables, replacement string) bool {
	metadataValue, exists := variables.Get("metadata")
	if !exists {
		return false
	}
	metadata, ok := metadataValue.(*config.MapVariable)
	if !ok || metadata.Set("value", &config.StringVariable{Value: replacement}) != nil {
		return false
	}

	itemsValue, exists := variables.Get("items")
	if !exists {
		return false
	}
	items, ok := itemsValue.(*config.SliceVariable)
	if !ok || len(items.Values) != 1 {
		return false
	}
	item, ok := items.Values[0].(*config.MapVariable)
	return ok && item.Set("value", &config.StringVariable{Value: replacement}) == nil
}

func TestPropertyCloneOwnsNestedValueGraph(t *testing.T) {
	property := func(value string) bool {
		original := nestedTemplateVariables(value)
		before := original.Flatten()
		clone := original.Clone()

		if !replaceNestedValues(clone, value+"-clone") {
			return false
		}
		return reflect.DeepEqual(original.Flatten(), before) && !reflect.DeepEqual(clone.Flatten(), before)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 500}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyMergeDoesNotBorrowMutableState(t *testing.T) {
	property := func(value string) bool {
		source := nestedTemplateVariables(value)
		destination := generation.NewTemplateVariables(nil)
		destination.Merge(source)
		merged := destination.Flatten()

		if !replaceNestedValues(source, value+"-source") {
			return false
		}
		return reflect.DeepEqual(destination.Flatten(), merged) && !reflect.DeepEqual(source.Flatten(), merged)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 500}); err != nil {
		t.Fatal(err)
	}
}
