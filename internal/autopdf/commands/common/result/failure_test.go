// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package result_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/configs"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/commands/common/result"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/strict"
)

func refusal(keys ...string) *strict.MissingVariablesError {
	verdict := strict.Verdict{Template: "doc.tex"}
	for index, key := range keys {
		verdict.Offences = append(verdict.Offences, strict.Offence{
			Reference: strict.Reference{Key: key, Line: index + 10, Column: index + 3},
			Cause:     strict.CauseAbsent,
		})
	}
	var missing *strict.MissingVariablesError
	if !errors.As(verdict.Err(), &missing) {
		panic("verdict must produce a MissingVariablesError")
	}
	return missing
}

func TestBuildFailureCarriesEveryOffendingKeyAndItsLocation(t *testing.T) {
	err := result.BuildFailure(refusal("alpha", "beta", "gamma"))
	if err == nil {
		t.Fatal("BuildFailure must not swallow a refusal")
	}

	message := err.Error()
	for _, fragment := range []string{
		"alpha", "beta", "gamma", // every key, not only the first
		"doc.tex:10:3", "doc.tex:11:4", "doc.tex:12:5", // every location
	} {
		if !strings.Contains(message, fragment) {
			t.Fatalf("message must contain %q, got:\n%s", fragment, message)
		}
	}
}

func TestBuildFailureNamesTheTemplatePhaseForARefusal(t *testing.T) {
	err := result.BuildFailure(refusal("alpha"))

	if !errors.Is(err, configs.TemplateError) {
		t.Fatalf("a refusal must be reported as a template failure, got: %v", err)
	}
	if errors.Is(err, configs.BuildError) {
		t.Fatalf("a refusal must not be reported as a config-build failure, got: %v", err)
	}
}

func TestBuildFailureKeepsTheRefusalRecoverable(t *testing.T) {
	err := result.BuildFailure(refusal("alpha", "beta"))

	var missing *strict.MissingVariablesError
	if !errors.As(err, &missing) {
		t.Fatalf("the cause must survive wrapping, got %T", err)
	}
	if got, want := len(missing.Keys()), 2; got != want {
		t.Fatalf("keys = %d, want %d", got, want)
	}
}

func TestBuildFailurePreservesANonTemplateCause(t *testing.T) {
	cause := fmt.Errorf("pdflatex exited 1: Undefined control sequence")
	err := result.BuildFailure(cause)

	if !errors.Is(err, configs.BuildError) {
		t.Fatalf("an unclassified failure keeps the build phase, got: %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("the cause must survive wrapping, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Undefined control sequence") {
		t.Fatalf("the cause's detail must reach the message, got: %v", err)
	}
}

func TestBuildFailureOfNilIsNil(t *testing.T) {
	if err := result.BuildFailure(nil); err != nil {
		t.Fatalf("BuildFailure(nil) = %v", err)
	}
}
