package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

var (
	ErrProjectionEngineRequired  = errors.New("api: projection engine is required")
	ErrProjectionMainMissing     = errors.New("api: projection main file is missing")
	ErrProjectionResolverMissing = errors.New("api: projection asset resolver is required")
	ErrProjectionPathUnsafe      = errors.New("api: unsafe projection path")
	ErrProjectionPathDuplicate   = errors.New("api: duplicate projection path")
	ErrProjectionDigestMismatch  = errors.New("api: projection asset digest mismatch")
)

// ProjectionAssetResolver materializes one inert projection asset.
type ProjectionAssetResolver interface {
	Resolve(context.Context, latex.AssetBinding) ([]byte, error)
}

// ProjectionGeneratorOptions configures production projection generation.
// Each call still receives a private ephemeral workspace; preview workspaces
// are deliberately managed by pkg/preview instead.
type ProjectionGeneratorOptions struct {
	Engine        *Engine
	Resolver      ProjectionAssetResolver
	WorkspaceRoot string
	KeepWorkspace bool
	OutputPath    string
	LaTeXEngine   string
	Passes        int
	UseLatexmk    bool
	Debug         bool
	Conversion    ConversionOptions
}

// ProjectionDiagnostic maps a compiler message back to generated source and,
// when possible, its component identity.
type ProjectionDiagnostic struct {
	ComponentID string
	Path        string
	Line        int
	Message     string
}

// ProjectionGenerationError retains the compiler error plus mapped messages.
type ProjectionGenerationError struct {
	Err         error
	Diagnostics []ProjectionDiagnostic
}

func (failure *ProjectionGenerationError) Error() string { return failure.Err.Error() }
func (failure *ProjectionGenerationError) Unwrap() error { return failure.Err }

// LocalProjectionGenerator is the supported production adapter from a pure
// LaTeX projection to the canonical Engine.Generate boundary.
type LocalProjectionGenerator struct {
	options ProjectionGeneratorOptions
}

// NewProjectionGenerator constructs a production adapter. A nil Engine uses
// AutoPDF's canonical built-in generator.
func NewProjectionGenerator(options ProjectionGeneratorOptions) (*LocalProjectionGenerator, error) {
	if options.Engine == nil {
		engine, err := NewEngine()
		if err != nil {
			return nil, fmt.Errorf("api: create projection engine: %w", err)
		}
		options.Engine = engine
	}
	if options.Engine.generator == nil {
		return nil, ErrProjectionEngineRequired
	}
	options.Conversion.Formats = append([]string(nil), options.Conversion.Formats...)
	return &LocalProjectionGenerator{options: options}, nil
}

