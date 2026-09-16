// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package strict_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/strict"
)

func ref(key string, line, column int) strict.Reference {
	return strict.Reference{Key: key, Line: line, Column: column}
}

func TestClassifyDistinguishesEveryVoidKind(t *testing.T) {
	policy := strict.Strict()

	cases := []struct {
		name  string
		value any
		found bool
		cause strict.Cause
		void  bool
	}{
		{"absent key", nil, false, strict.CauseAbsent, true},
		{"nil value", nil, true, strict.CauseNil, true},
		{"empty string", "", true, strict.CauseEmpty, true},
		{"whitespace only", "  \t\n ", true, strict.CauseEmpty, true},
		{"non-empty string", "x", true, "", false},
		{"zero number", 0, true, "", false},
		{"zero float", 0.0, true, "", false},
		{"false boolean", false, true, "", false},
		{"empty slice", []any{}, true, "", false},
		{"empty map", map[string]any{}, true, "", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			cause, void := policy.Classify(testCase.value, testCase.found)
			if void != testCase.void {
				t.Fatalf("void = %v, want %v", void, testCase.void)
			}
			if void && cause != testCase.cause {
				t.Fatalf("cause = %q, want %q", cause, testCase.cause)
			}
		})
	}
}

func TestPolicyRefusesOnlyTheKindsItDeclares(t *testing.T) {
	absentOnly := strict.Policy{Void: strict.VoidAbsent}

	if _, void := absentOnly.Classify("", true); void {
		t.Fatal("a policy that does not refuse VoidEmpty must accept an empty string")
	}
	if _, void := absentOnly.Classify(nil, false); !void {
		t.Fatal("a policy that refuses VoidAbsent must refuse an absent key")
	}
}

func TestAuditReportsEveryOffenceNotTheFirst(t *testing.T) {
	auditor := strict.Default()
	vars := strict.Bindings{"present": "yes", "empty": "", "nilled": nil}

	verdict := auditor.Audit("doc.tex", []strict.Reference{
		ref("present", 1, 1),
		ref("missing_a", 2, 5),
		ref("empty", 3, 9),
		ref("nilled", 4, 2),
		ref("missing_b", 5, 7),
	}, vars)

	if verdict.OK() {
		t.Fatal("verdict must not be OK")
	}
	if got, want := len(verdict.Offences), 4; got != want {
		t.Fatalf("offences = %d, want %d: %v", got, want, verdict.Keys())
	}
	wantKeys := []string{"missing_a", "empty", "nilled", "missing_b"}
	got := verdict.Keys()
	for index, key := range wantKeys {
		if got[index] != key {
			t.Fatalf("keys = %v, want %v", got, wantKeys)
		}
	}
}

func TestAuditOrdersOffencesBySourcePosition(t *testing.T) {
	verdict := strict.Default().Audit("t.tex", []strict.Reference{
		ref("c", 9, 1), ref("a", 2, 7), ref("b", 2, 3),
	}, strict.Bindings{})

	if got := verdict.Keys(); got[0] != "b" || got[1] != "a" || got[2] != "c" {
		t.Fatalf("keys = %v, want [b a c]", got)
	}
}

func TestNestedKeysResolveThroughTheBindings(t *testing.T) {
	vars := strict.Bindings{
		"outer": map[string]any{
			"inner": map[string]any{"first": "08:00", "second": ""},
		},
	}
	verdict := strict.Default().Audit("t.tex", []strict.Reference{
		ref("outer.inner.first", 1, 1),
		ref("outer.inner.second", 2, 1),
		ref("outer.inner.missing", 3, 1),
		ref("outer.missing.first", 4, 1),
	}, vars)

	if got, want := len(verdict.Offences), 3; got != want {
		t.Fatalf("offences = %d, want %d: %v", got, want, verdict.Keys())
	}
	if verdict.Offences[0].Cause != strict.CauseEmpty {
		t.Fatalf("cause = %q, want %q", verdict.Offences[0].Cause, strict.CauseEmpty)
	}
	if verdict.Offences[1].Cause != strict.CauseAbsent {
		t.Fatalf("cause = %q, want %q", verdict.Offences[1].Cause, strict.CauseAbsent)
	}
}

