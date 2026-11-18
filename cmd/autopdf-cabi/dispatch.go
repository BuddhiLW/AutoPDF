// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Command autopdf-cabi exposes AutoPDF to hosts that load it in-process
// instead of running the autopdf CLI. Every entry point is the same call:
// an operation name and a JSON request in, a JSON envelope out.
//
//	{"ok":true,"value":...}   the operation succeeded
//	{"ok":false,"error":"…"}  it did not; nothing panics across the boundary
//
// One source tree builds three artifacts, one per kind of host:
//
//	go build -buildmode=c-shared -o libautopdf.so ./cmd/autopdf-cabi
//	    C ABI: hive_call(op, json) -> json, hive_free(ptr).
//	    JVM (Panama FFM), Rust (dlopen), C, Zig.
//	GOOS=js GOARCH=wasm go build -o autopdf.js.wasm ./cmd/autopdf-cabi
//	    globalThis.hive_autopdf(op, json) -> json. Node and browsers.
//	GOOS=wasip1 GOARCH=wasm go build -o autopdf.wasi.wasm ./cmd/autopdf-cabi
//	    WASI command: argv[1] = op, stdin = request, stdout = envelope.
//	    Any WASI runtime, cljw's wasm/run among them.
//
// "generate" runs a LaTeX engine as a child process, so it only succeeds
// where the host can start processes: the C ABI build. The other operations
// are pure and answer identically in all three.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// handler runs one operation on its decoded request.
type handler func(ctx context.Context, request json.RawMessage) (any, error)

// operation is one entry of the catalog a host can list and call.
type operation struct {
	Doc    string `json:"doc"`
	Pure   bool   `json:"pure"`
	handle handler
}

// envelope is the only shape that crosses the boundary.
type envelope struct {
	OK    bool   `json:"ok"`
	Value any    `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

// defaultTimeout bounds an operation whose request names no timeout_ms.
const defaultTimeout = 2 * time.Minute

var operations = map[string]operation{
	"capabilities": {
		Doc:    "LaTeX engines and conversion formats the embedded engine advertises.",
		Pure:   true,
		handle: capabilities,
	},
	"audit": {
		Doc:    "Parse a template and report every variable read the supplied variables leave unsatisfied.",
		Pure:   true,
		handle: audit,
	},
	"generate": {
		Doc:    "Render a template file to PDF with a LaTeX engine.",
		Pure:   false,
		handle: generate,
	},
}

// "ops" reads the table it belongs to, so it joins after initialization.
func init() {
	operations["ops"] = operation{
		Doc:    "List the operations this library answers.",
		Pure:   true,
		handle: func(context.Context, json.RawMessage) (any, error) { return catalog(), nil },
	}
}

// catalog describes every operation, sorted by name so it reads the same on
// every host.
func catalog() []map[string]any {
	names := make([]string, 0, len(operations))
	for name := range operations {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]map[string]any, 0, len(names))
	for _, name := range names {
		op := operations[name]
		out = append(out, map[string]any{"name": name, "doc": op.Doc, "pure": op.Pure})
	}
	return out
}

// dispatch runs op on request and always returns an encoded envelope.
func dispatch(op string, request []byte) []byte {
	return encode(run(op, request))
}

func run(name string, request []byte) (env envelope) {
	defer func() {
		if r := recover(); r != nil {
			env = envelope{Error: fmt.Sprintf("autopdf: %s panicked: %v", name, r)}
		}
	}()

	op, ok := operations[name]
	if !ok {
		return envelope{Error: fmt.Sprintf("autopdf: unknown operation %q", name)}
	}
	if len(request) == 0 {
		request = []byte("{}")
	}
	if !json.Valid(request) {
		return envelope{Error: fmt.Sprintf("autopdf: %s: request is not valid JSON", name)}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutOf(request))
	defer cancel()
	value, err := op.handle(ctx, request)
	if err != nil {
		return envelope{Error: fmt.Sprintf("autopdf: %s: %v", name, err)}
	}
	return envelope{OK: true, Value: value}
}

// timeoutOf reads an optional timeout_ms from any request.
func timeoutOf(request []byte) time.Duration {
	var bounded struct {
		TimeoutMS int64 `json:"timeout_ms"`
	}
	if json.Unmarshal(request, &bounded) == nil && bounded.TimeoutMS > 0 {
		return time.Duration(bounded.TimeoutMS) * time.Millisecond
	}
	return defaultTimeout
}

func encode(env envelope) []byte {
	out, err := json.Marshal(env)
	if err != nil {
		out, _ = json.Marshal(envelope{Error: fmt.Sprintf("autopdf: encoding the result: %v", err)})
	}
	return out
}
