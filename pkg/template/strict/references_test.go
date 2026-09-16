// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package strict_test

import (
	"testing"
	"text/template"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/strict"
)

// parseSource parses source under the engine's delimiters and func map.
func parseSource(t *testing.T, source string) *template.Template {
	t.Helper()
	tmpl, err := template.New("t.tex").
		Funcs(strict.Funcs()).
		Delims("delim[[", "]]").
		Option(strict.MissingKeyOption).
		Parse(source)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return tmpl
}

func referencesOf(t *testing.T, source string) []strict.Reference {
	t.Helper()
	return strict.References(parseSource(t, source), source)
}

func keysOf(refs []strict.Reference) []string {
	keys := make([]string, 0, len(refs))
	for _, ref := range refs {
		keys = append(keys, ref.Key)
	}
	return keys
}

func equalKeys(got []string, want ...string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func TestReferencesCollectsPlainAndNestedReads(t *testing.T) {
	refs := referencesOf(t, `A delim[[.title]] B delim[[ .outer.inner.first ]]`)

	if got := keysOf(refs); !equalKeys(got, "title", "outer.inner.first") {
		t.Fatalf("keys = %v", got)
	}
}

func TestReferencesRecordsSourcePositions(t *testing.T) {
	refs := referencesOf(t, "line one\nline two delim[[.alpha]]\n")

	if len(refs) != 1 {
		t.Fatalf("refs = %v", keysOf(refs))
	}
	if refs[0].Line != 2 {
		t.Fatalf("line = %d, want 2", refs[0].Line)
	}
	if refs[0].Column != 17 {
		t.Fatalf("column = %d, want 17", refs[0].Column)
	}
}

func TestReferencesSeesReadsInsideConditionsAndPipelines(t *testing.T) {
	refs := referencesOf(t, `delim[[if .flag]]delim[[.body]]delim[[else]]delim[[.alternative]]delim[[end]]`)

	if got := keysOf(refs); !equalKeys(got, "flag", "body", "alternative") {
		t.Fatalf("keys = %v", got)
	}
}

func TestReferencesSkipsReadsWhoseDotWasRebound(t *testing.T) {
	refs := referencesOf(t, `delim[[range .items]]delim[[.name]]delim[[end]]`)

	if got := keysOf(refs); !equalKeys(got, "items") {
		t.Fatalf("keys = %v, want only the range subject", got)
	}
}

func TestReferencesKeepsTheElseBranchOfARangeInScope(t *testing.T) {
	refs := referencesOf(t, `delim[[range .items]]delim[[.name]]delim[[else]]delim[[.empty]]delim[[end]]`)

	if got := keysOf(refs); !equalKeys(got, "items", "empty") {
		t.Fatalf("keys = %v", got)
	}
}

func TestReferencesSkipsTheBodyOfWithButNotItsSubject(t *testing.T) {
	refs := referencesOf(t, `delim[[with .record]]delim[[.number]]delim[[end]]`)

	if got := keysOf(refs); !equalKeys(got, "record") {
		t.Fatalf("keys = %v", got)
	}
}

func TestReferencesResolvesDollarAgainstTheRootEvenInsideARange(t *testing.T) {
	refs := referencesOf(t, `delim[[range .items]]delim[[$.footer]]delim[[end]]`)

	if got := keysOf(refs); !equalKeys(got, "items", "footer") {
		t.Fatalf("keys = %v", got)
	}
}

func TestOptionalMarksTheReferenceWaivedAndCarriesItsReason(t *testing.T) {
	refs := referencesOf(t, `delim[[ optional .note "blank when the clause does not apply" ]]`)

	if len(refs) != 1 {
		t.Fatalf("refs = %v", keysOf(refs))
	}
	if !refs[0].Waived || refs[0].Malformed {
		t.Fatalf("reference must be waived and well formed: %+v", refs[0])
	}
	if refs[0].Key != "note" {
		t.Fatalf("key = %q", refs[0].Key)
	}
	if refs[0].Reason != "blank when the clause does not apply" {
		t.Fatalf("reason = %q", refs[0].Reason)
	}
}

func TestOptionalWithoutAReasonIsWaivedButUnexplained(t *testing.T) {
	refs := referencesOf(t, `delim[[ optional .note ]]`)

	if len(refs) != 1 {
		t.Fatalf("refs = %v", keysOf(refs))
	}
	if !refs[0].Malformed {
		t.Fatalf("a one-argument optional must be reported: %+v", refs[0])
	}
	if refs[0].Key != "note" {
		t.Fatalf("key = %q, the report must still name the field", refs[0].Key)
	}
}

func TestThePipeFormOfOptionalIsRefusedRatherThanHonoured(t *testing.T) {
	refs := referencesOf(t, `delim[[ .note | optional "reason" ]]`)

	var malformed int
	for _, reference := range refs {
		if reference.Malformed {
			malformed++
		}
	}
	if malformed != 1 {
		t.Fatalf("the pipe form must be reported malformed exactly once: %+v", refs)
	}
}

func TestOptionalOnSomethingThatIsNotAReadIsRefused(t *testing.T) {
	refs := referencesOf(t, `delim[[ optional "literal" "reason" ]]`)

	if len(refs) != 1 || !refs[0].Malformed {
		t.Fatalf("optional over a literal must be reported: %+v", refs)
	}
}

func TestAWaivedReadIsNotAlsoReportedAsAPlainRead(t *testing.T) {
	refs := referencesOf(t, `delim[[ optional .note "reason" ]]`)

	if len(refs) != 1 {
		t.Fatalf("the waived site must be recorded once, got %+v", refs)
	}
}

func TestReferencesOfATemplateWithNoActionsIsEmpty(t *testing.T) {
	if refs := referencesOf(t, `\documentclass{article}`); len(refs) != 0 {
		t.Fatalf("refs = %v", keysOf(refs))
	}
}

func TestReferencesOfANilTemplateIsEmpty(t *testing.T) {
	if refs := strict.References(nil, ""); refs != nil {
		t.Fatalf("refs = %v", refs)
	}
}

func TestEndToEndAuditOfAParsedTemplate(t *testing.T) {
	source := "Titulo: delim[[.title]]\n" +
		"Parcelas: delim[[.alpha]]\n" +
		"Obs: delim[[ optional .note \"blank when the clause does not apply\" ]]\n"

	tmpl := parseSource(t, source)
	refs := strict.References(tmpl, source)
	verdict := strict.Default().Audit("doc.tex", refs, strict.Bindings{"title": "X"})

	if got := verdict.Keys(); !equalKeys(got, "alpha") {
		t.Fatalf("keys = %v, want only the genuinely missing one", got)
	}
}

func TestFillWaivedBindsOnlyTheAbsentWaivedKeys(t *testing.T) {
	vars := strict.Bindings{"present": "x"}
	refs := []strict.Reference{
		{Key: "present", Waived: true, Reason: "r"},
		{Key: "absent", Waived: true, Reason: "r"},
		{Key: "nested.absent", Waived: true, Reason: "r"},
		{Key: "not_waived", Waived: false},
		{Key: "malformed", Waived: true, Malformed: true},
	}

	strict.FillWaived(vars, refs)

	if value, found := vars.Lookup("present"); !found || value != "x" {
		t.Fatalf("an existing binding must not be overwritten: %v", value)
	}
	if value, found := vars.Lookup("absent"); !found || value != "" {
		t.Fatalf("absent = %v, %v", value, found)
	}
	if value, found := vars.Lookup("nested.absent"); !found || value != "" {
		t.Fatalf("nested.absent = %v, %v", value, found)
	}
	if _, found := vars.Lookup("not_waived"); found {
		t.Fatal("a reference that was not waived must not be filled")
	}
	if _, found := vars.Lookup("malformed"); found {
		t.Fatal("a malformed waiver must not be filled")
	}
}

func TestAWaivedTemplateRendersBlankWithoutFailing(t *testing.T) {
	source := `Obs: delim[[ optional .note "blank when the clause does not apply" ]].`
	tmpl := parseSource(t, source)
	refs := strict.References(tmpl, source)
	vars := strict.FillWaived(strict.Bindings{}, refs)

	var out stringSink
	if err := tmpl.Execute(&out, map[string]any(vars)); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if out.text != "Obs: ." {
		t.Fatalf("output = %q", out.text)
	}
}

type stringSink struct{ text string }

func (s *stringSink) Write(p []byte) (int, error) {
	s.text += string(p)
	return len(p), nil
}
