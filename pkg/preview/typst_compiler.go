package preview

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	typstadapter "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/typst"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

// ErrTypstPagesMissing reports a PNG run that produced no page images.
var ErrTypstPagesMissing = errors.New("preview: typst produced no page images")

// TypstCompilerOptions configures a preview compiler for Typst workspaces.
type TypstCompilerOptions struct {
	// Binary is the typst executable; default "typst".
	Binary string
	// PPI is the pixel density of the page images; default 144.
	PPI int
	// FontPaths are extra font directories passed as --font-path.
	FontPaths []string
	// Inputs are passed as --input key=value (sys.inputs in the document).
	Inputs map[string]string
	// Runner executes typst; default ExecProcessRunner.
	Runner ProcessRunner
}

// TypstCompiler compiles a workspace's main .typ file to a PDF and rasterizes
// its pages with typst itself, so no external rasterizer is involved. The
// workspace is the typst --root: the document can read every file in it and
// nothing outside it.
type TypstCompiler struct {
	binary    string
	ppi       int
	fontPaths []string
	inputs    map[string]string
	runner    ProcessRunner
}

var _ Compiler = (*TypstCompiler)(nil)

// NewTypstCompiler creates a Typst preview compiler.
func NewTypstCompiler(options TypstCompilerOptions) (*TypstCompiler, error) {
	if options.Runner == nil {
		options.Runner = ExecProcessRunner{}
	}
	if options.Binary == "" {
		options.Binary = typstadapter.DefaultBinary
	}
	if options.PPI <= 0 {
		options.PPI = 144
	}
	inputs := make(map[string]string, len(options.Inputs))
	for key, value := range options.Inputs {
		inputs[key] = value
	}
	// Refuse unrepresentable inputs now rather than on every revision.
	if _, err := typstadapter.CompileArgs("", "", "", ports.TypstOptions{Inputs: inputs}); err != nil {
		return nil, err
	}
	return &TypstCompiler{
		binary: options.Binary, ppi: options.PPI,
		fontPaths: append([]string(nil), options.FontPaths...), inputs: inputs,
		runner: options.Runner,
	}, nil
}

// Compile runs typst twice in the workspace: once to PDF, once to one PNG per
// page. Every page carries its bytes and a content hash; the session reports
// only the pages whose hash changed.
func (compiler *TypstCompiler) Compile(ctx context.Context, request CompileRequest) (CompileOutput, error) {
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
	output, err := os.MkdirTemp(request.Workspace, ".autopdf-typst-")
	if err != nil {
		return CompileOutput{}, fmt.Errorf("preview: create typst output directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(output) }()

	pdfPath := filepath.Join(output, "preview.pdf")
	compileStarted := time.Now()
	processOutput, processErr := compiler.run(ctx, request.Workspace, mainPath, pdfPath)
	compileElapsed := time.Since(compileStarted)
	diagnostics := typstDiagnostics(joinStreams(processOutput), request.Workspace, request.SourceMap)
	if err := compileFailure(compiler.binary, processOutput, processErr); err != nil {
		return CompileOutput{Diagnostics: diagnostics, Timings: StageTimings{TeX: compileElapsed}}, err
	}
	pdf, err := os.ReadFile(pdfPath)
	if err != nil || len(pdf) < 5 || string(pdf[:5]) != "%PDF-" {
		if err == nil {
			err = ErrPreviewPDFMissing
		}
		return CompileOutput{Diagnostics: diagnostics, Timings: StageTimings{TeX: compileElapsed}}, fmt.Errorf("%w: %v", ErrPreviewPDFMissing, err)
	}

	rasterStarted := time.Now()
	pagePattern := filepath.Join(output, "page-{p}.png")
	processOutput, processErr = compiler.run(ctx, request.Workspace, mainPath, pagePattern,
		"--format", "png", "--ppi", strconv.Itoa(compiler.ppi))
	rasterElapsed := time.Since(rasterStarted)
	timings := StageTimings{TeX: compileElapsed, Raster: rasterElapsed}
	if err := compileFailure(compiler.binary, processOutput, processErr); err != nil {
		return CompileOutput{PDF: pdf, Diagnostics: diagnostics, Timings: timings}, fmt.Errorf("preview: typst raster: %w", err)
	}
	fingerprintStarted := time.Now()
	pages, err := readTypstPages(output)
	timings.Fingerprint = time.Since(fingerprintStarted)
	if err != nil {
		return CompileOutput{PDF: pdf, Diagnostics: diagnostics, Timings: timings}, err
	}
	return CompileOutput{PDF: pdf, Pages: pages, Diagnostics: diagnostics, Timings: timings}, nil
}

func (compiler *TypstCompiler) run(ctx context.Context, workspace, source, target string, extra ...string) (ProcessOutput, error) {
	args, err := typstadapter.CompileArgs(workspace, source, target, ports.TypstOptions{
		FontPaths: compiler.fontPaths, Inputs: compiler.inputs,
	})
	if err != nil {
		return ProcessOutput{}, err
	}
	return compiler.runner.Run(ctx, Process{
		Executable: compiler.binary, Args: append(args, extra...), Dir: workspace,
	})
}

// compileFailure turns a failed or non-zero typst run into an error carrying
// typst's diagnostics; cancellation is returned as the context error.
func compileFailure(binary string, output ProcessOutput, err error) error {
	if err == nil && output.ExitCode == 0 {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("preview: typst compile: %w", err)
	}
	text := joinStreams(output)
	return fmt.Errorf("preview: typst compile: %w", &typstadapter.CompileError{
		Command: binary + " compile", ExitCode: output.ExitCode,
		Diagnostics: typstadapter.ParseDiagnostics(text), Output: text, Err: err,
	})
}

func joinStreams(output ProcessOutput) string {
	switch {
	case output.Stderr == "":
		return output.Stdout
	case output.Stdout == "":
		return output.Stderr
	default:
		return strings.TrimRight(output.Stderr, "\n") + "\n" + output.Stdout
	}
}

func typstDiagnostics(output, workspace string, sourceMap map[string]SourceLocation) []Diagnostic {
	parsed := typstadapter.ParseDiagnostics(output)
	if len(parsed) == 0 {
		return nil
	}
	componentByPath := componentsByPath(sourceMap)
	result := make([]Diagnostic, 0, len(parsed))
	for _, entry := range parsed {
		diagnostic := Diagnostic{
			Line: entry.Line, Column: entry.Column, Severity: string(entry.Severity),
			Message: entry.Message, Hints: append([]string(nil), entry.Hints...),
		}
		if entry.Path != "" {
			path := workspaceRelative(workspace, entry.Path)
			diagnostic.Path = filepath.ToSlash(path)
			diagnostic.ComponentID = componentByPath[path]
		}
		result = append(result, diagnostic)
	}
	return result
}

var typstPagePattern = regexp.MustCompile(`^page-(\d+)\.png$`)

func readTypstPages(directory string) ([]Page, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("preview: list typst pages: %w", err)
	}
	pages := make([]Page, 0, len(entries))
	for _, entry := range entries {
		match := typstPagePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		number, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("preview: invalid typst page %q", entry.Name())
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("preview: read typst page %d: %w", number, err)
		}
		pages = append(pages, Page{Number: number, MediaType: "image/png", Data: data, Hash: pageHash("image/png", data)})
	}
	if len(pages) == 0 {
		return nil, ErrTypstPagesMissing
	}
	sort.Slice(pages, func(i, j int) bool { return pages[i].Number < pages[j].Number })
	return pages, nil
}
