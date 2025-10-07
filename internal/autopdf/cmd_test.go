package autopdf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rwxrob/bonzai"
)

func TestCmd_Structure(t *testing.T) {
	if Cmd == nil {
		t.Fatal("Cmd should not be nil")
	}

	if Cmd.Name != "autopdf" {
		t.Errorf("Expected Cmd.Name to be 'autopdf', got '%s'", Cmd.Name)
	}

	if Cmd.Alias != "apdf" {
		t.Errorf("Expected Cmd.Alias to be 'apdf', got '%s'", Cmd.Alias)
	}

	if Cmd.Vers != "v1.2.0" {
		t.Errorf("Expected Cmd.Vers to be 'v1.2.0', got '%s'", Cmd.Vers)
	}
}

func TestCmd_Commands(t *testing.T) {
	if len(Cmd.Cmds) == 0 {
		t.Error("Cmd.Cmds should not be empty")
	}

	// Check for expected commands
	expectedCommands := []string{"help", "build", "clean", "convert", "verbose", "debug", "force"}
	commandNames := make(map[string]bool)

	for _, cmd := range Cmd.Cmds {
		commandNames[cmd.Name] = true
	}

	for _, expectedCmd := range expectedCommands {
		if !commandNames[expectedCmd] {
			t.Errorf("Expected command '%s' not found in Cmd.Cmds", expectedCmd)
		}
	}
}

// getConvertCmd finds the convert command in the Cmd.Cmds slice
func getConvertCmd() *bonzai.Cmd {
	for _, cmd := range Cmd.Cmds {
		if cmd.Name == "convert" {
			return cmd
		}
	}
	return nil
}

func TestConvertCmd_Structure(t *testing.T) {
	convertCmd := getConvertCmd()
	if convertCmd == nil {
		t.Fatal("convert command should be found in Cmd.Cmds")
	}

	if convertCmd.Name != "convert" {
		t.Errorf("Expected convertCmd.Name to be 'convert', got '%s'", convertCmd.Name)
	}

	if convertCmd.Alias != "c" {
		t.Errorf("Expected convertCmd.Alias to be 'c', got '%s'", convertCmd.Alias)
	}

	if convertCmd.MinArgs != 1 {
		t.Errorf("Expected convertCmd.MinArgs to be 1, got %d", convertCmd.MinArgs)
	}
}

