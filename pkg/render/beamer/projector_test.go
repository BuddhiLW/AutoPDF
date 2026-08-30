package beamer_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/beamer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// engine wires the beamer target the way an adopter does: the target supplies
// its own catalog and projector, and pkg/api names neither.
func engine(t *testing.T, options beamer.Options) *api.DocumentEngine {
	t.Helper()
	catalog, err := beamer.Catalog()
	require.NoError(t, err)
	engine, err := api.NewDocumentEngine(catalog, api.DocumentEngineConfig{
		Target:     beamer.Target,
		MaxWorkers: 2,
		Projector:  beamer.NewProjector(options),
	})
	require.NoError(t, err)
	return engine
}

func decode(t *testing.T, raw string) document.DocumentSpec {
	t.Helper()
	spec, err := document.Decode([]byte(raw))
	require.NoError(t, err)
	return spec
}

func project(t *testing.T, spec document.DocumentSpec, options beamer.Options, focus ...string) map[string]string {
	t.Helper()
	projection, err := engine(t, options).Project(context.Background(), spec, api.ProjectionOptions{FocusSections: focus})
	require.NoError(t, err)
	files := make(map[string]string, len(projection.Files))
	for _, file := range projection.Files {
		files[file.Path] = string(file.Content)
	}
	return files
}

func main(files map[string]string) string { return files["main.tex"] }

func frames(files map[string]string) []string {
	var contents []string
	for path, content := range files {
		if strings.HasPrefix(path, "frames/") {
			contents = append(contents, content)
		}
	}
	return contents
}

const twoFrames = `{
  "schemaVersion": 1,
  "id": "Talk",
  "blocks": [
    {"id":"s1","kind":"section","variant":"default","props":{"title":"One"},
     "children":[{"id":"t1","kind":"text","variant":"default",
       "children":[{"id":"sp1","kind":"span","variant":"default","props":{"text":"alpha "}},
                   {"id":"sp2","kind":"span","variant":"default","props":{"text":"beta","marks":["strong"]}}]}]},
    {"id":"s2","kind":"section","variant":"default","props":{"title":"Two"},
     "children":[{"id":"b1","kind":"bullets","variant":"default","props":{"items":["x","y"]}}]}
  ]
}`

func TestChildrenAreSplicedIntoTheFrame(t *testing.T) {
	files := project(t, decode(t, twoFrames), beamer.Options{})
	var one string
	for _, frame := range frames(files) {
		if strings.Contains(frame, "{One}") {
			one = frame
		}
	}
	require.NotEmpty(t, one)
	assert.NotContains(t, one, beamer.ChildrenMarker, "the marker must be substituted, not shipped")
	assert.Contains(t, one, "alpha ")
	assert.Contains(t, one, "\\textbf{beta}")
	assert.Less(t, strings.Index(one, "alpha"), strings.Index(one, "\\end{frame}"))
}

func TestInlineRunsAreNotSeparatedByInput(t *testing.T) {
	files := project(t, decode(t, twoFrames), beamer.Options{})
	for _, frame := range frames(files) {
		assert.NotContains(t, frame, "\\input{",
			"a flow fragment written as its own file would inject a line break mid-sentence")
	}
	var one string
	for _, frame := range frames(files) {
		if strings.Contains(frame, "alpha") {
			one = frame
		}
	}
	assert.Contains(t, one, "alpha \\textbf{beta}",
		"adjacent spans must abut; spacing belongs to the text, not the assembler")
}

func TestEachSectionGetsItsOwnFileSoFocusedBuildsWork(t *testing.T) {
	files := project(t, decode(t, twoFrames), beamer.Options{})
	assert.Len(t, frames(files), 2)
	assert.Equal(t, 2, strings.Count(main(files), "\\include{frames/"))
}

func TestFocusSectionsEmitIncludeonly(t *testing.T) {
	files := project(t, decode(t, twoFrames), beamer.Options{}, "s2")
	body := main(files)
	assert.Contains(t, body, "\\includeonly{")
	line := body[strings.Index(body, "\\includeonly{"):]
	line = line[:strings.Index(line, "\n")]
	assert.Equal(t, 1, strings.Count(line, "frames/"), "only the focused frame may be listed")
}

