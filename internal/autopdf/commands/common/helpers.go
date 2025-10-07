// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/configs"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/rwxrob/bonzai"
)

// ExecuteServiceOperation is a generic helper for executing service operations
func ExecuteServiceOperation(
	ctx context.Context,
	operationName string,
	args []string,
	serviceBuilder func() interface{},
	handler func(interface{}, ...string) (interface{}, error),
	resultHandler func(interface{}) error,
) error {
	logger := configs.GetLoggerFromContext(ctx)
	logger.InfoWithFields("Handling operation", "operation", operationName, "args", args)

	// Build service
	svc := serviceBuilder()

	// Execute operation
	result, err := handler(svc, args...)
	if err != nil {
		logger.ErrorWithFields("Operation failed", "operation", operationName, "error", err)
		return fmt.Errorf("%s: %w", operationName, err)
	}

	// Handle result
	return resultHandler(result)
}

// HandleSubcommandDelegation manages subcommand delegation using Bonzai Cmd structs
func HandleSubcommandDelegation(
	ctx context.Context,
	subcommand string,
	remainingArgs []string,
	availableCommands map[string]*bonzai.Cmd,
) error {
	logger := configs.GetLoggerFromContext(ctx)

	// Check if the subcommand exists in available commands
	cmd, exists := availableCommands[subcommand]
	if !exists {
		logger.WarnWithFields("Unknown subcommand", "subcommand", subcommand, "available", getAvailableCommands(availableCommands))
		return fmt.Errorf("%w: %s", configs.UnknownSubcommandError, subcommand)
	}

	// Log the delegation
	logger.InfoWithFields("Delegating to subcommand", "subcommand", subcommand, "args", remainingArgs)

	// Execute the command using Bonzai's Do method
	return cmd.Do(cmd, remainingArgs...)
}

// getAvailableCommands returns a list of available command names
func getAvailableCommands(commands map[string]*bonzai.Cmd) []string {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	return names
}

// CreateCommandMap creates a map of available commands for delegation
func CreateCommandMap(commands ...*bonzai.Cmd) map[string]*bonzai.Cmd {
	commandMap := make(map[string]*bonzai.Cmd)
	for _, cmd := range commands {
		if cmd != nil {
			commandMap[cmd.Name] = cmd
		}
	}
	return commandMap
}

// CreateStandardLoggerContext creates a standardized logger context
func CreateStandardLoggerContext() (context.Context, *logger.LoggerAdapter) {
	return configs.CreateLoggerContext()
}

// LogOperationStart logs the start of an operation
func LogOperationStart(logger *logger.LoggerAdapter, operation string, args []string) {
	logger.InfoWithFields("Starting operation", "operation", operation, "args", args)
}

// LogOperationSuccess logs the successful completion of an operation
func LogOperationSuccess(logger *logger.LoggerAdapter, operation string, result interface{}) {
	logger.InfoWithFields("Operation completed successfully", "operation", operation, "result", result)
}

// LogOperationError logs an operation error
func LogOperationError(logger *logger.LoggerAdapter, operation string, err error) {
	logger.ErrorWithFields("Operation failed", "operation", operation, "error", err)
}
