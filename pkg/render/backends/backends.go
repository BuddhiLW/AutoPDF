// Package backends is the composition root for projection backends: the one
// place that names every concrete render target pkg/api can select by name.
package backends

import (
	"fmt"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/typst"
)

// Typst names the Typst projection backend.
const Typst = typst.Target

// TypstBackend projects manifests into a Typst workspace (main.typ plus
// fragments and assets) and maps `typst compile` short diagnostics.
func TypstBackend() api.ProjectionBackend {
	return api.ProjectionBackend{
		Name:          Typst,
		Engine:        Typst,
		MainExtension: ".typ",
		NewProjector: func(config api.DocumentEngineConfig) api.ManifestProjector {
			return typst.NewProjector(typst.Options{TrustedPreamble: config.TrustedPreamble})
		},
		Builtins: func(target string) ([]component.Definition, error) {
			if target != typst.Target {
				return nil, fmt.Errorf("typst builtins register under target %q, not %q", typst.Target, target)
			}
			return typst.Definitions()
		},
		Diagnostics:       typst.DiagnosticPattern,
		RootRelativePaths: true,
	}
}

// Default returns the LaTeX and Typst projection backends, keyed by name.
func Default() api.ProjectionBackends {
	selected := api.DefaultProjectionBackends()
	selected[Typst] = TypstBackend()
	return selected
}
