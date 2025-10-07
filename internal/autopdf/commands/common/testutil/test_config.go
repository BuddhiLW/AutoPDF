// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"os"
	"path/filepath"
)

// TestConfig represents test configuration settings
type TestConfig struct {
	UseTestDataDir bool
	TestDataDir    string
	CleanupAfter   bool
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		UseTestDataDir: true,
		TestDataDir:    "internal/autopdf/test-data",
		CleanupAfter:   true,
	}
}

// GetTestDataDir returns the test data directory path
func (tc *TestConfig) GetTestDataDir() string {
	if tc.UseTestDataDir {
		// Try to find the test-data directory relative to current working directory
		wd, err := os.Getwd()
		if err == nil {
			testDataPath := filepath.Join(wd, tc.TestDataDir)
			if _, err := os.Stat(testDataPath); err == nil {
				return testDataPath
			}
		}
	}
	return ""
}

// ShouldCleanup returns whether cleanup should be performed
func (tc *TestConfig) ShouldCleanup() bool {
	return tc.CleanupAfter
}
