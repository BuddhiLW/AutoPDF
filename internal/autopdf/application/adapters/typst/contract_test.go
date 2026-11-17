// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package typst_test

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/typst"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports/compilertest"
)

const contractDocument = "= compilertest\n\nHello from #emph[Typst].\n"

var typstTemplateExtensions = []string{".typ"}

func TestTypstCompilerAdapterHonoursTheDocumentCompilerContract(t *testing.T) {
	for name, backend := range map[string]ports.BackendOptions{
		"DefaultOptions": nil,
		"WithOptions": ports.TypstOptions{
			FontPaths: []string{"/fonts"},
			Inputs:    map[string]string{"title": "contract"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			compilertest.Run(t, func(t *testing.T, fs ports.FileSystem, exec ports.CommandExecutor) compilertest.Subject {
				return compilertest.Subject{
					Compiler:           typst.NewTypstCompilerAdapter(fs, exec, typst.WithWorkingDir(t.TempDir())),
					Engine:             typst.EngineName,
					Backend:            backend,
					Content:            contractDocument,
					Binaries:           []string{"typst"},
					TemplateExtensions: typstTemplateExtensions,
				}
			})
		})
	}
}
