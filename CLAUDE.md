# AutoPDF

Go CLI tool (Bonzai framework) for template→PDF generation (LaTeX or Typst) with YAML configuration.

## Commands

```bash
go build -o autopdf ./cmd/autopdf
./autopdf build     # Process template, compile with the selected backend
./autopdf multiple  # Compile several templates in parallel
./autopdf clean     # Remove the backends' auxiliary files
./autopdf convert   # PDF to image
make typst          # Pinned typst binary into ./bin (sha256-verified)
```

## Key Features

- YAML-based configuration
- Custom delimiters: `delim[[`, `]]` (avoids LaTeX conflicts)
- Backends: latex (pdflatex/xelatex/lualatex) and typst; engine name wins, else template extension, else pdflatex
- Optional PDF→image conversion

## Structure

```
cmd/autopdf/main.go                              # Bonzai CLI entry
internal/autopdf/commands/                       # build, clean, convert, deck commands
internal/autopdf/application/ports/              # DocumentCompiler + LaTeXOptions/TypstOptions; compilertest contract suite
internal/autopdf/application/backend/            # Backend registry: engines, extensions, aux files, compiler factory
internal/autopdf/application/adapters/{latex,typst}/  # DocumentCompiler adapters
pkg/api/                                         # Embedding API (Request.Engine)
pkg/preview/                                     # Preview compilers (LaTeX, Typst)
pkg/render/{latex,typst,backends}/               # Component projections; backends.Default() registers both
pkg/template/{strict,escape}/                    # Strict rendering; escapeTypst
```

Add a backend by registering a `backend.Backend`; never hardcode engine or extension lists. Typst e2e: `test/model_typst` (skips without typst on PATH).

## Template Processing

```go
template.New(file).Funcs(funcMap).Delims("delim[[", "]]")
```

Rendering is strict by default (`pkg/template/strict`): a key the template reads
but the variables do not supply, or supply as nil or as a blank string, refuses
the render and names every offending key at once. A blank that is intended is
declared at the site:

```latex
delim[[ optional .field "why the blank is acceptable" ]]
```

## Bonzai Pattern

Uses rwxrob/bonzai for composable command trees with help, vars, and completion.

## Full Documentation

Run `/catchup` to load complete details from hive-mcp memory (tag: autopdf).
