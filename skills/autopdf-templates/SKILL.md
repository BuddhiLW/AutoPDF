---
name: autopdf-templates
description: >
  Write AutoPDF LaTeX templates and YAML configuration — the custom delim[[ ]]
  delimiters, variable substitution, nested objects, and range loops. Use when
  authoring or debugging a .tex template for AutoPDF, when variables render
  literally instead of substituting, when a template must iterate over a list or
  reach into nested config, or when someone asks about AutoPDF template syntax,
  delimiters, or the variables block.
---

# AutoPDF template syntax

Templates are Go `text/template` with **custom delimiters**, compiled by LaTeX
afterwards.

## The delimiters

```
delim[[ ... ]]
```

Not `{{ ... }}`. LaTeX uses braces for nearly every command, so Go's default
delimiters would collide with the document itself. The engine configures
`.Delims("delim[[", "]]")`; there is no option to change it, so a template
written with `{{ }}` renders those characters literally and substitutes nothing.

**If variables appear verbatim in the output, check the delimiters first.** The
second thing to check is that the name exists in the config's `variables:` block
— a missing key renders empty rather than erroring.

## Basic substitution

```yaml
variables:
  title: "Quarterly Report"
  author: "Ada Lovelace"
  date: "2026-08-29"
```

```latex
\documentclass{article}
\title{delim[[.title]]}
\author{delim[[.author]]}
\date{delim[[.date]]}
\begin{document}
\maketitle
\end{document}
```

## Nested values

```yaml
variables:
  metadata:
    version: "2.0.0"
    settings:
      verbose: true
```

```latex
Version: delim[[.metadata.version]]
Verbose: delim[[.metadata.settings.verbose]]
```

Dotted paths reach arbitrarily deep. Nested data must come from the YAML config;
the Go `api.Request.Variables` field is `map[string]string` and cannot express
it — drive generation from a config file when the document needs structure.

## Loops

```yaml
variables:
  tags: ["example", "advanced", "complex"]
  items:
    - name: "Feature 1"
      enabled: true
      priority: 1
    - name: "Feature 2"
      enabled: false
      priority: 2
```

```latex
delim[[range .tags]]
delim[[.]]\par
delim[[end]]

delim[[range .items]]
\subsection{delim[[.name]]}
Enabled: delim[[.enabled]] \\
Priority: delim[[.priority]]
delim[[end]]
```

Inside `range`, `.` is the current element. Every `range` needs a matching
`end`, and an unclosed one fails template parsing before LaTeX ever runs — so
that error names a template line, not a `.tex` line.

## Conditionals

```latex
delim[[if .metadata.settings.verbose]]
\textbf{Verbose mode}
delim[[end]]

delim[[if eq .status "paid"]]
\textcolor{green}{PAID}
delim[[else]]
\textcolor{red}{OUTSTANDING}
delim[[end]]
```

## LaTeX escaping

Substitution is textual. A value containing `&`, `%`, `$`, `#`, `_`, `{` or `}`
reaches LaTeX raw and will break compilation — `Q&A` and `100%` are the usual
culprits. Escape at the data source:

```yaml
variables:
  company: "Smith \\& Sons"
  margin: "40\\%"
```

The doubled backslash is YAML escaping; LaTeX receives `\&` and `\%`. Values
that came from user input rather than a hand-written config should be escaped
programmatically before they reach the template.

## Whitespace

`delim[[range]]` and `delim[[end]]` leave blank lines behind, and in LaTeX a
blank line is a paragraph break. Trim with a leading dash when the spacing
matters:

```latex
delim[[range .items -]]
\item delim[[.name]]
delim[[- end]]
```

## Multi-pass documents

`\ref`, `\cite`, `\tableofcontents` and page-number cross-references need more
than one compile. Set `passes: 2` (or 3), or `use_latexmk: true` to let latexmk
decide. Symptoms of too few passes are `??` in place of references and a table
of contents that is empty or one revision stale.

## Debugging order

1. Are the delimiters `delim[[ ]]`?
2. Is the key present in `variables:`?
3. Does the path match the YAML nesting exactly?
4. Are `range`/`if` blocks closed with `end`?
5. Does any value contain unescaped LaTeX specials?
6. Read the **first** error in the LaTeX log — later ones cascade from it.

Run with `autopdf build template.tex config.yaml` plus `autopdf debug switch` to
keep intermediates and see the processed `.tex` before compilation. Inspecting
that intermediate separates "the template substituted wrong" from "LaTeX
rejected the result", which are different bugs.
