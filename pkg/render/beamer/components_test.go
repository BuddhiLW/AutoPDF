package beamer

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// render dispatches through the real catalog, so a test cannot pass against a
// renderer the catalog would never select.
func render(t *testing.T, node document.Component) string {
	t.Helper()
	definitions, err := Definitions()
	require.NoError(t, err)
	catalog, err := component.NewCatalog(definitions...)
	require.NoError(t, err)
	definition, err := catalog.Lookup(component.Key{Kind: node.Kind, Variant: "default", Target: Target})
	require.NoError(t, err)
	output, err := definition.Render(context.Background(), component.RenderContext{}, node)
	require.NoError(t, err)
	return string(output.Content)
}

func node(kind string, props document.Props) document.Component {
	return document.Component{ID: "c1", Kind: kind, Variant: "default", Props: props}
}

// jsonProps round-trips props through JSON, because that is how they reach a
// renderer in production: numbers arrive as json.Number, not int.
func jsonProps(t *testing.T, raw string) document.Props {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	var props document.Props
	require.NoError(t, decoder.Decode(&props))
	return props
}

func TestEveryVocabularyKindIsRegistered(t *testing.T) {
	definitions, err := Definitions()
	require.NoError(t, err)
	registered := make(map[string]bool, len(definitions))
	for _, definition := range definitions {
		assert.Equal(t, Target, definition.Key().Target)
		assert.Equal(t, "default", definition.Key().Variant)
		registered[definition.Key().Kind] = true
	}
	// The vocabulary plato's projection emits. A kind missing here is a
	// dispatch failure at compose time, not a rendering blemish.
	for _, kind := range []string{
		"section", "text", "span", "link", "heading", "bullets", "code", "quote",
		"table", "image", "columns", "cards", "card", "callout", "kicker",
		"media-placeholder", "notes", "rule", "scene",
	} {
		assert.True(t, registered[kind], "kind %q must be registered for the beamer target", kind)
	}
}

func TestBeamerAndLatexTargetsCoexist(t *testing.T) {
	builder := component.NewBuilder()
	latexDefinitions, err := component.Builtins("latex")
	require.NoError(t, err)
	for _, definition := range latexDefinitions {
		require.NoError(t, builder.Register(definition))
	}
	require.NoError(t, Register(builder), "beamer must register alongside latex, not collide with it")

	catalog, err := builder.Freeze()
	require.NoError(t, err)
	latexText, err := catalog.Lookup(component.Key{Kind: "text", Variant: "default", Target: "latex"})
	require.NoError(t, err)
	beamerText, err := catalog.Lookup(component.Key{Kind: "text", Variant: "default", Target: Target})
	require.NoError(t, err)
	assert.NotSame(t, latexText, beamerText, "the target dimension must keep the two definitions apart")
}

// ── escaping ────────────────────────────────────────────────────────────────

func TestDocumentTextIsEscaped(t *testing.T) {
	content := render(t, node("span", document.Props{"text": `100% & <a_b> #1 $x$ {y}`}))
	assert.NotContains(t, content, "100% ", "a bare % comments out the rest of the line")
	assert.Contains(t, content, `100\%`)
	assert.Contains(t, content, `\&`)
	assert.Contains(t, content, `\_`)
	assert.Contains(t, content, `\#`)
	assert.Contains(t, content, `\$`)
	assert.Contains(t, content, `\{`)
}

func TestSectionTitleIsEscaped(t *testing.T) {
	content := render(t, node("section", document.Props{"title": "Q3 & Q4: 100% growth"}))
	assert.Contains(t, content, `Q3 \& Q4: 100\% growth`)
}

func TestCodeIsNotEscaped(t *testing.T) {
	content := render(t, node("code", document.Props{"source": "if x & y { return 100% }", "language": "go"}))
	assert.Contains(t, content, "if x & y { return 100% }",
		"a code block is verbatim; escaping it would corrupt the source it exists to show")
	assert.Contains(t, content, "language=Go")
}

func TestCodeCannotEscapeItsEnvironment(t *testing.T) {
	content := render(t, node("code", document.Props{"source": "a\n\\end{lstlisting}\nrogue"}))
	assert.Equal(t, 1, strings.Count(content, "\\end{lstlisting}"),
		"source that spells the terminator must not close the block early")
}

// ── structure ───────────────────────────────────────────────────────────────

