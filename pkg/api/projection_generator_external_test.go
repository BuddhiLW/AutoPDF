package api_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/render/latex"
)

type assetResolverFunc func(context.Context, latex.AssetBinding) ([]byte, error)

func (function assetResolverFunc) Resolve(ctx context.Context, asset latex.AssetBinding) ([]byte, error) {
	return function(ctx, asset)
}

func TestProjectionGeneratorUsesCanonicalEngineAndCleansWorkspace(t *testing.T) {
	var workspace string
	engine, err := api.NewEngine(api.WithGenerator(api.GeneratorFunc(func(_ context.Context, request api.Request) (api.Result, error) {
		workspace = request.WorkingDir
		main, readErr := os.ReadFile(request.TemplatePath)
		if readErr != nil {
			return api.Result{}, readErr
		}
		fragment, readErr := os.ReadFile(filepath.Join(request.WorkingDir, "fragments", "body.tex"))
		if readErr != nil {
			return api.Result{}, readErr
		}
		asset, readErr := os.ReadFile(filepath.Join(request.WorkingDir, "assets", "logo.png"))
		if readErr != nil {
			return api.Result{}, readErr
		}
		return api.Result{PDF: append(append(main, fragment...), asset...)}, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	assetData := []byte("png")
	digest := sha256.Sum256(assetData)
	generator, err := api.NewProjectionGenerator(api.ProjectionGeneratorOptions{
		Engine: engine,
		Resolver: assetResolverFunc(func(context.Context, latex.AssetBinding) ([]byte, error) {
			return assetData, nil
		}),
		WorkspaceRoot: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	projection := latex.Projection{
		Main: "main.tex",
		Files: []latex.File{
			{Path: "main.tex", Content: []byte("main")},
			{Path: "fragments/body.tex", Content: []byte("body")},
		},
		Assets: []latex.AssetBinding{{AssetID: "logo", Path: "assets/logo.png", Digest: "sha256:" + hex.EncodeToString(digest[:])}},
	}
	result, err := generator.GenerateProjection(context.Background(), projection)
	if err != nil {
		t.Fatal(err)
	}
	if string(result.PDF) != "mainbodypng" {
		t.Fatalf("unexpected generated data %q", result.PDF)
	}
	if _, err := os.Stat(workspace); !os.IsNotExist(err) {
		t.Fatalf("projection workspace leaked: %s", workspace)
	}
}

func TestProjectionGeneratorRejectsTraversalAndDigestMismatch(t *testing.T) {
	generator, err := api.NewProjectionGenerator(api.ProjectionGeneratorOptions{
		Resolver:      assetResolverFunc(func(context.Context, latex.AssetBinding) ([]byte, error) { return []byte("wrong"), nil }),
		WorkspaceRoot: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = generator.GenerateProjection(context.Background(), latex.Projection{Main: "../main.tex", Files: []latex.File{{Path: "../main.tex"}}})
	if !errors.Is(err, api.ErrProjectionPathUnsafe) {
		t.Fatalf("expected unsafe path, got %v", err)
	}
	_, err = generator.GenerateProjection(context.Background(), latex.Projection{
		Main: "main.tex", Files: []latex.File{{Path: "main.tex"}},
		Assets: []latex.AssetBinding{{AssetID: "logo", Path: "logo.png", Digest: strings.Repeat("0", 64)}},
	})
	if !errors.Is(err, api.ErrProjectionDigestMismatch) {
		t.Fatalf("expected digest mismatch, got %v", err)
	}
}

func TestDocumentEngineGeneratesPDFThroughProductionProjectionAdapter(t *testing.T) {
	pdflatex, err := exec.LookPath("pdflatex")
	if err != nil {
		t.Skip("pdflatex unavailable")
	}
	documentEngine, err := api.NewDefaultDocumentEngine("latex", api.DocumentEngineConfig{})
	if err != nil {
		t.Fatal(err)
	}
	generator, err := api.NewProjectionGenerator(api.ProjectionGeneratorOptions{LaTeXEngine: pdflatex, Passes: 1})
	if err != nil {
		t.Fatal(err)
	}
	result, err := documentEngine.Generate(context.Background(), documentSpec(), api.ProjectionOptions{}, generator)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.PDF) < 5 || string(result.PDF[:5]) != "%PDF-" {
		t.Fatalf("not a PDF: %q", result.PDF)
	}
}
