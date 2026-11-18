// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"errors"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/typst"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

// EngineTypst is the engine name of the Typst backend.
const EngineTypst = typst.EngineName

// ErrUnsupportedSetting is matched (errors.Is) by every *UnsupportedSettingError.
var ErrUnsupportedSetting = errors.New("unsupported backend setting")

// UnsupportedSettingError reports a configuration value a backend cannot
// honour, e.g. passes > 1 or use_latexmk with the Typst backend.
type UnsupportedSettingError struct {
	Backend string
	Setting string
	Value   any
}

func (e *UnsupportedSettingError) Error() string {
	return fmt.Sprintf("backend %q does not support %s=%v", e.Backend, e.Setting, e.Value)
}

// Is makes errors.Is(err, ErrUnsupportedSetting) hold.
func (e *UnsupportedSettingError) Is(target error) bool { return target == ErrUnsupportedSetting }

// Typst returns the Typst backend descriptor. Typst compiles in one pass and
// leaves no auxiliary files.
func Typst() Backend {
	return Backend{
		Name:               ports.TypstBackendName,
		EngineNames:        []string{EngineTypst},
		TemplateExtensions: []string{".typ"},
		StagedSuffix:       typst.StagedSuffix,
		AuxExtensions:      nil,
		NewCompiler:        newTypstCompiler,
		NewOptions:         typstOptions,
	}
}

func typstOptions(s Settings) (ports.BackendOptions, error) {
	if s.Passes > 1 {
		return nil, &UnsupportedSettingError{Backend: ports.TypstBackendName, Setting: "passes", Value: s.Passes}
	}
	if s.UseLatexmk {
		return nil, &UnsupportedSettingError{Backend: ports.TypstBackendName, Setting: "use_latexmk", Value: true}
	}
	return ports.DefaultTypstOptions(), nil
}

func newTypstCompiler(deps CompilerDeps, opts ports.BackendOptions) (ports.DocumentCompiler, error) {
	if _, err := ports.TypstOptionsFrom(ports.CompileOptions{Backend: opts}); err != nil {
		return nil, err
	}
	return typst.NewTypstCompilerAdapter(deps.FileSystem, deps.Executor, typst.WithWorkingDir(deps.WorkingDir)), nil
}
