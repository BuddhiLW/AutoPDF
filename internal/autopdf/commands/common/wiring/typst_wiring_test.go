// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package wiring

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/cleaner"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports/compilertest"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/commands/common/args"
	configPkg "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/commands/common/config"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/domain/options"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTypstProject writes a .typ template and a config.yaml beside it.
func writeTypstProject(t *testing.T, configYAML string) (templatePath, configPath string) {
	t.Helper()
	dir := t.TempDir()
	templatePath = filepath.Join(dir, "report.typ")
	require.NoError(t, os.WriteFile(templatePath, []byte("= delim[[.title]]\nBody with #strong[bold] and ]] literal.\n"), 0o644))
	configPath = filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configYAML), 0o644))
	return templatePath, configPath
}

func loadCLIConfig(t *testing.T, templatePath, configPath string) (*config.Config, error) {
	t.Helper()
	return configPkg.NewConfigResolver().LoadConfigWithLogging(context.Background(), templatePath, configPath)
}

// TestCLIBuildSelectsTypst runs the CLI build path (config resolution,
// ServiceBuilder, DocumentService) over a fake typst toolchain, once with the
// engine named in config.yaml and once chosen by the .typ extension.
func TestCLIBuildSelectsTypst(t *testing.T) {
	for name, engineLine := range map[string]string{
		"engine: typst in config": "engine: typst\n",
		"by .typ extension":       "",
	} {
		t.Run(name, func(t *testing.T) {
			templatePath, configPath := writeTypstProject(t,
				"template: report.typ\noutput: out/report.pdf\nvariables:\n  title: Hello\n"+engineLine)
			cfg, err := loadCLIConfig(t, templatePath, configPath)
			require.NoError(t, err)
			assert.Equal(t, "typst", cfg.Engine.String())

			fs := compilertest.NewMemFS()
			exec := compilertest.NewFakeExecutor(fs)
			outputPath := cfg.Output.String() // relative to the CLI's working directory
			require.Equal(t, "out/report.pdf", outputPath)
			exec.SetPDFPath(outputPath)

			builder := NewServiceBuilder().WithToolchain(fs, exec)
			svc := builder.BuildDocumentServiceWithWorkingDir(cfg, filepath.Dir(templatePath))
			req, err := builder.BuildRequest(&args.BuildArgs{ConfigFile: configPath, Options: options.NewBuildOptions()}, cfg)
			require.NoError(t, err)
			assert.Equal(t, ports.DefaultTypstOptions(), req.Backend)

			result, err := svc.Build(context.Background(), req)
			require.NoError(t, err)
			assert.Equal(t, outputPath, result.PDFPath)

			compiles := exec.CompileCommands()
			require.Len(t, compiles, 1, "typst compiles in a single pass")
			assert.Equal(t, "typst", compiles[0].Executable)
			assert.Equal(t, "compile", compiles[0].Args[0])
			source := compiles[0].Args[len(compiles[0].Args)-2]
			assert.Equal(t, ".typ", filepath.Ext(source))
			assert.Contains(t, filepath.Base(source), ".autopdf.typ")
		})
	}
}

func TestCLIConfigRefusesLaTeXOnlySettingsWithTypst(t *testing.T) {
	for name, extra := range map[string]string{
		"passes":      "engine: typst\npasses: 3\n",
		"use_latexmk": "use_latexmk: true\n",
	} {
		t.Run(name, func(t *testing.T) {
			templatePath, configPath := writeTypstProject(t, "template: report.typ\nvariables:\n  title: x\n"+extra)
			_, err := loadCLIConfig(t, templatePath, configPath)
			require.Error(t, err)
			assert.ErrorIs(t, err, backend.ErrUnsupportedSetting)
		})
	}

	builder := NewServiceBuilder()
	_, err := builder.BuildRequest(&args.BuildArgs{Options: options.NewBuildOptions()},
		&config.Config{Template: "doc.typ", Engine: "typst", Passes: 2})
	assert.ErrorIs(t, err, backend.ErrUnsupportedSetting)
}

func TestCLIConfigKeepsLaTeXDefaultForTex(t *testing.T) {
	dir := t.TempDir()
	templatePath := filepath.Join(dir, "doc.tex")
	require.NoError(t, os.WriteFile(templatePath, []byte("x"), 0o644))
	configPath := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("template: doc.tex\npasses: 2\n"), 0o644))
	cfg, err := loadCLIConfig(t, templatePath, configPath)
	require.NoError(t, err)
	assert.Equal(t, "pdflatex", cfg.Engine.String())
	assert.Equal(t, 2, cfg.Passes)
}

func TestCleanerIsPerBackend(t *testing.T) {
	builder := NewServiceBuilder()
	typstSvc := builder.BuildDocumentServiceWithWorkingDir(&config.Config{Template: "doc.typ"}, t.TempDir())
	typstCleaner, ok := typstSvc.Cleaner.(*cleaner.CleanerAdapter)
	require.True(t, ok)
	assert.Empty(t, typstCleaner.AuxExtensions(), "typst leaves no auxiliary files")

	latexSvc := builder.BuildDocumentServiceWithWorkingDir(&config.Config{Template: "doc.tex", Engine: "xelatex"}, t.TempDir())
	latexCleaner, ok := latexSvc.Cleaner.(*cleaner.CleanerAdapter)
	require.True(t, ok)
	assert.Contains(t, latexCleaner.AuxExtensions(), ".aux")

	// A .log beside typst output is not the typst backend's to remove.
	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "doc.pdf")
	logPath := filepath.Join(dir, "doc.log")
	require.NoError(t, os.WriteFile(logPath, []byte("keep"), 0o644))
	require.NoError(t, typstCleaner.Clean(context.Background(), pdfPath))
	assert.FileExists(t, logPath)
}
