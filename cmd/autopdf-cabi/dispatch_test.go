// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type decoded struct {
	OK    bool            `json:"ok"`
	Value json.RawMessage `json:"value"`
	Error string          `json:"error"`
}

func call(t *testing.T, op, request string) decoded {
	t.Helper()
	var d decoded
	if err := json.Unmarshal(dispatch(op, []byte(request)), &d); err != nil {
		t.Fatalf("%s: envelope is not JSON: %v", op, err)
	}
	return d
}

func TestUnknownOperationIsAnErrorEnvelope(t *testing.T) {
	d := call(t, "nope", "{}")
	if d.OK || !strings.Contains(d.Error, `unknown operation "nope"`) {
		t.Fatalf("got %+v", d)
	}
}

func TestInvalidJSONIsAnErrorEnvelope(t *testing.T) {
	d := call(t, "audit", "{not json")
	if d.OK || !strings.Contains(d.Error, "not valid JSON") {
		t.Fatalf("got %+v", d)
	}
}

func TestOpsListsEveryOperationSorted(t *testing.T) {
	d := call(t, "ops", "")
	var ops []struct {
		Name string `json:"name"`
		Pure bool   `json:"pure"`
	}
	if err := json.Unmarshal(d.Value, &ops); err != nil || !d.OK {
		t.Fatalf("got %+v, %v", d, err)
	}
	var names []string
	for _, op := range ops {
		names = append(names, op.Name)
	}
	if got := strings.Join(names, ","); got != "audit,capabilities,generate,ops" {
		t.Fatalf("names = %s", got)
	}
}

func TestCapabilitiesNamesTheDefaultEngines(t *testing.T) {
	d := call(t, "capabilities", "{}")
	if !d.OK || !strings.Contains(string(d.Value), `"pdflatex"`) {
		t.Fatalf("got %+v", d)
	}
}

func TestAuditReportsEveryUnsatisfiedRead(t *testing.T) {
	request := `{"name":"letter","template":"Dear delim[[ .name ]], you owe delim[[ .amount ]] by delim[[ .due ]].",
	             "variables":{"name":"Ada","amount":""}}`
	d := call(t, "audit", request)
	var v struct {
		Satisfied bool `json:"satisfied"`
		Offences  []struct {
			Key   string `json:"key"`
			Cause string `json:"cause"`
		} `json:"offences"`
		References []struct {
			Key string `json:"key"`
		} `json:"references"`
	}
	if err := json.Unmarshal(d.Value, &v); err != nil || !d.OK {
		t.Fatalf("got %+v, %v", d, err)
	}
	if v.Satisfied || len(v.References) != 3 || len(v.Offences) != 2 {
		t.Fatalf("got %+v", v)
	}
	if v.Offences[0].Key != "amount" || v.Offences[1].Key != "due" {
		t.Fatalf("offences = %+v", v.Offences)
	}
}

func TestAuditWithoutTemplateIsRefused(t *testing.T) {
	d := call(t, "audit", `{"template":"  "}`)
	if d.OK || !strings.Contains(d.Error, "template is required") {
		t.Fatalf("got %+v", d)
	}
}

func TestGenerateRendersAPDFAndReportsItsSize(t *testing.T) {
	if _, err := exec.LookPath("pdflatex"); err != nil {
		// generate runs the LaTeX engine as a child process; a CI image
		// without TeX cannot exercise it (card 20261007151754-603acd0b).
		t.Skip("pdflatex not installed: generate needs a LaTeX engine on PATH")
	}
	dir := t.TempDir()
	tex := filepath.Join(dir, "letter.tex")
	source := "\\documentclass{article}\\begin{document}Dear delim[[ .name ]].\\end{document}\n"
	if err := os.WriteFile(tex, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "letter.pdf")
	request, _ := json.Marshal(map[string]any{
		"template_path": tex, "output_path": out, "working_dir": dir,
		"variables": map[string]string{"name": "Ada"}, "engine": "pdflatex",
	})
	d := call(t, "generate", string(request))
	var v struct {
		PDFPath  string `json:"pdf_path"`
		FileSize int64  `json:"file_size"`
	}
	if err := json.Unmarshal(d.Value, &v); err != nil || !d.OK {
		t.Fatalf("got %+v, %v", d, err)
	}
	head, err := os.ReadFile(v.PDFPath)
	if err != nil || !strings.HasPrefix(string(head), "%PDF") {
		t.Fatalf("no PDF at %s: %v", v.PDFPath, err)
	}
	if v.FileSize != int64(len(head)) {
		t.Fatalf("file_size = %d, file has %d bytes", v.FileSize, len(head))
	}
}

func TestGenerateWithoutTemplatePathIsAnErrorNotAPanic(t *testing.T) {
	d := call(t, "generate", `{}`)
	if d.OK || d.Error == "" {
		t.Fatalf("got %+v", d)
	}
}
