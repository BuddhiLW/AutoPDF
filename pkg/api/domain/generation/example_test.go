// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package generation_test

import (
	"fmt"
	"sort"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
)

func ExampleNewTemplateVariablesFromMap() {
	vars, err := generation.NewTemplateVariablesFromMap(map[string]interface{}{
		"title": "My Document",
		"metadata": map[string]interface{}{
			"version": "1.0",
			"tags":    []interface{}{"example", "flattening"},
		},
		"items": []interface{}{
			"item1",
			map[string]interface{}{"name": "nested item", "value": 42},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	flat := vars.Flatten()
	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s = %s\n", k, flat[k])
	}
	// Output:
	// items[0] = item1
	// items[1].name = nested item
	// items[1].value = 42
	// metadata.tags[0] = example
	// metadata.tags[1] = flattening
	// metadata.version = 1.0
	// title = My Document
}
