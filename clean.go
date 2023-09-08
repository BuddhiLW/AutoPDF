// Copyright 2022 bonzai-example Pedro Gomes Branquinho
// SPDX-License-Identifier: Apache-2.0

package example

import (
	"log"
	"os/exec"

	Z "github.com/rwxrob/bonzai/z"
)

// exported leaf
var CleanCmd = &Z.Cmd{
	Name: `clean`,
	Call: func(caller *Z.Cmd, none ...string) error {
		cmd := exec.Command("/bin/bash", "-c", "rm -rf pdfs/*.log pdfs/*.aux pdfs/*.synctex.gz pdfs/*.out pdfs/*.toc pdfs/*.fls pdfs/*.fdb_latexmk")
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		return nil
	},
}
