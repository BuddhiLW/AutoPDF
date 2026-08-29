package document_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/BuddhiLW/AutoPDF/pkg/document"
)

func TestDocumentSpecRoundTripAndCloneOwnership(t *testing.T) {
	spec := document.DocumentSpec{
		SchemaVersion: document.CurrentSchemaVersion,
		ID:            "case-42",
		Theme:         "legal/default@1",
		Blocks: []document.Component{{
			ID:      "photo",
			Kind:    "image",
			Variant: "portrait",
			Props:   document.Props{"aspectRatio": "3:4", "tags": []any{"client", "primary"}},
			Style:   document.Style{"radius": "8pt"},
			Assets:  []document.AssetRef{{ID: "portrait", URI: "asset://photo-1", MediaType: "image/jpeg"}},
		}},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := document.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	clone, err := decoded.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, clone) {
		t.Fatalf("clone mismatch:\nwant %#v\n got %#v", decoded, clone)
	}

	clone.Blocks[0].Style["radius"] = "0pt"
	clone.Blocks[0].Props["tags"].([]any)[0] = "changed"
	if decoded.Blocks[0].Style["radius"] != "8pt" {
		t.Fatal("clone borrowed style map")
	}
	if decoded.Blocks[0].Props["tags"].([]any)[0] != "client" {
		t.Fatal("clone borrowed nested props")
	}
}

func TestValidationReportsStableFieldPaths(t *testing.T) {
	spec := document.DocumentSpec{
		SchemaVersion: 99,
		Blocks: []document.Component{
			{ID: "same", Kind: "paragraph"},
			{ID: "same", Variant: "default", Assets: []document.AssetRef{{ID: "x"}, {ID: "x"}}},
		},
	}
	problems := spec.Problems()
	if len(problems) != 5 {
		t.Fatalf("expected 5 problems, got %d: %v", len(problems), problems)
	}
	codes := make([]string, len(problems))
	for index, problem := range problems {
		codes[index] = problem.Code
	}
	want := []string{
		"document.schema_version.unsupported",
		"component.variant.required",
		"component.id.duplicate",
		"component.kind.required",
		"asset.id.duplicate",
	}
	if !reflect.DeepEqual(codes, want) {
		t.Fatalf("codes mismatch: want %v, got %v", want, codes)
	}
}

func TestCompositionModes(t *testing.T) {
	for _, mode := range []document.CompositionMode{document.Flow, document.Section, document.Artifact} {
		if !mode.Valid() {
			t.Fatalf("expected %q to be valid", mode)
		}
	}
	if document.CompositionMode("guessed").Valid() {
		t.Fatal("unknown mode must be invalid")
	}
}

func TestDocumentJSONRoundTripProperty(t *testing.T) {
	property := func(raw string) bool {
		id := strings.Map(func(r rune) rune {
			if r < 32 {
				return -1
			}
			return r
		}, raw)
		if id == "" {
			id = "component"
		}
		spec := document.DocumentSpec{
			SchemaVersion: document.CurrentSchemaVersion,
			Blocks: []document.Component{{
				ID: id, Kind: "text", Variant: "default",
				Props: document.Props{"value": raw},
			}},
		}
		data, err := json.Marshal(spec)
		if err != nil {
			return false
		}
		decoded, err := document.Decode(data)
		return err == nil && reflect.DeepEqual(spec, decoded)
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeRejectsUnknownFieldsAndTrailingValues(t *testing.T) {
	for _, input := range []string{
		`{"schemaVersion":1,"blocks":[],"surprise":true}`,
		`{"schemaVersion":1,"blocks":[]} {"schemaVersion":1,"blocks":[]}`,
	} {
		if _, err := document.Decode([]byte(input)); err == nil {
			t.Fatalf("expected decode failure for %s", input)
		}
	}
}
