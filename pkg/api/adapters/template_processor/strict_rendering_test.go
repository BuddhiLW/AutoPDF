// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package template_processor_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/adapters/template_processor"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/strict"
)

func writeTemplate(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.tex")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write template: %v", err)
	}
	return path
}

func quietLogger() *logger.LoggerAdapter {
	return logger.NewLoggerAdapter(logger.Silent, "stdout")
}

func TestRenderRefusesATemplateWhoseKeyIsAbsent(t *testing.T) {
	path := writeTemplate(t, "A: delim[[.alpha]]\nB: delim[[.beta]]\n")
	adapter := template_processor.NewTemplateProcessorAdapter(&config.Config{}, quietLogger())

	_, err := adapter.Process(context.Background(), path, map[string]string{"alpha": "12"})
	if err == nil {
		t.Fatal("a template reading an absent key must not render")
	}
	if !strings.Contains(err.Error(), "beta") {
		t.Fatalf("error must name the absent key: %v", err)
	}
}

func TestRenderNamesEveryOffendingKeyNotOnlyTheFirst(t *testing.T) {
	path := writeTemplate(t, "delim[[.one]] delim[[.two]] delim[[.three]]\n")
	adapter := template_processor.NewTemplateProcessorAdapter(&config.Config{}, quietLogger())

	_, err := adapter.Process(context.Background(), path, map[string]string{})
	if err == nil {
		t.Fatal("expected a refusal")
	}
	for _, key := range []string{"one", "two", "three"} {
		if !strings.Contains(err.Error(), key) {
			t.Fatalf("error must name %q: %v", key, err)
		}
	}
}

// A bound-but-empty value is the case Go's own missingkey=error cannot see.
func TestRenderRefusesAKeyBoundToAnEmptyString(t *testing.T) {
	path := writeTemplate(t, "B: delim[[.beta]]\n")
	adapter := template_processor.NewTemplateProcessorAdapter(&config.Config{}, quietLogger())

	_, err := adapter.Process(context.Background(), path, map[string]string{"beta": "   "})
	if err == nil {
		t.Fatal("a key bound to a blank string must not render")
	}
	if !strings.Contains(err.Error(), "beta") {
		t.Fatalf("error must name the key: %v", err)
	}
}

func TestRenderAcceptsADeclaredBlank(t *testing.T) {
	path := writeTemplate(t, "N: delim[[ optional .note \"blank when the clause does not apply\" ]].\n")
	adapter := template_processor.NewTemplateProcessorAdapter(&config.Config{}, quietLogger())

	out, err := adapter.Process(context.Background(), path, map[string]string{})
	if err != nil {
		t.Fatalf("a declared blank must render: %v", err)
	}
	if strings.TrimSpace(out) != "N: ." {
		t.Fatalf("output = %q", out)
	}
}

func TestRenderRefusesAWaiverThatStatesNoReason(t *testing.T) {
	path := writeTemplate(t, "N: delim[[ optional .note ]].\n")
	adapter := template_processor.NewTemplateProcessorAdapter(&config.Config{}, quietLogger())

	if _, err := adapter.Process(context.Background(), path, map[string]string{}); err == nil {
		t.Fatal("a waiver without a reason must not render")
	}
}

func TestRenderSucceedsWhenEveryKeyIsSupplied(t *testing.T) {
	path := writeTemplate(t, "T: delim[[.title]]\n")
	adapter := template_processor.NewTemplateProcessorAdapter(&config.Config{}, quietLogger())

	out, err := adapter.Process(context.Background(), path, map[string]string{"title": "Example"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Example") {
		t.Fatalf("output = %q", out)
	}
}

// recordingAuditor is a stub standing in for the Auditor port. The renderer
// depends on the port, so a test substitutes an implementation rather than
// reaching for global state.
type recordingAuditor struct {
	sawTemplate string
	sawKeys     []string
	verdict     strict.Verdict
}

func (r *recordingAuditor) Audit(template string, refs []strict.Reference, vars strict.Bindings) strict.Verdict {
	r.sawTemplate = template
	for _, ref := range refs {
		r.sawKeys = append(r.sawKeys, ref.Key)
	}
	r.verdict.Template = template
	return r.verdict
}

func TestTheRendererTakesItsVerdictFromTheInjectedAuditor(t *testing.T) {
	path := writeTemplate(t, "delim[[.anything]]\n")
	stub := &recordingAuditor{}
	adapter := template_processor.NewTemplateProcessorAdapterWithAuditor(&config.Config{}, quietLogger(), stub)

	// The stub returns a clean verdict, so a key the strict policy would refuse
	// renders anyway: the decision belongs to the injected auditor.
	if _, err := adapter.Process(context.Background(), path, map[string]string{"anything": "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.sawTemplate != path {
		t.Fatalf("auditor saw template %q, want %q", stub.sawTemplate, path)
	}
	if len(stub.sawKeys) != 1 || stub.sawKeys[0] != "anything" {
		t.Fatalf("auditor saw keys %v", stub.sawKeys)
	}
}

func TestAnInjectedAuditorCanRefuseWhatStrictWouldAllow(t *testing.T) {
	path := writeTemplate(t, "delim[[.anything]]\n")
	stub := &recordingAuditor{verdict: strict.Verdict{Offences: []strict.Offence{{
		Reference: strict.Reference{Key: "anything", Line: 1, Column: 1},
		Cause:     strict.CauseAbsent,
	}}}}
	adapter := template_processor.NewTemplateProcessorAdapterWithAuditor(&config.Config{}, quietLogger(), stub)

	_, err := adapter.Process(context.Background(), path, map[string]string{"anything": "present"})
	if err == nil {
		t.Fatal("the injected auditor's refusal must stop the render")
	}
	var missing *strict.MissingVariablesError
	if !errors.As(err, &missing) && !strings.Contains(err.Error(), "anything") {
		t.Fatalf("error must carry the verdict: %v", err)
	}
}
