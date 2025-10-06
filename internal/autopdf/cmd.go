// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package autopdf provides the Bonzai command branch of the same name.
// This package now uses SOLID principles, Domain-Driven Design, and Gang of Four patterns.
package autopdf

import (
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/bonzai/vars"
)

// 🏗️  Architecture: SOLID + DDD + GoF Patterns
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
// This version uses a refactored architecture following:
//
// SOLID Principles:
//   • Single Responsibility: Each service has one clear purpose
//   • Open/Closed: Extensible through interfaces and strategies
//   • Liskov Substitution: All implementations are interchangeable
//   • Interface Segregation: Small, focused interfaces
//   • Dependency Inversion: Depends on abstractions, not concretions
//
// Domain-Driven Design:
//   • Domain Services: Business logic orchestration
//   • Value Objects: Immutable data with validation
//   • Entities: Objects with identity and lifecycle
//   • Factories: Complex object creation
//   • Events: Loose coupling through event-driven architecture
//
// Gang of Four Patterns:
//   • Factory Pattern: Engine creation and selection
//   • Strategy Pattern: Template processing strategies
//   • Observer Pattern: Event-driven architecture

// # Features:
// - Generate pdfs from latex templates
// - Generate images from pdfs
// - Clean up auxiliary files generated during compilation
// - Define templates, variables, LaTeX engine, output settings, and conversion options
// - Event-driven architecture for extensibility
// - Multiple engine support (pdflatex, xelatex, lualatex)

var Cmd = &bonzai.Cmd{
	Name:  `autopdf`,
	Alias: `apdf`,
	Vers:  `v2.0.0`,
	Short: `generate pdfs from latex templates`,
	Long: `
The autopdf tool helps generate pdfs from latex templates. It simplifies common latex
operations and workflow management.

# Commands:
- build:    Process template and compile to PDF (uses Application Layer)
- debug:    Enable verbose debug output for build operations
- clean:    Remove LaTeX auxiliary files (uses Domain Layer)
- convert:  Convert PDF to images (uses Strategy Pattern)
- compile:  Compile LaTeX to PDF (uses Factory Pattern)
- vars:     View and set configuration variables

Use 'autopdf help <command> <subcommand>...' for detailed information
about each command.
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
		vars.Cmd,
		commands.BuildCmd,
		commands.DebugCmd,
		commands.CleanCmd,
		commands.ConvertCmd,
		commands.CompileCmd,
	},
	Def: help.Cmd,
}
