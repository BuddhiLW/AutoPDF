package test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/internal/template"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/domain"
)

// TestMinimumBar establishes the minimum bar for backward compatibility
// This test ensures that the original AutoPDF API continues to work exactly as before
func TestMinimumBar(t *testing.T) {
	fmt.Println("AutoPDF Minimum Bar Test")
	fmt.Println("========================")
	fmt.Println("This test establishes the minimum bar for backward compatibility.")
	fmt.Println("If this test fails, we have broken backward compatibility.")
	fmt.Println()

	// Test 1: Original Engine API
	fmt.Println("1. Testing Original Engine API...")
	if err := testOriginalEngineAPI(); err != nil {
		log.Fatalf("Original Engine API test failed: %v", err)
	}
	fmt.Println("   ✅ Original Engine API works")

	// Test 2: Original Config
	fmt.Println("2. Testing Original Config...")
	if err := testOriginalConfig(); err != nil {
		log.Fatalf("Original Config test failed: %v", err)
	}
	fmt.Println("   ✅ Original Config works")

	// Test 3: Original Variables
	fmt.Println("3. Testing Original Variables...")
	if err := testOriginalVariables(); err != nil {
		log.Fatalf("Original Variables test failed: %v", err)
	}
	fmt.Println("   ✅ Original Variables work")

	// Test 4: Enhanced Engine with Simple Variables
	fmt.Println("4. Testing Enhanced Engine with Simple Variables...")
	if err := testEnhancedEngineWithSimpleVariablesMinBar(t); err != nil {
		log.Fatalf("Enhanced Engine with Simple Variables test failed: %v", err)
	}
	fmt.Println("   ✅ Enhanced Engine with Simple Variables works")

	// Test 5: Enhanced Engine with Complex Variables
	fmt.Println("5. Testing Enhanced Engine with Complex Variables...")
	if err := testEnhancedEngineWithComplexVariablesMinBar(t); err != nil {
		log.Fatalf("Enhanced Engine with Complex Variables test failed: %v", err)
	}
	fmt.Println("   ✅ Enhanced Engine with Complex Variables works")

	// Test 6: CartasBackend Compatibility
	fmt.Println("6. Testing CartasBackend Compatibility...")
	if err := testCartasBackendCompatibility(); err != nil {
		log.Fatalf("CartasBackend Compatibility test failed: %v", err)
	}
	fmt.Println("   ✅ CartasBackend Compatibility works")

	fmt.Println()
	fmt.Println("🎉 ALL MINIMUM BAR TESTS PASSED!")
	fmt.Println("✅ Backward compatibility is maintained")
	fmt.Println("✅ Enhanced features work alongside original features")
	fmt.Println("✅ CartasBackend patterns are supported")
}

func testOriginalEngineAPI() error {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-minimum-bar")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Create simple template
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
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return err
	}

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
	engine := template.NewEngine(cfg)
	result, err := engine.Process(templatePath)
	if err != nil {
		return err
	}

	// Verify the result contains expected content
	expected := []string{
		`\title{Original API Test}`,
		`\author{John Doe}`,
		`\date{2024-01-15}`,
		`This is a test using the original API.`,
	}

	for _, exp := range expected {
		if !containsMinBar(result, exp) {
			return fmt.Errorf("expected content not found: %s", exp)
		}
	}

	// Test ProcessToFile
	outputPath := filepath.Join(tempDir, "output.tex")
	cfg.Output = config.Output(outputPath)

	err = engine.ProcessToFile(templatePath, outputPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("output file was not created")
	}

	// Test AddFunction
	engine.AddFunction("upper", func(s string) string {
		return "UPPERCASE_" + s
	})

	templateContent = `\title{delim[[.title | upper]]}`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return err
	}

	result, err = engine.Process(templatePath)
	if err != nil {
		return err
	}

	if !containsMinBar(result, `\title{UPPERCASE_Original API Test}`) {
		return fmt.Errorf("custom function not applied correctly")
	}

	return nil
}

func testOriginalConfig() error {
	// Test original config structure
	cfg := &config.Config{
		Template: "test.tex",
		Output:   "output.pdf",
		Variables: map[string]interface{}{
			"title":   "Test Document",
			"author":  "John Doe",
			"content": "This is a test document.",
		},
		Engine: "pdflatex",
		Conversion: config.Conversion{
			Enabled: false,
			Formats: []string{},
		},
	}

	// Test that all fields are accessible
	if cfg.Template.String() != "test.tex" {
		return fmt.Errorf("template field not accessible")
	}
	if cfg.Output.String() != "output.pdf" {
		return fmt.Errorf("output field not accessible")
	}
	if cfg.Engine.String() != "pdflatex" {
		return fmt.Errorf("engine field not accessible")
	}
	if cfg.Conversion.Enabled != false {
		return fmt.Errorf("conversion enabled field not accessible")
	}
	if len(cfg.Conversion.Formats) != 0 {
		return fmt.Errorf("conversion formats field not accessible")
	}

	// Test variables access
	if cfg.Variables["title"] != "Test Document" {
		return fmt.Errorf("title variable not accessible")
	}
	if cfg.Variables["author"] != "John Doe" {
		return fmt.Errorf("author variable not accessible")
	}
	if cfg.Variables["content"] != "This is a test document." {
		return fmt.Errorf("content variable not accessible")
	}

	// Test JSON conversion
	jsonStr, err := cfg.ToJSON()
	if err != nil {
		return err
	}
	if !containsMinBar(jsonStr, `"template": "test.tex"`) {
		return fmt.Errorf("JSON conversion failed for template")
	}
	if !containsMinBar(jsonStr, `"output": "output.pdf"`) {
		return fmt.Errorf("JSON conversion failed for output")
	}
	if !containsMinBar(jsonStr, `"engine": "pdflatex"`) {
		return fmt.Errorf("JSON conversion failed for engine")
	}

	// Test default config
	defaultCfg := config.GetDefaultConfig()
	if defaultCfg.Engine.String() != "pdflatex" {
		return fmt.Errorf("default config engine not correct")
	}
	if defaultCfg.Conversion.Enabled != false {
		return fmt.Errorf("default config conversion enabled not correct")
	}
	if len(defaultCfg.Conversion.Formats) != 0 {
		return fmt.Errorf("default config conversion formats not correct")
	}
	if defaultCfg.Variables == nil {
		return fmt.Errorf("default config variables not initialized")
	}

	return nil
}

