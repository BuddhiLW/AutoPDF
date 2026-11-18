---
name: autopdf
description: >
  Generate PDFs from LaTeX or Typst templates with AutoPDF: install, CLI commands,
  and YAML configuration. Use when a project needs to produce PDFs, invoices,
  reports, certificates, letters or typeset documents from templates; when someone
  mentions AutoPDF, `autopdf build`, LaTeX-to-PDF or Typst-to-PDF generation,
  pdflatex/xelatex/lualatex/typst automation, or converting a PDF to images. Start here, then load autopdf-embed for
  Go integration, autopdf-templates for template syntax, or autopdf-preview for live
  preview and streaming.
---

# AutoPDF

A Go tool and library that fills LaTeX or Typst templates from YAML and compiles
them to PDF. Module path is `github.com/BuddhiLW/AutoPDF/v2` (v2.0.0+).

## Pick the integration first

| You want | Do this | Load |
| --- | --- | --- |
| PDFs from a shell script or CI | `autopdf build` | this skill |
| PDFs from Go code | `pkg/api.Engine` | `autopdf-embed` |
| To author the `.tex`/`.typ` and YAML | custom delimiters | `autopdf-templates` |
| Live preview in an editor/browser | `api.DocumentEngine` + `pkg/api/rest` | `autopdf-preview` |

**The backend's toolchain must be installed.** AutoPDF shells out to
`pdflatex`/`xelatex`/`lualatex` for LaTeX and to `typst` for Typst; it vendors
neither. Check with `which pdflatex` or `which typst` before diagnosing anything
else. In the repo, `make typst` fetches a pinned, sha256-verified typst into
`./bin`.

**The backend is chosen per document**: an explicit `engine` wins, otherwise the
template extension (`.tex`/`.latex` -> LaTeX, `.typ` -> Typst), otherwise
`pdflatex`.

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
autopdf build report.typ config.yaml            # Typst: same command, .typ template

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
engine: "pdflatex"           # pdflatex | xelatex | lualatex | typst
passes: 1                    # LaTeX only: compile passes; raise for refs/ToC
use_latexmk: false           # LaTeX only: delegate passes to latexmk
variables:
  title: "Quarterly Report"
  author: "Ada Lovelace"
conversion:
  enabled: false
  formats: ["png"]
```

Only `template` is really required; the rest have defaults. `passes: 1` is the
default and is **not** enough for documents with `\ref`, `\cite` or a table of
contents; those need 2 or 3, or `use_latexmk: true`. Typst resolves references
in one pass, and `passes > 1` or `use_latexmk: true` with `engine: typst` is
refused with a named error rather than ignored.

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
| `toolchain missing: typst not found` | `typst` not on `PATH`; `make typst` or a package manager |
| `x is absent from the supplied variables` | Variable missing from `variables:`; supply it or `optional .x "why"` |
| `backend "typst" does not support passes=N` | LaTeX-only setting on a Typst document |
| `??` or missing section numbers | `passes: 1`; raise it or set `use_latexmk` |
| Output lands somewhere unexpected | `output: ""` derives the name from the template |
| Conversion produces nothing | `conversion.enabled` is false, or no ImageMagick/pdftoppm |

Compilation failures surface the LaTeX log. Read the **first** error in it; TeX
cascades, so later errors are usually consequences of the first. Typst failures
arrive as parsed `file:line:col: error: message` diagnostics; with `debug`, the
staged `<job>.autopdf.typ` is kept so the line numbers can be read against it.
