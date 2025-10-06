package domain

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplatePath_NewTemplatePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TemplatePath
		wantErr error
	}{
		{
			name:    "valid template path",
			input:   "template.tex",
			want:    TemplatePath{value: "template.tex"},
			wantErr: nil,
		},
		{
			name:    "valid template path with directory",
			input:   "templates/document.tex",
			want:    TemplatePath{value: "templates/document.tex"},
			wantErr: nil,
		},
		{
			name:    "valid template path with spaces",
			input:   "  template.tex  ",
			want:    TemplatePath{value: "template.tex"},
			wantErr: nil,
		},
		{
			name:    "valid template path with absolute path",
			input:   "/home/user/template.tex",
			want:    TemplatePath{value: "/home/user/template.tex"},
			wantErr: nil,
		},
		{
			name:    "empty path",
			input:   "",
			want:    TemplatePath{},
			wantErr: ErrEmptyPath,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			want:    TemplatePath{},
			wantErr: ErrEmptyPath,
		},
		{
			name:    "invalid extension",
			input:   "template.txt",
			want:    TemplatePath{},
			wantErr: ErrInvalidTemplateExtension,
		},
		{
			name:    "no extension",
			input:   "template",
			want:    TemplatePath{},
			wantErr: ErrInvalidTemplateExtension,
		},
		{
			name:    "uppercase extension",
			input:   "template.TEX",
			want:    TemplatePath{value: "template.TEX"},
			wantErr: nil,
		},
		{
			name:    "path with invalid characters - less than",
			input:   "template<.tex",
			want:    TemplatePath{},
			wantErr: ErrInvalidPath,
		},
		{
			name:    "path with invalid characters - greater than",
			input:   "template>.tex",
			want:    TemplatePath{},
			wantErr: ErrInvalidPath,
		},
		{
			name:    "path with invalid characters - pipe",
			input:   "template|.tex",
			want:    TemplatePath{},
			wantErr: ErrInvalidPath,
		},
		{
			name:    "path with invalid characters - asterisk",
			input:   "template*.tex",
			want:    TemplatePath{},
			wantErr: ErrInvalidPath,
		},
		{
			name:    "path with invalid characters - question mark",
			input:   "template?.tex",
			want:    TemplatePath{},
			wantErr: ErrInvalidPath,
		},
		{
			name:    "path with invalid characters - quotes",
			input:   "template\".tex",
			want:    TemplatePath{},
			wantErr: ErrInvalidPath,
		},
		{
			name:    "path too long",
			input:   strings.Repeat("a", 256) + ".tex",
			want:    TemplatePath{},
			wantErr: ErrPathTooLong,
		},
		{
			name:    "path with dots",
			input:   "template..tex",
			want:    TemplatePath{value: "template..tex"},
			wantErr: nil,
		},
		{
			name:    "path with hyphens",
			input:   "template-file.tex",
			want:    TemplatePath{value: "template-file.tex"},
			wantErr: nil,
		},
		{
			name:    "path with underscores",
			input:   "template_file.tex",
			want:    TemplatePath{value: "template_file.tex"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTemplatePath(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewTemplatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewTemplatePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplatePath_String(t *testing.T) {
	path, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	assert.Equal(t, "template.tex", path.String())
}

func TestTemplatePath_Value(t *testing.T) {
	path, err := NewTemplatePath("templates/document.tex")
	require.NoError(t, err)

	assert.Equal(t, "templates/document.tex", path.Value())
}

func TestTemplatePath_IsEmpty(t *testing.T) {
	path, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	assert.False(t, path.IsEmpty())

	emptyPath := TemplatePath{}
	assert.True(t, emptyPath.IsEmpty())
}

func TestTemplatePath_Extension(t *testing.T) {
	path, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	assert.Equal(t, ".tex", path.Extension())
}

func TestTemplatePath_BaseName(t *testing.T) {
	path, err := NewTemplatePath("templates/document.tex")
	require.NoError(t, err)

	assert.Equal(t, "document.tex", path.BaseName())
}

func TestTemplatePath_Dir(t *testing.T) {
	path, err := NewTemplatePath("templates/document.tex")
	require.NoError(t, err)

	assert.Equal(t, "templates", path.Dir())
}

func TestTemplatePath_Equals(t *testing.T) {
	path1, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	path2, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	path3, err := NewTemplatePath("other.tex")
	require.NoError(t, err)

	assert.True(t, path1.Equals(path2))
	assert.False(t, path1.Equals(path3))
}

func TestTemplatePath_IsAbsolute(t *testing.T) {
	// Test relative path
	relativePath, err := NewTemplatePath("template.tex")
	require.NoError(t, err)
	assert.False(t, relativePath.IsAbsolute())

	// Test absolute path
	absolutePath, err := NewTemplatePath("/home/user/template.tex")
	require.NoError(t, err)
	assert.True(t, absolutePath.IsAbsolute())
}

func TestTemplatePath_Join(t *testing.T) {
	path, err := NewTemplatePath("templates/base.tex")
	require.NoError(t, err)

	joined, err := path.Join("document.tex")
	require.NoError(t, err)

	assert.Equal(t, "templates/base.tex/document.tex", joined.String())
}

func TestTemplatePath_Join_InvalidExtension(t *testing.T) {
	path, err := NewTemplatePath("templates/base.tex")
	require.NoError(t, err)

	_, err = path.Join("document.txt")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidTemplateExtension, err)
}

func TestTemplatePath_MarshalJSON(t *testing.T) {
	path, err := NewTemplatePath("template.tex")
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(path)
	require.NoError(t, err)

	assert.Equal(t, `"template.tex"`, string(jsonBytes))
}

func TestTemplatePath_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    TemplatePath
		wantErr bool
	}{
		{
			name:    "valid JSON",
			json:    `"template.tex"`,
			want:    TemplatePath{value: "template.tex"},
			wantErr: false,
		},
		{
			name:    "invalid JSON - empty",
			json:    `""`,
			want:    TemplatePath{},
			wantErr: true,
		},
		{
			name:    "invalid JSON - wrong extension",
			json:    `"template.txt"`,
			want:    TemplatePath{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TemplatePath
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTemplatePath_EdgeCases(t *testing.T) {
	t.Run("path with multiple dots", func(t *testing.T) {
		path, err := NewTemplatePath("template.backup.tex")
		require.NoError(t, err)
		assert.Equal(t, "template.backup.tex", path.String())
	})

	t.Run("path with spaces in directory", func(t *testing.T) {
		path, err := NewTemplatePath("my templates/document.tex")
		require.NoError(t, err)
		assert.Equal(t, "my templates/document.tex", path.String())
	})

	t.Run("path with special characters in directory", func(t *testing.T) {
		path, err := NewTemplatePath("templates-2024/document.tex")
		require.NoError(t, err)
		assert.Equal(t, "templates-2024/document.tex", path.String())
	})
}

func TestTemplatePath_PathOperations(t *testing.T) {
	t.Run("clean path with dots", func(t *testing.T) {
		path, err := NewTemplatePath("./templates/../templates/document.tex")
		require.NoError(t, err)
		assert.Equal(t, "templates/document.tex", path.String())
	})

	t.Run("path with trailing slash", func(t *testing.T) {
		_, err := NewTemplatePath("templates/document.tex/")
		require.Error(t, err)
		assert.Equal(t, ErrInvalidTemplateExtension, err)
	})
}

func TestTemplatePath_PropertyBased(t *testing.T) {
	// Test that valid paths are always cleaned
	validPaths := []string{
		"  template.tex  ",
		"./template.tex",
		"templates//document.tex",
		"templates/./document.tex",
	}

	for _, path := range validPaths {
		t.Run("cleaned_"+path, func(t *testing.T) {
			templatePath, err := NewTemplatePath(path)
			require.NoError(t, err)
			assert.Equal(t, filepath.Clean(strings.TrimSpace(path)), templatePath.String())
		})
	}
}
