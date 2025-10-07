// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/testutil"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigResolver_ResolveTemplatePath(t *testing.T) {
	resolver := NewConfigResolver()

	tests := []struct {
		name         string
		cfg          *config.Config
		templateFile string
		configFile   string
		expectedPath string
		expectError  bool
	}{
		{
			name: "no template in config - use provided template",
			cfg: &config.Config{
				Template: "",
			},
			templateFile: "/absolute/path/template.tex",
			configFile:   "/absolute/path/config.yaml",
			expectedPath: "/absolute/path/template.tex",
			expectError:  false,
		},
		{
			name: "absolute template in config",
			cfg: &config.Config{
				Template: config.Template("/absolute/path/template.tex"),
			},
			templateFile: "ignored.tex",
			configFile:   "/config/dir/config.yaml",
			expectedPath: "/absolute/path/template.tex",
			expectError:  false,
		},
		{
			name: "relative template in config - resolve relative to config dir",
			cfg: &config.Config{
				Template: config.Template("./template.tex"),
			},
			templateFile: "ignored.tex",
			configFile:   "/config/dir/config.yaml",
			expectedPath: "/config/dir/template.tex",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resolver.ResolveTemplatePath(tt.cfg, tt.templateFile, tt.configFile)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPath, tt.cfg.Template.String())
			}
		})
	}
}

func TestConfigResolver_ResolveConfigFile(t *testing.T) {
	resolver := NewConfigResolver()

	tests := []struct {
		name           string
		templateFile   string
		providedConfig string
		expectedConfig string
		expectError    bool
	}{
		{
			name:           "config file provided",
			templateFile:   "template.tex",
			providedConfig: "config.yaml",
			expectedConfig: "config.yaml",
			expectError:    false,
		},
		{
			name:           "no config file provided",
			templateFile:   "template.tex",
			providedConfig: "",
			expectedConfig: "autopdf.yaml", // Default config name
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use test environment for consistent setup
			env := testutil.SetupTestEnvironment(t)
			defer env.Cleanup()

			// Store original working directory
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				env.RestoreWorkingDir(originalDir)
			}()

			// Change to test directory
			err = env.ChangeToTestDir()
			require.NoError(t, err)

			// Use test template file
			templateFile := env.TemplateFile
			if tt.providedConfig != "" {
				templateFile = tt.templateFile
			}

			configFile, err := resolver.ResolveConfigFile(templateFile, tt.providedConfig)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedConfig, configFile)
			}
		})
	}
}
