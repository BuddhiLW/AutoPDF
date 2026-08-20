# Embedding AutoPDF

`pkg/api.Engine` is the supported library entry point. It owns request isolation,
context cancellation, artifact publication, and optional logging.

## Generate a PDF

```go
engine, err := api.NewEngine()
if err != nil {
    return err
}

result, err := engine.Generate(ctx, api.Request{
    TemplatePath: "invoice.tex",
    OutputPath:   "build/invoice.pdf", // optional when only PDF bytes are needed
    Variables: map[string]string{
        "customer.name": "Ada Lovelace",
        "invoice.total": "125.00",
    },
    LaTeXEngine: "pdflatex",
    Passes:      2,
})
if err != nil {
    return err
}

send(result.PDF)
```

Each call compiles in a unique temporary workspace. Cancellation of `ctx` stops
the compiler, temporary files are removed, and requested artifacts are published
atomically. An empty `OutputPath` returns bytes without publishing a file.

## Extend the engine

Implement the small `api.Generator` capability, or adapt a function with
`api.GeneratorFunc`:

```go
custom := api.GeneratorFunc(func(ctx context.Context, req api.Request) (api.Result, error) {
    // Delegate to a remote renderer, add caching, enforce policy, or provide a test fake.
    return api.Result{PDF: renderedBytes, PDFPath: req.OutputPath}, nil
})

engine, err := api.NewEngine(
    api.WithGenerator(custom),
    api.WithCapabilities(api.Capabilities{Engines: []string{"remote"}}),
    api.WithLogger(myLogger),
)
```

The logger implements `api.Logger`; no internal AutoPDF or zap type is required.
AutoPDF does not log template contents, variable names, variable values, or full
configuration objects by default.

## Compatibility migration

| Previous entry point | Supported replacement |
| --- | --- |
| `api.GeneratePDF(config, template)` | `api.NewEngine` + `Engine.Generate` |
| `services.PDFGenerationAPIService` | `api.Engine` |
| `generation.PDFGenerationRequest` | `api.Request` |
| `generation.PDFGenerationResult` | `api.Result` |
| `generation.ConversionOptions` | `api.ConversionOptions` |

Compatibility functions remain available for existing callers and route through
the canonical engine. Compatibility subpackages may be deprecated in a future
major release; new programs should import only `github.com/BuddhiLW/AutoPDF/pkg/api`.