func TestUnknownFocusSectionIsAnError(t *testing.T) {
	_, err := engine(t, beamer.Options{}).Project(context.Background(), decode(t, twoFrames),
		api.ProjectionOptions{FocusSections: []string{"nope"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "focus")
}

func TestProjectionIsDeterministic(t *testing.T) {
	spec := decode(t, twoFrames)
	first, err := engine(t, beamer.Options{}).Project(context.Background(), spec, api.ProjectionOptions{})
	require.NoError(t, err)
	second, err := engine(t, beamer.Options{}).Project(context.Background(), spec, api.ProjectionOptions{})
	require.NoError(t, err)
	assert.Equal(t, first.Hash, second.Hash)
	require.Equal(t, len(first.Files), len(second.Files))
	for index := range first.Files {
		assert.Equal(t, first.Files[index].Path, second.Files[index].Path)
		assert.Equal(t, first.Files[index].Content, second.Files[index].Content)
	}
}

func TestGraphicsPathAlsoSetsInputPath(t *testing.T) {
	body := main(project(t, decode(t, twoFrames), beamer.Options{GraphicsPath: "/assets"}))
	assert.Contains(t, body, "\\graphicspath{{/assets/}}")
	assert.Contains(t, body, "\\input@path",
		"\\IfFileExists searches the input path, not \\graphicspath: without both, every asset guard reports missing")
}

func TestThemeOptionsCannotInjectLatex(t *testing.T) {
	body := main(project(t, decode(t, twoFrames), beamer.Options{
		Theme:      `Madrid}\input{/etc/passwd}\usetheme{`,
		ColorTheme: `x{}\relax`,
	}))
	assert.NotContains(t, body, "\\input{/etc/passwd}")
	assert.Contains(t, body, "\\usetheme{Madridinputetcpasswdusetheme}")
}

func TestANonSectionRootIsRejected(t *testing.T) {
	spec := decode(t, `{"schemaVersion":1,"blocks":[{"id":"t","kind":"text","variant":"default"}]}`)
	_, err := engine(t, beamer.Options{}).Project(context.Background(), spec, api.ProjectionOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sequence of frames",
		"a Beamer document is frames at the top level; saying so beats a LaTeX error later")
}

func TestANestedSectionIsRejected(t *testing.T) {
	spec := decode(t, `{"schemaVersion":1,"blocks":[
	  {"id":"s1","kind":"section","variant":"default","children":[
	    {"id":"s2","kind":"section","variant":"default"}]}]}`)
	_, err := engine(t, beamer.Options{}).Project(context.Background(), spec, api.ProjectionOptions{})
	require.ErrorIs(t, err, beamer.ErrNestedSection)
}

func TestTitleProducesATitlePage(t *testing.T) {
	body := main(project(t, decode(t, twoFrames), beamer.Options{Title: "My Talk", Author: "Someone"}))
	assert.Contains(t, body, "\\title{My Talk}")
	assert.Contains(t, body, "\\frame{\\titlepage}")

	withoutTitle := main(project(t, decode(t, twoFrames), beamer.Options{}))
	assert.NotContains(t, withoutTitle, "\\titlepage", "no title, no title page")
}

func TestPreambleCarriesWhatTheRenderersNeed(t *testing.T) {
	body := main(project(t, decode(t, twoFrames), beamer.Options{}))
	// Each package here is required by a renderer; dropping one turns a
	// working deck into an undefined-control-sequence error at compile time.
	for _, required := range []string{
		"\\documentclass[aspectratio=169]{beamer}",
		"{graphicx}", "{listings}", "{tcolorbox}", "{ulem}", "{xcolor}", "{hyperref}",
	} {
		assert.Contains(t, body, required)
	}
}

// ── the compile gate ────────────────────────────────────────────────────────

func TestProjectionCompilesToAPDF(t *testing.T) {
	if _, err := exec.LookPath("pdflatex"); err != nil {
		t.Skip("pdflatex not installed")
	}
	spec := decode(t, richDeck)
	projection, err := engine(t, beamer.Options{Title: "Integration", Theme: "Madrid"}).
		Project(context.Background(), spec, api.ProjectionOptions{})
	require.NoError(t, err)

	workspace := t.TempDir()
	for _, file := range projection.Files {
		path := filepath.Join(workspace, file.Path)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, file.Content, 0o644))
	}
	command := exec.Command("pdflatex", "-interaction=nonstopmode", "-halt-on-error", projection.Main)
	command.Dir = workspace
	output, err := command.CombinedOutput()
	require.NoError(t, err, "pdflatex failed:\n%s", tail(string(output), 3000))

	info, err := os.Stat(filepath.Join(workspace, "main.pdf"))
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(1000))
}

