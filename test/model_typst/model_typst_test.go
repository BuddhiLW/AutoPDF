// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package model_typst_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/pdf"
)

// exampleFiles are the files of the test/model_typst example the CLI reads.
var exampleFiles = []string{"config.yaml", "main.typ"}

// TestModelTypstBuildsACompletePDF runs the real `autopdf build` on this
// example and asserts the published PDF is complete. It needs the typst
// toolchain on PATH (`make typst` puts one in ./bin).
func TestModelTypstBuildsACompletePDF(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst not found on PATH: skipping the Typst end-to-end build " +
			"(run `make typst` and add ./bin to PATH, or `cargo install typst-cli --locked`)")
	}
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not found on PATH: cannot build the autopdf binary")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	binary := filepath.Join(t.TempDir(), "autopdf")
	build := exec.CommandContext(ctx, goTool, "build", "-o", binary, "./cmd/autopdf")
	build.Dir = filepath.Join("..", "..")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/autopdf: %v\n%s", err, out)
	}

	workDir := t.TempDir()
	for _, name := range exampleFiles {
		content, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(workDir, name), content, 0o644); err != nil {
			t.Fatalf("copy %s: %v", name, err)
		}
	}

	run := exec.CommandContext(ctx, binary, "build", "main.typ", "config.yaml")
	run.Dir = workDir
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("autopdf build main.typ config.yaml: %v\n%s", err, out)
	}

	pdfPath := filepath.Join(workDir, "out", "output.pdf")
	content, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("read built PDF: %v\n%s", err, out)
	}
	if err := pdf.Check(pdfPath, content); err != nil {
		t.Fatalf("built PDF is not complete: %v", err)
	}

	entries, err := os.ReadDir(workDir)
	if err != nil {
		t.Fatalf("read work dir: %v", err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".autopdf.typ") {
			t.Fatalf("staged file %s left behind without debug", entry.Name())
		}
	}
}
