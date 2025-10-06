package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackwardCompatibility ensures that all existing template functionality continues to work
// This establishes the "minimum bar" for backward compatibility
func TestBackwardCompatibility(t *testing.T) {
	t.Run("Original Engine API", func(t *testing.T) {
		// Test that the original engine API continues to work exactly as before
		tempDir, err := os.MkdirTemp("", "autopdf-compat-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create a simple template
		templateContent := `
\documentclass{article}
\title{delim[[.title]]}
\author{delim[[.author]]}
\date{delim[[.date]]}

\begin{document}
\maketitle

delim[[.content]]
\end{document}
`

		templatePath := filepath.Join(tempDir, "test.tex")
		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		require.NoError(t, err)

		// Test original config with simple variables
		cfg := &config.Config{
			Template: config.Template(templatePath),
			Variables: map[string]interface{}{
				"title":   "Original API Test",
				"author":  "John Doe",
				"date":    "2024-01-15",
				"content": "This is a test using the original API.",
			},
		}

		// Test original engine
		engine := NewEngine(cfg)
		result, err := engine.Process(templatePath)
		require.NoError(t, err)

		// Verify the result contains expected content
		expected := []string{
			`\title{Original API Test}`,
			`\author{John Doe}`,
			`\date{2024-01-15}`,
			`This is a test using the original API.`,
		}

		for _, exp := range expected {
			assert.Contains(t, result, exp)
		}
	})

	t.Run("Original Engine ProcessToFile", func(t *testing.T) {
		// Test that ProcessToFile continues to work
		tempDir, err := os.MkdirTemp("", "autopdf-compat-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		templateContent := `\title{delim[[.title]]}`
		templatePath := filepath.Join(tempDir, "test.tex")
		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		require.NoError(t, err)

		outputPath := filepath.Join(tempDir, "output.tex")

		cfg := &config.Config{
			Template: config.Template(templatePath),
			Output:   config.Output(outputPath),
			Variables: map[string]interface{}{
				"title": "ProcessToFile Test",
			},
		}

		engine := NewEngine(cfg)
		err = engine.ProcessToFile(templatePath, outputPath)
		require.NoError(t, err)

		// Verify output file exists and has correct content
		require.FileExists(t, outputPath)

		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), `\title{ProcessToFile Test}`)
	})

	t.Run("Original Engine AddFunction", func(t *testing.T) {
		// Test that AddFunction continues to work
		tempDir, err := os.MkdirTemp("", "autopdf-compat-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		templateContent := `\title{delim[[.title | upper]]}`
		templatePath := filepath.Join(tempDir, "test.tex")
		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		require.NoError(t, err)

		cfg := &config.Config{
			Template: config.Template(templatePath),
			Variables: map[string]interface{}{
				"title": "test document",
			},
		}

		engine := NewEngine(cfg)
		engine.AddFunction("upper", func(s string) string {
			return "UPPERCASE_" + s
		})

		result, err := engine.Process(templatePath)
		require.NoError(t, err)
		assert.Contains(t, result, `\title{UPPERCASE_test document}`)
	})

	t.Run("Enhanced Engine with Simple Variables", func(t *testing.T) {
		// Test that enhanced engine works with simple variables (backward compatibility)
		tempDir, err := os.MkdirTemp("", "autopdf-compat-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		templateContent := `
\documentclass{article}
\title{delim[[.title]]}
\author{delim[[.author]]}

\begin{document}
\maketitle

delim[[.content]]
\end{document}
`

		templatePath := filepath.Join(tempDir, "enhanced-simple.tex")
		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		require.NoError(t, err)

		config := &EnhancedConfig{
			TemplatePath: templatePath,
			Delimiters: DelimiterConfig{
				Left:  "delim[[",
				Right: "]]",
			},
		}

		engine := NewEnhancedEngine(config)

		// Set simple variables (same as original format)
		variables := map[string]interface{}{
			"title":   "Enhanced Engine with Simple Variables",
			"author":  "Jane Smith",
			"content": "This document uses the enhanced engine with simple variables.",
		}

		err = engine.SetVariablesFromMap(variables)
		require.NoError(t, err)

		result, err := engine.Process(templatePath)
		require.NoError(t, err)

		expected := []string{
			`\title{Enhanced Engine with Simple Variables}`,
			`\author{Jane Smith}`,
			`This document uses the enhanced engine with simple variables.`,
		}

		for _, exp := range expected {
			assert.Contains(t, result, exp)
		}
	})

	t.Run("Enhanced Engine with Complex Variables", func(t *testing.T) {
		// Test that enhanced engine works with complex variables (new functionality)
		tempDir, err := os.MkdirTemp("", "autopdf-compat-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		templateContent := `
\documentclass{article}
\title{delim[[.document.title]]}
\author{delim[[.document.author]]}

\begin{document}
\maketitle

\section{Company Information}
Company: delim[[.company.name]]
Address: delim[[.company.address.street]], delim[[.company.address.city]]

\section{Team Members}
delim[[range .team]]
\subsection{delim[[.name]]}
Role: delim[[.role]]
Skills: delim[[join ", " .skills]]
delim[[end]]
\end{document}
`

		templatePath := filepath.Join(tempDir, "enhanced-complex.tex")
		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		require.NoError(t, err)

		config := &EnhancedConfig{
			TemplatePath: templatePath,
			Delimiters: DelimiterConfig{
				Left:  "delim[[",
				Right: "]]",
			},
		}

		engine := NewEnhancedEngine(config)

		// Set complex variables
		variables := map[string]interface{}{
			"document": map[string]interface{}{
				"title":  "Complex Document with Enhanced Engine",
				"author": "AutoPDF Team",
			},
			"company": map[string]interface{}{
				"name": "AutoPDF Solutions",
				"address": map[string]interface{}{
					"street": "123 Tech Drive",
					"city":   "San Francisco",
				},
			},
			"team": []interface{}{
				map[string]interface{}{
					"name":  "John Doe",
					"role":  "Developer",
					"skills": []interface{}{"Go", "LaTeX"},
				},
				map[string]interface{}{
					"name":  "Jane Smith",
					"role":  "Writer",
					"skills": []interface{}{"Documentation", "Markdown"},
				},
			},
		}

		err = engine.SetVariablesFromMap(variables)
		require.NoError(t, err)

		result, err := engine.Process(templatePath)
		require.NoError(t, err)

		expected := []string{
			`\title{Complex Document with Enhanced Engine}`,
			`\author{AutoPDF Team}`,
			`Company: AutoPDF Solutions`,
			`Address: 123 Tech Drive, San Francisco`,
			`\subsection{John Doe}`,
			`Role: Developer`,
			`Skills: Go, LaTeX`,
			`\subsection{Jane Smith}`,
			`Role: Writer`,
			`Skills: Documentation, Markdown`,
		}

		for _, exp := range expected {
			assert.Contains(t, result, exp)
		}
	})

	t.Run("Enhanced Engine Validation", func(t *testing.T) {
		// Test that enhanced engine validation works
		config := &EnhancedConfig{}
		engine := NewEnhancedEngine(config)

		// Set some variables
		variables := map[string]interface{}{
			"title":  "Test Document",
			"author": "Test Author",
		}

		err := engine.SetVariablesFromMap(variables)
		require.NoError(t, err)

		// Test validation with existing variables
		requiredVars := []string{"title", "author"}
		err = engine.ValidateTemplate(requiredVars)
		assert.NoError(t, err)

		// Test validation with missing variables
		requiredVars = []string{"title", "author", "missing"}
		err = engine.ValidateTemplate(requiredVars)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required variable 'missing' not found")
	})

	t.Run("Enhanced Engine Clone", func(t *testing.T) {
		// Test that enhanced engine cloning works
		config := &EnhancedConfig{}
		engine := NewEnhancedEngine(config)

		// Set some variables
		variables := map[string]interface{}{
			"title": "Original Document",
		}

		err := engine.SetVariablesFromMap(variables)
		require.NoError(t, err)

		// Clone the engine
		clone := engine.Clone()

		// Verify the clone has the same variables
		originalVar, err := engine.GetVariable("title")
		require.NoError(t, err)

		cloneVar, err := clone.GetVariable("title")
		require.NoError(t, err)

		assert.Equal(t, originalVar.Value, cloneVar.Value)

		// Modify the clone and verify it doesn't affect the original
		clone.SetVariable("title", domain.NewStringVariable("Modified Document"))

		originalVar, err = engine.GetVariable("title")
		require.NoError(t, err)
		originalStr, err := originalVar.AsString()
		require.NoError(t, err)
		assert.Equal(t, "Original Document", originalStr)
	})
}

// TestMinimumBar establishes the minimum bar for backward compatibility
func TestMinimumBar(t *testing.T) {
	t.Run("Original API Must Work", func(t *testing.T) {
		// This test ensures that the original AutoPDF API continues to work
		// This is our "minimum bar" - if this breaks, we've broken backward compatibility

		tempDir, err := os.MkdirTemp("", "autopdf-minimum-bar")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Test 1: Original engine with simple variables
		templateContent := `\title{delim[[.title]]}`
		templatePath := filepath.Join(tempDir, "test.tex")
		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		require.NoError(t, err)

		cfg := &config.Config{
			Template: config.Template(templatePath),
			Variables: map[string]interface{}{
				"title": "Minimum Bar Test",
			},
		}

		engine := NewEngine(cfg)
		result, err := engine.Process(templatePath)
		require.NoError(t, err)
		assert.Contains(t, result, `\title{Minimum Bar Test}`)

		// Test 2: Original engine ProcessToFile
		outputPath := filepath.Join(tempDir, "output.tex")
		cfg.Output = config.Output(outputPath)

		err = engine.ProcessToFile(templatePath, outputPath)
		require.NoError(t, err)
		require.FileExists(t, outputPath)

		// Test 3: Original engine with custom functions
		engine.AddFunction("upper", func(s string) string {
			return "UPPERCASE_" + s
		})

		templateContent = `\title{delim[[.title | upper]]}`
		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		require.NoError(t, err)

		result, err = engine.Process(templatePath)
		require.NoError(t, err)
		assert.Contains(t, result, `\title{UPPERCASE_Minimum Bar Test}`)
	})

	t.Run("Enhanced Features Must Not Break Original", func(t *testing.T) {
		// This test ensures that enhanced features don't break original functionality

		tempDir, err := os.MkdirTemp("", "autopdf-minimum-bar")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Test that we can use both original and enhanced engines side by side
		templateContent := `\title{delim[[.title]]}`
		templatePath := filepath.Join(tempDir, "test.tex")
		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		require.NoError(t, err)

		// Original engine
		cfg := &config.Config{
			Template: config.Template(templatePath),
			Variables: map[string]interface{}{
				"title": "Original Engine",
			},
		}
		originalEngine := NewEngine(cfg)
		originalResult, err := originalEngine.Process(templatePath)
		require.NoError(t, err)

		// Enhanced engine with simple variables
		enhancedConfig := &EnhancedConfig{
			TemplatePath: templatePath,
			Delimiters: DelimiterConfig{
				Left:  "delim[[",
				Right: "]]",
			},
		}
		enhancedEngine := NewEnhancedEngine(enhancedConfig)
		enhancedEngine.SetVariablesFromMap(map[string]interface{}{
			"title": "Enhanced Engine",
		})
		enhancedResult, err := enhancedEngine.Process(templatePath)
		require.NoError(t, err)

		// Both should work
		assert.Contains(t, originalResult, `\title{Original Engine}`)
		assert.Contains(t, enhancedResult, `\title{Enhanced Engine}`)
	})

	t.Run("Error Handling Must Remain Consistent", func(t *testing.T) {
		// This test ensures that error handling remains consistent

		// Test original engine error handling
		cfg := &config.Config{
			Template: config.Template("nonexistent.tex"),
			Variables: map[string]interface{}{
				"title": "Test",
			},
		}
		engine := NewEngine(cfg)
		_, err := engine.Process("nonexistent.tex")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file")

		// Test enhanced engine error handling
		enhancedConfig := &EnhancedConfig{
			TemplatePath: "nonexistent.tex",
		}
		enhancedEngine := NewEnhancedEngine(enhancedConfig)
		_, err = enhancedEngine.Process("nonexistent.tex")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file")
	})
}

// TestCartasBackendCompatibility tests compatibility with cartas-backend patterns
func TestCartasBackendCompatibility(t *testing.T) {
	t.Run("Letter Processing Pattern", func(t *testing.T) {
		// This test follows the cartas-backend pattern for letter processing
		// but uses AutoPDF enhanced features

		tempDir, err := os.MkdirTemp("", "autopdf-cartas-compat")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create a template similar to cartas-backend
		templateContent := `
\documentclass{article}
\title{delim[[.document.title]]}
\author{delim[[.document.author]]}

\begin{document}
\maketitle

\section{Personal Information}
Name: delim[[.person.name]]
Birth Year: delim[[.person.birth_year]]
Death Year: delim[[.person.death_year]]

\section{Event Details}
delim[[if .event.has_wake]]
Wake Location: delim[[.event.wake_location]]
delim[[end]]

delim[[if .event.has_graveyard]]
Graveyard: delim[[.event.graveyard]]
delim[[end]]

\section{Timing}
delim[[if .event.two_days]]
First Day: delim[[.event.first_date]] at delim[[.event.first_time]]
Second Day: delim[[.event.second_date]] at delim[[.event.second_time]]
delim[[else]]
Date: delim[[.event.date]] at delim[[.event.time]]
delim[[end]]
\end{document}
`

		templatePath := filepath.Join(tempDir, "letter.tex")
		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		require.NoError(t, err)

		// Create enhanced engine with cartas-backend-like data structure
		config := &EnhancedConfig{
			TemplatePath: templatePath,
			Delimiters: DelimiterConfig{
				Left:  "delim[[",
				Right: "]]",
			},
		}

		engine := NewEnhancedEngine(config)

		// Set variables in cartas-backend-like structure
		variables := map[string]interface{}{
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
		}

		err = engine.SetVariablesFromMap(variables)
		require.NoError(t, err)

		result, err := engine.Process(templatePath)
		require.NoError(t, err)

		// Verify the result contains expected content
		expected := []string{
			`\title{Participation in Death}`,
			`\author{Funeral Home}`,
			`Name: João Silva`,
			`Birth Year: 1980`,
			`Death Year: 2024`,
			`Wake Location: Central Wake`,
			`Graveyard: Municipal Cemetery`,
			`First Day: 15/03/2024 at 14h30`,
			`Second Day: 16/03/2024 at 09h00`,
		}

		for _, exp := range expected {
			assert.Contains(t, result, exp)
		}
	})

	t.Run("Formatter Pattern Compatibility", func(t *testing.T) {
		// This test ensures that AutoPDF can work with cartas-backend formatter patterns

		// Create a mock formatter similar to cartas-backend
		type MockFormatter struct{}

		func (f *MockFormatter) FormatTime(hour, minute int) string {
			return "formatted_time"
		}

		func (f *MockFormatter) FormatDate(day, month, year int) string {
			return "formatted_date"
		}

		formatter := &MockFormatter{}

		// Test that we can use formatters with AutoPDF
		config := &EnhancedConfig{}
		engine := NewEnhancedEngine(config)

		// Add formatter functions to the engine
		engine.AddFunction("format_time", func(hour, minute int) string {
			return formatter.FormatTime(hour, minute)
		})
		engine.AddFunction("format_date", func(day, month, year int) string {
			return formatter.FormatDate(day, month, year)
		})

		// Set variables that would use the formatter
		variables := map[string]interface{}{
			"time": "14:30",
			"date": "15/03/2024",
		}

		err := engine.SetVariablesFromMap(variables)
		require.NoError(t, err)

		// Verify the functions were added
		assert.Contains(t, engine.Context.Functions, "format_time")
		assert.Contains(t, engine.Context.Functions, "format_date")
	})
}
