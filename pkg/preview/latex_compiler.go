package preview

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/render/latex"
)

var (
	ErrProcessRunnerRequired = errors.New("preview: process runner is required")
	ErrRasterizerRequired    = errors.New("preview: page rasterizer is required")
	ErrPreviewPDFMissing     = errors.New("preview: compiler produced no PDF")
)

// Process is one cancellation-aware external command.
type Process struct {
	Executable string
	Args       []string
	Dir        string
}

// ProcessOutput retains both output streams for diagnostic parsing.
type ProcessOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// ProcessRunner is injectable so process lifecycle can be tested without TeX.
type ProcessRunner interface {
	Run(context.Context, Process) (ProcessOutput, error)
}

// ExecProcessRunner uses exec.CommandContext; cancellation terminates the child.
type ExecProcessRunner struct{}

func (ExecProcessRunner) Run(ctx context.Context, process Process) (ProcessOutput, error) {
	command := exec.CommandContext(ctx, process.Executable, process.Args...)
	command.Dir = process.Dir
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	output := ProcessOutput{Stdout: stdout.String(), Stderr: stderr.String()}
	if command.ProcessState != nil {
		output.ExitCode = command.ProcessState.ExitCode()
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return output, ctxErr
	}
	return output, err
}

// RasterOutput reports complete page identities; unchanged pages intentionally
// carry no Data, avoiding needless rasterization and transport.
type RasterOutput struct {
	Pages       []Page
	Fingerprint time.Duration
	Raster      time.Duration
}

// PageRasterizer fingerprints pages and rasterizes only changed identities.
type PageRasterizer interface {
	Rasterize(context.Context, string, string, map[int]string) (RasterOutput, error)
}

// LaTeXCompilerOptions configures a warm compiler bound to session workspaces.
type LaTeXCompilerOptions struct {
	Engine     string
	JobName    string
	DPI        int
	Runner     ProcessRunner
	Rasterizer PageRasterizer
}

// LaTeXCompiler retains per-workspace page fingerprints while TeX retains aux
// state in the workspace itself.
type LaTeXCompiler struct {
	engine     string
	jobName    string
	runner     ProcessRunner
	rasterizer PageRasterizer

	mu       sync.Mutex
	previous map[string]map[int]string
}

// NewLaTeXCompiler creates a concrete preview compiler. Defaults: xelatex,
// job name autopdf-preview, 144 DPI, and Poppler rasterization.
func NewLaTeXCompiler(options LaTeXCompilerOptions) (*LaTeXCompiler, error) {
	if options.Runner == nil {
		options.Runner = ExecProcessRunner{}
	}
	if options.Engine == "" {
		options.Engine = "xelatex"
	}
	if options.JobName == "" {
		options.JobName = "autopdf-preview"
	}
	if options.DPI <= 0 {
		options.DPI = 144
	}
	if options.Rasterizer == nil {
		options.Rasterizer = NewPopplerRasterizer(options.Runner, options.DPI)
	}
	if options.Runner == nil {
		return nil, ErrProcessRunnerRequired
	}
	if options.Rasterizer == nil {
		return nil, ErrRasterizerRequired
	}
	return &LaTeXCompiler{
		engine: options.Engine, jobName: options.JobName, runner: options.Runner,
		rasterizer: options.Rasterizer, previous: make(map[string]map[int]string),
	}, nil
}

