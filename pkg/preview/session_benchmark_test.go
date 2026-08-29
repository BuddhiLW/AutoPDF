package preview_test

import (
	"context"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/preview"
)

func BenchmarkWarmPreviewFeedback(b *testing.B) {
	compiler := compilerFunc(func(context.Context, preview.CompileRequest) (preview.CompileOutput, error) {
		return preview.CompileOutput{Pages: []preview.Page{{Number: 1, MediaType: "image/png", Hash: "stable"}}}, nil
	})
	session, err := preview.NewSession(compiler, preview.Options{WorkspaceRoot: b.TempDir()})
	if err != nil {
		b.Fatal(err)
	}
	defer session.Close()
	value := projection("stable", "document")
	if result := <-session.Submit(context.Background(), value); result.Err != nil {
		b.Fatal(result.Err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if result := <-session.Submit(context.Background(), value); result.Err != nil {
			b.Fatal(result.Err)
		}
	}
}
