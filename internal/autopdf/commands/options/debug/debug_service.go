// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package debug

import (
	"fmt"
	"log"

	resultPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/result"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// DebugServiceCmd is the debug command that enables debug information output
var DebugServiceCmd = &bonzai.Cmd{
	Name:    `debug`,
	Alias:   `d`,
	Short:   `enable debug information output`,
	Usage:   `[OUTPUT]`,
	MinArgs: 0,
	MaxArgs: 1,
	Long: `
The debug command enables debug information output for AutoPDF operations.
This provides detailed diagnostic information useful for troubleshooting.

Output options:
  stdout - Output debug info to standard output (default)
  stderr - Output debug info to standard error
  file   - Output debug info to a log file
  both   - Output debug info to both stdout and file

Examples:
  autopdf debug
  autopdf debug stdout
  autopdf debug file
  autopdf debug both
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		log.Println("Enabling debug information output...")

		// Default to stdout
		output := "stdout"
		if len(args) > 0 {
			output = args[0]
		}

		// Validate output option
		validOutputs := map[string]bool{
			"stdout": true,
			"stderr": true,
			"file":   true,
			"both":   true,
		}
		if !validOutputs[output] {
			return fmt.Errorf("invalid debug output '%s': must be one of stdout, stderr, file, both", output)
		}

		// Build the debug service
		serviceBuilder := wiring.NewServiceBuilder()
		debugService := serviceBuilder.BuildDebugService(output)

		// Execute the debug operation
		result, err := debugService.EnableDebug()
		if err != nil {
			log.Printf("Error enabling debug: %s", err)
			return err
		}

		// Handle the result
		resultHandler := resultPkg.NewResultHandler()
		return resultHandler.HandleDebugResult(result)
	},
}