// Compile reuses workspace aux files, compiles the current main source, then
// returns full page identities with bytes only for changed pages.
func (compiler *LaTeXCompiler) Compile(ctx context.Context, request CompileRequest) (CompileOutput, error) {
	if compiler == nil || compiler.runner == nil {
		return CompileOutput{}, ErrProcessRunnerRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}
	mainPath, err := safeWorkspacePath(request.Workspace, request.Main)
	if err != nil {
		return CompileOutput{}, err
	}
	pdfPath := filepath.Join(request.Workspace, compiler.jobName+".pdf")
	_ = os.Remove(pdfPath) // never accept a stale PDF after a failed compile
	texStarted := time.Now()
	processOutput, processErr := compiler.runner.Run(ctx, Process{
		Executable: compiler.engine,
		Args: []string{
			"-interaction=nonstopmode", "-halt-on-error", "-file-line-error",
			"-jobname=" + compiler.jobName,
			"-output-directory=" + request.Workspace,
			mainPath,
		},
		Dir: request.Workspace,
	})
	texElapsed := time.Since(texStarted)
	diagnostics := parseLaTeXDiagnostics(processOutput.Stdout+"\n"+processOutput.Stderr, request.Workspace, request.SourceMap)
	if processErr != nil {
		return CompileOutput{Diagnostics: diagnostics, Timings: StageTimings{TeX: texElapsed}}, fmt.Errorf("preview: TeX compile: %w", processErr)
	}
	pdf, err := os.ReadFile(pdfPath)
	if err != nil || len(pdf) < 5 || string(pdf[:5]) != "%PDF-" {
		if err == nil {
			err = ErrPreviewPDFMissing
		}
		return CompileOutput{Diagnostics: diagnostics, Timings: StageTimings{TeX: texElapsed}}, fmt.Errorf("%w: %v", ErrPreviewPDFMissing, err)
	}
	previous := compiler.previousFor(request.Workspace)
	raster, err := compiler.rasterizer.Rasterize(ctx, pdfPath, request.Workspace, previous)
	if err != nil {
		return CompileOutput{PDF: pdf, Diagnostics: diagnostics, Timings: StageTimings{TeX: texElapsed, Fingerprint: raster.Fingerprint, Raster: raster.Raster}}, err
	}
	compiler.remember(request.Workspace, raster.Pages)
	return CompileOutput{
		PDF: pdf, Pages: raster.Pages, Diagnostics: diagnostics,
		Timings: StageTimings{TeX: texElapsed, Fingerprint: raster.Fingerprint, Raster: raster.Raster},
	}, nil
}

func (compiler *LaTeXCompiler) previousFor(workspace string) map[int]string {
	compiler.mu.Lock()
	defer compiler.mu.Unlock()
	result := make(map[int]string, len(compiler.previous[workspace]))
	for page, hash := range compiler.previous[workspace] {
		result[page] = hash
	}
	return result
}

func (compiler *LaTeXCompiler) remember(workspace string, pages []Page) {
	current := make(map[int]string, len(pages))
	for _, page := range pages {
		current[page.Number] = page.Hash
	}
	compiler.mu.Lock()
	compiler.previous[workspace] = current
	compiler.mu.Unlock()
}

// PopplerRasterizer uses a cheap low-DPI visual pass for stable page identities
// and full-DPI pdftoppm only for pages whose identity changed.
type PopplerRasterizer struct {
	runner ProcessRunner
	dpi    int
}

func NewPopplerRasterizer(runner ProcessRunner, dpi int) *PopplerRasterizer {
	if runner == nil {
		runner = ExecProcessRunner{}
	}
	if dpi <= 0 {
		dpi = 144
	}
	return &PopplerRasterizer{runner: runner, dpi: dpi}
}

