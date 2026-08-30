// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package deck builds a PDF presentation from a DocumentSpec.
package deck

import (
	"fmt"
	"sort"
	"strings"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/beamer"
)

// Target names one rendering target and supplies the two halves an engine
// needs: a component catalog and a manifest projector.
type Target struct {
	Name      string
	Catalog   func() (*component.Catalog, error)
	Projector func(Settings) api.ManifestProjector
}

// Settings are the presentation choices a command line can express.
type Settings struct {
	Title        string
	Author       string
	Date         string
	Theme        string
	ColorTheme   string
	StyleFile    string
	GraphicsPath string
	AspectRatio  string
	ShowNotes    bool
}

// Targets is the registry the CLI dispatches over. This is the composition
// root: naming concrete targets here is correct, and adding one is an entry in
// this map rather than a change to any package under pkg/.
var Targets = map[string]Target{
	beamer.Target: {
		Name:    beamer.Target,
		Catalog: beamer.Catalog,
		Projector: func(settings Settings) api.ManifestProjector {
			return beamer.NewProjector(beamer.Options{
				Title:        settings.Title,
				Author:       settings.Author,
				Date:         settings.Date,
				Theme:        settings.Theme,
				ColorTheme:   settings.ColorTheme,
				StyleFile:    settings.StyleFile,
				GraphicsPath: settings.GraphicsPath,
				AspectRatio:  settings.AspectRatio,
				ShowNotes:    settings.ShowNotes,
			})
		},
	},
}

// TargetNames lists the registered targets in a stable order.
func TargetNames() []string {
	names := make([]string, 0, len(Targets))
	for name := range Targets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// LookupTarget resolves a target by name.
func LookupTarget(name string) (Target, error) {
	target, exists := Targets[strings.TrimSpace(name)]
	if !exists {
		return Target{}, fmt.Errorf("unknown render target %q; known targets: %s",
			name, strings.Join(TargetNames(), ", "))
	}
	return target, nil
}

// NewEngine wires a target's catalog and projector into a document engine.
func NewEngine(target Target, settings Settings, maxWorkers int) (*api.DocumentEngine, error) {
	catalog, err := target.Catalog()
	if err != nil {
		return nil, fmt.Errorf("build %s catalog: %w", target.Name, err)
	}
	return api.NewDocumentEngine(catalog, api.DocumentEngineConfig{
		Target:     target.Name,
		MaxWorkers: maxWorkers,
		Projector:  target.Projector(settings),
	})
}
