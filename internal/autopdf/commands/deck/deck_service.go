// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package deck

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/commands/common"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// DeckServiceCmd builds a PDF presentation from a DocumentSpec JSON file.
var DeckServiceCmd = &bonzai.Cmd{
	Name:    `deck`,
	Alias:   `d`,
	Short:   `build a PDF presentation from a DocumentSpec`,
	Usage:   `SPEC.json [OUTPUT.pdf] [KEY=VALUE...]`,
	MinArgs: 1,
	MaxArgs: 12,
	Long: `
The deck command compiles a DocumentSpec into a PDF presentation.

A DocumentSpec is a renderer-independent document tree. plato produces one
from Markdown or Org:

  plato spec talk.org -o talk.json
  autopdf deck talk.json talk.pdf assets=./public theme=Madrid

Options are KEY=VALUE pairs:

  target=NAME       render target (default: beamer)
  title=TEXT        title page; omitted when absent
  author=TEXT       title page author
  date=TEXT         title page date
  theme=NAME        Beamer theme, e.g. Madrid, metropolis
  colortheme=NAME   Beamer colour theme
  style=FILE        a .sty to load after the theme
  assets=DIR        directory images are resolved against
  aspect=RATIO      aspect ratio (default: 169)
  notes=on          render speaker notes
  focus=ID[,ID]     rebuild only these sections
  engine=NAME       LaTeX engine (default: pdflatex)
  passes=N          compiler passes (default: 2)

Examples:
  autopdf deck talk.json
  autopdf deck talk.json out.pdf theme=metropolis assets=./public
  autopdf deck talk.json out.pdf focus=intro
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		ctx, logger := common.CreateStandardLoggerContext()
		defer func() { _ = logger.Sync() }()
		return executeDeckProcess(ctx, args)
	},
}

// request is everything the pipeline needs, with no I/O performed yet.
type request struct {
	specPath   string
	outputPath string
	stylePath  string
	target     Target
	settings   Settings
	focus      []string
	engine     string
	passes     int
}

// parseRequest promotes raw arguments into a validated request.
func parseRequest(args []string) (request, error) {
	if len(args) == 0 {
		return request{}, fmt.Errorf("a DocumentSpec path is required")
	}

	parsed := request{specPath: args[0], engine: "pdflatex", passes: 2}
	options := map[string]string{}
	for _, argument := range args[1:] {
		key, value, isPair := strings.Cut(argument, "=")
		if !isPair {
			if parsed.outputPath != "" {
				return request{}, fmt.Errorf("unexpected argument %q; options are KEY=VALUE", argument)
			}
			parsed.outputPath = argument
			continue
		}
		options[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}

	if parsed.outputPath == "" {
		parsed.outputPath = strings.TrimSuffix(parsed.specPath, filepath.Ext(parsed.specPath)) + ".pdf"
	}

	targetName := options["target"]
	if targetName == "" {
		targetName = "beamer"
	}
	target, err := LookupTarget(targetName)
	if err != nil {
		return request{}, err
	}
	parsed.target = target

	parsed.settings = Settings{
		Title:        options["title"],
		Author:       options["author"],
		Date:         options["date"],
		Theme:        options["theme"],
		ColorTheme:   options["colortheme"],
		GraphicsPath: options["assets"],
		AspectRatio:  options["aspect"],
		ShowNotes:    isEnabled(options["notes"]),
	}
	if parsed.settings.GraphicsPath != "" {
		absolute, err := filepath.Abs(parsed.settings.GraphicsPath)
		if err != nil {
			return request{}, fmt.Errorf("resolve assets directory: %w", err)
		}
		parsed.settings.GraphicsPath = absolute
	}
	parsed.stylePath = options["style"]
	parsed.settings.StyleFile = strings.TrimSuffix(filepath.Base(parsed.stylePath), ".sty")

	if focus := options["focus"]; focus != "" {
		for _, id := range strings.Split(focus, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				parsed.focus = append(parsed.focus, trimmed)
			}
		}
	}
	if engine := options["engine"]; engine != "" {
		parsed.engine = engine
	}
	if passes := options["passes"]; passes != "" {
		count, err := strconv.Atoi(passes)
		if err != nil || count < 1 {
			return request{}, fmt.Errorf("passes must be a positive integer, got %q", passes)
		}
		parsed.passes = count
	}
	return parsed, nil
}

func isEnabled(value string) bool {
	switch strings.ToLower(value) {
	case "on", "true", "yes", "1":
		return true
	default:
		return false
	}
}

// executeDeckProcess is the effect boundary: it reads the spec, runs the pure
// pipeline, and writes the PDF.
func executeDeckProcess(ctx context.Context, args []string) error {
	parsed, err := parseRequest(args)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(parsed.specPath)
	if err != nil {
		return fmt.Errorf("read document spec: %w", err)
	}
	spec, err := document.Decode(data)
	if err != nil {
		return fmt.Errorf("decode document spec: %w", err)
	}

	// The compile runs in a private workspace, so a style file is carried into
	// the projection rather than referenced where it sits.
	if parsed.stylePath != "" {
		style, err := os.ReadFile(parsed.stylePath)
		if err != nil {
			return fmt.Errorf("read style file: %w", err)
		}
		parsed.settings.StyleContent = style
	}

	engine, err := NewEngine(parsed.target, parsed.settings, 0)
	if err != nil {
		return err
	}
	generator, err := api.NewProjectionGenerator(api.ProjectionGeneratorOptions{
		OutputPath:  parsed.outputPath,
		LaTeXEngine: parsed.engine,
		Passes:      parsed.passes,
	})
	if err != nil {
		return fmt.Errorf("create projection generator: %w", err)
	}

	result, err := engine.Generate(ctx, spec,
		api.ProjectionOptions{FocusSections: parsed.focus}, generator)
	if err != nil {
		return err
	}

	fmt.Printf("wrote %s (%d bytes, %d frames)\n",
		parsed.outputPath, len(result.PDF), len(spec.Blocks))
	return nil
}
