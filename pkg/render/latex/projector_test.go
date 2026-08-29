package latex_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/composition"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

func manifest() composition.RenderManifest {
	return composition.RenderManifest{
		RootOrder: []string{"intro", "appendix"},
		Fragments: []composition.Fragment{
			{ComponentID: "intro", Mode: document.Section, Content: []byte("Intro"), Children: []string{"body"}, Hash: "111111111111"},
			{ComponentID: "body", Mode: document.Flow, Content: []byte("Body"), Children: []string{"chart"}, Hash: "222222222222"},
			{ComponentID: "chart", Mode: document.Artifact, Assets: []document.AssetRef{{ID: "plot", URI: "store://plot.pdf", MediaType: "application/pdf"}}, Hash: "333333333333"},
			{ComponentID: "appendix", Mode: document.Section, Content: []byte("Appendix"), Hash: "444444444444"},
		},
	}
}

func TestProjectMapsModesAndSourceLocations(t *testing.T) {
	projection, err := latex.Project(manifest(), latex.Options{FocusSections: []string{"intro"}})
	if err != nil {
		t.Fatal(err)
	}
	main := fileContent(t, projection, "main.tex")
	if !strings.Contains(main, `\includeonly{fragments/intro-111111111111}`) ||
		!strings.Contains(main, `\include{fragments/intro-111111111111}`) {
		t.Fatalf("section projection missing: %s", main)
	}
	body := fileContent(t, projection, projection.SourceMap["body"].Path)
	if !strings.Contains(body, `\input{fragments/chart-333333333333}`) {
		t.Fatalf("flow child projection missing: %s", body)
	}
	chart := fileContent(t, projection, projection.SourceMap["chart"].Path)
	if !strings.Contains(chart, `\includepdf[pages=-]{assets/`) || len(projection.Assets) != 1 {
		t.Fatalf("artifact projection missing: %s %+v", chart, projection.Assets)
	}
	if projection.SourceMap["intro"].Line != 1 || projection.Hash == "" {
		t.Fatalf("source map/hash missing: %+v", projection)
	}
}

func TestProjectIsDeterministicAndSanitizesPaths(t *testing.T) {
	value := manifest()
	value.RootOrder = []string{"../../ evil"}
	value.Fragments = []composition.Fragment{{
		ComponentID: "../../ evil", Mode: document.Artifact, Hash: "hash",
		Assets: []document.AssetRef{{ID: "../asset", URI: "../../secret.png", MediaType: "image/png"}},
	}}
	first, err := latex.Project(value, latex.Options{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := latex.Project(value, latex.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("projection is not deterministic")
	}
	for _, file := range first.Files {
		if strings.Contains(file.Path, "..") {
			t.Fatalf("unsafe generated path: %s", file.Path)
		}
	}
	if strings.Contains(first.Assets[0].Path, "..") {
		t.Fatalf("unsafe asset path: %s", first.Assets[0].Path)
	}
}

func TestProjectRejectsNestedSectionsAndInvalidFocus(t *testing.T) {
	value := manifest()
	value.Fragments[1].Mode = document.Section
	_, err := latex.Project(value, latex.Options{})
	if !errors.Is(err, latex.ErrNestedSection) {
		t.Fatalf("expected nested section error, got %v", err)
	}
	_, err = latex.Project(manifest(), latex.Options{FocusSections: []string{"body"}})
	if !errors.Is(err, latex.ErrInvalidManifest) {
		t.Fatalf("expected invalid focus error, got %v", err)
	}
}

func fileContent(t *testing.T, projection latex.Projection, path string) string {
	t.Helper()
	for _, file := range projection.Files {
		if file.Path == path {
			return string(file.Content)
		}
	}
	t.Fatalf("missing file %s", path)
	return ""
}