func TestAWaivedReferenceWithAReasonIsNotAnOffence(t *testing.T) {
	verdict := strict.Default().Audit("t.tex", []strict.Reference{
		{Key: "note", Line: 4, Column: 2, Waived: true, Reason: "blank when the clause does not apply"},
	}, strict.Bindings{})

	if !verdict.OK() {
		t.Fatalf("verdict must be OK, got %v", verdict.Offences)
	}
}

func TestAWaivedReferenceWithoutAReasonIsAnOffence(t *testing.T) {
	verdict := strict.Default().Audit("t.tex", []strict.Reference{
		{Key: "note", Line: 4, Column: 2, Waived: true},
		{Key: "outro", Line: 5, Column: 2, Waived: true, Reason: "   "},
	}, strict.Bindings{})

	if got, want := len(verdict.Offences), 2; got != want {
		t.Fatalf("offences = %d, want %d", got, want)
	}
	for _, offence := range verdict.Offences {
		if offence.Cause != strict.CauseUnexplainedWaiver {
			t.Fatalf("cause = %q, want %q", offence.Cause, strict.CauseUnexplainedWaiver)
		}
	}
}

func TestAPolicyMayAcceptAnUnexplainedWaiver(t *testing.T) {
	lenientOnReasons := strict.NewAuditor(strict.Policy{Void: strict.VoidAbsent})
	verdict := lenientOnReasons.Audit("t.tex", []strict.Reference{
		{Key: "note", Waived: true},
	}, strict.Bindings{})

	if !verdict.OK() {
		t.Fatalf("verdict must be OK, got %v", verdict.Offences)
	}
}

func TestAMalformedWaiverIsAnOffenceEvenWhenTheKeyIsBound(t *testing.T) {
	verdict := strict.Default().Audit("t.tex", []strict.Reference{
		{Key: "title", Line: 1, Column: 1, Waived: true, Malformed: true},
	}, strict.Bindings{"title": "present"})

	if got, want := len(verdict.Offences), 1; got != want {
		t.Fatalf("offences = %d, want %d", got, want)
	}
	if verdict.Offences[0].Cause != strict.CauseMalformedWaiver {
		t.Fatalf("cause = %q, want %q", verdict.Offences[0].Cause, strict.CauseMalformedWaiver)
	}
}

func TestVerdictErrNamesEveryKeyAndItsLocation(t *testing.T) {
	verdict := strict.Default().Audit("doc.tex", []strict.Reference{
		ref("alpha", 12, 30),
		ref("beta", 40, 3),
	}, strict.Bindings{})

	err := verdict.Err()
	if err == nil {
		t.Fatal("Err must not be nil for a failed verdict")
	}

	var missing *strict.MissingVariablesError
	if !errors.As(err, &missing) {
		t.Fatalf("Err must be a *MissingVariablesError, got %T", err)
	}
	if got, want := len(missing.Keys()), 2; got != want {
		t.Fatalf("keys = %d, want %d", got, want)
	}

	message := err.Error()
	for _, fragment := range []string{"doc.tex", "alpha", "beta", "12:30", "40:3"} {
		if !strings.Contains(message, fragment) {
			t.Fatalf("message %q must mention %q", message, fragment)
		}
	}
}

func TestVerdictErrIsNilWhenNothingIsRefused(t *testing.T) {
	verdict := strict.Default().Audit("t.tex", []strict.Reference{ref("a", 1, 1)},
		strict.Bindings{"a": "x"})

	if !verdict.OK() || verdict.Err() != nil {
		t.Fatalf("verdict must be clean, got %v", verdict.Offences)
	}
}

func TestAuditOfNoReferencesIsClean(t *testing.T) {
	if !strict.Default().Audit("t.tex", nil, nil).OK() {
		t.Fatal("a template that reads nothing cannot be unsatisfied")
	}
}

// StrictAuditor must satisfy the port so a renderer can depend on the port.
var _ strict.Auditor = strict.StrictAuditor{}
