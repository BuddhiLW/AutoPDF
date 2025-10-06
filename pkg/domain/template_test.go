package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateID_NewTemplateID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TemplateID
		wantErr error
	}{
		{
			name:    "valid template ID",
			input:   "template-123",
			want:    TemplateID{Value: "template-123"},
			wantErr: nil,
		},
		{
			name:    "empty template ID",
			input:   "",
			want:    TemplateID{},
			wantErr: errors.New("template ID cannot be empty"),
		},
		{
			name:    "template ID with spaces",
			input:   "  template-123  ",
			want:    TemplateID{Value: "  template-123  "},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTemplateID(tt.input)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("NewTemplateID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("NewTemplateID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewTemplateID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateID_String(t *testing.T) {
	id, err := NewTemplateID("template-123")
	require.NoError(t, err)

	assert.Equal(t, "template-123", id.String())
}

func TestTemplateID_Equals(t *testing.T) {
	id1, err := NewTemplateID("template-123")
	require.NoError(t, err)

	id2, err := NewTemplateID("template-123")
	require.NoError(t, err)

	id3, err := NewTemplateID("template-456")
	require.NoError(t, err)

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}

func TestTemplateMetadata_NewTemplateMetadata(t *testing.T) {
	metadata := NewTemplateMetadata("John Doe", "Test template", TemplateTypeLaTeX)

	assert.Equal(t, "John Doe", metadata.Author)
	assert.Equal(t, "Test template", metadata.Description)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.NotZero(t, metadata.CreatedAt)
	assert.NotZero(t, metadata.UpdatedAt)
	assert.Empty(t, metadata.Tags)
}

func TestTemplate_NewTemplate(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		path    string
		content string
		wantErr error
	}{
		{
			name:    "valid template",
			id:      "template-123",
			path:    "template.tex",
			content: "Hello {{.Name}}",
			wantErr: nil,
		},
		{
			name:    "empty content",
			id:      "template-123",
			path:    "template.tex",
			content: "",
			wantErr: ErrEmptyTemplate,
		},
		{
			name:    "content too large",
			id:      "template-123",
			path:    "template.tex",
			content: string(make([]byte, 1000001)),
			wantErr: ErrTemplateTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templateID, err := NewTemplateID(tt.id)
			require.NoError(t, err)

			templatePath, err := NewTemplatePath(tt.path)
			require.NoError(t, err)

			metadata := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
			template, err := NewTemplate(templateID, templatePath, tt.content, TemplateTypeLaTeX, metadata)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				assert.NotNil(t, template)
				assert.Equal(t, tt.content, template.Content)
				assert.Equal(t, templateID, template.ID)
				assert.Equal(t, templatePath, template.Path)
			}
		})
	}
}

