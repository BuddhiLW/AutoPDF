// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package typst_template_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/logger"
	clitemplate "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/template"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/adapters/template_processor"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

const fixture = "document.typ"

// renderer is the shape both template adapters share.
type renderer interface {
	Process(ctx context.Context, templatePath string, variables map[string]string) (string, error)
}

func renderers() map[string]renderer {
	return map[string]renderer{
		"cli": clitemplate.NewTemplateProcessorAdapter(&config.Config{}),
		"api": template_processor.NewTemplateProcessorAdapter(
			&config.Config{}, logger.NewLoggerAdapter(logger.Silent, "stdout")),
	}
}

func fixtureVariables() map[string]string {
	return map[string]string{
		"title":  "Typst Report",
		"author": "Ada #1 _Lovelace_ [a]",
	}
}

func TestTypstFixtureRendersActionsAndKeepsTypstBracketsUntouched(t *testing.T) {
	want, err := os.ReadFile("document.expected.typ")
	if err != nil {
		t.Fatalf("read expected: %v", err)
	}
	for name, adapter := range renderers() {
		t.Run(name, func(t *testing.T) {
			got, err := adapter.Process(context.Background(), fixture, fixtureVariables())
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != string(want) {
				t.Fatalf("output mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
			}
			for _, literal := range []string{"#f[a[b]]", "stay put: ]]", "#f[x[y[z]]]"} {
				if !strings.Contains(got, literal) {
					t.Fatalf("literal %q did not survive: %s", literal, got)
				}
			}
			if strings.Contains(got, "delim[[") {
				t.Fatalf("an action was left unrendered: %s", got)
			}
		})
	}
}

func TestTypstFixtureRefusesAMissingKey(t *testing.T) {
	for name, adapter := range renderers() {
		t.Run(name, func(t *testing.T) {
			variables := fixtureVariables()
			delete(variables, "author")
			_, err := adapter.Process(context.Background(), fixture, variables)
			if err == nil {
				t.Fatal("a .typ template reading an absent key must not render")
			}
			if !strings.Contains(err.Error(), "author") {
				t.Fatalf("error must name the absent key: %v", err)
			}
		})
	}
}

func TestTypstTemplateWithoutActionsSurvivesVerbatim(t *testing.T) {
	body := "#let g(x) = [#x]\n#g[a[b]] ]] [[ ]]\n"
	path := filepath.Join(t.TempDir(), "plain.typ")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	for name, adapter := range renderers() {
		t.Run(name, func(t *testing.T) {
			got, err := adapter.Process(context.Background(), path, map[string]string{})
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got != body {
				t.Fatalf("output = %q, want %q", got, body)
			}
		})
	}
}
