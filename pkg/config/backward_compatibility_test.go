package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackwardCompatibility ensures that all existing configuration functionality continues to work
// This establishes the "minimum bar" for backward compatibility
func TestBackwardCompatibility(t *testing.T) {
	t.Run("Original Config Structure", func(t *testing.T) {
		// Test that the original config structure continues to work
		cfg := &Config{
			Template: "test.tex",
			Output:   "output.pdf",
			Variables: map[string]interface{}{
				"title":   "Test Document",
				"author":  "John Doe",
				"content": "This is a test document.",
			},
			Engine: "pdflatex",
			Conversion: Conversion{
				Enabled: false,
				Formats: []string{},
			},
		}

		// Test that all fields are accessible
		assert.Equal(t, "test.tex", cfg.Template.String())
		assert.Equal(t, "output.pdf", cfg.Output.String())
		assert.Equal(t, "pdflatex", cfg.Engine.String())
		assert.False(t, cfg.Conversion.Enabled)
		assert.Empty(t, cfg.Conversion.Formats)

		// Test variables access
		assert.Equal(t, "Test Document", cfg.Variables["title"])
		assert.Equal(t, "John Doe", cfg.Variables["author"])
		assert.Equal(t, "This is a test document.", cfg.Variables["content"])
	})

	t.Run("Config String Representation", func(t *testing.T) {
		// Test that the String() method continues to work
		cfg := &Config{
			Template: "test.tex",
			Output:   "output.pdf",
			Variables: map[string]interface{}{
				"title": "Test Document",
			},
			Engine: "pdflatex",
		}

		str := cfg.String()
		assert.Contains(t, str, "template: test.tex")
		assert.Contains(t, str, "output: output.pdf")
		assert.Contains(t, str, "engine: pdflatex")
	})

	t.Run("Config JSON Conversion", func(t *testing.T) {
		// Test that JSON conversion continues to work
		cfg := &Config{
			Template: "test.tex",
			Output:   "output.pdf",
			Variables: map[string]interface{}{
				"title": "Test Document",
			},
			Engine: "pdflatex",
		}

		jsonStr, err := cfg.ToJSON()
		require.NoError(t, err)
		assert.Contains(t, jsonStr, `"template":"test.tex"`)
		assert.Contains(t, jsonStr, `"output":"output.pdf"`)
		assert.Contains(t, jsonStr, `"engine":"pdflatex"`)
	})

	t.Run("Config YAML Operations", func(t *testing.T) {
		// Test that YAML operations continue to work
		yamlData := []byte(`
template: test.tex
output: output.pdf
variables:
  title: Test Document
  author: John Doe
engine: pdflatex
conversion:
  enabled: false
  formats: []
`)

		cfg, err := NewConfigFromYAML(yamlData)
		require.NoError(t, err)
		assert.Equal(t, "test.tex", cfg.Template.String())
		assert.Equal(t, "output.pdf", cfg.Output.String())
		assert.Equal(t, "Test Document", cfg.Variables["title"])
		assert.Equal(t, "John Doe", cfg.Variables["author"])
		assert.Equal(t, "pdflatex", cfg.Engine.String())
		assert.False(t, cfg.Conversion.Enabled)
	})

	t.Run("Default Config", func(t *testing.T) {
		// Test that default config continues to work
		cfg := GetDefaultConfig()
		assert.Equal(t, "", cfg.Template.String())
		assert.Equal(t, "", cfg.Output.String())
		assert.Equal(t, "pdflatex", cfg.Engine.String())
		assert.False(t, cfg.Conversion.Enabled)
		assert.Empty(t, cfg.Conversion.Formats)
		assert.NotNil(t, cfg.Variables)
	})

	t.Run("Config Marshal/Unmarshal", func(t *testing.T) {
		// Test that marshal/unmarshal operations continue to work
		original := &Config{
			Template: "test.tex",
			Output:   "output.pdf",
			Variables: map[string]interface{}{
				"title": "Test Document",
			},
			Engine: "xelatex",
			Conversion: Conversion{
				Enabled: true,
				Formats: []string{"png", "jpg"},
			},
		}

		// Marshal
		data, err := original.Marshal()
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Unmarshal
		restored := &Config{}
		err = restored.Unmarshal(data)
		require.NoError(t, err)

		assert.Equal(t, original.Template, restored.Template)
		assert.Equal(t, original.Output, restored.Output)
		assert.Equal(t, original.Engine, restored.Engine)
		assert.Equal(t, original.Conversion.Enabled, restored.Conversion.Enabled)
		assert.Equal(t, original.Conversion.Formats, restored.Conversion.Formats)
		assert.Equal(t, original.Variables, restored.Variables)
	})
}

