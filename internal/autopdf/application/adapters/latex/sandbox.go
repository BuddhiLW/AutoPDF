// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex

import "os"

// noShellEscape is the engine flag that disables \write18 outright, instead
// of TeX Live's default restricted whitelist.
const noShellEscape = "-no-shell-escape"

// sandboxEnv is the environment every LaTeX run gets: the caller's, plus
// kpathsea's paranoid file access. openin_any=p refuses absolute paths and
// ".." outside the output directory, so a template that \input's
// /proc/self/environ cannot read it; the staged input must therefore be
// passed relative to the working directory.
func sandboxEnv() []string {
	return append(os.Environ(),
		"openin_any=p",
		"openout_any=p",
		"shell_escape=f")
}