func TestSectionOpensAFragileFrameAroundItsChildren(t *testing.T) {
	content := render(t, node("section", document.Props{"title": "Intro"}))
	assert.Contains(t, content, "\\begin{frame}[fragile]")
	assert.Contains(t, content, "\\frametitle{Intro}")
	assert.Contains(t, content, ChildrenMarker)
	assert.Contains(t, content, "\\end{frame}")
	assert.Less(t, strings.Index(content, ChildrenMarker), strings.Index(content, "\\end{frame}"),
		"children belong inside the frame")
}

func TestTheTitleIsNotAFrameArgument(t *testing.T) {
	// A fragile frame scans for an optional subtitle after \begin{frame}{...}.
	// A child whose LaTeX opens with a brace — a kicker, a sized heading —
	// would then be read as that subtitle and fail on the \par inside it.
	content := render(t, node("section", document.Props{"title": "Intro"}))
	assert.NotContains(t, content, "\\begin{frame}[fragile]{",
		"the title must go through \\frametitle, or the first braced child is swallowed")
}

func TestOnlyHexBackgroundsAreEmitted(t *testing.T) {
	withHex := render(t, document.Component{
		ID: "c1", Kind: "section", Variant: "default",
		Style: document.Style{"backgroundColor": "#1a2b3c"},
	})
	assert.Contains(t, withHex, "\\definecolor{platoframebg}{HTML}{1A2B3C}")
	assert.Contains(t, withHex, "background canvas")

	// An unknown xcolor name is a compile error; a missing tint is not.
	named := render(t, document.Component{
		ID: "c1", Kind: "section", Variant: "default",
		Style: document.Style{"backgroundColor": "rebeccapurple"},
	})
	assert.NotContains(t, named, "definecolor")
	assert.NotContains(t, named, "background canvas")
}

func TestFragileIsUnconditional(t *testing.T) {
	// A section renderer cannot see whether its subtree holds a verbatim
	// environment, and a frame containing one without [fragile] fails to
	// compile. Paying for it always is the only safe choice.
	assert.Contains(t, render(t, node("section", nil)), "[fragile]")
}

func TestSpanMarksNest(t *testing.T) {
	props := jsonProps(t, `{"text":"both","marks":["strong","emph"]}`)
	content := render(t, document.Component{ID: "c1", Kind: "span", Variant: "default", Props: props})
	assert.Contains(t, content, "\\textbf{")
	assert.Contains(t, content, "\\emph{")
	assert.Contains(t, content, "both")
}

func TestBulletsBecomeOverlays(t *testing.T) {
	props := jsonProps(t, `{"items":["a","b"],"fragments":true}`)
	content := render(t, document.Component{ID: "c1", Kind: "bullets", Variant: "default", Props: props})
	assert.Contains(t, content, "\\begin{itemize}[<+->]",
		"Reveal reveals on click, Beamer on advance: the same idea")
	assert.Contains(t, content, "\\item a")
}

func TestOrderedBulletsUseEnumerate(t *testing.T) {
	props := jsonProps(t, `{"items":["a"],"ordered":true}`)
	content := render(t, document.Component{ID: "c1", Kind: "bullets", Variant: "default", Props: props})
	assert.Contains(t, content, "\\begin{enumerate}")
}

func TestAFragmentStyleBecomesAnOverlaySpec(t *testing.T) {
	component := document.Component{
		ID: "c1", Kind: "bullets", Variant: "default",
		Props: jsonProps(t, `{"items":["a"]}`),
		Style: document.Style{"overlay": "+-"},
	}
	assert.Contains(t, render(t, component), "[<+->]")
}

func TestShortTableRowsArePadded(t *testing.T) {
	props := jsonProps(t, `{"head":["A","B","C"],"rows":[["1"],["1","2","3"]]}`)
	content := render(t, document.Component{ID: "c1", Kind: "table", Variant: "default", Props: props})
	assert.Contains(t, content, "\\begin{tabular}{lll}")
	for _, line := range strings.Split(content, "\n") {
		if strings.HasSuffix(line, `\\`) {
			assert.Equal(t, 2, strings.Count(line, "&"),
				"a tabular row with too few ampersands is a compile error, not a ragged edge: %q", line)
		}
	}
}

func TestHeadingIsBodyTextNotASectioningCommand(t *testing.T) {
	props := jsonProps(t, `{"level":2}`)
	content := render(t, document.Component{ID: "c1", Kind: "heading", Variant: "default", Props: props})
	assert.NotContains(t, content, "\\section", "\\section inside a frame is an error")
	assert.Contains(t, content, "\\large")
}

