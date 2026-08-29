package preview_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

func TestWarmPreviewOrchestrationLatencyBudget(t *testing.T) {
	compiler := compilerFunc(func(context.Context, preview.CompileRequest) (preview.CompileOutput, error) {
		return preview.CompileOutput{Pages: []preview.Page{{Number: 1, MediaType: "image/png", Hash: "stable"}}}, nil
	})
	session, err := preview.NewSession(compiler, preview.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	value := projection("stable", "document")
	if result := <-session.Submit(context.Background(), value); result.Err != nil {
		t.Fatal(result.Err)
	}
	const iterations = 200
	started := time.Now()
	for index := 0; index < iterations; index++ {
		if result := <-session.Submit(context.Background(), value); result.Err != nil {
			t.Fatal(result.Err)
		}
	}
	average := time.Since(started) / iterations
	if average > 10*time.Millisecond {
		t.Fatalf("warm preview orchestration budget exceeded: average=%s budget=10ms", average)
	}
}

func TestPreviewMaterializationLatencyBudget100Files(t *testing.T) {
	compiler := compilerFunc(func(context.Context, preview.CompileRequest) (preview.CompileOutput, error) {
		return preview.CompileOutput{}, nil
	})
	session, err := preview.NewSession(compiler, preview.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	value := latex.Projection{Main: "main.tex", Hash: "hundred"}
	value.Files = append(value.Files, latex.File{Path: "main.tex", Content: []byte("main")})
	for index := 0; index < 100; index++ {
		value.Files = append(value.Files, latex.File{Path: fmt.Sprintf("fragments/%03d.tex", index), Content: []byte("fragment")})
	}
	first := <-session.Submit(context.Background(), value)
	if first.Err != nil {
		t.Fatal(first.Err)
	}
	second := <-session.Submit(context.Background(), value)
	if second.Err != nil {
		t.Fatal(second.Err)
	}
	if first.Timings.Materialize > 250*time.Millisecond {
		t.Fatalf("cold materialization budget exceeded: %s", first.Timings.Materialize)
	}
	if second.Timings.Materialize > 25*time.Millisecond {
		t.Fatalf("warm materialization budget exceeded: %s", second.Timings.Materialize)
	}
}
