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

func TestOutputPath_NewOutputPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    OutputPath
		wantErr error
	}{
		{
			name:    "valid output path",
			input:   "output.pdf",
			want:    OutputPath{value: "output.pdf"},
			wantErr: nil,
		},
		{
			name:    "valid output path with directory",
			input:   "outputs/document.pdf",
			want:    OutputPath{value: "outputs/document.pdf"},
			wantErr: nil,
		},
		{
			name:    "valid output path with spaces",
			input:   "  output.pdf  ",
			want:    OutputPath{value: "output.pdf"},
			wantErr: nil,
		},
		{
			name:    "valid output path with absolute path",
			input:   "/home/user/output.pdf",
			want:    OutputPath{value: "/home/user/output.pdf"},
			wantErr: nil,
		},
		{
			name:    "empty path",
			input:   "",
			want:    OutputPath{},
			wantErr: ErrEmptyOutputPath,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			want:    OutputPath{},
			wantErr: ErrEmptyOutputPath,
		},
		{
			name:    "invalid extension",
			input:   "output.txt",
			want:    OutputPath{},
			wantErr: ErrInvalidOutputExtension,
		},
		{
			name:    "no extension",
			input:   "output",
			want:    OutputPath{},
			wantErr: ErrInvalidOutputExtension,
		},
		{
			name:    "uppercase extension",
			input:   "output.PDF",
			want:    OutputPath{value: "output.PDF"},
			wantErr: nil,
		},
		{
			name:    "path with invalid characters - less than",
			input:   "output<.pdf",
			want:    OutputPath{},
			wantErr: ErrInvalidOutputPath,
		},
		{
			name:    "path with invalid characters - greater than",
			input:   "output>.pdf",
			want:    OutputPath{},
			wantErr: ErrInvalidOutputPath,
		},
		{
			name:    "path with invalid characters - pipe",
			input:   "output|.pdf",
			want:    OutputPath{},
			wantErr: ErrInvalidOutputPath,
		},
		{
			name:    "path with invalid characters - asterisk",
			input:   "output*.pdf",
			want:    OutputPath{},
			wantErr: ErrInvalidOutputPath,
		},
		{
			name:    "path with invalid characters - question mark",
			input:   "output?.pdf",
			want:    OutputPath{},
			wantErr: ErrInvalidOutputPath,
		},
		{
			name:    "path with invalid characters - quotes",
			input:   "output\".pdf",
			want:    OutputPath{},
			wantErr: ErrInvalidOutputPath,
		},
		{
			name:    "path too long",
			input:   strings.Repeat("a", 256) + ".pdf",
			want:    OutputPath{},
			wantErr: ErrOutputPathTooLong,
		},
		{
			name:    "path with dots",
			input:   "output..pdf",
			want:    OutputPath{value: "output..pdf"},
			wantErr: nil,
		},
		{
			name:    "path with hyphens",
			input:   "output-file.pdf",
			want:    OutputPath{value: "output-file.pdf"},
			wantErr: nil,
		},
		{
			name:    "path with underscores",
			input:   "output_file.pdf",
			want:    OutputPath{value: "output_file.pdf"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOutputPath(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewOutputPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewOutputPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputPath_String(t *testing.T) {
	path, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	assert.Equal(t, "output.pdf", path.String())
}

func TestOutputPath_Value(t *testing.T) {
	path, err := NewOutputPath("outputs/document.pdf")
	require.NoError(t, err)

	assert.Equal(t, "outputs/document.pdf", path.Value())
}

func TestOutputPath_IsEmpty(t *testing.T) {
	path, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	assert.False(t, path.IsEmpty())

	emptyPath := OutputPath{}
	assert.True(t, emptyPath.IsEmpty())
}

func TestOutputPath_Extension(t *testing.T) {
	path, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	assert.Equal(t, ".pdf", path.Extension())
}

func TestOutputPath_BaseName(t *testing.T) {
	path, err := NewOutputPath("outputs/document.pdf")
	require.NoError(t, err)

	assert.Equal(t, "document.pdf", path.BaseName())
}

func TestOutputPath_Dir(t *testing.T) {
	path, err := NewOutputPath("outputs/document.pdf")
	require.NoError(t, err)

	assert.Equal(t, "outputs", path.Dir())
}

func TestOutputPath_Equals(t *testing.T) {
	path1, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	path2, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	path3, err := NewOutputPath("other.pdf")
	require.NoError(t, err)

	assert.True(t, path1.Equals(path2))
	assert.False(t, path1.Equals(path3))
}

func TestOutputPath_IsAbsolute(t *testing.T) {
	// Test relative path
	relativePath, err := NewOutputPath("output.pdf")
	require.NoError(t, err)
	assert.False(t, relativePath.IsAbsolute())

	// Test absolute path
	absolutePath, err := NewOutputPath("/home/user/output.pdf")
	require.NoError(t, err)
	assert.True(t, absolutePath.IsAbsolute())
}

func TestOutputPath_Join(t *testing.T) {
	path, err := NewOutputPath("outputs/base.pdf")
	require.NoError(t, err)

	joined, err := path.Join("document.pdf")
	require.NoError(t, err)

	assert.Equal(t, "outputs/base.pdf/document.pdf", joined.String())
}

func TestOutputPath_Join_InvalidExtension(t *testing.T) {
	path, err := NewOutputPath("outputs/base.pdf")
	require.NoError(t, err)

	_, err = path.Join("document.txt")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidOutputExtension, err)
}

func TestOutputPath_WithExtension(t *testing.T) {
	path, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	// Test with valid extension
	newPath, err := path.WithExtension(".pdf")
	require.NoError(t, err)
	assert.Equal(t, "output.pdf", newPath.String())

	// Test with invalid extension
	_, err = path.WithExtension(".txt")
	assert.Error(t, err)
}

func TestOutputPath_MarshalJSON(t *testing.T) {
	path, err := NewOutputPath("output.pdf")
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(path)
	require.NoError(t, err)

	assert.Equal(t, `"output.pdf"`, string(jsonBytes))
}

func TestOutputPath_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    OutputPath
		wantErr bool
	}{
		{
			name:    "valid JSON",
			json:    `"output.pdf"`,
			want:    OutputPath{value: "output.pdf"},
			wantErr: false,
		},
		{
			name:    "invalid JSON - empty",
			json:    `""`,
			want:    OutputPath{},
			wantErr: true,
		},
		{
			name:    "invalid JSON - wrong extension",
			json:    `"output.txt"`,
			want:    OutputPath{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got OutputPath
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

func TestOutputPath_EdgeCases(t *testing.T) {
	t.Run("path with multiple dots", func(t *testing.T) {
		path, err := NewOutputPath("output.backup.pdf")
		require.NoError(t, err)
		assert.Equal(t, "output.backup.pdf", path.String())
	})

	t.Run("path with spaces in directory", func(t *testing.T) {
		path, err := NewOutputPath("my outputs/document.pdf")
		require.NoError(t, err)
		assert.Equal(t, "my outputs/document.pdf", path.String())
	})

	t.Run("path with special characters in directory", func(t *testing.T) {
		path, err := NewOutputPath("outputs-2024/document.pdf")
		require.NoError(t, err)
		assert.Equal(t, "outputs-2024/document.pdf", path.String())
	})
}

func TestOutputPath_PathOperations(t *testing.T) {
	t.Run("clean path with dots", func(t *testing.T) {
		path, err := NewOutputPath("./outputs/../outputs/document.pdf")
		require.NoError(t, err)
		assert.Equal(t, "outputs/document.pdf", path.String())
	})

	t.Run("path with trailing slash", func(t *testing.T) {
		_, err := NewOutputPath("outputs/document.pdf/")
		require.Error(t, err)
		assert.Equal(t, ErrInvalidOutputExtension, err)
	})
}

func TestOutputPath_PropertyBased(t *testing.T) {
	// Test that valid paths are always cleaned
	validPaths := []string{
		"  output.pdf  ",
		"./output.pdf",
		"outputs//document.pdf",
		"outputs/./document.pdf",
	}

	for _, path := range validPaths {
		t.Run("cleaned_"+path, func(t *testing.T) {
			outputPath, err := NewOutputPath(path)
			require.NoError(t, err)
			assert.Equal(t, filepath.Clean(strings.TrimSpace(path)), outputPath.String())
		})
	}
}