func TestTemplate_Validate(t *testing.T) {
	tests := []struct {
		name     string
		template *Template
		wantErr  error
	}{
		{
			name: "valid template",
			template: &Template{
				ID:      TemplateID{Value: "template-123"},
				Content: "Hello {{.Name}}",
			},
			wantErr: nil,
		},
		{
			name: "empty content",
			template: &Template{
				ID:      TemplateID{Value: "template-123"},
				Content: "",
			},
			wantErr: ErrEmptyTemplate,
		},
		{
			name: "content too large",
			template: &Template{
				ID:      TemplateID{Value: "template-123"},
				Content: string(make([]byte, 1000001)),
			},
			wantErr: ErrTemplateTooLarge,
		},
		{
			name: "empty ID",
			template: &Template{
				ID:      TemplateID{Value: ""},
				Content: "Hello {{.Name}}",
			},
			wantErr: errors.New("template ID cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplate_CanBeProcessed(t *testing.T) {
	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	templatePath, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	metadata := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template, err := NewTemplate(templateID, templatePath, "Hello {{.Name}}", TemplateTypeLaTeX, metadata)
	require.NoError(t, err)

	assert.True(t, template.CanBeProcessed())

	// Test with empty content
	template.Content = ""
	assert.False(t, template.CanBeProcessed())

	// Test with nil variables
	template.Content = "Hello {{.Name}}"
	template.Variables = nil
	assert.False(t, template.CanBeProcessed())
}

func TestTemplate_HasVariables(t *testing.T) {
	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	templatePath, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	metadata := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template, err := NewTemplate(templateID, templatePath, "Hello {{.Name}}", TemplateTypeLaTeX, metadata)
	require.NoError(t, err)

	// Initially no variables
	assert.False(t, template.HasVariables())

	// Add a variable
	nameVar, _ := NewStringVariable("John")
	template.Variables.Set("name", nameVar)
	assert.True(t, template.HasVariables())

	// Test with nil variables
	template.Variables = nil
	assert.False(t, template.HasVariables())
}

func TestTemplate_GetVariableNames(t *testing.T) {
	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	templatePath, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	metadata := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template, err := NewTemplate(templateID, templatePath, "Hello {{.Name}}", TemplateTypeLaTeX, metadata)
	require.NoError(t, err)

	// Initially no variables
	assert.Empty(t, template.GetVariableNames())

	// Add variables
	nameVar, _ := NewStringVariable("John")
	ageVar, _ := NewNumberVariable(30)
	template.Variables.Set("name", nameVar)
	template.Variables.Set("age", ageVar)

	names := template.GetVariableNames()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "name")
	assert.Contains(t, names, "age")

	// Test with nil variables
	template.Variables = nil
	assert.Empty(t, template.GetVariableNames())
}

func TestTemplate_UpdateContent(t *testing.T) {
	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	templatePath, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	metadata := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template, err := NewTemplate(templateID, templatePath, "Hello {{.Name}}", TemplateTypeLaTeX, metadata)
	require.NoError(t, err)

	originalUpdatedAt := template.Metadata.UpdatedAt

	// Update with valid content
	newContent := "Hello {{.Name}}, welcome to {{.Company}}"
	err = template.UpdateContent(newContent)
	require.NoError(t, err)

	assert.Equal(t, newContent, template.Content)
	assert.True(t, template.Metadata.UpdatedAt.After(originalUpdatedAt))

	// Update with empty content
	err = template.UpdateContent("")
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyTemplate, err)

	// Update with content too large
	err = template.UpdateContent(string(make([]byte, 1000001)))
	assert.Error(t, err)
	assert.Equal(t, ErrTemplateTooLarge, err)
}

func TestTemplate_UpdateMetadata(t *testing.T) {
	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	templatePath, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	metadata := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template, err := NewTemplate(templateID, templatePath, "Hello {{.Name}}", TemplateTypeLaTeX, metadata)
	require.NoError(t, err)

	originalUpdatedAt := template.Metadata.UpdatedAt

	newMetadata := TemplateMetadata{
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     "2.0.0",
		Author:      "Jane Doe",
		Description: "Updated template",
		Tags:        []string{"updated", "v2"},
	}

	template.UpdateMetadata(newMetadata)

	assert.Equal(t, newMetadata.Version, template.Metadata.Version)
	assert.Equal(t, newMetadata.Author, template.Metadata.Author)
	assert.Equal(t, newMetadata.Description, template.Metadata.Description)
	assert.Equal(t, newMetadata.Tags, template.Metadata.Tags)
	assert.True(t, template.Metadata.UpdatedAt.After(originalUpdatedAt))
}

func TestTemplate_AddTag(t *testing.T) {
	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	templatePath, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	metadata := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template, err := NewTemplate(templateID, templatePath, "Hello {{.Name}}", TemplateTypeLaTeX, metadata)
	require.NoError(t, err)

	originalUpdatedAt := template.Metadata.UpdatedAt

	// Add first tag
	template.AddTag("important")
	assert.Contains(t, template.Metadata.Tags, "important")
	assert.True(t, template.Metadata.UpdatedAt.After(originalUpdatedAt))

	// Add second tag
	template.AddTag("draft")
	assert.Contains(t, template.Metadata.Tags, "draft")
	assert.Len(t, template.Metadata.Tags, 2)

	// Add duplicate tag (should not be added)
	template.AddTag("important")
	assert.Len(t, template.Metadata.Tags, 2)

	// Add empty tag (should not be added)
	template.AddTag("")
	assert.Len(t, template.Metadata.Tags, 2)
}

func TestTemplate_RemoveTag(t *testing.T) {
	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	templatePath, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	metadata := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template, err := NewTemplate(templateID, templatePath, "Hello {{.Name}}", TemplateTypeLaTeX, metadata)
	require.NoError(t, err)

	// Add tags
	template.AddTag("important")
	template.AddTag("draft")

	originalUpdatedAt := template.Metadata.UpdatedAt

	// Remove existing tag
	template.RemoveTag("important")
	assert.NotContains(t, template.Metadata.Tags, "important")
	assert.Contains(t, template.Metadata.Tags, "draft")
	assert.True(t, template.Metadata.UpdatedAt.After(originalUpdatedAt))

	// Remove non-existing tag
	template.RemoveTag("nonexistent")
	assert.Len(t, template.Metadata.Tags, 1)
}

func TestTemplate_GetSize(t *testing.T) {
	templateID, err := NewTemplateID("template-123")
	require.NoError(t, err)

	templatePath, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	content := "Hello {{.Name}}"
	metadata := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template, err := NewTemplate(templateID, templatePath, content, TemplateTypeLaTeX, metadata)
	require.NoError(t, err)

	assert.Equal(t, len(content), template.GetSize())
}

func TestTemplate_Equals(t *testing.T) {
	templateID1, err := NewTemplateID("template-123")
	require.NoError(t, err)

	templatePath1, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	metadata1 := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template1, err := NewTemplate(templateID1, templatePath1, "Hello {{.Name}}", TemplateTypeLaTeX, metadata1)
	require.NoError(t, err)

	templateID2, err := NewTemplateID("template-123")
	require.NoError(t, err)

	templatePath2, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	metadata2 := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template2, err := NewTemplate(templateID2, templatePath2, "Hello {{.Name}}", TemplateTypeLaTeX, metadata2)
	require.NoError(t, err)

	templateID3, err := NewTemplateID("template-456")
	require.NoError(t, err)

	metadata3 := NewTemplateMetadata("Test Author", "Test template", TemplateTypeLaTeX)
	template3, err := NewTemplate(templateID3, templatePath1, "Hello {{.Name}}", TemplateTypeLaTeX, metadata3)
	require.NoError(t, err)

	assert.True(t, template1.Equals(template2))
	assert.False(t, template1.Equals(template3))
}
