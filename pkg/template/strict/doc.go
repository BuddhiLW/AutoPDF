// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package strict decides whether a template's variable reads are satisfied by
// the variables supplied to it.
//
// The package is pure: it performs no file I/O, runs no compiler and writes no
// logs. Its inputs are a parsed template and the data that would be handed to
// Execute; its output is a Verdict naming every unsatisfied read at once.
//
// The vocabulary:
//
//	Policy     what counts as void, and what a waiver must carry
//	Reference  one site in a template that reads a variable
//	Bindings   the data a template would be executed against
//	Verdict    the offences found, or none
//	Auditor    the port a renderer depends on to obtain a Verdict
//
// A site that is deliberately allowed to render blank declares itself:
//
//	delim[[ optional .field "reason the blank is acceptable" ]]
//
// What counts as void, precisely:
//
//	absent      the key is in no binding                        REFUSED
//	nil         the key is bound and carries no value           REFUSED
//	empty       the key is bound to an empty or blank string    REFUSED
//	zero        0, 0.0, false                                   allowed
//	empty list  an empty slice or map                           allowed
//
// Scope of the audit: References resolves reads against the root data only.
// Inside the body of a range or a with the dot is rebound, so those reads are
// left to the template option missingkey=error, which fails the execution on
// the first miss and cannot be waived by optional.
package strict
