package preview_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
)

func BenchmarkLaTeXPreviewWarmTwoPage(b *testing.B) {
	requirePreviewTools(b)
	compiler, err := preview.NewLaTeXCompiler(preview.LaTeXCompilerOptions{Engine: "pdflatex", DPI: 72})
	if err != nil {
		b.Fatal(err)
	}
	workspace := b.TempDir()
	source := []byte("\\documentclass{article}\n\\begin{document}\nFirst page.\\newpage\nSecond page.\n\\end{document}\n")
	if err := os.WriteFile(filepath.Join(workspace, "main.tex"), source, 0o644); err != nil {
		b.Fatal(err)
	}
	request := preview.CompileRequest{Workspace: workspace, Main: "main.tex"}
	if _, err := compiler.Compile(context.Background(), request); err != nil {
		b.Fatal(err)
	}
	var tex, fingerprint, raster time.Duration
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		request.Revision = uint64(index + 2)
		output, err := compiler.Compile(context.Background(), request)
		if err != nil {
			b.Fatal(err)
		}
		tex += output.Timings.TeX
		fingerprint += output.Timings.Fingerprint
		raster += output.Timings.Raster
	}
	b.StopTimer()
	if b.N > 0 {
		b.ReportMetric(float64(tex.Microseconds())/float64(b.N)/1000, "tex-ms/op")
		b.ReportMetric(float64(fingerprint.Microseconds())/float64(b.N)/1000, "fingerprint-ms/op")
		b.ReportMetric(float64(raster.Microseconds())/float64(b.N)/1000, "raster-ms/op")
	}
}

func BenchmarkLaTeXPreviewFocusedSectionEdit(b *testing.B) {
	requirePreviewTools(b)
	compiler, err := preview.NewLaTeXCompiler(preview.LaTeXCompilerOptions{Engine: "pdflatex", DPI: 72})
	if err != nil {
		b.Fatal(err)
	}
	workspace := b.TempDir()
	main := []byte("\\documentclass{article}\n\\includeonly{section-a}\n\\begin{document}\n\\include{section-a}\n\\include{section-b}\n\\end{document}\n")
	if err := os.WriteFile(filepath.Join(workspace, "main.tex"), main, 0o644); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "section-b.tex"), []byte("Unchanged section."), 0o644); err != nil {
		b.Fatal(err)
	}
	request := preview.CompileRequest{Workspace: workspace, Main: "main.tex"}
	var tex, fingerprint, raster time.Duration
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		content := []byte("Focused revision A.")
		if index%2 == 1 {
			content = []byte("Focused revision B.")
		}
		if err := os.WriteFile(filepath.Join(workspace, "section-a.tex"), content, 0o644); err != nil {
			b.Fatal(err)
		}
		request.Revision = uint64(index + 1)
		output, err := compiler.Compile(context.Background(), request)
		if err != nil {
			b.Fatal(err)
		}
		tex += output.Timings.TeX
		fingerprint += output.Timings.Fingerprint
		raster += output.Timings.Raster
	}
	b.StopTimer()
	if b.N > 0 {
		b.ReportMetric(float64(tex.Microseconds())/float64(b.N)/1000, "tex-ms/op")
		b.ReportMetric(float64(fingerprint.Microseconds())/float64(b.N)/1000, "fingerprint-ms/op")
		b.ReportMetric(float64(raster.Microseconds())/float64(b.N)/1000, "raster-ms/op")
	}
}

func requirePreviewTools(b *testing.B) {
	b.Helper()
	for _, executable := range []string{"pdflatex", "pdftoppm"} {
		if _, err := exec.LookPath(executable); err != nil {
			b.Skip(executable + " unavailable")
		}
	}
}
