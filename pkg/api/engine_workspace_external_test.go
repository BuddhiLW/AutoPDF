// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultEngineIsolatesConcurrentRequests(t *testing.T) {
	testDir := t.TempDir()
	templatePath := filepath.Join(testDir, "template.tex")
	require.NoError(t, os.WriteFile(templatePath, []byte("request=delim[[.request]]"), 0644))
	compilerPath := writeFakeCompiler(t, testDir, 0)
	before := workspaceSnapshot(t)

	engine, err := api.NewEngine()
	require.NoError(t, err)

	const requestCount = 6
	type outcome struct {
		index  int
		result api.Result
		err    error
	}
	outcomes := make(chan outcome, requestCount)
	var group sync.WaitGroup
	for index := 0; index < requestCount; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			result, err := engine.Generate(context.Background(), api.Request{
				TemplatePath: templatePath,
				OutputPath:   filepath.Join(testDir, "out", fmt.Sprintf("request-%d", index)),
				Variables:    map[string]string{"request": fmt.Sprintf("value-%d", index)},
				LaTeXEngine:  compilerPath,
			})
			outcomes <- outcome{index: index, result: result, err: err}
		}(index)
	}
	group.Wait()
	close(outcomes)

	for outcome := range outcomes {
		require.NoError(t, outcome.err)
		assert.Contains(t, string(outcome.result.PDF), fmt.Sprintf("request=value-%d", outcome.index))
		assert.Equal(t, filepath.Join(testDir, "out", fmt.Sprintf("request-%d.pdf", outcome.index)), outcome.result.PDFPath)
		published, err := os.ReadFile(outcome.result.PDFPath)
		require.NoError(t, err)
		assert.Equal(t, outcome.result.PDF, published)
	}
	assert.Equal(t, before, workspaceSnapshot(t))
}

func TestDefaultEngineCancelsCompilerAndCleansWorkspace(t *testing.T) {
	testDir := t.TempDir()
	templatePath := filepath.Join(testDir, "template.tex")
	require.NoError(t, os.WriteFile(templatePath, []byte("cancel me"), 0644))
	compilerPath := writeFakeCompiler(t, testDir, 5*time.Second)
	before := workspaceSnapshot(t)

	engine, err := api.NewEngine()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	started := time.Now()
	_, err = engine.Generate(ctx, api.Request{
		TemplatePath: templatePath,
		OutputPath:   filepath.Join(testDir, "cancelled.pdf"),
		LaTeXEngine:  compilerPath,
	})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(started), 2*time.Second)
	assert.NoFileExists(t, filepath.Join(testDir, "cancelled.pdf"))
	assert.Equal(t, before, workspaceSnapshot(t))
}

func writeFakeCompiler(t *testing.T, directory string, delay time.Duration) string {
	t.Helper()
	path := filepath.Join(directory, fmt.Sprintf("fake-latex-%d", delay))
	delayCommand := ":"
	if delay > 0 {
		delayCommand = "while :; do :; done"
	}
	script := fmt.Sprintf(`#!/bin/sh
job=document
out=.
source_file=
for arg in "$@"; do
  case "$arg" in
    -jobname=*) job=${arg#-jobname=} ;;
    -output-directory=*) out=${arg#-output-directory=} ;;
    *.tex) source_file=$arg ;;
  esac
done
%s
mkdir -p "$out"
{
  printf '%%%%PDF-1.4\n'
  test -z "$source_file" || sed -n '1,40p' "$source_file"
} > "$out/$job.pdf"
`, delayCommand)
	require.NoError(t, os.WriteFile(path, []byte(script), 0755))
	return path
}

func workspaceSnapshot(t *testing.T) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(os.TempDir(), "autopdf-request-*"))
	require.NoError(t, err)
	return strings.Join(paths, "\n")
}
