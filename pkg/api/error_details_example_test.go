// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"errors"
	"fmt"
	"sort"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
)

func ExampleNewErrorDetails() {
	details := api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
		WithTemplatePath("templates/missing-vars.tex").
		AddContext("variable_name", "title").
		AddContext("available_variables", "author,date").
		WithError(errors.New("required variable 'title' is missing"))

	details.LogErrorWithMessage(api.NewNoopLogger(), "variable resolution failed")

	fmt.Println(details.Category, details.Severity)
	keys := make([]string, 0, len(details.Context))
	for k := range details.Context {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s: %s\n", k, details.Context[k])
	}
	// Output:
	// variable_resolution high
	// available_variables: author,date
	// error: required variable 'title' is missing
	// template_path: templates/missing-vars.tex
	// variable_name: title
}