func (rasterizer *PopplerRasterizer) Rasterize(ctx context.Context, pdfPath, workspace string, previous map[int]string) (RasterOutput, error) {
	temporary, err := os.MkdirTemp(workspace, ".autopdf-pages-")
	if err != nil {
		return RasterOutput{}, fmt.Errorf("preview: create page workspace: %w", err)
	}
	defer func() { _ = os.RemoveAll(temporary) }()
	fingerprintStarted := time.Now()
	fingerprintPrefix := filepath.Join(temporary, "fingerprint")
	if _, err := rasterizer.runner.Run(ctx, Process{
		Executable: "pdftoppm",
		Args:       []string{"-png", "-r", "12", pdfPath, fingerprintPrefix},
		Dir:        workspace,
	}); err != nil {
		return RasterOutput{Fingerprint: time.Since(fingerprintStarted)}, fmt.Errorf("preview: fingerprint pages: %w", err)
	}
	paths, err := filepath.Glob(filepath.Join(temporary, "fingerprint-*.png"))
	if err != nil {
		return RasterOutput{Fingerprint: time.Since(fingerprintStarted)}, err
	}
	type identity struct {
		number int
		hash   string
	}
	identities := make([]identity, 0, len(paths))
	for _, pagePath := range paths {
		data, readErr := os.ReadFile(pagePath)
		if readErr != nil {
			return RasterOutput{Fingerprint: time.Since(fingerprintStarted)}, readErr
		}
		number, parseErr := pageNumber(pagePath)
		if parseErr != nil {
			return RasterOutput{Fingerprint: time.Since(fingerprintStarted)}, parseErr
		}
		digest := sha256.Sum256(data)
		identities = append(identities, identity{number: number, hash: hex.EncodeToString(digest[:])})
	}
	sort.Slice(identities, func(i, j int) bool { return identities[i].number < identities[j].number })
	fingerprintElapsed := time.Since(fingerprintStarted)
	rasterStarted := time.Now()
	pages := make([]Page, 0, len(identities))
	for _, identity := range identities {
		page := Page{Number: identity.number, MediaType: "image/png", Hash: identity.hash}
		if previous[identity.number] != identity.hash {
			prefix := filepath.Join(temporary, "raster-"+strconv.Itoa(identity.number))
			_, err := rasterizer.runner.Run(ctx, Process{
				Executable: "pdftoppm",
				Args:       []string{"-f", strconv.Itoa(identity.number), "-l", strconv.Itoa(identity.number), "-singlefile", "-png", "-r", strconv.Itoa(rasterizer.dpi), pdfPath, prefix},
				Dir:        workspace,
			})
			if err != nil {
				return RasterOutput{Pages: pages, Fingerprint: fingerprintElapsed, Raster: time.Since(rasterStarted)}, fmt.Errorf("preview: raster page %d: %w", identity.number, err)
			}
			page.Data, err = os.ReadFile(prefix + ".png")
			if err != nil {
				return RasterOutput{Pages: pages, Fingerprint: fingerprintElapsed, Raster: time.Since(rasterStarted)}, fmt.Errorf("preview: read raster page %d: %w", identity.number, err)
			}
		}
		pages = append(pages, page)
	}
	return RasterOutput{Pages: pages, Fingerprint: fingerprintElapsed, Raster: time.Since(rasterStarted)}, nil
}

var pagePattern = regexp.MustCompile(`fingerprint-(\d+)\.png$`)

func pageNumber(path string) (int, error) {
	match := pagePattern.FindStringSubmatch(filepath.Base(path))
	if len(match) != 2 {
		return 0, fmt.Errorf("preview: invalid separated page path %q", path)
	}
	return strconv.Atoi(match[1])
}

var latexDiagnosticPattern = regexp.MustCompile(`(?m)([^:\n]+\.tex):(\d+):\s*([^\n]+)`)

func parseLaTeXDiagnostics(output, workspace string, sourceMap map[string]latex.SourceLocation) []Diagnostic {
	componentByPath := make(map[string]string, len(sourceMap))
	for componentID, location := range sourceMap {
		componentByPath[filepath.Clean(filepath.FromSlash(location.Path))] = componentID
	}
	matches := latexDiagnosticPattern.FindAllStringSubmatch(output, -1)
	result := make([]Diagnostic, 0, len(matches))
	for _, match := range matches {
		path := filepath.Clean(strings.TrimSpace(match[1]))
		if filepath.IsAbs(path) {
			if relative, err := filepath.Rel(workspace, path); err == nil {
				path = relative
			}
		}
		line, _ := strconv.Atoi(match[2])
		result = append(result, Diagnostic{
			ComponentID: componentByPath[path], Path: filepath.ToSlash(path), Line: line,
			Severity: "error", Message: strings.TrimSpace(match[3]),
		})
	}
	return result
}
