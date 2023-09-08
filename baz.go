// Copyright 2022 bonzai-example Authors
// SPDX-License-Identifier: Apache-2.0

package example

import (
	"log"
	"os/exec"

	Z "github.com/rwxrob/bonzai/z"
)

// exported leaf
var CompileCmd = &Z.Cmd{
	Name: `compile`,
	Call: func(caller *Z.Cmd, none ...string) error {
		log.Print("Baz, suncreen song")
		cmd := exec.Command("/bin/bash", "-c", "pdflatex -synctex=1 -interaction=nonstopmode -output-directory=pdfs tex-models/main.tex")
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		return nil
	},
}
