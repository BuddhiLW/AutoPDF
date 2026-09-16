// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package strict

import (
	"fmt"
	"sort"
	"strings"
)

// Cause names why a reference is an offence.
type Cause string

const (
	// CauseAbsent: the key is in no binding.
	CauseAbsent Cause = "absent from the supplied variables"
	// CauseNil: the key is bound to nothing.
	CauseNil Cause = "bound to nil"
	// CauseEmpty: the key is bound to an empty or whitespace-only string.
	CauseEmpty Cause = "bound to an empty string"
	// CauseUnexplainedWaiver: the site waived strictness without stating why.
	CauseUnexplainedWaiver Cause = "waived without a stated reason"
	// CauseMalformedWaiver: the site used the waiver builtin in a shape the
	// engine cannot honour.
	CauseMalformedWaiver Cause = "an unusable " + FuncName + " form"
)

// Offence is one reference the policy refuses, with the cause.
type Offence struct {
	Reference
	Cause Cause
}

// Verdict is the whole result of auditing one template: every offence, not the
// first one.
type Verdict struct {
	// Template names the template the references were read from.
	Template string
	// Offences holds every refused reference, ordered by position.
	Offences []Offence
}

// OK reports that nothing was refused.
func (v Verdict) OK() bool { return len(v.Offences) == 0 }

// Keys lists the distinct offending keys in order of first appearance.
func (v Verdict) Keys() []string {
	seen := make(map[string]bool, len(v.Offences))
	keys := make([]string, 0, len(v.Offences))
	for _, offence := range v.Offences {
		if seen[offence.Key] {
			continue
		}
		seen[offence.Key] = true
		keys = append(keys, offence.Key)
	}
	return keys
}

// Err returns a *MissingVariablesError describing every offence, or nil.
func (v Verdict) Err() error {
	if v.OK() {
		return nil
	}
	return &MissingVariablesError{Template: v.Template, Offences: v.Offences}
}

// MissingVariablesError reports every unsatisfied variable read in a template.
type MissingVariablesError struct {
	Template string
	Offences []Offence
}

// Error lists every offence, one per line, with its source position.
func (e *MissingVariablesError) Error() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "template %q: %d unsatisfied variable reference(s); refusing to render",
		e.Template, len(e.Offences))
	for _, offence := range e.Offences {
		fmt.Fprintf(&builder, "\n  %s:%d:%d: %s is %s",
			e.Template, offence.Line, offence.Column, offence.Key, offence.Cause)
		switch offence.Cause {
		case CauseUnexplainedWaiver, CauseMalformedWaiver:
			fmt.Fprintf(&builder, " (write: %s .%s \"why the blank is acceptable\")",
				FuncName, offence.Key)
		case CauseAbsent, CauseNil, CauseEmpty:
			fmt.Fprintf(&builder, " (supply it, or write: %s .%s \"why the blank is acceptable\")",
				FuncName, offence.Key)
		}
	}
	return builder.String()
}

// Keys lists the distinct offending keys.
func (e *MissingVariablesError) Keys() []string {
	return Verdict{Template: e.Template, Offences: e.Offences}.Keys()
}

// sortOffences orders offences by source position, then by key.
func sortOffences(offences []Offence) {
	sort.SliceStable(offences, func(i, j int) bool {
		left, right := offences[i], offences[j]
		switch {
		case left.Line != right.Line:
			return left.Line < right.Line
		case left.Column != right.Column:
			return left.Column < right.Column
		default:
			return left.Key < right.Key
		}
	})
}
