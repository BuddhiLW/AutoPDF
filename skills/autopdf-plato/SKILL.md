---
name: autopdf-plato
description: >
  Build a PDF presentation from Markdown or Org by pairing AutoPDF with plato —
  plato parses the source to a DocumentSpec, AutoPDF's beamer target compiles it
  to a Beamer deck, and `autopdf deck watch` rebuilds on every save. Also covers
  adding a new render target to AutoPDF. Use when someone wants slides from
  .md/.org, mentions plato, Beamer, `autopdf deck`, a DocumentSpec from a text
  front end, sharing theme tokens between a web deck and a PDF, or registering a
  render target other than the built-in LaTeX one.
---

# Presentations: plato parses, AutoPDF compiles

Two commands. plato owns Markdown and Org; AutoPDF owns LaTeX. The wire format
between them is a `document.DocumentSpec` in JSON.

```bash
plato spec talk.org -o talk.json
autopdf deck talk.json talk.pdf assets=./public theme=metropolis
```

Live, rebuilding on every save:

```bash
autopdf deck watch talk.json talk.pdf \
  watch=talk.org command="plato spec talk.org -o talk.json"
```

`autopdf deck watch` debounces, so a burst of editor writes is one rebuild, and
a failed build leaves the watch running. **AutoPDF never parses Markdown or
Org** — `command` is a shell hook, so any front end that emits a DocumentSpec
works the same way.

## `autopdf deck` options

All are `KEY=VALUE`, in any position after the spec:

| Option | Meaning |
| --- | --- |
| `target=NAME` | render target, default `beamer` |
| `title=` `author=` `date=` | title page; omitted entirely when `title` is absent |
| `theme=` `colortheme=` | Beamer theme names, e.g. `Madrid`, `metropolis` |
| `style=FILE.sty` | a style file, carried into the compile workspace |
| `assets=DIR` | directory images resolve against |
| `aspect=169` | class aspect ratio |
| `notes=on` | render speaker notes |
| `focus=ID[,ID]` | rebuild only those frames via `\includeonly` |
| `engine=` `passes=` | LaTeX engine and pass count (default `pdflatex`, 2) |

`watch` adds `watch=FILE`, `command=CMD`, `poll=`, `debounce=`.

## Embedding it

A render target supplies a catalog and a projector; `pkg/api` names neither.

```go
import (
    "github.com/BuddhiLW/AutoPDF/v2/pkg/api"
    "github.com/BuddhiLW/AutoPDF/v2/pkg/document"
    "github.com/BuddhiLW/AutoPDF/v2/pkg/render/beamer"
)

catalog, err := beamer.Catalog()
engine, err := api.NewDocumentEngine(catalog, api.DocumentEngineConfig{
    Target:    beamer.Target,
    Projector: beamer.NewProjector(beamer.Options{Theme: "metropolis"}),
})

spec, err := document.Decode(data)
projection, err := engine.Project(ctx, spec, api.ProjectionOptions{})
```

`beamer.Project` returns the same `latex.Projection` the LaTeX target does, so
`api.NewProjectionGenerator` and the whole preview pipeline work unchanged.

**Adding a target is a new package, never an edit to `pkg/api`.** Implement
`api.ManifestProjector` (satisfied structurally — your package need not import
`pkg/api`) and register your own component definitions.
`test/architecture/render_target_openness_test.go` enforces this.

## Styling is data, not code

`beamer.Options` carries `DocumentClass`, `ClassOptions`, `Packages`,
`Preamble`, `Theme`, `ColorTheme`, `StyleFile` and `Files`. Package-level tables
extend the renderers without forking them:

```go
beamer.ListingLanguages["nim"] = "Python"
beamer.MarkCommands["spoiler"] = "underline"
beamer.CalloutTones["danger"] = struct{ Background, Frame string }{"red!10!white", "red!70!black"}
beamer.PrintableExtensions[".svg"] = true // only if your preamble loads the svg package
```

Register your catalog after mutating these; `Definitions()` reads them.

## Sharing one palette between the deck and the PDF

plato's theme tokens project to a Beamer style file, so both surfaces carry the
same colours from one source:

```bash
plato theme theme/acme.tokens.edn -o public/css/acme-theme.css --sty theme/acme-beamer.sty
autopdf deck talk.json talk.pdf style=theme/acme-beamer.sty
```

Only hex values reach `\definecolor`. `rgba()` and named CSS colours are
dropped, because an unknown xcolor name is a compile error.

## The four things that actually bite

**A `.sty` beside your shell is invisible to the build.** Generation runs in a
private temp workspace. `style=` reads the file and carries it in; if you are
driving the library directly, put it in `beamer.Options.Files`.

**GIF and SVG abort the whole document.** pdflatex reads PDF, PNG, JPEG and EPS;
an unknown graphics extension is fatal, not a warning, so one animated GIF means
no PDF at all. AutoPDF filters by extension and renders a visible placeholder
instead — which is why a Desargues scene prints only after something converts
its SVG snapshot to PDF.

**Video, audio and iframes never play.** `\movie` and `media9` only work in
Adobe Reader. They render as a framed placeholder naming the medium, caption and
source. Do not "fix" this by emitting `\movie`.

**Vertical stacks flatten.** Beamer has no second axis. A stack's slides become
consecutive frames carrying `Style["stackOf"]`.

## Latency

About 1.5s for a small deck: a Clojure CLI start for the reparse plus two
pdflatex passes. Treat the browser deck as the instant surface and the PDF as
the pane that trails by a compile. `focus=ID` rebuilds one frame; plato's native
ClojureWasm binary removes the JVM start.

## Keeping the two sides in step

`test/plato_integration` holds a shared fixture: `deck.md` and plato's committed
projection `deck.json`. `TestGoldenMatchesPlatosCurrentProjection` regenerates
it from a plato checkout and reports a one-line outline diff. After an
intentional vocabulary change, regenerate the golden file rather than editing
it by hand.

Reference: `docs/plato-integration.md` in the AutoPDF repository.
