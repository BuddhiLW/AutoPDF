package config

import (
	"os"
	"strconv"
)

// APIDebugConfig holds configuration for API debug settings
type APIDebugConfig struct {
	Enabled         bool   // AUTOPDF_API_DEBUG=true
	LogDirectory    string // AUTOPDF_API_LOG_DIR=/var/log/autopdf
	ConcreteFileDir string // AUTOPDF_API_CONCRETE_DIR=/tmp/autopdf
	DefaultVerbose  int    // AUTOPDF_API_DEFAULT_VERBOSE=1
}

// LoadDebugConfigFromEnv loads debug configuration from environment variables
func LoadDebugConfigFromEnv() *APIDebugConfig {
	return &APIDebugConfig{
		Enabled:         getEnvBool("AUTOPDF_API_DEBUG", false),
		LogDirectory:    getEnvOrDefault("AUTOPDF_API_LOG_DIR", "/tmp/autopdf/logs"),
		ConcreteFileDir: getEnvOrDefault("AUTOPDF_API_CONCRETE_DIR", "/tmp/autopdf/concrete"),
		DefaultVerbose:  getEnvInt("AUTOPDF_API_DEFAULT_VERBOSE", 1),
	}
}

// IsDebugEnabled returns true if debug mode is enabled
func (c *APIDebugConfig) IsDebugEnabled() bool {
	return c.Enabled
}

// GetLogDirectory returns the configured log directory
func (c *APIDebugConfig) GetLogDirectory() string {
	return c.LogDirectory
}

// GetConcreteFileDirectory returns the configured concrete file directory
func (c *APIDebugConfig) GetConcreteFileDirectory() string {
	return c.ConcreteFileDir
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns the environment variable as a boolean or a default
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvInt returns the environment variable as an integer or a default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
