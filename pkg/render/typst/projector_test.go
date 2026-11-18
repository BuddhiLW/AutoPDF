package typst_test

import (
	"bytes"
	"context"
	"errors"
	"flag"
	imagepkg "image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/composition"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/typst"
)

var update = flag.Bool("update", false, "rewrite golden .typ files")

func compose(t *testing.T, blocks ...document.Component) composition.RenderManifest {
	t.Helper()
	catalog, err := typst.Catalog()
	if err != nil {
		t.Fatal(err)
	}
	composer, err := composition.New(catalog, typst.Target, composition.Options{MaxWorkers: 2})
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := composer.Compose(context.Background(), document.DocumentSpec{
		SchemaVersion: document.CurrentSchemaVersion, ID: "doc", Blocks: blocks,
	})
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

// workspace renders every projected file and asset binding as one golden text.
func workspace(projection latex.Projection) string {
	var text strings.Builder
	text.WriteString("main: " + projection.Main + "\n")
	for _, file := range projection.Files {
		text.WriteString("--- " + file.Path + "\n")
		text.Write(file.Content)
	}
	for _, asset := range projection.Assets {
		text.WriteString("=== asset " + asset.AssetID + " " + asset.SourceURI + " -> " + asset.Path + "\n")
	}
	return text.String()
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run with -update): %v", err)
	}
	if string(want) != got {
		t.Fatalf("%s mismatch\n--- want\n%s\n--- got\n%s", name, want, got)
	}
}

func TestBuiltinProjectionsMatchGolden(t *testing.T) {
	cases := map[string][]document.Component{
		"text": {{
			ID: "intro", Kind: "text", Variant: "default",
			Props: document.Props{"content": "Price: $5 #1 *bold* _x_ <a> @ref [b] `c` ~ = - + / \\ done"},
		}},
		"section": {{
			ID: "chapter", Kind: "section", Variant: "default",
			Props: document.Props{"title": "Results & #findings", "content": "Opening."},
			Children: []document.Component{
				{ID: "para", Kind: "text", Variant: "default", Props: document.Props{"content": "Nested _flow_."}},
			},
		}},
		"artifact": {{
			ID: "chart", Kind: "artifact", Variant: "default",
			Props:  document.Props{"content": "Figure 1"},
			Assets: []document.AssetRef{{ID: "plot", URI: "store://plot.png", MediaType: "image/png"}},
		}},
	}
	for name, blocks := range cases {
		t.Run(name, func(t *testing.T) {
			projection, err := typst.Project(compose(t, blocks...), typst.Options{})
			if err != nil {
				t.Fatal(err)
			}
			assertGolden(t, name, workspace(projection))
		})
	}
}

func TestDefinitionsCoverEveryComponentBuiltin(t *testing.T) {
	builtins, err := component.Builtins(typst.Target)
	if err != nil {
		t.Fatal(err)
	}
	definitions, err := typst.Definitions()
	if err != nil {
		t.Fatal(err)
	}
	modes := make(map[component.Key]document.CompositionMode, len(definitions))
	for _, definition := range definitions {
		modes[definition.Key()] = definition.Mode()
	}
	for _, builtin := range builtins {
		mode, ok := modes[builtin.Key()]
		if !ok {
			t.Fatalf("builtin %s has no typst projection", builtin.Key())
		}
		if mode != builtin.Mode() {
			t.Fatalf("builtin %s mode %s, typst %s", builtin.Key(), builtin.Mode(), mode)
		}
	}
}

func TestProjectRefusesComponentWithoutTypstProjectionByName(t *testing.T) {
	manifest := compose(t, document.Component{ID: "intro", Kind: "text", Variant: "default"})
	manifest.Fragments = append(manifest.Fragments, composition.Fragment{
		ComponentID: "agenda", Mode: document.Flow, Hash: "555555555555",
		Key: component.Key{Kind: "bullets", Variant: "default", Target: "beamer"},
	})
	manifest.RootOrder = append(manifest.RootOrder, "agenda")

	_, err := typst.Project(manifest, typst.Options{})
	if !errors.Is(err, typst.ErrUnsupportedComponent) {
		t.Fatalf("expected unsupported component, got %v", err)
	}
	var refused *typst.UnsupportedComponentError
	if !errors.As(err, &refused) || refused.ComponentID != "agenda" || refused.Key.Kind != "bullets" {
		t.Fatalf("refusal does not name the component: %#v", err)
	}
	if !strings.Contains(err.Error(), `"agenda"`) || !strings.Contains(err.Error(), "bullets/default@beamer") {
		t.Fatalf("refusal message does not name the component: %v", err)
	}

	custom := component.Key{Kind: "bullets", Variant: "default", Target: "beamer"}
	if _, err := typst.Project(manifest, typst.Options{Supported: append(typst.DefaultKeys(), custom)}); err != nil {
		t.Fatalf("explicitly supported key refused: %v", err)
	}
}

