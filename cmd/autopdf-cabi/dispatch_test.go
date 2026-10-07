// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
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

func TestGenerateWithoutTemplatePathIsAnErrorNotAPanic(t *testing.T) {
	d := call(t, "generate", `{}`)
	if d.OK || d.Error == "" {
		t.Fatalf("got %+v", d)
	}
}
