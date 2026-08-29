package preview

import (
	"context"
	"errors"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/render/latex"
)

var (
	ErrClosed           = errors.New("preview session is closed")
	ErrSuperseded       = errors.New("preview revision was superseded")
	ErrCompilerRequired = errors.New("preview compiler is required")
	ErrResolverRequired = errors.New("preview asset resolver is required")
	ErrUnsafePath       = errors.New("unsafe preview workspace path")
)

// Diagnostic is a compiler message addressable to generated source.
type Diagnostic struct {
	ComponentID string `json:"componentId,omitempty"`
	Path        string `json:"path,omitempty"`
	Line        int    `json:"line,omitempty"`
	Severity    string `json:"severity,omitempty"`
	Message     string `json:"message"`
}

// Page is browser-ready output for one page. Hash may be omitted by compiler;
// the session fills it from MIME and Data.
type Page struct {
	Number    int    `json:"number"`
	MediaType string `json:"mediaType"`
	Data      []byte `json:"data"`
	Hash      string `json:"hash"`
}

// CompileRequest identifies one revision within a persistent workspace.
type CompileRequest struct {
	Revision   uint64
	Workspace  string
	Main       string
	DirtyFiles []string
	SourceMap  map[string]latex.SourceLocation
}

// StageTimings separates deterministic orchestration from environment-sensitive
// TeX and raster work.
type StageTimings struct {
	Materialize time.Duration `json:"materialize"`
	TeX         time.Duration `json:"tex"`
	Fingerprint time.Duration `json:"fingerprint"`
	Raster      time.Duration `json:"raster"`
	Total       time.Duration `json:"total"`
}

// CompileOutput is returned by a target-specific compiler/rasterizer adapter.
type CompileOutput struct {
	PDF         []byte
	Pages       []Page
	Diagnostics []Diagnostic
	Timings     StageTimings
}

// Compiler is the effect port around LaTeX execution and page rasterization.
type Compiler interface {
	Compile(context.Context, CompileRequest) (CompileOutput, error)
}

// AssetResolver collects one inert asset at the workspace effect boundary.
type AssetResolver interface {
	Resolve(context.Context, latex.AssetBinding) ([]byte, error)
}

// Result contains only pages changed from the last accepted revision.
type Result struct {
	Revision     uint64       `json:"revision"`
	Projection   string       `json:"projectionHash,omitempty"`
	PDF          []byte       `json:"pdf,omitempty"`
	ChangedPages []Page       `json:"changedPages,omitempty"`
	RemovedPages []int        `json:"removedPages,omitempty"`
	Diagnostics  []Diagnostic `json:"diagnostics,omitempty"`
	DirtyFiles   []string     `json:"dirtyFiles,omitempty"`
	Timings      StageTimings `json:"timings"`
	Err          error        `json:"-"`
}