func TestProjectRefusesUnsupportedAssetAndNestedSection(t *testing.T) {
	pdf := compose(t, document.Component{
		ID: "doc", Kind: "artifact", Variant: "default",
		Assets: []document.AssetRef{{ID: "paper", URI: "store://paper.pdf"}},
	})
	if _, err := typst.Project(pdf, typst.Options{}); !errors.Is(err, typst.ErrUnsupportedAsset) {
		t.Fatalf("expected unsupported asset, got %v", err)
	}
	nested := compose(t, document.Component{
		ID: "outer", Kind: "text", Variant: "default",
		Children: []document.Component{{ID: "inner", Kind: "section", Variant: "default"}},
	})
	if _, err := typst.Project(nested, typst.Options{}); !errors.Is(err, typst.ErrNestedSection) {
		t.Fatalf("expected nested section, got %v", err)
	}
}

func TestProjectFocusKeepsFlowAndSelectedSections(t *testing.T) {
	manifest := compose(t,
		document.Component{ID: "lead", Kind: "text", Variant: "default", Props: document.Props{"content": "Lead"}},
		document.Component{ID: "one", Kind: "section", Variant: "default", Props: document.Props{"title": "One"}},
		document.Component{ID: "two", Kind: "section", Variant: "default", Props: document.Props{"title": "Two"}},
	)
	projection, err := typst.NewProjector(typst.Options{TrustedPreamble: []byte(`#set page(paper: "a4")`)}).
		ProjectManifest(manifest, []string{"two"})
	if err != nil {
		t.Fatal(err)
	}
	main := string(projection.Files[len(projection.Files)-1].Content)
	if projection.Files[len(projection.Files)-1].Path != "main.typ" {
		main = ""
		for _, file := range projection.Files {
			if file.Path == "main.typ" {
				main = string(file.Content)
			}
		}
	}
	if !strings.HasPrefix(main, "#set page(paper: \"a4\")\n") ||
		!strings.Contains(main, projection.SourceMap["lead"].Path) ||
		!strings.Contains(main, projection.SourceMap["two"].Path) ||
		strings.Contains(main, projection.SourceMap["one"].Path) {
		t.Fatalf("focus projection wrong:\n%s", main)
	}
	if _, err := typst.Project(manifest, typst.Options{FocusSections: []string{"lead"}}); !errors.Is(err, typst.ErrInvalidManifest) {
		t.Fatalf("expected invalid focus, got %v", err)
	}
}

func TestProjectIsDeterministicAndSanitizesPaths(t *testing.T) {
	manifest := compose(t, document.Component{
		ID: "../../ evil", Kind: "artifact", Variant: "default",
		Assets: []document.AssetRef{{ID: "../asset", URI: "../../secret.png", MediaType: "image/png"}},
	})
	first, err := typst.Project(manifest, typst.Options{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := typst.Project(manifest, typst.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) || first.Hash == "" {
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

func TestProjectedWorkspaceCompilesWithRealTypst(t *testing.T) {
	binary, err := exec.LookPath("typst")
	if err != nil {
		t.Skip("typst not on PATH")
	}
	manifest := compose(t,
		document.Component{ID: "lead", Kind: "text", Variant: "default", Props: document.Props{"content": "Cost: $5 #tag *not bold* [x] <y> @z"}},
		document.Component{
			ID: "results", Kind: "section", Variant: "default", Props: document.Props{"title": "Results #1"},
			Children: []document.Component{
				{ID: "body", Kind: "text", Variant: "default", Props: document.Props{"content": "Body_text"}},
				{ID: "chart", Kind: "artifact", Variant: "default", Assets: []document.AssetRef{{ID: "plot", URI: "store://plot.png"}}},
			},
		},
	)
	projection, err := typst.Project(manifest, typst.Options{})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, file := range projection.Files {
		writeFile(t, filepath.Join(root, file.Path), file.Content)
	}
	var image bytes.Buffer
	canvas := imagepkg.NewRGBA(imagepkg.Rect(0, 0, 4, 4))
	if err := png.Encode(&image, canvas); err != nil {
		t.Fatal(err)
	}
	for _, asset := range projection.Assets {
		writeFile(t, filepath.Join(root, asset.Path), image.Bytes())
	}
	output := filepath.Join(root, "out.pdf")
	command := exec.Command(binary, "compile", "--root", root, "--diagnostic-format", "short", filepath.Join(root, projection.Main), output)
	if combined, err := command.CombinedOutput(); err != nil {
		t.Fatalf("typst compile failed: %v\n%s", err, combined)
	}
	pdf, err := os.ReadFile(output)
	if err != nil || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("no PDF produced: %v", err)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
