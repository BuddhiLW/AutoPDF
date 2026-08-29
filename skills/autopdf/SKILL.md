---
name: autopdf
description: >
  Generate PDFs from LaTeX templates with AutoPDF — install, CLI commands, and YAML
  configuration. Use when a project needs to produce PDFs, invoices, reports,
  certificates, letters or typeset documents from templates; when someone mentions
  AutoPDF, `autopdf build`, LaTeX-to-PDF generation, pdflatex/xelatex/lualatex
  automation, or converting a PDF to images. Start here, then load autopdf-embed for
  Go integration, autopdf-templates for template syntax, or autopdf-preview for live
  preview and streaming.
---

# AutoPDF

A Go tool and library that fills LaTeX templates from YAML and compiles them to
PDF. Module path is `github.com/BuddhiLW/AutoPDF/v2` (v2.0.0+).

## Pick the integration first

| You want | Do this | Load |
| --- | --- | --- |
| PDFs from a shell script or CI | `autopdf build` | this skill |
| PDFs from Go code | `pkg/api.Engine` | `autopdf-embed` |
| To author the `.tex` and YAML | custom delimiters | `autopdf-templates` |
| Live preview in an editor/browser | `api.DocumentEngine` + `pkg/api/rest` | `autopdf-preview` |

**A LaTeX distribution must be installed.** AutoPDF shells out to
`pdflatex`/`xelatex`/`lualatex`; it does not vendor a TeX engine. Check with
`which pdflatex` before diagnosing anything else.

## Install

```bash
go install github.com/BuddhiLW/AutoPDF/v2/cmd/autopdf@latest
```

As a library dependency:

```bash
go get github.com/BuddhiLW/AutoPDF/v2
```

The `/v2` suffix is required. An import without it resolves to the v1 line,
which lacks the streaming transports and uses different logger constructors.

## CLI

```bash
autopdf build template.tex config.yaml          # process template, compile PDF
autopdf build template.tex config.yaml clean    # and remove .aux/.log/.toc after
autopdf build template.tex config.yaml --convert png,jpeg

autopdf convert document.pdf png                # PDF -> images
autopdf clean ./build                           # remove LaTeX aux files
```

Settings that persist across invocations:

```bash
autopdf verbose 3        # log verbosity
autopdf debug switch     # toggle debug output
autopdf clean on         # always clean after a build
autopdf force on         # overwrite existing output
```

`autopdf watch` rebuilds on file change. `autopdf vars` inspects stored
settings. Every command takes `-h`.

Batch compilation exists but is reachable only as a post-build delegation
(`autopdf build <template> <config> multiple ...`), not as a top-level command.
From Go, drive `parallel.ParallelCompilationService` directly instead.

## Configuration

```yaml
template: "document.tex"     # path to the LaTeX template
output: "output.pdf"         # "" derives from the template name
engine: "pdflatex"           # pdflatex | xelatex | lualatex
passes: 1                    # compile passes; raise for refs/ToC
use_latexmk: false           # delegate passes to latexmk when available
variables:
  title: "Quarterly Report"
  author: "Ada Lovelace"
conversion:
  enabled: false
  formats: ["png"]
```

Only `template` is really required; the rest have defaults. `passes: 1` is the
default and is **not** enough for documents with `\ref`, `\cite` or a table of
contents — those need 2 or 3, or `use_latexmk: true`.

## The one thing that surprises everyone

Templates use `delim[[` and `]]`, not Go's `{{` and `}}`:

```latex
\title{delim[[.title]]}
```

LaTeX uses `{}` heavily, so the default Go delimiters collide. See
`autopdf-templates` for the full syntax.

## Troubleshooting

| Symptom | Cause |
| --- | --- |
| `executable file not found` | No TeX distribution installed |
| Variables render literally as `delim[[.x]]` | Variable absent from `variables:` |
| `??` or missing section numbers | `passes: 1`; raise it or set `use_latexmk` |
| Output lands somewhere unexpected | `output: ""` derives the name from the template |
| Conversion produces nothing | `conversion.enabled` is false, or no ImageMagick/pdftoppm |

Compilation failures surface the LaTeX log. Read the **first** error in it; TeX
cascades, so later errors are usually consequences of the first.
