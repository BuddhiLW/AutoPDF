# AutoPDF

Go CLI tool (Bonzai framework) for LaTeX→PDF generation with YAML configuration.

## Commands

```bash
go build -o autopdf ./cmd/autopdf
./autopdf build     # Process template, compile LaTeX
./autopdf clean     # Remove .aux, .log, .toc
./autopdf convert   # PDF to image
```

## Key Features

- YAML-based configuration
- Custom delimiters: `delim[[`, `]]` (avoids LaTeX conflicts)
- Compiles via pdflatex/xelatex
- Optional PDF→image conversion

## Structure

```
cmd/autopdf/main.go       # Bonzai CLI entry
internal/cli/             # build, clean, convert commands
internal/template/        # Go template engine with custom delims
internal/compiler/        # LaTeX compilation via os/exec
```

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
