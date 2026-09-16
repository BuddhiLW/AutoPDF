// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/template"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/strict"
)

func writeTemplate(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "document.tex")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write template: %v", err)
	}
	return path
}

func TestCLIRendererRefusesAnAbsentKeyAndNamesThemAll(t *testing.T) {
	path := writeTemplate(t, "delim[[.title]] delim[[.alpha]] delim[[.beta]]\n")
	adapter := template.NewTemplateProcessorAdapter(&config.Config{})

	_, err := adapter.Process(context.Background(), path, map[string]string{"title": "Example"})
	if err == nil {
		t.Fatal("a template reading absent keys must not render")
	}
	for _, key := range []string{"alpha", "beta"} {
		if !strings.Contains(err.Error(), key) {
			t.Fatalf("error must name %q: %v", key, err)
		}
	}
	if strings.Contains(err.Error(), "title") {
		t.Fatalf("error must not name a satisfied key: %v", err)
	}
}

func TestCLIRendererRefusesAKeyBoundToABlankString(t *testing.T) {
	path := writeTemplate(t, "delim[[.beta]]\n")
	adapter := template.NewTemplateProcessorAdapter(&config.Config{})

	if _, err := adapter.Process(context.Background(), path, map[string]string{"beta": ""}); err == nil {
		t.Fatal("a key bound to an empty string must not render")
	}
}

func TestCLIRendererAcceptsADeclaredBlank(t *testing.T) {
	path := writeTemplate(t, "N: delim[[ optional .note \"blank when the clause does not apply\" ]].\n")
	adapter := template.NewTemplateProcessorAdapter(&config.Config{})

	out, err := adapter.Process(context.Background(), path, map[string]string{})
	if err != nil {
		t.Fatalf("a declared blank must render: %v", err)
	}
	if strings.TrimSpace(out) != "N: ." {
		t.Fatalf("output = %q", out)
	}
}

// refuseEverything is a stub standing in for the Auditor port.
type refuseEverything struct{ seen []string }

func (r *refuseEverything) Audit(name string, refs []strict.Reference, _ strict.Bindings) strict.Verdict {
	verdict := strict.Verdict{Template: name}
	for _, reference := range refs {
		r.seen = append(r.seen, reference.Key)
		verdict.Offences = append(verdict.Offences,
			strict.Offence{Reference: reference, Cause: strict.CauseAbsent})
	}
	return verdict
}

func TestCLIRendererTakesItsVerdictFromTheInjectedAuditor(t *testing.T) {
	path := writeTemplate(t, "delim[[.title]]\n")
	stub := &refuseEverything{}
	adapter := template.NewTemplateProcessorAdapterWithAuditor(&config.Config{}, stub)

	// Every key is supplied, so the strict policy would allow this render.
	// The injected auditor refuses, and its verdict is what decides.
	_, err := adapter.Process(context.Background(), path, map[string]string{"title": "Example"})
	if err == nil {
		t.Fatal("the injected auditor's refusal must stop the render")
	}
	if len(stub.seen) != 1 || stub.seen[0] != "title" {
		t.Fatalf("auditor saw keys %v", stub.seen)
	}
}

// Even when an auditor is bypassed, a waiver that states no reason cannot
// execute: the builtin takes the reason as an argument.
func TestAReasonlessWaiverCannotExecuteEvenWithoutTheAudit(t *testing.T) {
	path := writeTemplate(t, "delim[[ optional .note ]]\n")
	adapter := template.NewTemplateProcessorAdapterWithAuditor(&config.Config{}, permitEverything{})

	if _, err := adapter.Process(context.Background(), path, map[string]string{"note": "x"}); err == nil {
		t.Fatal("a reasonless waiver must not execute")
	}
}

type permitEverything struct{}

func (permitEverything) Audit(name string, _ []strict.Reference, _ strict.Bindings) strict.Verdict {
	return strict.Verdict{Template: name}
}
