// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package strict_test

import (
	"sort"
	"strings"
	"testing"
	"testing/quick"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/strict"
)

// sample turns arbitrary generated data into a reference set and a binding set
// that overlap in interesting ways.
func sample(keys []string, supplied []string, values []string) ([]strict.Reference, strict.Bindings) {
	refs := make([]strict.Reference, 0, len(keys))
	for index, key := range keys {
		key = sanitize(key, index)
		refs = append(refs, strict.Reference{Key: key, Line: index + 1, Column: 1})
	}
	vars := strict.Bindings{}
	for index, key := range supplied {
		value := ""
		if index < len(values) {
			value = values[index]
		}
		vars.Set(sanitize(key, index), value)
	}
	return refs, vars
}

// sanitize keeps generated keys usable as dotted paths.
func sanitize(key string, index int) string {
	key = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		default:
			return -1
		}
	}, key)
	if key == "" {
		return string(rune('a' + index%26))
	}
	return key
}

func TestPropertyVerdictIsCleanExactlyWhenNoReadIsVoid(t *testing.T) {
	law := func(keys, supplied, values []string) bool {
		refs, vars := sample(keys, supplied, values)
		verdict := strict.Default().Audit("t.tex", refs, vars)

		var expectedOffences int
		for _, ref := range refs {
			value, found := vars.Lookup(ref.Key)
			if _, void := strict.Strict().Classify(value, found); void {
				expectedOffences++
			}
		}
		return verdict.OK() == (expectedOffences == 0) && len(verdict.Offences) == expectedOffences
	}
	if err := quick.Check(law, nil); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyEveryOffenceNamesAReferenceThatWasGiven(t *testing.T) {
	law := func(keys, supplied, values []string) bool {
		refs, vars := sample(keys, supplied, values)
		given := make(map[string]bool, len(refs))
		for _, ref := range refs {
			given[ref.Key] = true
		}
		for _, offence := range strict.Default().Audit("t.tex", refs, vars).Offences {
			if !given[offence.Key] {
				return false
			}
		}
		return true
	}
	if err := quick.Check(law, nil); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyTheVerdictDoesNotDependOnTheOrderOfTheReferences(t *testing.T) {
	law := func(keys, supplied, values []string) bool {
		refs, vars := sample(keys, supplied, values)
		reversed := make([]strict.Reference, len(refs))
		for index := range refs {
			reversed[len(refs)-1-index] = refs[index]
		}

		forward := offenceKeys(strict.Default().Audit("t.tex", refs, vars))
		backward := offenceKeys(strict.Default().Audit("t.tex", reversed, vars))
		sort.Strings(forward)
		sort.Strings(backward)
		return strings.Join(forward, "\x00") == strings.Join(backward, "\x00")
	}
	if err := quick.Check(law, nil); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyAWaivedReferenceWithAReasonIsNeverAnOffence(t *testing.T) {
	law := func(keys []string, reason string) bool {
		if strings.TrimSpace(reason) == "" {
			return true
		}
		refs, _ := sample(keys, nil, nil)
		for index := range refs {
			refs[index].Waived = true
			refs[index].Reason = reason
		}
		return strict.Default().Audit("t.tex", refs, strict.Bindings{}).OK()
	}
	if err := quick.Check(law, nil); err != nil {
		t.Fatal(err)
	}
}

func offenceKeys(verdict strict.Verdict) []string {
	keys := make([]string, 0, len(verdict.Offences))
	for _, offence := range verdict.Offences {
		keys = append(keys, offence.Key)
	}
	return keys
}
