# Precompiled preamble: `format_file`

`format_file: true` makes the LaTeX backend dump a template's static preamble
into a TeX format file and start each later compile from it, skipping the
package loading the preamble would otherwise repeat. It is off by default.

## Surfaces

| Surface | Setting |
|---|---|
| config YAML / JSON | `format_file: true` |
| Go embedding (`pkg/api`) | `api.Request{FormatFile: true}` |
| internal ports | `ports.LaTeXOptions{FormatFile: true}` / `.WithFormatFile(true)` |
| backend registry | `backend.Settings{FormatFile: true}` |

Typst refuses it with `UnsupportedSettingError{Setting: "format_file"}`. LaTeX
refuses it together with `use_latexmk`.

## What happens on a build

1. The preamble is everything before the first `\begin{document}`,
   `\endofdump` or `\csname endofdump\endcsname`. Put a marker before
   preamble lines that change per document (for example a `\newcommand`
   holding a variable) to keep them out of the format.
2. The cache key is `<engine>-<hash(engine, preamble)>`. The format and its
   dependency manifest live in `<working dir>/.autopdf-fmt/`.
3. On a miss AutoPDF runs, in the working directory and with the same sandbox
   as every compile (`-no-shell-escape`, `openin_any=p`, `openout_any=p`,
   `shell_escape=f`):

   ```
   pdflatex -ini -interaction=nonstopmode -no-shell-escape -recorder \
     -jobname=<key>-<tmp> -output-directory=.autopdf-fmt \
     '&pdflatex' mylatexformat.ltx <key>-<tmp>.autopdf.tex
   ```

   The `-recorder` output names the local files the preamble read (a `.sty`,
   `.cls` or `\input` file in the template directory). Their hashes go into
   `<key>.deps`, so editing one of them rebuilds the format. Files of the TeX
   distribution are not tracked: rebuild by deleting `.autopdf-fmt/` after a
   TeX Live upgrade (a format the engine refuses is also discarded on its own,
   see below).
4. Each compile pass adds `-fmt=.autopdf-fmt/<key>`, relative to the working
   directory, as the sandbox requires.

## Failure is never fatal

- If the dump fails (missing `mylatexformat`, a preamble the engine cannot
  dump), the build logs a warning and compiles normally. A `<key>.failed`
  marker holding the engine's version stops it retrying the same dump on every
  build. It is retried after an engine upgrade.
- If a compile with the format fails, the build retries once without it, and
  when that succeeds it deletes the format so the next build rebuilds it.

## Engines

| Engine | Result |
|---|---|
| pdflatex | works |
| xelatex | dumps fail for preambles with native fonts (`fontspec`, the default font setup): falls back |
| lualatex | `lualatex -ini` does not find the LaTeX base files on TeX Live 2025: falls back |

## Measurement

N=5 sequential builds of a one-page document whose preamble loads `fontenc`,
`lmodern`, `amsmath`, `amssymb`, `graphicx`, `xcolor`, `tikz` (three
libraries), `pgfplots`, `booktabs`, `longtable`, `geometry`, `fancyhdr` and
`hyperref`, one pass, pdflatex (TeX Live 2025), through the LaTeX compiler
adapter. Three runs on a loaded host (load average about 20):

| | builds 1-5 | mean of builds 2-5 |
|---|---|---|
| plain | 0.91 0.97 0.79 0.82 0.82 s | 0.85 s |
| `format_file` | 1.09 0.19 0.22 0.20 0.29 s | 0.22 s |
| plain | 0.77 0.78 0.92 1.00 0.97 s | 0.92 s |
| `format_file` | 1.45 0.43 0.21 0.20 0.27 s | 0.28 s |
| plain | 0.80 0.92 1.10 0.92 0.90 s | 0.96 s |
| `format_file` | 1.30 0.24 0.31 0.25 0.31 s | 0.28 s |

The first build costs about 0.3-0.5 s more (it builds the format). Each later
build is about 3-4x faster. A template with a short preamble gains little.

## Output

Text and layout match a plain compile. The bytes differ, as for any two
compiles: internal font resource names and the document `/ID` change.
