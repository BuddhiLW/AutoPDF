// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package typst_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/typst"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	infra "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/infrastructure/adapters"
)

func requireTypst(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst binary not on PATH; skipping real-toolchain test (install typst-cli to run it)")
	}
}

func realAdapter(t *testing.T, dir string) *typst.TypstCompilerAdapter {
	t.Helper()
	return typst.NewTypstCompilerAdapter(infra.NewOSFileSystem(), infra.NewOSCommandExecutor(),
		typst.WithWorkingDir(dir), typst.WithTimeout(2*time.Minute))
}

func TestRealTypstCompilesAPDF(t *testing.T) {
	requireTypst(t)
	dir := t.TempDir()
	out := filepath.Join(dir, "out", "doc.pdf")
	opts := ports.NewCompileOptions("typst", out, dir).WithJobName("doc").
		WithBackendOptions(ports.TypstOptions{Inputs: map[string]string{"who": "autopdf"}})

	path, err := realAdapter(t, dir).Compile(t.Context(),
		"= Real typst\n\nHello #sys.inputs.at(\"who\").\n", opts)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read PDF: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatalf("output does not start with %%PDF-")
	}
	if _, err := os.Stat(filepath.Join(dir, "doc"+typst.StagedSuffix)); !os.IsNotExist(err) {
		t.Errorf("staged file survived a non-debug compile (stat err %v)", err)
	}
}

func TestRealTypstReportsDiagnostics(t *testing.T) {
	requireTypst(t)
	dir := t.TempDir()
	out := filepath.Join(dir, "doc.pdf")
	opts := ports.NewCompileOptions("typst", out, dir).WithJobName("doc")

	_, err := realAdapter(t, dir).Compile(t.Context(), "= Broken\n\n#undefined-function()\n", opts)
	var ce *typst.CompileError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v, want *typst.CompileError", err)
	}
	errs := ce.Errors()
	if len(errs) == 0 {
		t.Fatalf("no error diagnostics parsed from %q", ce.Output)
	}
	if errs[0].Line != 3 {
		t.Errorf("first error at line %d, want 3: %+v", errs[0].Line, errs[0])
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Error("a failed compile left a PDF")
	}
}

func TestRealMissingBinaryIsToolchainMissing(t *testing.T) {
	dir := t.TempDir()
	adapter := typst.NewTypstCompilerAdapter(infra.NewOSFileSystem(), infra.NewOSCommandExecutor(),
		typst.WithWorkingDir(dir), typst.WithBinary("typst-autopdf-definitely-absent"))
	_, err := adapter.Compile(t.Context(), "= x\n",
		ports.NewCompileOptions("typst", filepath.Join(dir, "x.pdf"), dir))
	var missing *ports.ToolchainMissingError
	if !errors.As(err, &missing) || missing.Binary != "typst-autopdf-definitely-absent" {
		t.Fatalf("err = %v, want *ToolchainMissingError naming the binary", err)
	}
}
