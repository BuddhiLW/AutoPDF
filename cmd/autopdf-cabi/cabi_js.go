// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

//go:build js && wasm

package main

import "syscall/js"

// jsGlobal is the function the module installs on globalThis.
const jsGlobal = "hive_autopdf"

// main installs globalThis.hive_autopdf(op, json) -> json and then parks, so
// the function stays callable for as long as the host keeps the module.
func main() {
	js.Global().Set(jsGlobal, js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 2 {
			return string(encode(envelope{Error: "autopdf: hive_autopdf takes (op, request)"}))
		}
		return string(dispatch(args[0].String(), []byte(args[1].String())))
	}))
	select {}
}