// ── the degradations ────────────────────────────────────────────────────────

func TestUnprintableFormatsDegradeRatherThanAbortTheBuild(t *testing.T) {
	// pdflatex treats an unknown graphics extension as FATAL, so one animated
	// GIF would otherwise produce no PDF at all.
	for _, testCase := range []struct{ source, expect string }{
		{"loop.gif", "ANIMATED IMAGE"},
		{"diagram.svg", "VECTOR IMAGE"},
		{"thing.bmp", "UNSUPPORTED FORMAT"},
	} {
		content := render(t, node("image", document.Props{"src": testCase.source}))
		assert.NotContains(t, content, "\\includegraphics",
			"%s cannot be included by pdflatex", testCase.source)
		assert.Contains(t, strings.ToUpper(content), testCase.expect)
		assert.Contains(t, content, strings.ReplaceAll(testCase.source, ".", "."),
			"the placeholder must name the source so a reader can find it")
	}
}

func TestPrintableImagesAreGuardedNotAssumed(t *testing.T) {
	content := render(t, node("image", document.Props{"src": "chart.png", "caption": "Revenue"}))
	assert.Contains(t, content, "\\IfFileExists{chart.png}",
		"a preview compiles while the author is still adding assets")
	assert.Contains(t, content, "\\includegraphics")
	assert.Contains(t, content, "missing: chart.png")
	assert.Contains(t, content, "Revenue")
}

func TestMediaDegradesVisiblyAndKeepsItsSource(t *testing.T) {
	content := render(t, node("media-placeholder", document.Props{
		"mediaType": "video", "src": "demo.mp4", "caption": "Five seconds",
	}))
	assert.Contains(t, strings.ToUpper(content), "VIDEO")
	assert.Contains(t, content, "Five seconds")
	assert.Contains(t, content, "demo.mp4")
	assert.NotContains(t, content, "\\movie",
		"\\movie and media9 only play in Adobe Reader; relying on them is silently broken for most readers")
}

func TestAnUnprintablePosterIsSkippedNotIncluded(t *testing.T) {
	content := render(t, node("media-placeholder", document.Props{
		"mediaType": "video", "src": "demo.mp4", "poster": "poster.svg",
	}))
	assert.NotContains(t, content, "\\includegraphics", "an SVG poster would abort the build")
}

func TestASceneSaysWhyItCannotPrintYet(t *testing.T) {
	content := render(t, node("scene", document.Props{"svg": "scene.svg"}))
	assert.Contains(t, strings.ToUpper(content), "SCENE")
	assert.Contains(t, content, "scene.svg")
	assert.NotContains(t, content, "\\includegraphics")

	converted := render(t, node("scene", document.Props{"svg": "scene.pdf"}))
	assert.Contains(t, converted, "\\includegraphics",
		"once the asset boundary has converted it, a scene prints like any figure")
}

// ── robustness ──────────────────────────────────────────────────────────────

func TestWrongPropTypesRenderAsAbsentRatherThanPanicking(t *testing.T) {
	for _, testCase := range []struct {
		kind  string
		props string
	}{
		{"span", `{"text":{"nested":"object"},"marks":"not-an-array"}`},
		{"bullets", `{"items":"not-an-array"}`},
		{"table", `{"head":"no","rows":"no"}`},
		{"image", `{"src":123}`},
		{"heading", `{"level":"two"}`},
	} {
		props := jsonProps(t, testCase.props)
		assert.NotPanics(t, func() {
			render(t, document.Component{ID: "c1", Kind: testCase.kind, Variant: "default", Props: props})
		}, "kind %q with props %s", testCase.kind, testCase.props)
	}
}

func TestValidatorsRejectShapesTheyRequire(t *testing.T) {
	definitions, err := Definitions()
	require.NoError(t, err)
	catalog, err := component.NewCatalog(definitions...)
	require.NoError(t, err)

	bullets, err := catalog.Lookup(component.Key{Kind: "bullets", Variant: "default", Target: Target})
	require.NoError(t, err)
	problems := bullets.Validate(document.Component{
		ID: "c1", Kind: "bullets", Variant: "default",
		Props: jsonProps(t, `{"items":"not-an-array"}`),
	})
	assert.NotEmpty(t, problems, "a wrong shape should be reported by name, not silently dropped")
}
