// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package strict

import "strings"

// VoidKind classifies one way a variable read can put nothing on the page.
// The kinds are distinct facts and are named separately so a report can say
// which one happened.
type VoidKind uint8

const (
	// VoidAbsent: no binding carries the key at all.
	VoidAbsent VoidKind = 1 << iota
	// VoidNil: a binding carries the key and its value is nil.
	VoidNil
	// VoidEmpty: a binding carries the key and its value is a string that is
	// empty or contains only whitespace.
	VoidEmpty
)

// Policy is the value that decides what counts as void and what a waiver must
// carry. It is data: build one, pass it, compare it.
//
// Policy deliberately has no "lenient" setting. A render either holds every
// read to the policy or it does not; a blank that is intended is declared at
// the site that produces it, never by relaxing the policy for the document.
type Policy struct {
	// Void is the set of VoidKind values this policy refuses.
	Void VoidKind
	// RequireWaiverReason makes a waiver that states no reason an offence.
	RequireWaiverReason bool
}

// Strict is the default policy: every void kind is refused and every waiver
// must state a reason.
func Strict() Policy {
	return Policy{
		Void:                VoidAbsent | VoidNil | VoidEmpty,
		RequireWaiverReason: true,
	}
}

// Refuses reports whether the policy treats kind as void.
func (p Policy) Refuses(kind VoidKind) bool { return p.Void&kind != 0 }

// Classify names the void kind of a looked-up value and reports whether this
// policy refuses it. A value that is not void, or a void kind this policy does
// not refuse, yields ok == false.
//
// Values that are NOT void under any policy: a zero number, a false boolean,
// an empty collection. Each is a stated fact, not an unstated one.
func (p Policy) Classify(value any, found bool) (cause Cause, ok bool) {
	switch {
	case !found:
		return CauseAbsent, p.Refuses(VoidAbsent)
	case value == nil:
		return CauseNil, p.Refuses(VoidNil)
	default:
		if s, isString := value.(string); isString && strings.TrimSpace(s) == "" {
			return CauseEmpty, p.Refuses(VoidEmpty)
		}
		return "", false
	}
}
