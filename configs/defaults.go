package configs

import (
	"context"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters"
)

var DefaultConfigName = "autopdf.yaml"

// Standardized error types
var (
	ConfigFileExistsError = fmt.Errorf("config file already exists")
	ConfigFileWriteError  = fmt.Errorf("failed to write config file")
	BuildError            = fmt.Errorf("failed to build config")
	WriteError            = fmt.Errorf("failed to write config")
	ReadError             = fmt.Errorf("failed to read config")
	ParseError            = fmt.Errorf("failed to parse config")
	TemplateError         = fmt.Errorf("failed to process template")
	ConversionError       = fmt.Errorf("failed to convert PDF to images")
	CleanError            = fmt.Errorf("failed to clean up auxiliary files")

	// Command-specific errors
	UnknownSubcommandError = fmt.Errorf("unknown subcommand")
	ConfigOperationError   = fmt.Errorf("config operation failed")
	ConvertOperationError  = fmt.Errorf("convert operation failed")
	BuildOperationError    = fmt.Errorf("build operation failed")
)

// Context key type for logger
type ContextKey string

const LoggerKey ContextKey = "logger"

// GetLoggerFromContext extracts logger from context with fallback
func GetLoggerFromContext(ctx context.Context) *adapters.LoggerAdapter {
	if logger, ok := ctx.Value(LoggerKey).(*adapters.LoggerAdapter); ok {
		return logger
	}
	// Fallback to default logger
	return adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
}

// CreateLoggerContext creates a new context with logger
func CreateLoggerContext() (context.Context, *adapters.LoggerAdapter) {
	logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
	ctx := context.WithValue(context.Background(), LoggerKey, logger)
	return ctx, logger
}
