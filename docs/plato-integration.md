# AutoPDF × plato: one source, a Reveal deck and a Beamer PDF

[plato](https://github.com/BuddhiLW/plato) is a data-driven ClojureScript
presentation engine. It parses Markdown and Org into a document IR and renders a
Reveal.js deck. AutoPDF compiles LaTeX. Neither has what the other does: plato's
PDF story is Reveal's print-pdf, which is a browser screenshot; AutoPDF has no
Markdown or Org front end.

Pairing them means one `.org` or `.md` file produces both surfaces.

```bash
plato spec talk.org -o talk.json
autopdf deck talk.json talk.pdf assets=./public theme=metropolis
```

And live, rebuilding on every save:

```bash
autopdf deck watch talk.json talk.pdf \
  watch=talk.org command="plato spec talk.org -o talk.json"
```

## The seam

plato parses; AutoPDF compiles; the wire format is a `document.DocumentSpec` in
JSON. Neither repository depends on the other's language or runtime.

```
talk.org ──plato.markdown/plato.org──> plato.doc ──┬── doc/document->deck ──> Reveal deck
                                                    └── spec/document->spec ─> DocumentSpec JSON
                                                                                     │
                                                                        AutoPDF beamer target
                                                                                     │
                                                                                    PDF
```

The join is at `plato.doc`, the IR both front ends already parse to — **not** at
`plato.source/->deck`, because a deck is Reveal-shaped, and **not** at
`plato.content/render`, because hiccup is HTML-shaped. plato already had
`doc/document->deck`; the integration adds its sibling `spec/document->spec`.

## Adding a render target

AutoPDF's component dispatch key is three-dimensional:

```go
// pkg/component/definition.go
type Key struct { Kind, Variant, Target string }
```

`Target` is a first-class axis, folded into the fragment cache hash, so several
targets coexist in one catalog. A target supplies two things — a catalog and a
projector — and `pkg/api` names neither:

```go
catalog, _ := beamer.Catalog()
engine, _ := api.NewDocumentEngine(catalog, api.DocumentEngineConfig{
    Target:    beamer.Target,
    Projector: beamer.NewProjector(beamer.Options{Theme: "metropolis"}),
})
```

`api.ManifestProjector` is satisfied structurally, so `pkg/render/beamer` does
not import `pkg/api` either. `test/architecture/render_target_openness_test.go`
fails if the core ever imports a concrete target.

## The shared vocabulary

Every kind is a `document.Component` with `Variant: "default"`. plato's
projection emits these; AutoPDF's beamer target registers a renderer for each;
`test/plato_integration` holds both sides to it.

| Kind | Mode | Props | Beamer |
| --- | --- | --- | --- |
| `section` | Section | `title`, `level` | `\begin{frame}` + `\frametitle` |
| `text` | Flow | `content`, `format` | paragraph |
| `span` | Flow | `text`, `marks[]` | `\textbf`, `\emph`, `\texttt`, … |
| `link` | Flow | `text`, `href` | `\href` |
| `heading` | Flow | `level` | sized bold text, never `\section` |
| `bullets` | Flow | `items[]`, `ordered`, `fragments` | `itemize` / `enumerate` |
| `code` | Flow | `language`, `source`, `highlight` | `lstlisting` |
| `quote` | Flow | `cite` | `quote` + attribution |
| `table` | Flow | `head[]`, `rows[][]`, `caption` | `tabular` |
| `image` | Flow | `src`, `alt`, `caption`, `width` | guarded `\includegraphics` |
| `columns` | Flow | `widths[]` | `\begin{columns}` |
| `cards` / `card` | Flow | `columns`; `title`, `icon` | `tcolorbox` |
| `callout` | Flow | `tone`, `title` | `tcolorbox` |
| `kicker` | Flow | `text` | small-caps lead line |
| `media-placeholder` | Flow | `mediaType`, `src`, `poster`, `caption` | poster + framed caption |
| `notes` | Flow | `text` | `\note` |
| `rule` | Flow | — | `\rule` |
| `scene` | Artifact | `svg` | guarded `\includegraphics` |

Inline structure is a `text` component's children: `span`s carrying `marks`, and
`link`s. Two plato constructs do not become kinds:

- `:group` **flattens** — its items become siblings.
- `:fragment` **wraps** — it contributes `Style["overlay"]` to what it wraps,
  because a Beamer overlay annotates an element rather than being one.

### Style keys

| Key | Source | Beamer |
| --- | --- | --- |
| `overlay` | `:fragment` | overlay specification, e.g. `<+->` |
| `stackOf` | a flattened vertical stack | the parent frame's title |
| `backgroundColor` | slide `:opts` | `\setbeamercolor{background canvas}` |

Reveal-only options — transitions, auto-animate, visibility — have no Beamer
meaning and are dropped rather than carried.

## Fragments and overlays are the same idea

Reveal reveals on click, Beamer on advance. `content/bullets` with
`:fragments? true` becomes `\begin{itemize}[<+->]`, and a `content/fragment`
with an `:index` becomes `<n->`. This part of the mapping is native on both
sides rather than approximated.

## Theming is single-source

`theme/*.tokens.edn` already projected to CSS, a Clojure namespace and a JSON
manifest. It now also projects to a Beamer style file, so the deck and the PDF
carry identical colours:

```bash
plato theme theme/acme.tokens.edn -o public/css/acme-theme.css --sty theme/acme-beamer.sty
autopdf deck talk.json talk.pdf style=theme/acme-beamer.sty
```

`plato.tokens/beamer-roles` maps token keys to Beamer colour elements and is
data a caller can replace. Only hex values reach `\definecolor`: `rgba()` and
named CSS colours have no LaTeX spelling, and an unknown xcolor name is a
compile error, so they are dropped.

The `.sty` is carried into the compile workspace through the projection
(`beamer.Options.Files`), because generation runs in a private temp directory
where a file beside your shell is invisible.

## What does not map, and how it degrades

**Video, audio and embedded iframes have no PDF equivalent.** Beamer's `\movie`
and `media9` only play in Adobe Reader, so a deck relying on them is silently
broken for most readers. They render as a framed placeholder naming the medium,
the caption and the source, plus the poster frame when there is one.

**GIF and SVG cannot be included at all.** pdflatex reads PDF, PNG, JPEG and
EPS; an unknown graphics extension is a *fatal* error, so one animated GIF would
otherwise produce no PDF. They degrade to the same visible placeholder. This
also means a Desargues scene — which `plato.snapshot/write-svg!` writes as SVG —
prints only once an asset boundary has converted it.

**Vertical stacks flatten.** Reveal has a second navigation axis; Beamer does
not. A stack's slides become consecutive frames carrying `Style["stackOf"]`.

## Drift is the standing risk

Two IRs that evolve independently rot silently. `test/plato_integration` is the
guard: `deck.md` is the shared source, `deck.json` is plato's committed
projection of it, and `TestGoldenMatchesPlatosCurrentProjection` regenerates it
from a plato checkout beside this repository and reports a one-line outline
diff. It skips when plato or the Clojure CLI is absent.

Regenerate after an intentional change:

```bash
cd ../play/plato && clojure -M:cli spec $(realpath ../../AutoPDF/test/plato_integration/deck.md) \
  -o $(realpath ../../AutoPDF/test/plato_integration/deck.json)
```

## Latency: what "live" actually means

A two-frame deck rebuilds in about 1.5 seconds end to end — a Clojure CLI start
for the reparse, then two pdflatex passes. `autopdf deck watch` debounces so a
burst of editor writes is one rebuild, and a failed build leaves the watch
running.

That is a viable loop, not an instant one. The honest arrangement is **the
browser deck stays the instant surface, and the PDF trails by one compile**,
which is the right trade when the artifact you are shipping is print. Two ways
to close the gap: plato's native ClojureWasm binary removes the JVM start, and
`focus=ID` rebuilds a single frame through `\includeonly`.

For a browser-driven preview with per-page updates over SSE or WebSocket, see
[streaming transports](streaming-transports.md); the beamer target produces the
same `latex.Projection` that pipeline consumes.