// GenerateProjection materializes one immutable projection, invokes the
// canonical generation pipeline, then removes its private workspace.
func (generator *LocalProjectionGenerator) GenerateProjection(ctx context.Context, projection latex.Projection) (Result, error) {
	if generator == nil || generator.options.Engine == nil {
		return Result{}, ErrProjectionEngineRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	workspace, err := os.MkdirTemp(generator.options.WorkspaceRoot, "autopdf-projection-")
	if err != nil {
		return Result{}, fmt.Errorf("api: create projection workspace: %w", err)
	}
	if !generator.options.KeepWorkspace {
		defer func() { _ = os.RemoveAll(workspace) }()
	}
	mainPath, err := materializeProjection(ctx, workspace, projection, generator.options.Resolver)
	if err != nil {
		return Result{}, err
	}
	result, err := generator.options.Engine.Generate(ctx, Request{
		TemplatePath: mainPath,
		OutputPath:   generator.options.OutputPath,
		WorkingDir:   workspace,
		LaTeXEngine:  generator.options.LaTeXEngine,
		Passes:       generator.options.Passes,
		UseLatexmk:   generator.options.UseLatexmk,
		Debug:        generator.options.Debug,
		Conversion: ConversionOptions{
			Enabled: generator.options.Conversion.Enabled,
			Formats: append([]string(nil), generator.options.Conversion.Formats...),
		},
	})
	if err != nil {
		return Result{}, &ProjectionGenerationError{Err: err, Diagnostics: mapProjectionDiagnostics(workspace, projection, err)}
	}
	return result, nil
}

func materializeProjection(ctx context.Context, workspace string, projection latex.Projection, resolver ProjectionAssetResolver) (string, error) {
	written := make(map[string]struct{}, len(projection.Files)+len(projection.Assets))
	mainAbsolute, err := projectionPath(workspace, projection.Main)
	if err != nil {
		return "", err
	}
	mainFound := false
	for _, file := range projection.Files {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		absolute, err := projectionPath(workspace, file.Path)
		if err != nil {
			return "", err
		}
		if _, duplicate := written[absolute]; duplicate {
			return "", fmt.Errorf("%w: %q", ErrProjectionPathDuplicate, file.Path)
		}
		written[absolute] = struct{}{}
		if absolute == mainAbsolute {
			mainFound = true
		}
		if err := writeProjectionFile(absolute, file.Content); err != nil {
			return "", err
		}
	}
	if !mainFound {
		return "", fmt.Errorf("%w: %q", ErrProjectionMainMissing, projection.Main)
	}
	for _, asset := range projection.Assets {
		if resolver == nil {
			return "", ErrProjectionResolverMissing
		}
		absolute, err := projectionPath(workspace, asset.Path)
		if err != nil {
			return "", err
		}
		if _, duplicate := written[absolute]; duplicate {
			return "", fmt.Errorf("%w: %q", ErrProjectionPathDuplicate, asset.Path)
		}
		written[absolute] = struct{}{}
		data, err := resolver.Resolve(ctx, asset)
		if err != nil {
			return "", fmt.Errorf("api: resolve projection asset %s: %w", asset.AssetID, err)
		}
		if !matchesSHA256(asset.Digest, data) {
			return "", fmt.Errorf("%w: %s", ErrProjectionDigestMismatch, asset.AssetID)
		}
		if err := writeProjectionFile(absolute, data); err != nil {
			return "", err
		}
	}
	return mainAbsolute, nil
}

func projectionPath(workspace, relative string) (string, error) {
	if relative == "" || filepath.IsAbs(relative) || strings.ContainsRune(relative, '\x00') {
		return "", fmt.Errorf("%w: %q", ErrProjectionPathUnsafe, relative)
	}
	clean := filepath.Clean(filepath.FromSlash(relative))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %q", ErrProjectionPathUnsafe, relative)
	}
	return filepath.Join(workspace, clean), nil
}

func writeProjectionFile(absolute string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
		return fmt.Errorf("api: create projection directory: %w", err)
	}
	temporary := absolute + ".autopdf-new"
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		return fmt.Errorf("api: write projection file: %w", err)
	}
	if err := os.Rename(temporary, absolute); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("api: publish projection file: %w", err)
	}
	return nil
}

func matchesSHA256(declared string, data []byte) bool {
	value := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(declared)), "sha256:")
	if len(value) != sha256.Size*2 {
		return true // opaque version token; still participates in projection identity
	}
	if _, err := hex.DecodeString(value); err != nil {
		return true
	}
	digest := sha256.Sum256(data)
	return value == hex.EncodeToString(digest[:])
}

var projectionDiagnosticPattern = regexp.MustCompile(`(?m)([^:\n]+\.tex):(\d+):\s*([^\n]+)`)

func mapProjectionDiagnostics(workspace string, projection latex.Projection, compileErr error) []ProjectionDiagnostic {
	componentByPath := make(map[string]string, len(projection.SourceMap))
	for componentID, location := range projection.SourceMap {
		componentByPath[filepath.Clean(filepath.FromSlash(location.Path))] = componentID
	}
	matches := projectionDiagnosticPattern.FindAllStringSubmatch(compileErr.Error(), -1)
	diagnostics := make([]ProjectionDiagnostic, 0, len(matches))
	for _, match := range matches {
		path := filepath.Clean(match[1])
		if filepath.IsAbs(path) {
			if relative, err := filepath.Rel(workspace, path); err == nil {
				path = relative
			}
		}
		line, _ := strconv.Atoi(match[2])
		diagnostics = append(diagnostics, ProjectionDiagnostic{ComponentID: componentByPath[path], Path: filepath.ToSlash(path), Line: line, Message: strings.TrimSpace(match[3])})
	}
	if len(diagnostics) == 0 {
		diagnostics = append(diagnostics, ProjectionDiagnostic{Path: projection.Main, Message: compileErr.Error()})
	}
	return diagnostics
}
