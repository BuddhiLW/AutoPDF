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

## Bonzai Pattern

Uses rwxrob/bonzai for composable command trees with help, vars, and completion.

## Full Documentation

Run `/catchup` to load complete details from hive-mcp memory (tag: autopdf).