func testOriginalVariables() error {
	// Test that original variable types still work
	stringVar, _ := domain.NewStringVariable("test")
	if stringVar.Type != domain.VariableTypeString {
		return fmt.Errorf("string variable type not correct")
	}
	if stringVar.Value != "test" {
		return fmt.Errorf("string variable value not correct")
	}

	numberVar, _ := domain.NewNumberVariable(42)
	if numberVar.Type != domain.VariableTypeNumber {
		return fmt.Errorf("number variable type not correct")
	}
	if numberVar.Value != float64(42) {
		return fmt.Errorf("number variable value not correct")
	}

	booleanVar := domain.NewBooleanVariable(true)
	if booleanVar.Type != domain.VariableTypeBoolean {
		return fmt.Errorf("boolean variable type not correct")
	}
	if booleanVar.Value != true {
		return fmt.Errorf("boolean variable value not correct")
	}

	// Test variable collection
	vc := domain.NewVariableCollection()
	vc.Set("key", stringVar)

	retrieved, exists := vc.Get("key")
	if !exists {
		return fmt.Errorf("variable not found in collection")
	}
	if retrieved != stringVar {
		return fmt.Errorf("retrieved variable not the same as set")
	}

	// Test template context
	context := domain.NewTemplateContext()
	err := context.SetVariable("key", stringVar)
	if err != nil {
		return err
	}

	retrieved, err = context.GetVariable("key")
	if err != nil {
		return err
	}
	if retrieved != stringVar {
		return fmt.Errorf("retrieved variable from context not the same as set")
	}

	return nil
}

func testEnhancedEngineWithSimpleVariablesMinBar(t *testing.T) error {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-minimum-bar")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Create simple template
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
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return err
	}

	// Test enhanced engine with simple variables
	config := &template.EnhancedConfig{
		TemplatePath: templatePath,
		Delimiters: template.DelimiterConfig{
			Left:  "delim[[",
			Right: "]]",
		},
	}

	engine := template.NewEnhancedEngine(config)

	// Set simple variables (same as original format)
	variables := map[string]interface{}{
		"title":   "Enhanced Engine with Simple Variables",
		"author":  "Jane Smith",
		"content": "This document uses the enhanced engine with simple variables.",
	}

	err = engine.SetVariablesFromMap(variables)
	if err != nil {
		return err
	}

	result, err := engine.Process(templatePath)
	if err != nil {
		return err
	}

	// Verify the result contains expected content
	expected := []string{
		`\title{Enhanced Engine with Simple Variables}`,
		`\author{Jane Smith}`,
		`This document uses the enhanced engine with simple variables.`,
	}

	for _, exp := range expected {
		if !containsMinBar(result, exp) {
			return fmt.Errorf("expected content not found: %s", exp)
		}
	}

	return nil
}

func testEnhancedEngineWithComplexVariablesMinBar(t *testing.T) error {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-minimum-bar")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Create complex template
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
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return err
	}

	// Test enhanced engine with complex variables
	config := &template.EnhancedConfig{
		TemplatePath: templatePath,
		Delimiters: template.DelimiterConfig{
			Left:  "delim[[",
			Right: "]]",
		},
	}

	engine := template.NewEnhancedEngine(config)

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
				"name":   "John Doe",
				"role":   "Developer",
				"skills": []interface{}{"Go", "LaTeX"},
			},
			map[string]interface{}{
				"name":   "Jane Smith",
				"role":   "Writer",
				"skills": []interface{}{"Documentation", "Markdown"},
			},
		},
	}

	err = engine.SetVariablesFromMap(variables)
	if err != nil {
		return err
	}

	result, err := engine.Process(templatePath)
	if err != nil {
		return err
	}

	// Verify the result contains expected content
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
		if !containsMinBar(result, exp) {
			return fmt.Errorf("expected content not found: %s", exp)
		}
	}

	return nil
}

func testCartasBackendCompatibility() error {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-minimum-bar")
	if err != nil {
		return err
	}
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
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		return err
	}

	// Create enhanced engine with cartas-backend-like data structure
	config := &template.EnhancedConfig{
		TemplatePath: templatePath,
		Delimiters: template.DelimiterConfig{
			Left:  "delim[[",
			Right: "]]",
		},
	}

	engine := template.NewEnhancedEngine(config)

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
	if err != nil {
		return err
	}

	result, err := engine.Process(templatePath)
	if err != nil {
		return err
	}

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
		if !containsMinBar(result, exp) {
			return fmt.Errorf("expected content not found: %s", exp)
		}
	}

	return nil
}

func containsMinBar(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
