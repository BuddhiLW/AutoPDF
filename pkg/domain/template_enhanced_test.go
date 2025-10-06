package domain

import (
	"testing"
)

func TestTemplateType_IsValid(t *testing.T) {
	tests := []struct {
		name         string
		templateType TemplateType
		want         bool
	}{
		{
			name:         "LaTeX type",
			templateType: TemplateTypeLaTeX,
			want:         true,
		},
		{
			name:         "ABNTeX type",
			templateType: TemplateTypeABNTeX,
			want:         true,
		},
		{
			name:         "Markdown type",
			templateType: TemplateTypeMarkdown,
			want:         true,
		},
		{
			name:         "HTML type",
			templateType: TemplateTypeHTML,
			want:         true,
		},
		{
			name:         "Invalid type",
			templateType: TemplateType("invalid"),
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidTemplateType(tt.templateType)
			if got != tt.want {
				t.Errorf("isValidTemplateType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplate_NewTemplate_WithType(t *testing.T) {
	tests := []struct {
		name         string
		id           TemplateID
		path         TemplatePath
		content      string
		templateType TemplateType
		metadata     TemplateMetadata
		wantErr      bool
	}{
		{
			name:         "valid LaTeX template",
			id:           TemplateID{Value: "template_123"},
			path:         TemplatePath{value: "templates/document.tex"},
			content:      "\\documentclass{article}\\begin{document}Hello World\\end{document}",
			templateType: TemplateTypeLaTeX,
			metadata:     NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
			wantErr:      false,
		},
		{
			name:         "valid ABNTeX template",
			id:           TemplateID{Value: "template_123"},
			path:         TemplatePath{value: "templates/document.tex"},
			content:      "\\documentclass{article}\\usepackage{abntex2}\\begin{document}Hello World\\end{document}",
			templateType: TemplateTypeABNTeX,
			metadata:     NewTemplateMetadata("Test", "Test template", TemplateTypeABNTeX),
			wantErr:      false,
		},
		{
			name:         "empty content",
			id:           TemplateID{Value: "template_123"},
			path:         TemplatePath{value: "templates/document.tex"},
			content:      "",
			templateType: TemplateTypeLaTeX,
			metadata:     NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
			wantErr:      true,
		},
		{
			name:         "invalid template type",
			id:           TemplateID{Value: "template_123"},
			path:         TemplatePath{value: "templates/document.tex"},
			content:      "\\documentclass{article}\\begin{document}Hello World\\end{document}",
			templateType: TemplateType("invalid"),
			metadata:     NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTemplate(tt.id, tt.path, tt.content, tt.templateType, tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("NewTemplate() returned nil template")
			}
			if !tt.wantErr && got.Type != tt.templateType {
				t.Errorf("NewTemplate() type = %v, want %v", got.Type, tt.templateType)
			}
		})
	}
}

func TestTemplate_GetType(t *testing.T) {
	template, err := NewTemplate(
		TemplateID{Value: "template_123"},
		TemplatePath{value: "templates/document.tex"},
		"\\documentclass{article}\\begin{document}Hello World\\end{document}",
		TemplateTypeLaTeX,
		NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
	)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if template.GetType() != TemplateTypeLaTeX {
		t.Errorf("GetType() = %v, want %v", template.GetType(), TemplateTypeLaTeX)
	}
}

func TestTemplate_SetType(t *testing.T) {
	template, err := NewTemplate(
		TemplateID{Value: "template_123"},
		TemplatePath{value: "templates/document.tex"},
		"\\documentclass{article}\\begin{document}Hello World\\end{document}",
		TemplateTypeLaTeX,
		NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
	)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Test valid type change
	err = template.SetType(TemplateTypeABNTeX)
	if err != nil {
		t.Errorf("SetType() error = %v", err)
	}
	if template.Type != TemplateTypeABNTeX {
		t.Errorf("SetType() type = %v, want %v", template.Type, TemplateTypeABNTeX)
	}

	// Test invalid type change
	err = template.SetType(TemplateType("invalid"))
	if err == nil {
		t.Errorf("SetType() should return error for invalid type")
	}
}

func TestTemplate_AddVariable(t *testing.T) {
	template, err := NewTemplate(
		TemplateID{Value: "template_123"},
		TemplatePath{value: "templates/document.tex"},
		"\\documentclass{article}\\begin{document}Hello World\\end{document}",
		TemplateTypeLaTeX,
		NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
	)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Add a variable
	variable, err := NewStringVariable("test value")
	if err != nil {
		t.Fatalf("Failed to create variable: %v", err)
	}

	template.AddVariable("test_key", variable)

	// Check if variable was added
	if !template.HasVariable("test_key") {
		t.Errorf("HasVariable() should return true for added variable")
	}
}

func TestTemplate_RemoveVariable(t *testing.T) {
	template, err := NewTemplate(
		TemplateID{Value: "template_123"},
		TemplatePath{value: "templates/document.tex"},
		"\\documentclass{article}\\begin{document}Hello World\\end{document}",
		TemplateTypeLaTeX,
		NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
	)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Add a variable
	variable, err := NewStringVariable("test value")
	if err != nil {
		t.Fatalf("Failed to create variable: %v", err)
	}

	template.AddVariable("test_key", variable)

	// Remove the variable
	err = template.RemoveVariable("test_key")
	if err != nil {
		t.Errorf("RemoveVariable() error = %v", err)
	}

	// Check if variable was removed
	if template.HasVariable("test_key") {
		t.Errorf("HasVariable() should return false for removed variable")
	}
}

func TestTemplate_GetVariable(t *testing.T) {
	template, err := NewTemplate(
		TemplateID{Value: "template_123"},
		TemplatePath{value: "templates/document.tex"},
		"\\documentclass{article}\\begin{document}Hello World\\end{document}",
		TemplateTypeLaTeX,
		NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
	)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Add a variable
	variable, err := NewStringVariable("test value")
	if err != nil {
		t.Fatalf("Failed to create variable: %v", err)
	}

	template.AddVariable("test_key", variable)

	// Get the variable
	retrievedVariable, err := template.GetVariable("test_key")
	if err != nil {
		t.Errorf("GetVariable() error = %v", err)
	}
	if retrievedVariable == nil {
		t.Errorf("GetVariable() returned nil variable")
	}
}

func TestTemplate_ValidateTemplate(t *testing.T) {
	template, err := NewTemplate(
		TemplateID{Value: "template_123"},
		TemplatePath{value: "templates/document.tex"},
		"\\documentclass{article}\\begin{document}Hello World\\end{document}",
		TemplateTypeLaTeX,
		NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
	)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Test valid template
	err = template.ValidateTemplate()
	if err != nil {
		t.Errorf("ValidateTemplate() error = %v", err)
	}

	// Test template with invalid content
	err = template.UpdateContent("")
	if err == nil {
		t.Errorf("UpdateContent() should return error for empty content")
	}

	// Manually set empty content to test ValidateTemplate
	template.Content = ""
	err = template.ValidateTemplate()
	if err == nil {
		t.Errorf("ValidateTemplate() should return error for invalid content")
	}
}

func TestDetermineTemplateType(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected TemplateType
	}{
		{
			name:     "LaTeX content",
			content:  "\\documentclass{article}\\begin{document}Hello World\\end{document}",
			expected: TemplateTypeLaTeX,
		},
		{
			name:     "ABNTeX content",
			content:  "\\documentclass{article}\\usepackage{abntex2}\\begin{document}Hello World\\end{document}",
			expected: TemplateTypeABNTeX,
		},
		{
			name:     "Markdown content",
			content:  "# Hello World\n\nThis is a test document.",
			expected: TemplateTypeMarkdown,
		},
		{
			name:     "HTML content",
			content:  "<html><body><h1>Hello World</h1></body></html>",
			expected: TemplateTypeHTML,
		},
		{
			name:     "Unknown content",
			content:  "Some random text",
			expected: TemplateTypeLaTeX, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineTemplateType(tt.content)
			if got != tt.expected {
				t.Errorf("DetermineTemplateType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultTemplateFactory_CreateTemplate(t *testing.T) {
	factory := NewDefaultTemplateFactory()

	tests := []struct {
		name     string
		id       TemplateID
		path     TemplatePath
		content  string
		metadata TemplateMetadata
		wantErr  bool
	}{
		{
			name:     "LaTeX template",
			id:       TemplateID{Value: "template_123"},
			path:     TemplatePath{value: "templates/document.tex"},
			content:  "\\documentclass{article}\\begin{document}Hello World\\end{document}",
			metadata: NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
			wantErr:  false,
		},
		{
			name:     "ABNTeX template",
			id:       TemplateID{Value: "template_123"},
			path:     TemplatePath{value: "templates/document.tex"},
			content:  "\\documentclass{article}\\usepackage{abntex2}\\begin{document}Hello World\\end{document}",
			metadata: NewTemplateMetadata("Test", "Test template", TemplateTypeABNTeX),
			wantErr:  false,
		},
		{
			name:     "Markdown template",
			id:       TemplateID{Value: "template_123"},
			path:     TemplatePath{value: "templates/document.md"},
			content:  "# Hello World\n\nThis is a test document.",
			metadata: NewTemplateMetadata("Test", "Test template", TemplateTypeMarkdown),
			wantErr:  false,
		},
		{
			name:     "HTML template",
			id:       TemplateID{Value: "template_123"},
			path:     TemplatePath{value: "templates/document.html"},
			content:  "<html><body><h1>Hello World</h1></body></html>",
			metadata: NewTemplateMetadata("Test", "Test template", TemplateTypeHTML),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := factory.CreateTemplate(tt.id, tt.path, tt.content, tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("CreateTemplate() returned nil template")
			}
		})
	}
}

func TestDefaultTemplateFactory_CreateTemplateFromType(t *testing.T) {
	factory := NewDefaultTemplateFactory()

	tests := []struct {
		name         string
		templateType TemplateType
		id           TemplateID
		path         TemplatePath
		content      string
		metadata     TemplateMetadata
		wantErr      bool
	}{
		{
			name:         "LaTeX template",
			templateType: TemplateTypeLaTeX,
			id:           TemplateID{Value: "template_123"},
			path:         TemplatePath{value: "templates/document.tex"},
			content:      "\\documentclass{article}\\begin{document}Hello World\\end{document}",
			metadata:     NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
			wantErr:      false,
		},
		{
			name:         "ABNTeX template",
			templateType: TemplateTypeABNTeX,
			id:           TemplateID{Value: "template_123"},
			path:         TemplatePath{value: "templates/document.tex"},
			content:      "\\documentclass{article}\\begin{document}Hello World\\end{document}",
			metadata:     NewTemplateMetadata("Test", "Test template", TemplateTypeABNTeX),
			wantErr:      false,
		},
		{
			name:         "Markdown template",
			templateType: TemplateTypeMarkdown,
			id:           TemplateID{Value: "template_123"},
			path:         TemplatePath{value: "templates/document.md"},
			content:      "# Hello World",
			metadata:     NewTemplateMetadata("Test", "Test template", TemplateTypeMarkdown),
			wantErr:      false,
		},
		{
			name:         "HTML template",
			templateType: TemplateTypeHTML,
			id:           TemplateID{Value: "template_123"},
			path:         TemplatePath{value: "templates/document.html"},
			content:      "<html><body>Hello World</body></html>",
			metadata:     NewTemplateMetadata("Test", "Test template", TemplateTypeHTML),
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := factory.CreateTemplateFromType(tt.templateType, tt.id, tt.path, tt.content, tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTemplateFromType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("CreateTemplateFromType() returned nil template")
			}
			if !tt.wantErr && got.Type != tt.templateType {
				t.Errorf("CreateTemplateFromType() type = %v, want %v", got.Type, tt.templateType)
			}
		})
	}
}

func TestLaTeXProcessingStrategy_Process(t *testing.T) {
	delimiters := DelimiterConfig{Left: "{{", Right: "}}"}
	strategy := NewLaTeXProcessingStrategy(delimiters)

	template, err := NewTemplate(
		TemplateID{Value: "template_123"},
		TemplatePath{value: "templates/document.tex"},
		"\\documentclass{article}\\begin{document}Hello {{.name}}\\end{document}",
		TemplateTypeLaTeX,
		NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
	)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	variables := NewVariableCollection()
	nameVar, err := NewStringVariable("John Doe")
	if err != nil {
		t.Fatalf("Failed to create variable: %v", err)
	}
	variables.Set("name", nameVar)

	result, err := strategy.Process(template, variables)
	if err != nil {
		t.Errorf("Process() error = %v", err)
	}

	expected := "\\documentclass{article}\\begin{document}Hello John Doe\\end{document}"
	if result != expected {
		t.Errorf("Process() result = %v, want %v", result, expected)
	}
}

func TestLaTeXProcessingStrategy_Supports(t *testing.T) {
	delimiters := DelimiterConfig{Left: "{{", Right: "}}"}
	strategy := NewLaTeXProcessingStrategy(delimiters)

	if !strategy.Supports(TemplateTypeLaTeX) {
		t.Errorf("Supports() should return true for LaTeX")
	}
	if strategy.Supports(TemplateTypeABNTeX) {
		t.Errorf("Supports() should return false for ABNTeX")
	}
}

func TestLaTeXProcessingStrategy_GetName(t *testing.T) {
	delimiters := DelimiterConfig{Left: "{{", Right: "}}"}
	strategy := NewLaTeXProcessingStrategy(delimiters)

	expected := "LaTeX Processing Strategy"
	if strategy.GetName() != expected {
		t.Errorf("GetName() = %v, want %v", strategy.GetName(), expected)
	}
}

func TestTemplateProcessingContext_ProcessTemplate(t *testing.T) {
	context := NewTemplateProcessingContext()

	template, err := NewTemplate(
		TemplateID{Value: "template_123"},
		TemplatePath{value: "templates/document.tex"},
		"\\documentclass{article}\\begin{document}Hello {{.name}}\\end{document}",
		TemplateTypeLaTeX,
		NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
	)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	variables := NewVariableCollection()
	nameVar, err := NewStringVariable("John Doe")
	if err != nil {
		t.Fatalf("Failed to create variable: %v", err)
	}
	variables.Set("name", nameVar)

	result, err := context.ProcessTemplate(template, variables)
	if err != nil {
		t.Errorf("ProcessTemplate() error = %v", err)
	}

	expected := "\\documentclass{article}\\begin{document}Hello John Doe\\end{document}"
	if result != expected {
		t.Errorf("ProcessTemplate() result = %v, want %v", result, expected)
	}
}

func TestTemplateProcessingContext_AddStrategy(t *testing.T) {
	context := NewTemplateProcessingContext()

	// Create a custom strategy
	delimiters := DelimiterConfig{Left: "{{", Right: "}}"}
	customStrategy := NewLaTeXProcessingStrategy(delimiters)

	// Add the strategy
	context.AddStrategy(TemplateTypeLaTeX, customStrategy)

	// Test that the strategy was added
	template, err := NewTemplate(
		TemplateID{Value: "template_123"},
		TemplatePath{value: "templates/document.tex"},
		"\\documentclass{article}\\begin{document}Hello {{.name}}\\end{document}",
		TemplateTypeLaTeX,
		NewTemplateMetadata("Test", "Test template", TemplateTypeLaTeX),
	)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	variables := NewVariableCollection()
	nameVar, err := NewStringVariable("John Doe")
	if err != nil {
		t.Fatalf("Failed to create variable: %v", err)
	}
	variables.Set("name", nameVar)

	result, err := context.ProcessTemplate(template, variables)
	if err != nil {
		t.Errorf("ProcessTemplate() error = %v", err)
	}

	expected := "\\documentclass{article}\\begin{document}Hello John Doe\\end{document}"
	if result != expected {
		t.Errorf("ProcessTemplate() result = %v, want %v", result, expected)
	}
}
