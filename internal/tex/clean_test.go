package tex

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCleaner(t *testing.T) {
	dir := "/test/dir"
	cleaner := NewCleaner(dir)

	if cleaner == nil {
		t.Fatalf("NewCleaner returned nil")
	}

	if cleaner.Directory != dir {
		t.Errorf("Expected Directory to be '%s', got '%s'", dir, cleaner.Directory)
	}
}

func TestClean_InvalidInput(t *testing.T) {
	// Test with empty directory
	cleaner := NewCleaner("")
	err := cleaner.Clean()
	if err == nil {
		t.Errorf("Expected error for empty directory but got none")
	}

	// Test with non-existent directory
	cleaner = NewCleaner("/path/to/nonexistent/dir")
	err = cleaner.Clean()
	if err == nil {
		t.Errorf("Expected error for non-existent directory but got none")
	}
}

func TestClean(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create sample auxiliary files
	auxFiles := map[string]string{
		"document.aux":        "aux content",
		"document.log":        "log content",
		"document.toc":        "toc content",
		"document.synctex.gz": "synctex content",
	}

	// Create a non-auxiliary file that should not be deleted
	regularFile := filepath.Join(tempDir, "document.tex")
	if err := ioutil.WriteFile(regularFile, []byte("LaTeX content"), 0644); err != nil {
		t.Fatalf("Failed to write regular file: %v", err)
	}

	// Create the auxiliary files
	for filename, content := range auxFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write auxiliary file %s: %v", filename, err)
		}
	}

	// Create a cleaner and run it
	cleaner := NewCleaner(tempDir)
	err = cleaner.Clean()
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	// Check that auxiliary files were deleted
	for filename := range auxFiles {
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("Auxiliary file %s should have been deleted but still exists", filename)
		}
	}

	// Check that the regular file was not deleted
	if _, err := os.Stat(regularFile); os.IsNotExist(err) {
		t.Errorf("Regular file %s should not have been deleted", regularFile)
	}
}

func TestClean_Subdirectories(t *testing.T) {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create auxiliary files in main directory and subdirectory
	auxFiles := []struct {
		dir      string
		filename string
	}{
		{tempDir, "main.aux"},
		{tempDir, "main.log"},
		{subDir, "sub.aux"},
		{subDir, "sub.log"},
	}

	// Create the auxiliary files
	for _, file := range auxFiles {
		filePath := filepath.Join(file.dir, file.filename)
		if err := ioutil.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to write auxiliary file %s: %v", filePath, err)
		}
	}

	// Create a cleaner and run it (should clean both directories)
	cleaner := NewCleaner(tempDir)
	err = cleaner.Clean()
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	// Check that all auxiliary files were deleted
	for _, file := range auxFiles {
		filePath := filepath.Join(file.dir, file.filename)
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("Auxiliary file %s should have been deleted but still exists", filePath)
		}
	}
}
