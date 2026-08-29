// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
)

func ExampleEngine() {
	engine, _ := api.NewEngine(api.WithGenerator(api.GeneratorFunc(
		func(_ context.Context, req api.Request) (api.Result, error) {
			return api.Result{PDFPath: req.OutputPath, PDF: []byte("%PDF-demo")}, nil
		},
	)))
	result, _ := engine.Generate(context.Background(), api.Request{
		TemplatePath: "invoice.tex",
		OutputPath:   "invoice.pdf",
		Variables:    map[string]string{"customer.name": "Ada"},
	})
	fmt.Println(result.PDFPath, len(result.PDF))
	// Output: invoice.pdf 9
}

func ExampleGeneratorFunc() {
	remote := api.GeneratorFunc(func(_ context.Context, req api.Request) (api.Result, error) {
		return api.Result{PDFPath: req.OutputPath, PDF: []byte("%PDF-remote")}, nil
	})
	engine, _ := api.NewEngine(
		api.WithGenerator(remote),
		api.WithCapabilities(api.Capabilities{Engines: []string{"remote"}}),
	)
	fmt.Println(engine.Capabilities().Engines)
	// Output: [remote]
}
