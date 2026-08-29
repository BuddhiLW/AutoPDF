// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package parallel

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/compilation"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/domain/parallel"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParallelCompilationProducesPDFs exercises the whole path the `multiple`
// command uses: the engine strategy behind the CompilationStrategy port, the
// bounded orchestrator, and the aggregate result.
//
// Before strategies were injected, findCompilationStrategy returned nil for
// every template and this path could only ever report failures.
func TestParallelCompilationProducesPDFs(t *testing.T) {
	if _, err := exec.LookPath("pdflatex"); err != nil {
		t.Skip("pdflatex unavailable")
	}

	workspace := t.TempDir()
	configPath := filepath.Join(workspace, "autopdf.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("engine: \"pdflatex\"\n"), 0o644))

	templates := make([]string, 0, 3)
	for _, name := range []string{"alpha", "beta", "gamma"} {
		path := filepath.Join(workspace, name+".tex")
		body := fmt.Sprintf(
			"\\documentclass{article}\n\\begin{document}\nHello from %s.\n\\end{document}\n", name)
		require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
		templates = append(templates, path)
	}

	engine, err := api.NewEngine()
	require.NoError(t, err)
	strategy := compilation.NewEngineStrategy(engine, compilation.WithLaTeXEngine("pdflatex"))

	orchestrator := NewParallelExecutionOrchestrator(strategy)
	service := NewParallelCompilationService(
		orchestrator, nil, []parallel.CompilationStrategy{strategy})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	result, err := service.CompileTemplates(ctx, parallel.ParallelCompilationRequest{
		ConfigurationFile: configPath,
		TemplateFiles:     templates,
		MaxConcurrency:    2,
		Timeout:           90 * time.Second,
	})
	require.NoError(t, err)

	for _, failure := range result.FailedBuilds {
		t.Logf("failed build %s: %v", failure.TemplateFile, failure.Error)
	}
	assert.Equal(t, len(templates), result.SuccessCount,
		"every template should compile through the injected strategy")
	assert.Zero(t, result.FailureCount)

	for _, build := range result.SuccessfulBuilds {
		assert.FileExists(t, build.PDFPath)
		assert.NotZero(t, build.Duration, "per-build duration should be recorded")
	}
	assert.NotZero(t, result.TotalDuration)
}
