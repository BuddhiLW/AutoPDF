// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package services_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/services"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

func newConfig() *config.Config {
	return &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}
}

func ExamplePDFGenerationAPIService_GeneratePDF() {
	svc := services.NewPDFGenerationAPIService(newConfig(), nil, false)

	pdf, images, err := svc.GeneratePDF(context.Background(), "template.tex", "output.pdf", map[string]interface{}{
		"title": "My Document",
		"foo": map[string]interface{}{
			"bar": []string{"bar1", "bar2"},
			"zet": []int{1, 2, 3},
		},
		"foo_bar": []string{"foo", "bar"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(pdf), images)
}

func ExamplePDFGenerationAPIService_GeneratePDFWithOptions() {
	svc := services.NewPDFGenerationAPIService(newConfig(), nil, false)

	opts := services.NewPDFGenerationOptions("template.tex", "output.pdf").
		WithEngine("xelatex").
		WithVariable("title", "Builder Pattern Example").
		WithVariable("metadata", map[string]interface{}{
			"version": "1.0",
			"tags":    []string{"example", "builder"},
		}).
		WithConversion(true, "png").
		WithTimeout(30 * time.Second)

	pdf, images, err := svc.GeneratePDFWithOptions(context.Background(), *opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(pdf), images)
}

func ExamplePDFGenerationAPIService_GetTemplateVariables() {
	svc := services.NewPDFGenerationAPIService(newConfig(), nil, false)

	names, err := svc.GetTemplateVariables("template.tex")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(names)
}
