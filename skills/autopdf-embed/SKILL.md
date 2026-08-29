---
name: autopdf-embed
description: >
  Embed AutoPDF as a Go library to generate PDFs from application code via
  pkg/api.Engine. Use when wiring PDF generation into a Go service, HTTP handler,
  worker or CLI; when someone asks how to call AutoPDF from Go, use api.Engine /
  api.Request / api.Result, inject a custom api.Generator, route AutoPDF logs into
  an application logger, test code that generates PDFs without a LaTeX toolchain,
  or migrate from the v1 module path.
---

# Embedding AutoPDF in Go

`pkg/api` is the only package an embedding application should import. Everything
under `internal/` is unimportable from outside the module by design.

```go
import "github.com/BuddhiLW/AutoPDF/v2/pkg/api"
```

## Generate a PDF

```go
engine, err := api.NewEngine()
if err != nil {
    return err
}

result, err := engine.Generate(ctx, api.Request{
    TemplatePath: "invoice.tex",
    OutputPath:   "out/invoice.pdf",
    Variables: map[string]string{
        "customer": "Ada Lovelace",
        "total":    "1240.00",
    },
    LaTeXEngine: "pdflatex",
    Passes:      2,
})
if err != nil {
    return err
}

fmt.Println(result.PDFPath, len(result.PDF), result.Metadata.PageCount)
```

`Engine` is safe for concurrent use and every call takes a `context.Context`;
cancelling it aborts the compilation.

### Request

| Field | Notes |
| --- | --- |
| `TemplatePath` | Required. `api.ErrTemplateRequired` when empty |
| `OutputPath` | Empty derives from the template name |
| `Variables` | `map[string]string` — flat only; see below |
| `LaTeXEngine` | `pdflatex` (default), `xelatex`, `lualatex` |
| `WorkingDir` | Resolves relative paths and template includes |
| `Passes` | Raise to 2–3 for `\ref`, `\cite`, ToC |
| `UseLatexmk` | Let latexmk decide the pass count |
| `Debug` | Keep intermediates and emit verbose logs |
| `Conversion` | `api.ConversionOptions{Enabled, Formats}` |

`Variables` is `map[string]string`, so nested data must be shaped in the YAML
config rather than passed here. For nested structures and loops, drive
generation from a config file — see `autopdf-templates`.

### Result

```go
type Result struct {
    PDFPath    string
    PDF        []byte      // the bytes, so you can stream without touching disk
    ImagePaths []string    // populated when Conversion.Enabled
    Metadata   PDFMetadata // FileSize, PageCount, GeneratedAt, Engine, Template
}
```

Serving a generated PDF over HTTP needs no file handling:

```go
w.Header().Set("Content-Type", "application/pdf")
w.Write(result.PDF)
```

## Testing without LaTeX

`api.Generator` is the extension point, so tests substitute a fake and never
shell out. This is the single most useful thing to know when embedding.

```go
engine, _ := api.NewEngine(api.WithGenerator(
    api.GeneratorFunc(func(_ context.Context, req api.Request) (api.Result, error) {
        return api.Result{
            PDFPath: req.OutputPath,
            PDF:     []byte("%PDF-1.4 fake"),
        }, nil
    }),
))
```

The same seam takes a caching layer, a remote renderer, or a rate limiter —
anything implementing `Generate(context.Context, Request) (Result, error)`.

Guard genuine end-to-end tests so they skip on machines without a toolchain:

```go
if _, err := exec.LookPath("pdflatex"); err != nil {
    t.Skip("pdflatex unavailable")
}
```

## Logging

By default the engine is silent. Route its lifecycle messages into an existing
logger by implementing `api.Logger`:

```go
type Logger interface {
    Debug(ctx context.Context, msg string, fields ...api.LogField)
    Info(ctx context.Context, msg string, fields ...api.LogField)
    Warn(ctx context.Context, msg string, fields ...api.LogField)
    Error(ctx context.Context, msg string, fields ...api.LogField)
}

engine, _ := api.NewEngine(api.WithLogger(myLogger))
```

Fields are built with `api.NewLogField(key, value)`. Passing a nil logger is
accepted and yields the no-op, so a partially wired application degrades to
silence rather than panicking.

## Capabilities

```go
caps := engine.Capabilities() // Engines, ConversionFormats
```

Use it to populate a UI dropdown or validate a request before compiling. An
embedded implementation advertises its own with `api.WithCapabilities`.

## Migrating from v1

The module path gained a `/v2` suffix at v2.0.0. Beyond rewriting imports:

- `rest.SimpleStructConverterAPI` and its nine `Simple*` types are gone. Use
  `rest.StructConverterAPI`.
- `NewPDFGenerationApplicationService`, `NewPDFOrchestrationService` and
  `NewWatchModeManager` now take the `ports.Logger` port instead of a concrete
  `*logger.LoggerAdapter`. Bridge an existing one with
  `infrastructure/adapters.NewLoggerPortAdapter`.
- `adapters.NoOpLogger` moved to `ports.NoOpLogger`.

`api.Engine`, `api.Request` and `api.Result` are unchanged, so an application
that only used `pkg/api` needs nothing but the import rewrite.