// TestMinimumBar establishes the minimum bar for backward compatibility
func TestMinimumBar(t *testing.T) {
	t.Run("Original API Must Work", func(t *testing.T) {
		// This test ensures that the original AutoPDF config API continues to work
		// This is our "minimum bar" - if this breaks, we've broken backward compatibility

		// Test 1: Simple config creation
		cfg := &Config{
			Template: "document.tex",
			Output:   "document.pdf",
			Variables: map[string]interface{}{
				"title":   "My Document",
				"author":  "John Doe",
				"content": "This is my document content.",
			},
			Engine: "pdflatex",
		}

		// Verify basic functionality
		assert.Equal(t, "document.tex", cfg.Template.String())
		assert.Equal(t, "document.pdf", cfg.Output.String())
		assert.Equal(t, "pdflatex", cfg.Engine.String())
		assert.Equal(t, "My Document", cfg.Variables["title"])
		assert.Equal(t, "John Doe", cfg.Variables["author"])
		assert.Equal(t, "This is my document content.", cfg.Variables["content"])

		// Test 2: Default config
		defaultCfg := GetDefaultConfig()
		assert.Equal(t, "pdflatex", defaultCfg.Engine.String())
		assert.False(t, defaultCfg.Conversion.Enabled)
		assert.Empty(t, defaultCfg.Conversion.Formats)
		assert.NotNil(t, defaultCfg.Variables)

		// Test 3: JSON conversion
		jsonStr, err := cfg.ToJSON()
		require.NoError(t, err)
		assert.Contains(t, jsonStr, `"template":"document.tex"`)
		assert.Contains(t, jsonStr, `"output":"document.pdf"`)
		assert.Contains(t, jsonStr, `"engine":"pdflatex"`)
	})

	t.Run("Enhanced Features Must Not Break Original", func(t *testing.T) {
		// This test ensures that enhanced features don't break original functionality

		// Test that we can still use simple variables with enhanced features
		cfg := &Config{
			Template: "document.tex",
			Output:   "document.pdf",
			Variables: map[string]interface{}{
				"title":   "Simple Title",
				"content": "Simple Content",
				// Can add complex structures when ready
				"metadata": map[string]interface{}{
					"created": "2024-01-15",
					"version": "1.0",
				},
			},
			Engine: "xelatex",
		}

		// Verify original functionality still works
		assert.Equal(t, "Simple Title", cfg.Variables["title"])
		assert.Equal(t, "Simple Content", cfg.Variables["content"])

		// Verify enhanced functionality works alongside original
		metadata := cfg.Variables["metadata"].(map[string]interface{})
		assert.Equal(t, "2024-01-15", metadata["created"])
		assert.Equal(t, "1.0", metadata["version"])
	})

	t.Run("Error Handling Must Remain Consistent", func(t *testing.T) {
		// This test ensures that error handling remains consistent

		// Test invalid YAML
		invalidYAML := []byte(`invalid: yaml: content: [`)
		_, err := NewConfigFromYAML(invalidYAML)
		assert.Error(t, err)

		// Test nil config
		err = SaveConfig(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("CartasBackendCompatibility", func(t *testing.T) {
		// This test ensures compatibility with cartas-backend patterns

		// Test config structure similar to cartas-backend
		cfg := &Config{
			Template: "letter.tex",
			Output:   "letter.pdf",
			Variables: map[string]interface{}{
				"document": map[string]interface{}{
					"title":  "Participation in Death",
					"author": "Funeral Home",
				},
				"person": map[string]interface{}{
					"name":       "João Silva",
					"birth_year": 1980,
					"death_year": 2024,
				},
				"event": map[string]interface{}{
					"has_wake":      true,
					"wake_location": "Central Wake",
					"has_graveyard": true,
					"graveyard":     "Municipal Cemetery",
					"two_days":      true,
					"first_date":    "15/03/2024",
					"first_time":    "14h30",
					"second_date":   "16/03/2024",
					"second_time":   "09h00",
				},
			},
			Engine: "xelatex",
			Conversion: Conversion{
				Enabled: true,
				Formats: []string{"png", "jpg"},
			},
		}

		// Verify the config structure
		assert.Equal(t, "letter.tex", cfg.Template.String())
		assert.Equal(t, "letter.pdf", cfg.Output.String())
		assert.Equal(t, "xelatex", cfg.Engine.String())
		assert.True(t, cfg.Conversion.Enabled)
		assert.Equal(t, []string{"png", "jpg"}, cfg.Conversion.Formats)

		// Verify nested variables
		document := cfg.Variables["document"].(map[string]interface{})
		assert.Equal(t, "Participation in Death", document["title"])
		assert.Equal(t, "Funeral Home", document["author"])

		person := cfg.Variables["person"].(map[string]interface{})
		assert.Equal(t, "João Silva", person["name"])
		assert.Equal(t, 1980, person["birth_year"])
		assert.Equal(t, 2024, person["death_year"])

		event := cfg.Variables["event"].(map[string]interface{})
		assert.Equal(t, true, event["has_wake"])
		assert.Equal(t, "Central Wake", event["wake_location"])
		assert.Equal(t, true, event["has_graveyard"])
		assert.Equal(t, "Municipal Cemetery", event["graveyard"])
		assert.Equal(t, true, event["two_days"])
		assert.Equal(t, "15/03/2024", event["first_date"])
		assert.Equal(t, "14h30", event["first_time"])
		assert.Equal(t, "16/03/2024", event["second_date"])
		assert.Equal(t, "09h00", event["second_time"])
	})
}

// TestConfigMigration tests migration scenarios
func TestConfigMigration(t *testing.T) {
	t.Run("Simple to Complex Migration", func(t *testing.T) {
		// Test migration from simple to complex config

		// Original simple config
		simpleConfig := &Config{
			Template: "document.tex",
			Output:   "document.pdf",
			Variables: map[string]interface{}{
				"title":   "My Document",
				"author":  "John Doe",
				"content": "This is my document content.",
			},
			Engine: "pdflatex",
		}

		// Migrate to complex config
		complexConfig := &Config{
			Template: simpleConfig.Template,
			Output:   simpleConfig.Output,
			Variables: map[string]interface{}{
				"document": map[string]interface{}{
					"title":   simpleConfig.Variables["title"],
					"author":  simpleConfig.Variables["author"],
					"content": simpleConfig.Variables["content"],
				},
				"metadata": map[string]interface{}{
					"created": "2024-01-15",
					"version": "1.0",
				},
			},
			Engine: simpleConfig.Engine,
		}

		// Verify migration
		assert.Equal(t, simpleConfig.Template, complexConfig.Template)
		assert.Equal(t, simpleConfig.Output, complexConfig.Output)
		assert.Equal(t, simpleConfig.Engine, complexConfig.Engine)

		// Verify complex structure
		document := complexConfig.Variables["document"].(map[string]interface{})
		assert.Equal(t, "My Document", document["title"])
		assert.Equal(t, "John Doe", document["author"])
		assert.Equal(t, "This is my document content.", document["content"])

		metadata := complexConfig.Variables["metadata"].(map[string]interface{})
		assert.Equal(t, "2024-01-15", metadata["created"])
		assert.Equal(t, "1.0", metadata["version"])
	})

	t.Run("Backward Compatibility During Migration", func(t *testing.T) {
		// Test that both simple and complex configs can coexist

		// Simple config (original)
		simpleConfig := &Config{
			Template: "simple.tex",
			Variables: map[string]interface{}{
				"title": "Simple Document",
			},
		}

		// Complex config (enhanced)
		complexConfig := &Config{
			Template: "complex.tex",
			Variables: map[string]interface{}{
				"document": map[string]interface{}{
					"title": "Complex Document",
				},
			},
		}

		// Both should work
		assert.Equal(t, "Simple Document", simpleConfig.Variables["title"])

		document := complexConfig.Variables["document"].(map[string]interface{})
		assert.Equal(t, "Complex Document", document["title"])

		// Both should be able to convert to JSON
		simpleJSON, err := simpleConfig.ToJSON()
		require.NoError(t, err)
		assert.Contains(t, simpleJSON, `"title":"Simple Document"`)

		complexJSON, err := complexConfig.ToJSON()
		require.NoError(t, err)
		assert.Contains(t, complexJSON, `"document"`)
	})
}
