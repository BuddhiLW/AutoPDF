// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"context"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common"
	argsPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/args"
	resultPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/result"
	wiringPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// ConvertServiceCmd is the thin CLI layer that delegates to the converter service
var ConvertServiceCmd = &bonzai.Cmd{
	Name:    `convert`,
	Alias:   `c`,
	Short:   `convert PDF to images`,
	Usage:   `PDF [FORMAT...]`,
	MinArgs: 1,
	MaxArgs: 10, // Allow up to 10 formats
	Long: `
The convert command takes a PDF file and converts it to one or more image formats.

Supported formats: png, jpeg, jpg, gif, bmp, tiff, webp

Examples:
  autopdf convert document.pdf
  autopdf convert document.pdf png
  autopdf convert document.pdf png jpeg
  autopdf convert document.pdf png jpeg gif
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		// Create standardized logger and context
		ctx, logger := common.CreateStandardLoggerContext()
		defer logger.Sync()

		// Execute the streamlined convert process
		return executeConvertProcess(ctx, args)
	},
}

// executeConvertProcess orchestrates the convert process with minimal logging overhead
func executeConvertProcess(ctx context.Context, args []string) error {
	// Parse arguments with logging
	argsParser := argsPkg.NewArgsParser()
	convertArgs, err := argsParser.ParseConvertArgsWithLogging(ctx, args)
	if err != nil {
		return err
	}

	// Build and execute conversion with logging
	serviceBuilder := wiringPkg.NewConvertServiceBuilder()
	svc := serviceBuilder.BuildConverterService(convertArgs)

	imageFiles, err := svc.ConvertToImages(ctx, convertArgs.PDFFile, convertArgs.Formats)
	if err != nil {
		return fmt.Errorf("PDF to image conversion failed: %w", err)
	}

	// Handle result
	resultHandler := resultPkg.NewConvertResultHandler()
	return resultHandler.HandleConvertResult(imageFiles)
}
