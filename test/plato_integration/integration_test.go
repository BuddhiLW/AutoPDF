// Package plato_integration holds the fixture corpus shared with plato.
//
// deck.md is the source; deck.json is plato's projection of it, committed as a
// golden file. The tests here compile that projection and, when a plato
// checkout is reachable, regenerate it to detect drift between the two IRs.
package plato_integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/beamer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sourceFile = "deck.md"
	goldenFile = "deck.json"
	platoPath  = "../../../play/plato"
)

func golden(t *testing.T) document.DocumentSpec {
	t.Helper()
	data, err := os.ReadFile(goldenFile)
	require.NoError(t, err)
	spec, err := document.Decode(data)
	require.NoError(t, err, "the committed projection must satisfy the current DocumentSpec contract")
	return spec
}

func kinds(components []document.Component, seen map[string]int) map[string]int {
	for _, component := range components {
		seen[component.Kind]++
		kinds(component.Children, seen)
	}
	return seen
}

// TestGoldenProjectionSatisfiesTheContract fails when AutoPDF's document
// contract moves away from what plato emits.
func TestGoldenProjectionSatisfiesTheContract(t *testing.T) {
	spec := golden(t)
	assert.Empty(t, spec.Problems())
	assert.NotEmpty(t, spec.Blocks)
	for _, block := range spec.Blocks {
		assert.Equal(t, "section", block.Kind, "every root of a deck is a frame")
	}
}

// TestEveryKindInTheFixtureDispatches fails when the fixture uses a kind the
// beamer catalog does not implement, which is the shape a vocabulary
// disagreement takes.
func TestEveryKindInTheFixtureDispatches(t *testing.T) {
	catalog, err := beamer.Catalog()
	require.NoError(t, err)
	for kind := range kinds(golden(t).Blocks, map[string]int{}) {
		_, err := catalog.Lookup(component.Key{Kind: kind, Variant: "default", Target: beamer.Target})
		assert.NoError(t, err, "kind %q appears in the fixture but has no beamer definition", kind)
	}
}

// TestFixtureExercisesTheInterestingKinds keeps the corpus honest: a fixture
// that drifted down to plain text would pass everything while proving nothing.
func TestFixtureExercisesTheInterestingKinds(t *testing.T) {
	present := kinds(golden(t).Blocks, map[string]int{})
	for _, kind := range []string{
		"section", "text", "span", "link", "bullets", "code", "quote",
		"table", "image", "media-placeholder", "notes", "heading",
	} {
		assert.Positive(t, present[kind], "the fixture no longer exercises %q", kind)
	}
}

// TestGoldenProjectionCompiles is the end of the pipeline: plato's output
// becomes a real PDF.
func TestGoldenProjectionCompiles(t *testing.T) {
	if _, err := exec.LookPath("pdflatex"); err != nil {
		t.Skip("pdflatex not installed")
	}
	spec := golden(t)
	catalog, err := beamer.Catalog()
	require.NoError(t, err)
	engine, err := api.NewDocumentEngine(catalog, api.DocumentEngineConfig{
		Target:    beamer.Target,
		Projector: beamer.NewProjector(beamer.Options{Title: "Fixture", Theme: "Madrid"}),
	})
	require.NoError(t, err)

	projection, err := engine.Project(context.Background(), spec, api.ProjectionOptions{})
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

// TestGoldenMatchesPlatosCurrentProjection is the drift gate.
//
// It regenerates deck.json from deck.md with the plato checkout beside this
// repository and compares. It skips when plato or the Clojure CLI is absent,
// so CI without them still runs everything above.
func TestGoldenMatchesPlatosCurrentProjection(t *testing.T) {
	if _, err := os.Stat(platoPath); err != nil {
		t.Skip("plato checkout not found beside this repository")
	}
	if _, err := exec.LookPath("clojure"); err != nil {
		t.Skip("clojure CLI not installed")
	}

	source, err := filepath.Abs(sourceFile)
	require.NoError(t, err)
	regenerated := filepath.Join(t.TempDir(), goldenFile)

	command := exec.Command("clojure", "-M:cli", "spec", source, "-o", regenerated)
	command.Dir = platoPath
	if output, err := command.CombinedOutput(); err != nil {
		t.Skipf("plato could not project the fixture: %v\n%s", err, tail(string(output), 1500))
	}

	current, err := os.ReadFile(regenerated)
	require.NoError(t, err)
	committed, err := os.ReadFile(goldenFile)
	require.NoError(t, err)

	regenerateHint := fmt.Sprintf("Regenerate with:\n  cd %s && clojure -M:cli spec $(realpath %s) -o $(realpath %s)",
		platoPath, sourceFile, goldenFile)

	// Compare the tree's shape first. A renamed kind or a shifted id is the
	// usual drift, and it reports as a one-line diff instead of two documents.
	if !assert.Equal(t, outline(t, committed), outline(t, current),
		"plato's projection of %s changed shape.\n%s", sourceFile, regenerateHint) {
		return
	}
	assert.Equal(t, normalize(t, committed), normalize(t, current),
		"plato's projection of %s changed in its props or style.\n%s", sourceFile, regenerateHint)
}

// outline reduces a spec to one "id:kind" line per component, depth-first.
func outline(t *testing.T, data []byte) string {
	t.Helper()
	spec, err := document.Decode(data)
	require.NoError(t, err)
	var lines []string
	var walk func([]document.Component, string)
	walk = func(components []document.Component, indent string) {
		for _, component := range components {
			lines = append(lines, indent+component.ID+":"+component.Kind)
			walk(component.Children, indent+"  ")
		}
	}
	walk(spec.Blocks, "")
	return strings.Join(lines, "\n")
}

// normalize compares JSON as values, so formatting alone never fails the gate.
func normalize(t *testing.T, data []byte) string {
	t.Helper()
	var value any
	require.NoError(t, json.Unmarshal(data, &value))
	encoded, err := json.MarshalIndent(value, "", "  ")
	require.NoError(t, err)
	return string(encoded)
}

func tail(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[len(value)-limit:]
}
