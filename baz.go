// Copyright 2022 bonzai-example Authors
// SPDX-License-Identifier: Apache-2.0

package example

import (
	"log"

	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/help"
)

// exported leaf
var BazCmd = &Z.Cmd{
	Name: `baz`,

	// Aliases are not commands but will be replaced by their target names
	// during bash tab completion. Aliases show up in the COMMANDS section
	// of help, but do not display during tab completion so as to keep the
	// list shorter.
	Aliases: []string{"Bz", "notbaz"},

	// Commands are the main way to compose other commands into your
	// branch. When in doubt, add a command, even if it is in the same
	// file.
	Commands: []*Z.Cmd{help.Cmd, fileCmd},

	// Call first-class functions can be highly detailed, refer to an
	// existing function someplace else, or can call high-level package
	// library functions. Developers are encouraged to consider well where
	// they maintain the core logic of their applications. Often, it will
	// not be here within the Z.Cmd definition. One use case for
	// decoupled first-class Call functions is when creating multiple
	// binaries for different target languages. In such cases this
	// Z.Cmd definition is essentially just a wrapper for
	// documentation and other language-specific embedded assets.
	Call: func(caller *Z.Cmd, none ...string) error {
		log.Print("Baz, suncreen song")
		return nil
	},
}
