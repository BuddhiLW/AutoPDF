// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBuildOptions(t *testing.T) {
	options := NewBuildOptions()

	// Test default values
	assert.False(t, options.Clean.Enabled)
	assert.Equal(t, ".", options.Clean.Target)

	assert.False(t, options.Verbose.Enabled)
	assert.Equal(t, 1, options.Verbose.Level)

	assert.False(t, options.Debug.Enabled)
	assert.Equal(t, "stdout", options.Debug.Output)

	assert.False(t, options.Force.Enabled)
	assert.False(t, options.Force.Overwrite)
}

func TestBuildOptions_EnableClean(t *testing.T) {
	options := NewBuildOptions()

	options.EnableClean("/tmp")

	assert.True(t, options.Clean.Enabled)
	assert.Equal(t, "/tmp", options.Clean.Target)
}

func TestBuildOptions_EnableVerbose(t *testing.T) {
	options := NewBuildOptions()

	options.EnableVerbose(3)

	assert.True(t, options.Verbose.Enabled)
	assert.Equal(t, 3, options.Verbose.Level)
}

func TestBuildOptions_EnableDebug(t *testing.T) {
	options := NewBuildOptions()

	options.EnableDebug("debug.log")

	assert.True(t, options.Debug.Enabled)
	assert.Equal(t, "debug.log", options.Debug.Output)
}

func TestBuildOptions_EnableForce(t *testing.T) {
	options := NewBuildOptions()

	options.EnableForce(true)

	assert.True(t, options.Force.Enabled)
	assert.True(t, options.Force.Overwrite)
}

func TestBuildOptions_HasAnyEnabled(t *testing.T) {
	options := NewBuildOptions()

	// Initially no options enabled
	assert.False(t, options.HasAnyEnabled())

	// Enable one option
	options.EnableClean(".")
	assert.True(t, options.HasAnyEnabled())

	// Reset and enable another
	options = NewBuildOptions()
	options.EnableVerbose(2)
	assert.True(t, options.HasAnyEnabled())
}

func TestBuildOptions_GetEnabledOptions(t *testing.T) {
	options := NewBuildOptions()

	// Initially no options enabled
	enabled := options.GetEnabledOptions()
	assert.Empty(t, enabled)

	// Enable multiple options
	options.EnableClean(".")
	options.EnableVerbose(2)
	options.EnableDebug("debug.log")
	options.EnableForce(true)

	enabled = options.GetEnabledOptions()
	assert.Len(t, enabled, 4)
	assert.Contains(t, enabled, "clean")
	assert.Contains(t, enabled, "verbose")
	assert.Contains(t, enabled, "debug")
	assert.Contains(t, enabled, "force")
}

func TestBuildOptions_ComplexScenario(t *testing.T) {
	options := NewBuildOptions()

	// Enable clean with specific target
	options.EnableClean("/tmp/build")
	assert.True(t, options.Clean.Enabled)
	assert.Equal(t, "/tmp/build", options.Clean.Target)

	// Enable verbose with high level
	options.EnableVerbose(3)
	assert.True(t, options.Verbose.Enabled)
	assert.Equal(t, 3, options.Verbose.Level)

	// Enable debug to file
	options.EnableDebug("/tmp/debug.log")
	assert.True(t, options.Debug.Enabled)
	assert.Equal(t, "/tmp/debug.log", options.Debug.Output)

	// Enable force with overwrite
	options.EnableForce(true)
	assert.True(t, options.Force.Enabled)
	assert.True(t, options.Force.Overwrite)

	// Verify all are enabled
	assert.True(t, options.HasAnyEnabled())
	enabled := options.GetEnabledOptions()
	assert.Len(t, enabled, 4)
}

func TestBuildOptions_EdgeCases(t *testing.T) {
	options := NewBuildOptions()

	// Test empty target for clean
	options.EnableClean("")
	assert.True(t, options.Clean.Enabled)
	assert.Equal(t, "", options.Clean.Target)

	// Test zero level for verbose
	options.EnableVerbose(0)
	assert.True(t, options.Verbose.Enabled)
	assert.Equal(t, 0, options.Verbose.Level)

	// Test empty output for debug
	options.EnableDebug("")
	assert.True(t, options.Debug.Enabled)
	assert.Equal(t, "", options.Debug.Output)

	// Test force with false overwrite
	options.EnableForce(false)
	assert.True(t, options.Force.Enabled)
	assert.False(t, options.Force.Overwrite)
}

func TestBuildOptions_DisableOptions(t *testing.T) {
	options := NewBuildOptions()

	// Enable all options
	options.EnableClean(".")
	options.EnableVerbose(2)
	options.EnableDebug("debug.log")
	options.EnableForce(true)

	assert.True(t, options.HasAnyEnabled())
	assert.Len(t, options.GetEnabledOptions(), 4)

	// Create new instance to "disable" all
	options = NewBuildOptions()
	assert.False(t, options.HasAnyEnabled())
	assert.Empty(t, options.GetEnabledOptions())
}
