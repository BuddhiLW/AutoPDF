// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package valueobjects

import (
	"errors"
	"strings"
)

// DebugConfig encapsulates debug configuration
type DebugConfig struct {
	enabled     bool
	concreteDir string
	logDir      string
}

// NewDebugConfig creates a new debug configuration
func NewDebugConfig(enabled bool, concreteDir, logDir string) (DebugConfig, error) {
	if !enabled {
		return DebugConfig{enabled: false}, nil
	}

	if strings.TrimSpace(concreteDir) == "" {
		return DebugConfig{}, errors.New("concrete directory cannot be empty when debug is enabled")
	}

	if strings.TrimSpace(logDir) == "" {
		return DebugConfig{}, errors.New("log directory cannot be empty when debug is enabled")
	}

	return DebugConfig{
		enabled:     enabled,
		concreteDir: strings.TrimSpace(concreteDir),
		logDir:      strings.TrimSpace(logDir),
	}, nil
}

// IsEnabled returns true if debug mode is enabled
func (d DebugConfig) IsEnabled() bool {
	return d.enabled
}

// ConcreteDirectory returns the concrete file directory
func (d DebugConfig) ConcreteDirectory() string {
	return d.concreteDir
}

// LogDirectory returns the log directory
func (d DebugConfig) LogDirectory() string {
	return d.logDir
}
