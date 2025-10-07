// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package watch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileChangeEvent_Creation(t *testing.T) {
	now := time.Now()
	event := FileChangeEvent{
		FilePath:      "/path/to/file.tex",
		Operation:     WriteOp,
		Timestamp:     now,
		ShouldRebuild: true,
	}

	assert.Equal(t, "/path/to/file.tex", event.FilePath)
	assert.Equal(t, WriteOp, event.Operation)
	assert.Equal(t, now, event.Timestamp)
	assert.True(t, event.ShouldRebuild)
}

func TestFileOperation_Constants(t *testing.T) {
	assert.Equal(t, "write", string(WriteOp))
	assert.Equal(t, "create", string(CreateOp))
	assert.Equal(t, "remove", string(RemoveOp))
	assert.Equal(t, "rename", string(RenameOp))
}

func TestWatchConfiguration_Creation(t *testing.T) {
	config := WatchConfiguration{
		TemplateFile:      "template.tex",
		ConfigFile:        "config.yaml",
		DebounceInterval:  500 * time.Millisecond,
		ExclusionPatterns: []string{"*.aux", "*.log"},
		InclusionPatterns: []string{"*.tex", "*.yaml"},
	}

	assert.Equal(t, "template.tex", config.TemplateFile)
	assert.Equal(t, "config.yaml", config.ConfigFile)
	assert.Equal(t, 500*time.Millisecond, config.DebounceInterval)
	assert.Equal(t, []string{"*.aux", "*.log"}, config.ExclusionPatterns)
	assert.Equal(t, []string{"*.tex", "*.yaml"}, config.InclusionPatterns)
}

func TestWatchConfiguration_DefaultValues(t *testing.T) {
	config := WatchConfiguration{}

	assert.Empty(t, config.TemplateFile)
	assert.Empty(t, config.ConfigFile)
	assert.Zero(t, config.DebounceInterval)
	assert.Nil(t, config.ExclusionPatterns)
	assert.Nil(t, config.InclusionPatterns)
}

func TestWatchConfiguration_WithPatterns(t *testing.T) {
	exclusions := []string{"*.aux", "*.log", "*.out"}
	inclusions := []string{"*.tex", "*.yaml", "*.yml"}

	config := WatchConfiguration{
		ExclusionPatterns: exclusions,
		InclusionPatterns: inclusions,
	}

	assert.Equal(t, exclusions, config.ExclusionPatterns)
	assert.Equal(t, inclusions, config.InclusionPatterns)
}

func TestWatchConfiguration_EdgeCases(t *testing.T) {
	// Test with empty patterns
	config := WatchConfiguration{
		ExclusionPatterns: []string{},
		InclusionPatterns: []string{},
	}

	assert.Empty(t, config.ExclusionPatterns)
	assert.Empty(t, config.InclusionPatterns)

	// Test with nil patterns
	config = WatchConfiguration{
		ExclusionPatterns: nil,
		InclusionPatterns: nil,
	}

	assert.Nil(t, config.ExclusionPatterns)
	assert.Nil(t, config.InclusionPatterns)
}

func TestFileChangeEvent_EdgeCases(t *testing.T) {
	// Test with empty file path
	event := FileChangeEvent{
		FilePath:      "",
		Operation:     WriteOp,
		ShouldRebuild: false,
	}

	assert.Empty(t, event.FilePath)
	assert.False(t, event.ShouldRebuild)

	// Test with different operations
	operations := []FileOperation{WriteOp, CreateOp, RemoveOp, RenameOp}
	for _, op := range operations {
		event := FileChangeEvent{
			FilePath:  "/test/file",
			Operation: op,
		}
		assert.Equal(t, op, event.Operation)
	}
}

func TestWatchConfiguration_Validation(t *testing.T) {
	// Test with valid configuration
	config := WatchConfiguration{
		TemplateFile:      "template.tex",
		ConfigFile:        "config.yaml",
		DebounceInterval:  100 * time.Millisecond,
		ExclusionPatterns: []string{"*.aux"},
		InclusionPatterns: []string{"*.tex"},
	}

	// All fields should be set correctly
	assert.NotEmpty(t, config.TemplateFile)
	assert.NotEmpty(t, config.ConfigFile)
	assert.Greater(t, config.DebounceInterval, time.Duration(0))
	assert.NotEmpty(t, config.ExclusionPatterns)
	assert.NotEmpty(t, config.InclusionPatterns)
}

func TestFileChangeEvent_Timestamp(t *testing.T) {
	// Test that timestamp is set correctly
	before := time.Now()
	event := FileChangeEvent{
		FilePath:  "/test/file",
		Operation: WriteOp,
		Timestamp: time.Now(),
	}
	after := time.Now()

	assert.True(t, event.Timestamp.After(before) || event.Timestamp.Equal(before))
	assert.True(t, event.Timestamp.Before(after) || event.Timestamp.Equal(after))
}

func TestWatchConfiguration_Copy(t *testing.T) {
	original := WatchConfiguration{
		TemplateFile:      "template.tex",
		ConfigFile:        "config.yaml",
		DebounceInterval:  500 * time.Millisecond,
		ExclusionPatterns: []string{"*.aux", "*.log"},
		InclusionPatterns: []string{"*.tex", "*.yaml"},
	}

	// Create a copy
	copy := original
	copy.TemplateFile = "different.tex"

	// Original should be unchanged
	assert.Equal(t, "template.tex", original.TemplateFile)
	assert.Equal(t, "different.tex", copy.TemplateFile)

	// Other fields should be the same
	assert.Equal(t, original.ConfigFile, copy.ConfigFile)
	assert.Equal(t, original.DebounceInterval, copy.DebounceInterval)
}
