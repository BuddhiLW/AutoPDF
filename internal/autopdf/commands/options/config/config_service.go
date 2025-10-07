// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common"
	resultPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/result"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// ConfigServiceCmd is the config command that handles configuration operations
var ConfigServiceCmd = &bonzai.Cmd{
	Name:    `config`,
	Alias:   `cfg`,
	Short:   `handle configuration operations`,
	Usage:   `[CONFIG_FILE] [OPTIONS...]`,
	MinArgs: 0,
	MaxArgs: 10,
	Long: `
The config command handles configuration operations for AutoPDF.

This command can be used to:
- Set configuration file paths
- Validate configuration files
- Display configuration information
- Manage configuration settings

Examples:
  autopdf config
  autopdf config config.yaml
  autopdf config config.yaml validate
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		// Create standardized logger and context
		ctx, logger := common.CreateStandardLoggerContext()
		defer logger.Sync()

		// Execute the streamlined config process
		return executeConfigProcess(ctx, args)
	},
}

// executeConfigProcess orchestrates the config process with minimal logging overhead
func executeConfigProcess(ctx context.Context, args []string) error {
	return common.ExecuteServiceOperation(
		ctx,
		"config",
		args,
		func() interface{} {
			serviceBuilder := wiring.NewServiceBuilder()
			return serviceBuilder.BuildConfigService()
		},
		func(svc interface{}, args ...string) (interface{}, error) {
			configService := svc.(*wiring.ConfigService)
			return configService.HandleConfig(args...)
		},
		func(result interface{}) error {
			resultHandler := resultPkg.NewResultHandler()
			return resultHandler.HandleConfigResult(result.(*wiring.ConfigResult))
		},
	)
}
