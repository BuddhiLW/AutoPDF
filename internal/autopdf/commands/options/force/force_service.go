// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package force

import (
	"fmt"
	"log"
	"strconv"

	resultPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/result"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// ForceServiceCmd is the force command that enables force operations
var ForceServiceCmd = &bonzai.Cmd{
	Name:    `force`,
	Alias:   `f`,
	Short:   `enable force operations`,
	Usage:   `[ENABLED]`,
	MinArgs: 0,
	MaxArgs: 1,
	Long: `
The force command enables force operations for AutoPDF.
This allows operations to overwrite existing files and bypass certain safety checks.

Force operations include:
- Overwriting existing PDF files
- Overwriting existing image files
- Bypassing file existence checks
- Continuing operations despite warnings

Examples:
  autopdf force
  autopdf force true
  autopdf force false
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		log.Println("Setting force operations...")

		// Default to true (force enabled)
		enabled := true
		if len(args) > 0 {
			parsedEnabled, err := strconv.ParseBool(args[0])
			if err != nil {
				return fmt.Errorf("invalid force value '%s': must be true or false", args[0])
			}
			enabled = parsedEnabled
		}

		// Build the force service
		serviceBuilder := wiring.NewServiceBuilder()
		forceService := serviceBuilder.BuildForceService(enabled)

		// Execute the force operation
		result, err := forceService.SetForceMode()
		if err != nil {
			log.Printf("Error setting force mode: %s", err)
			return err
		}

		// Handle the result
		resultHandler := resultPkg.NewResultHandler()
		return resultHandler.HandleForceResult(result)
	},
}