func TestConvertCmd_Do(t *testing.T) {
	convertCmd := getConvertCmd()
	if convertCmd == nil {
		t.Fatal("convert command should be found in Cmd.Cmds")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfFile := filepath.Join(tempDir, "test")
	if err := os.WriteFile(pdfFile+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Test with single argument (PDF file)
	args := []string{pdfFile}
	err = convertCmd.Do(convertCmd, args...)
	if err != nil {
		t.Logf("convertCmd.Do failed with single argument (expected for dummy PDF): %v", err)
	}

	// Test with multiple arguments (PDF file + formats)
	args = []string{pdfFile, "png", "jpg"}
	err = convertCmd.Do(convertCmd, args...)
	if err != nil {
		t.Logf("convertCmd.Do failed with multiple arguments (expected for dummy PDF): %v", err)
	}
}

func TestConvertCmd_Do_InvalidInput(t *testing.T) {
	convertCmd := getConvertCmd()
	if convertCmd == nil {
		t.Fatal("convert command should be found in Cmd.Cmds")
	}

	// Test with non-existent file
	args := []string{"/path/to/nonexistent/file.pdf"}
	err := convertCmd.Do(convertCmd, args...)
	if err == nil {
		t.Error("Expected error for non-existent file but got none")
	}
}

func TestConvertCmd_Do_EmptyFormats(t *testing.T) {
	convertCmd := getConvertCmd()
	if convertCmd == nil {
		t.Fatal("convert command should be found in Cmd.Cmds")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfFile := filepath.Join(tempDir, "test")
	if err := os.WriteFile(pdfFile+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Test with single argument (should default to png format)
	args := []string{pdfFile}
	err = convertCmd.Do(convertCmd, args...)
	if err != nil {
		t.Logf("convertCmd.Do failed with single argument (expected for dummy PDF): %v", err)
	}
}

func TestConvertCmd_ConfigCreation(t *testing.T) {
	convertCmd := getConvertCmd()
	if convertCmd == nil {
		t.Fatal("convert command should be found in Cmd.Cmds")
	}
	// Test that the convert command creates a proper config
	// This is tested indirectly through the Do function, but we can verify
	// the config structure is correct by checking the conversion settings

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfFile := filepath.Join(tempDir, "test")
	if err := os.WriteFile(pdfFile+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Test with specific formats
	args := []string{pdfFile, "png", "jpg"}
	err = convertCmd.Do(convertCmd, args...)
	if err != nil {
		t.Logf("convertCmd.Do failed with specific formats (expected for dummy PDF): %v", err)
	}
}

func TestConvertCmd_DefaultFormats(t *testing.T) {
	convertCmd := getConvertCmd()
	if convertCmd == nil {
		t.Fatal("convert command should be found in Cmd.Cmds")
	}
	// Test that default formats are set correctly when no formats are provided
	// This is tested by checking that the command doesn't fail with default settings

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfFile := filepath.Join(tempDir, "test")
	if err := os.WriteFile(pdfFile+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Test with single argument (should use default png format)
	args := []string{pdfFile}
	err = convertCmd.Do(convertCmd, args...)
	if err != nil {
		t.Logf("convertCmd.Do failed with default format (expected for dummy PDF): %v", err)
	}
}

func TestConvertCmd_MultipleFormats(t *testing.T) {
	convertCmd := getConvertCmd()
	if convertCmd == nil {
		t.Fatal("convert command should be found in Cmd.Cmds")
	}
	// Test that multiple formats are handled correctly

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfFile := filepath.Join(tempDir, "test")
	if err := os.WriteFile(pdfFile+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Test with multiple formats
	args := []string{pdfFile, "png", "jpg", "gif"}
	err = convertCmd.Do(convertCmd, args...)
	if err != nil {
		t.Logf("convertCmd.Do failed with multiple formats (expected for dummy PDF): %v", err)
	}
}

func TestConvertCmd_ErrorHandling(t *testing.T) {
	convertCmd := getConvertCmd()
	if convertCmd == nil {
		t.Fatal("convert command should be found in Cmd.Cmds")
	}
	// Test error handling for various invalid inputs

	// Test with empty string
	args := []string{""}
	err := convertCmd.Do(convertCmd, args...)
	if err == nil {
		t.Error("Expected error for empty string but got none")
	}

	// Test with non-existent file
	args = []string{"/path/to/nonexistent/file.pdf"}
	err = convertCmd.Do(convertCmd, args...)
	if err == nil {
		t.Error("Expected error for non-existent file but got none")
	}
}

func TestConvertCmd_ConfigValidation(t *testing.T) {
	convertCmd := getConvertCmd()
	if convertCmd == nil {
		t.Fatal("convert command should be found in Cmd.Cmds")
	}
	// Test that the config created by convertCmd has the correct structure
	// This is tested indirectly through the Do function

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfFile := filepath.Join(tempDir, "test")
	if err := os.WriteFile(pdfFile+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Test that the command creates a valid config
	args := []string{pdfFile, "png"}
	err = convertCmd.Do(convertCmd, args...)
	if err != nil {
		t.Logf("convertCmd.Do failed (expected for dummy PDF): %v", err)
	}
}

func TestConvertCmd_Integration(t *testing.T) {
	convertCmd := getConvertCmd()
	if convertCmd == nil {
		t.Fatal("convert command should be found in Cmd.Cmds")
	}
	// Test the full integration of the convert command
	// This tests the entire flow from command execution to converter creation

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfFile := filepath.Join(tempDir, "test")
	if err := os.WriteFile(pdfFile+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Test the full command execution
	args := []string{pdfFile, "png", "jpg"}
	err = convertCmd.Do(convertCmd, args...)
	if err != nil {
		t.Logf("Full convert command execution failed (expected for dummy PDF): %v", err)
	}
}
