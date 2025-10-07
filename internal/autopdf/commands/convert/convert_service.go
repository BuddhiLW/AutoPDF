// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package convert

import (
	"context"
	"fmt"
	"log"

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
		log.Println("Converting PDF to images using service...")

		// Parse arguments
		argsParser := argsPkg.NewConvertArgsParser()
		convertArgs, err := argsParser.ParseConvertArgs(args)
		if err != nil {
			return err
		}

		// Build the converter service
		serviceBuilder := wiringPkg.NewConvertServiceBuilder()
		svc := serviceBuilder.BuildConverterService(convertArgs)

		// Execute the conversion
		ctx := context.Background()
		imageFiles, err := svc.ConvertToImages(ctx, convertArgs.PDFFile, convertArgs.Formats)
		if err != nil {
			log.Printf("Error converting PDF: %s", err)
			return fmt.Errorf("PDF to image conversion failed: %w", err)
		}

		// Handle the result
		resultHandler := resultPkg.NewConvertResultHandler()
		return resultHandler.HandleConvertResult(imageFiles)
	},
}
