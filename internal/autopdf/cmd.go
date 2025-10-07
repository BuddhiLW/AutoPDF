// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package autopdf provides the Bonzai command branch of the same name.
package autopdf

import (
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/build"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/convert"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/clean"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/debug"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/force"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/verbose"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/watch"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/vars"
)

var Cmd = &bonzai.Cmd{
	Name:  `autopdf`,
	Alias: `apdf`,
	Vers:  `v1.2.0`,
	Short: `generate pdfs from latex templates`,
	Long: `
The autopdf tool helps generate pdfs from latex templates. It simplifies common latex
operations and workflow management.

# Features:
- Generate pdfs from latex templates
- Generate images from pdfs
- Clean up auxiliary files generated during compilation
- Define templates, variables, LaTeX engine, output settings, and conversion options.

# Commands:
- build:    Process template and compile to PDF
- convert:  Convert PDF to images
- clean:    Remove LaTeX auxiliary files
- verbose:  Set verbose logging level
- debug:    Enable debug information output
- force:    Enable force operations
- vars:     View and set configuration variables

Use 'autopdf help <command> <subcommand>...' for detailed information
about each command.
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
		vars.Cmd,
		build.BuildServiceCmd,     // Use new service-based build command
		convert.ConvertServiceCmd, // Use new service-based convert command
		clean.CleanServiceCmd,     // Use new service-based clean command
		verbose.VerboseServiceCmd, // Use new service-based verbose command
		debug.DebugServiceCmd,     // Use new service-based debug command
		force.ForceServiceCmd,     // Use new service-based force command
		watch.WatchServiceCmd,     // Use new service-based watch command
	},
	Def: help.Cmd,
}
