// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package args

import (
	"context"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain"
)

// BuildArgs represents the parsed arguments for the build command
type BuildArgs struct {
	TemplateFile  string
	ConfigFile    string
	Options       domain.BuildOptions
	RemainingArgs []string
}

// ArgsParser handles parsing of command line arguments
type ArgsParser struct{}

// NewArgsParser creates a new argument parser
func NewArgsParser() *ArgsParser {
	return &ArgsParser{}
}

// ParseBuildArgs parses the build command arguments
func (ap *ArgsParser) ParseBuildArgs(args []string) (*BuildArgs, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("template file is required")
	}

	buildArgs := &BuildArgs{
		TemplateFile:  args[0],
		ConfigFile:    "",                       // Will be resolved by ConfigResolver
		Options:       domain.NewBuildOptions(), // Initialize with default values
		RemainingArgs: []string{},
	}

	// Parse arguments starting from index 1
	for i := 1; i < len(args); i++ {
		arg := args[i]

		// Check if it's a config file (not an option)
		if !ap.isOption(arg) {
			// This could be a config file, but validate it first
			if ap.isValidConfigFile(arg) {
				buildArgs.ConfigFile = arg
				continue
			} else {
				// Not a valid config file and not an option, treat as invalid
				return nil, fmt.Errorf("invalid argument '%s': not a valid config file or option", arg)
			}
		}

		// Parse the option
		option, err := ap.parseOption(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid option '%s': %w", arg, err)
		}

		// Set the option
		ap.setOption(&buildArgs.Options, option)
	}

	return buildArgs, nil
}

// ParseBuildArgsWithDelegation parses build arguments and handles delegation to other commands
func (ap *ArgsParser) ParseBuildArgsWithDelegation(args []string) (*BuildArgs, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("template file is required")
	}

	buildArgs := &BuildArgs{
		TemplateFile:  args[0],
		ConfigFile:    "",                       // Will be resolved by ConfigResolver
		Options:       domain.NewBuildOptions(), // Initialize with default values
		RemainingArgs: []string{},
	}

	// Parse arguments starting from index 1
	for i := 1; i < len(args); i++ {
		arg := args[i]

		// Check if this is a subcommand (like "convert")
		if ap.isSubcommand(arg) {
			// Everything from this point on is for the subcommand
			buildArgs.RemainingArgs = args[i:]
			break
		}

		// Check if it's a config file (not an option)
		if !ap.isOption(arg) {
			// This could be a config file, but validate it first
			if ap.isValidConfigFile(arg) {
				buildArgs.ConfigFile = arg
				continue
			} else {
				// Not a valid config file and not an option, treat as invalid
				return nil, fmt.Errorf("invalid argument '%s': not a valid config file or option", arg)
			}
		}

		// Parse the option
		option, err := ap.parseOption(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid option '%s': %w", arg, err)
		}

		// Set the option
		ap.setOption(&buildArgs.Options, option)
	}

	return buildArgs, nil
}

// GetRemainingArgs returns the remaining arguments for delegation
func (ba *BuildArgs) GetRemainingArgs() []string {
	return ba.RemainingArgs
}

// isSubcommand checks if an argument is a subcommand
func (ap *ArgsParser) isSubcommand(arg string) bool {
	knownSubcommands := map[string]bool{
		"convert": true,
		"config":  true,
	}
	return knownSubcommands[arg]
}

// ParseBuildArgsWithLogging parses build arguments with integrated logging
func (ap *ArgsParser) ParseBuildArgsWithLogging(ctx context.Context, args []string) (*BuildArgs, error) {
	logger := getLoggerFromContext(ctx)

	logger.Debug("Parsing command line arguments")
	buildArgs, err := ap.ParseBuildArgsWithDelegation(args)
	if err != nil {
		logger.ErrorWithFields("Failed to parse arguments", "error", err)
		return nil, err
	}

	logger.InfoWithFields("Arguments parsed successfully",
		"template_file", buildArgs.TemplateFile,
		"config_file", buildArgs.ConfigFile,
		"options", buildArgs.Options,
	)

	return buildArgs, nil
}

// ParseConvertArgsWithLogging parses convert arguments with integrated logging
func (ap *ArgsParser) ParseConvertArgsWithLogging(ctx context.Context, args []string) (*ConvertArgs, error) {
	logger := getLoggerFromContext(ctx)

	logger.Debug("Parsing convert arguments")
	convertParser := NewConvertArgsParser()
	convertArgs, err := convertParser.ParseConvertArgs(args)
	if err != nil {
		logger.ErrorWithFields("Failed to parse convert arguments", "error", err)
		return nil, err
	}

	logger.InfoWithFields("Convert arguments parsed successfully",
		"pdf_file", convertArgs.PDFFile,
		"formats", convertArgs.Formats,
	)

	return convertArgs, nil
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const loggerKey contextKey = "logger"

// getLoggerFromContext extracts logger from context
func getLoggerFromContext(ctx context.Context) *adapters.LoggerAdapter {
	if logger, ok := ctx.Value(loggerKey).(*adapters.LoggerAdapter); ok {
		return logger
	}
	// Fallback to default logger
	return adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
}

// isOption checks if an argument is an option (starts with known option names)
func (ap *ArgsParser) isOption(arg string) bool {
	knownOptions := map[string]bool{
		"clean":   true,
		"verbose": true,
		"debug":   true,
		"force":   true,
	}
	return knownOptions[arg]
}

// isValidConfigFile checks if an argument looks like a valid config file
func (ap *ArgsParser) isValidConfigFile(arg string) bool {
	// Basic validation: should have a file extension
	if len(arg) < 4 {
		return false
	}

	// Should end with common config file extensions
	validExtensions := []string{".yaml", ".yml", ".json", ".toml"}
	for _, ext := range validExtensions {
		if len(arg) >= len(ext) && arg[len(arg)-len(ext):] == ext {
			return true
		}
	}

	return false
}

// parseOption parses a single option string
func (ap *ArgsParser) parseOption(option string) (string, error) {
	// For now, we only support simple boolean options
	// Future: could support key=value options like "verbose=2"
	return option, nil
}

// setOption sets the appropriate option in BuildOptions
func (ap *ArgsParser) setOption(options *domain.BuildOptions, option string) {
	switch option {
	case "clean":
		options.EnableClean(".") // Default to current directory
	case "verbose":
		options.EnableVerbose(2) // Default to level 2 (verbose)
	case "debug":
		options.EnableDebug("stdout") // Default to stdout
	case "force":
		options.EnableForce(true) // Default to overwrite enabled
	}
}
