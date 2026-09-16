// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package strict

import "strings"

// Auditor decides whether a template's variable reads are satisfied.
//
// It is the port a renderer depends on. Implementations must be pure: no file
// I/O, no compilation, no logging.
type Auditor interface {
	// Audit returns every refused reference at once, never only the first.
	Audit(template string, refs []Reference, vars Bindings) Verdict
}

// StrictAuditor holds every reference to a Policy.
type StrictAuditor struct {
	Policy Policy
}

// NewAuditor returns an Auditor enforcing policy.
func NewAuditor(policy Policy) StrictAuditor { return StrictAuditor{Policy: policy} }

// Default returns the Auditor a renderer uses when no other is injected.
func Default() StrictAuditor { return NewAuditor(Strict()) }

// Audit checks every reference against vars and returns all offences.
func (a StrictAuditor) Audit(template string, refs []Reference, vars Bindings) Verdict {
	verdict := Verdict{Template: template}
	for _, ref := range refs {
		if ref.Malformed {
			verdict.Offences = append(verdict.Offences,
				Offence{Reference: ref, Cause: CauseMalformedWaiver})
			continue
		}
		if ref.Waived {
			if a.Policy.RequireWaiverReason && strings.TrimSpace(ref.Reason) == "" {
				verdict.Offences = append(verdict.Offences,
					Offence{Reference: ref, Cause: CauseUnexplainedWaiver})
			}
			continue
		}
		value, found := vars.Lookup(ref.Key)
		if cause, refused := a.Policy.Classify(value, found); refused {
			verdict.Offences = append(verdict.Offences, Offence{Reference: ref, Cause: cause})
		}
	}
	sortOffences(verdict.Offences)
	return verdict
}
