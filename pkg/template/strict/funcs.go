// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package strict

import "text/template"

// FuncName is the template builtin that declares a blank to be deliberate.
// It has exactly one spelling, the command form:
//
//	delim[[ optional .field "reason the blank is acceptable" ]]
//
// The pipe form is refused by the audit, because a pipe would hand optional
// its arguments in the opposite order and silently print the reason.
const FuncName = "optional"

// MissingKeyOption is the text/template option a strict render executes under.
// It is the backstop for reads the static audit cannot resolve, not the
// primary check: it stops at the first miss and cannot see a bound-but-empty
// value at all.
const MissingKeyOption = "missingkey=error"

// Funcs returns the template functions a strict render must install.
//
// optional renders a void value as nothing. The reason is read by the audit,
// not by the render; requiring it here keeps a waiver that states none from
// executing even if the audit was bypassed.
func Funcs() template.FuncMap {
	return template.FuncMap{
		FuncName: func(value any, reason string) any {
			if value == nil {
				return ""
			}
			return value
		},
	}
}

// FillWaived binds every waived reference that is absent to the empty string,
// so a declared blank survives execution under MissingKeyOption. It mutates
// and returns vars.
func FillWaived(vars Bindings, refs []Reference) Bindings {
	if vars == nil {
		return vars
	}
	for _, ref := range refs {
		if !ref.Waived || ref.Malformed {
			continue
		}
		if _, found := vars.Lookup(ref.Key); found {
			continue
		}
		vars.Set(ref.Key, "")
	}
	return vars
}
