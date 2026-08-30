# AutoPDF × plato: one source, a Reveal deck and a Beamer PDF

[plato](https://github.com/BuddhiLW/plato) is a data-driven ClojureScript
presentation engine. It parses Markdown and Org into a document IR and renders a
Reveal.js deck. AutoPDF compiles LaTeX and owns a live preview pipeline. Neither
has what the other does: plato's PDF story is Reveal's print-pdf, which is a
browser screenshot; AutoPDF has no Markdown or Org front end.

Pairing them means one `.org` or `.md` file produces both surfaces.

## The seam

plato parses; AutoPDF compiles; the wire format is a `document.DocumentSpec` in
JSON. Neither repository takes a dependency on the other's language or runtime.

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
`plato.content/render`, because hiccup is HTML-shaped. plato already has
`doc/document->deck`; the integration adds its sibling `document->spec`. One
parse, two projections.

## Why a `beamer` target rather than a fork

AutoPDF's component dispatch key is three-dimensional:

```go
// pkg/component/definition.go
type Key struct { Kind, Variant, Target string }
```

`pkg/composition/composer.go` builds that key per node from the composer's
target, registers builtins parameterized by target, and folds `Target` into the
fragment cache hash. `beamer` fragments and `latex` fragments therefore coexist
in one catalog without colliding, and adding the target is
`component.Builder.Register` — no edit to the composer, the catalog, or the
existing LaTeX projector.

## The shared vocabulary

Every kind below is a `document.Component` with `Variant: "default"`. This table
is the contract: plato's projection emits these kinds, AutoPDF's beamer target
registers a renderer for each, and a fixture corpus holds both sides to it.

| plato `:plato/type` | Kind | Mode | Props | Beamer projection |
| --- | --- | --- | --- | --- |
| *(section, from the doc IR)* | `section` | Section | `title`, `level` | `\begin{frame}{title}` |
| string, `:markdown`, `:html` | `text` | Flow | `content`, `format` | paragraph |
| `:bullets` | `bullets` | Flow | `items[]`, `ordered`, `fragment` | `itemize` / `enumerate` |
| `:code` | `code` | Flow | `language`, `source`, `highlight` | `lstlisting` |
| `:quote` | `quote` | Flow | `text`, `cite` | `quote` + attribution |
| `:table` | `table` | Flow | `head[]`, `rows[][]`, `caption` | `tabular` |
| `:image` | `image` | Flow | `src`, `alt`, `caption`, `width` | `figure` + `\includegraphics` |
| `:columns` | `columns` | Flow | children | `\begin{columns}` |
| `:cards` | `cards` | Flow | `items[]` | `tcolorbox` grid |
| `:note` | `callout` | Flow | `tone`, `content` | `tcolorbox` |
| `:kicker` | `kicker` | Flow | `text` | small-caps lead line |
| `:video`, `:audio`, `:embed` | `media-placeholder` | Flow | `mediaType`, `src`, `poster`, `caption` | poster + caption + URL |
| *(Desargues scene)* | `scene` | Artifact | `svg` | `\includegraphics` of the snapshot |

Two plato constructs do not become kinds of their own:

- `:group` **flattens** — its items become siblings in the parent's children.
- `:fragment` **wraps** — it contributes `Style["overlay"]` to the component it
  wraps rather than adding a node, because a Beamer overlay is an annotation on
  an element, not an element.

### Style keys

`Style` carries presentation choices that are not semantic content:

| Key | Source | Beamer |
| --- | --- | --- |
| `overlay` | `:fragment` wrapper | `<+->` overlay specification |
| `level` | section depth | frame vs. subsection placement |
| `background-color` | slide `:opts` | frame background |
| `align` | content opts | alignment environment |

## Fragments and overlays are the same idea

Reveal reveals elements one click at a time; Beamer reveals them one overlay at
a time. `content/bullets` with `:fragment true` becomes `\begin{itemize}[<+->]`,
and `content/code` with `:highlight "1|3-5"` — Reveal's stepped highlight — is
the same stepping Beamer expresses as overlay specifications. This is the part of
the mapping that is genuinely native on both sides rather than approximated.

## What does not map, and how it degrades

**Video, audio and embedded iframes have no PDF equivalent.** Beamer's `\movie`
and `media9` only play in Adobe Reader, so relying on them produces a deck that
is broken for most readers. These project to `media-placeholder`: the poster
frame if there is one, the caption, and the source URL. The degradation is
visible on the slide — a reader can see that something interactive lives here
and where to get it. It is never silent.

**Vertical stacks flatten.** Reveal has a second navigation axis; Beamer does
not. A stack's slides become consecutive frames, and the stack's own identity
survives as a subsection.

**Desargues scenes become their final frame.** plato's
`plato.snapshot/write-svg!` already renders a scene's terminal state to a
standalone SVG, which is exactly what a printed deck wants from an animation.

## Drift is the standing risk

Two IRs that evolve independently rot silently. The mitigation is not review: it
is `test/plato_integration`, a fixture corpus where one source produces both a
deck and a spec and a test asserts the two projections agree on section and
block counts. `document.Decode` uses `DisallowUnknownFields`, so drift in the
top-level shape fails loudly on the first run — but `Props` is `map[string]any`
and will not, which is precisely what the fixture gate is for.

## Latency: what "live" actually means

AutoPDF's preview session keeps TeX auxiliary state warm, supersedes obsolete
builds, fingerprints pages and transmits only the ones that changed. That makes
the loop viable, not instantaneous: a real Beamer deck compiles in roughly one
to three seconds, and cross-references need more than one pass.

So the honest arrangement is **the browser deck stays the instant surface, and
the PDF is a second pane that trails by one compile.** That is the right trade
when the artifact you are actually shipping is print.
