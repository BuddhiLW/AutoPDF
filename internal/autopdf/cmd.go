// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package autopdf provides the Bonzai command branch of the same name.
package autopdf

import (
	"fmt"

	"github.com/BuddhiLW/AutoPDF/internal/build"
	"github.com/BuddhiLW/AutoPDF/internal/config"
	"github.com/BuddhiLW/AutoPDF/internal/converter"
	"github.com/BuddhiLW/AutoPDF/internal/tex"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/cmds/vars"
	"github.com/rwxrob/bonzai/comp"
)

var Cmd = &bonzai.Cmd{
	Name:  `autopdf`,
	Alias: `apdf`,
	Vers:  `v0.1.0`,
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
- clean:    Remove LaTeX auxiliary files
- convert:  Convert PDF to images
- vars:     View and set configuration variables

Use 'autopdf help <command> <subcommand>...' for detailed information
about each command.
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
		vars.Cmd,
		build.Cmd,
		tex.CleanCmd,
		convertCmd,
	},
	Def: help.Cmd,
}

var convertCmd = &bonzai.Cmd{
	Name:    `convert`,
	Alias:   `c`,
	Short:   `convert PDF to images`,
	Usage:   `PDF [FORMAT...]`,
	MinArgs: 1,
	Long: `
The convert command takes a PDF file and converts it to one or more image formats.
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		pdfFile := args[0]
		formats := []string{"png"}

		if len(args) > 1 {
			formats = args[1:]
		}

		// Create a minimal config for the converter
		cfg := &config.Config{
			Conversion: struct {
				Enabled bool     `yaml:"enabled" json:"enabled"`
				Formats []string `yaml:"formats" json:"formats"`
			}{
				Enabled: true,
				Formats: formats,
			},
		}

		// Convert the PDF to images
		conv := converter.NewConverter(cfg)
		imageFiles, err := conv.ConvertPDFToImages(pdfFile)
		if err != nil {
			return fmt.Errorf("PDF to image conversion failed: %w", err)
		}

		if len(imageFiles) > 0 {
			fmt.Println("Generated image files:")
			for _, file := range imageFiles {
				fmt.Printf("  - %s\n", file)
			}
		} else {
			fmt.Println("No image files were generated")
		}

		return nil
	},
}