func tail(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[len(value)-limit:]
}

// richDeck exercises every kind that has non-trivial LaTeX, including the ones
// that abort a build when they are wrong: verbatim code, a padded table, an
// unprintable GIF, and text carrying LaTeX metacharacters.
const richDeck = `{
  "schemaVersion": 1,
  "id": "Rich",
  "blocks": [
    {"id":"f1","kind":"section","variant":"default","props":{"title":"Text & 100% escaping"},
     "children":[
       {"id":"k1","kind":"kicker","variant":"default","props":{"text":"ACME CORP"}},
       {"id":"h1","kind":"heading","variant":"default","props":{"level":2},
        "children":[{"id":"hs","kind":"span","variant":"default","props":{"text":"A heading"}}]},
       {"id":"p1","kind":"text","variant":"default","children":[
         {"id":"ps1","kind":"span","variant":"default","props":{"text":"Up "}},
         {"id":"ps2","kind":"span","variant":"default","props":{"text":"86%","marks":["strong"]}},
         {"id":"ps3","kind":"span","variant":"default","props":{"text":" with a_b & #1"}},
         {"id":"pl1","kind":"link","variant":"default","props":{"text":"a link","href":"https://example.com"}}]},
       {"id":"r1","kind":"rule","variant":"default"},
       {"id":"n1","kind":"notes","variant":"default","props":{"text":"mention the pipeline"}}]},

    {"id":"f2","kind":"section","variant":"default","props":{"title":"Lists, code, quote"},
     "children":[
       {"id":"b2","kind":"bullets","variant":"default","props":{"items":["first 50%","second"],"fragments":true}},
       {"id":"c2","kind":"code","variant":"default","props":{"language":"go","source":"func main() { x := a & b }"}},
       {"id":"q2","kind":"quote","variant":"default","props":{"cite":"Plato"},
        "children":[{"id":"qs","kind":"span","variant":"default","props":{"text":"the unexamined deck"}}]}]},

    {"id":"f3","kind":"section","variant":"default","props":{"title":"Table, media, columns"},
     "children":[
       {"id":"t3","kind":"table","variant":"default",
        "props":{"head":["Quarter","Revenue","Delta"],"rows":[["Q3","18.4"],["Q4","34.2","+86%"]],"caption":"USD m"}},
       {"id":"m3","kind":"media-placeholder","variant":"default",
        "props":{"mediaType":"video","src":"demo.mp4","caption":"Five seconds"}},
       {"id":"i3","kind":"image","variant":"default","props":{"src":"loop.gif","caption":"An animation"}},
       {"id":"col3","kind":"columns","variant":"default","children":[
         {"id":"cd1","kind":"card","variant":"default","props":{"title":"Left"},
          "children":[{"id":"cs1","kind":"span","variant":"default","props":{"text":"one"}}]},
         {"id":"cd2","kind":"card","variant":"default","props":{"title":"Right"},
          "children":[{"id":"cs2","kind":"span","variant":"default","props":{"text":"two"}}]}]},
       {"id":"cal3","kind":"callout","variant":"default","props":{"tone":"warn","title":"Careful"},
        "children":[{"id":"cals","kind":"span","variant":"default","props":{"text":"a warning"}}]}]}
  ]
}`

func TestRichDeckSpecIsValid(t *testing.T) {
	var raw map[string]any
	require.NoError(t, json.Unmarshal([]byte(richDeck), &raw), "the fixture must be valid JSON")
	assert.Empty(t, decode(t, richDeck).Problems())
}
