package deck

import (
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/beamer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputDefaultsToTheSpecName(t *testing.T) {
	parsed, err := parseRequest([]string{"talk.json"})
	require.NoError(t, err)
	assert.Equal(t, "talk.pdf", parsed.outputPath)
	assert.Equal(t, beamer.Target, parsed.target.Name)
	assert.Equal(t, "pdflatex", parsed.engine)
	assert.Equal(t, 2, parsed.passes)
}

func TestSecondBareArgumentIsTheOutput(t *testing.T) {
	parsed, err := parseRequest([]string{"talk.json", "out/deck.pdf"})
	require.NoError(t, err)
	assert.Equal(t, "out/deck.pdf", parsed.outputPath)
}

func TestAThirdBareArgumentIsRejected(t *testing.T) {
	_, err := parseRequest([]string{"talk.json", "a.pdf", "b.pdf"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "KEY=VALUE")
}

func TestOptionsArePositionIndependent(t *testing.T) {
	parsed, err := parseRequest([]string{
		"talk.json", "theme=metropolis", "out.pdf", "title=Q3", "notes=on",
	})
	require.NoError(t, err)
	assert.Equal(t, "out.pdf", parsed.outputPath)
	assert.Equal(t, "metropolis", parsed.settings.Theme)
	assert.Equal(t, "Q3", parsed.settings.Title)
	assert.True(t, parsed.settings.ShowNotes)
}

func TestAssetsResolveToAnAbsolutePath(t *testing.T) {
	parsed, err := parseRequest([]string{"talk.json", "assets=./public"})
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(parsed.settings.GraphicsPath),
		"the compile runs in a temp workspace, so a relative asset path finds nothing")
}

func TestFocusAcceptsAList(t *testing.T) {
	parsed, err := parseRequest([]string{"talk.json", "focus=intro, summary ,"})
	require.NoError(t, err)
	assert.Equal(t, []string{"intro", "summary"}, parsed.focus)
}

func TestPassesMustBePositive(t *testing.T) {
	for _, value := range []string{"0", "-1", "two", ""} {
		_, err := parseRequest([]string{"talk.json", "passes=" + value})
		if value == "" {
			require.NoError(t, err, "an empty value falls back to the default")
			continue
		}
		require.Error(t, err, "passes=%s", value)
	}
}

func TestAnUnknownTargetNamesTheKnownOnes(t *testing.T) {
	_, err := parseRequest([]string{"talk.json", "target=typst"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "typst")
	assert.Contains(t, err.Error(), beamer.Target)
}

func TestEveryRegisteredTargetWiresAnEngine(t *testing.T) {
	require.NotEmpty(t, TargetNames())
	for _, name := range TargetNames() {
		target, err := LookupTarget(name)
		require.NoError(t, err)
		assert.Equal(t, name, target.Name, "registry key and target name must agree")

		engine, err := NewEngine(target, Settings{Theme: "Madrid"}, 2)
		require.NoError(t, err, "target %q", name)
		require.NotNil(t, engine)
		assert.NotEmpty(t, engine.Catalog(), "target %q registers no components", name)
	}
}
