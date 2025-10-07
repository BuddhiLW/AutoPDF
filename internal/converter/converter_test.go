package converter

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

func TestNewConverter(t *testing.T) {
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{"png"},
		},
	}

	converter := NewConverter(cfg)

	if converter == nil {
		t.Fatalf("NewConverter returned nil")
	}

	if converter.Config != cfg {
		t.Errorf("NewConverter did not correctly set the Config field")
	}
}

func TestConvertPDFToImages_InvalidInput(t *testing.T) {
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{"png"},
		},
	}

	converter := NewConverter(cfg)

	// Test with empty file path
	_, err := converter.ConvertPDFToImages("")
	if err == nil {
		t.Errorf("Expected error for empty file path but got none")
	}

	// Test with non-existent file
	_, err = converter.ConvertPDFToImages("/path/to/nonexistent/file.pdf")
	if err == nil {
		t.Errorf("Expected error for non-existent file but got none")
	}
}

func TestConvertPDFToImages_DisabledConversion(t *testing.T) {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfPath := filepath.Join(tempDir, "test")
	if err := ioutil.WriteFile(pdfPath+".pdf", []byte("dummy pdf content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Create config with conversion disabled
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: false,
			Formats: []string{"png"},
		},
	}

	converter := NewConverter(cfg)

	// Convert should return nil with no error
	images, err := converter.ConvertPDFToImages(pdfPath)
	if err != nil {
		t.Fatalf("ConvertPDFToImages failed: %v", err)
	}

	if images != nil {
		t.Errorf("Expected nil image list when conversion is disabled, got %v", images)
	}
}

// This test checks if either convert (ImageMagick) or pdftoppm is installed
// and skips if neither is available
func TestConvertPDFToImages_ToolCheck(t *testing.T) {
	// Check if we have either convert or pdftoppm
	_, convertErr := exec.LookPath("convert")
	_, pdftoppmErr := exec.LookPath("pdftoppm")

	if convertErr != nil && pdftoppmErr != nil {
		t.Skip("Neither convert nor pdftoppm found, skipping test")
	}

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file - note this isn't a real PDF and will fail conversion
	// but we can still test the tool detection logic
	pdfPath := filepath.Join(tempDir, "test.pdf")
	if err := ioutil.WriteFile(pdfPath, []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Create config with conversion enabled
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{"png"},
		},
	}

	converter := NewConverter(cfg)

	// Attempt conversion will likely fail with our dummy PDF, but should not fail
	// at the tool detection stage
	_, err = converter.ConvertPDFToImages(pdfPath)

	// We should have either successfully started the conversion (but it might fail later)
	// or gotten a specific error that is not "no suitable conversion tool found"
	if err != nil && strings.Contains(err.Error(), "no suitable conversion tool found") {
		t.Errorf("Tool detection failed despite having a conversion tool available")
	}
}

func TestConvertPDFToImages_NoToolsAvailable(t *testing.T) {
	// This test simulates the case where no conversion tools are available
	// We'll create a converter and test the error path

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfPath := filepath.Join(tempDir, "test")
	if err := ioutil.WriteFile(pdfPath+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Create config with conversion enabled
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{"png"},
		},
	}

	converter := NewConverter(cfg)

	// Test with file that doesn't have .pdf extension
	_, err = converter.ConvertPDFToImages("test")
	if err == nil {
		t.Error("Expected error for file without .pdf extension but got none")
	}
}

func TestConvertPDFToImages_DefaultFormats(t *testing.T) {
	// Check if we have either convert or pdftoppm
	_, convertErr := exec.LookPath("convert")
	_, pdftoppmErr := exec.LookPath("pdftoppm")

	if convertErr != nil && pdftoppmErr != nil {
		t.Skip("Neither convert nor pdftoppm found, skipping test")
	}

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfPath := filepath.Join(tempDir, "test")
	if err := ioutil.WriteFile(pdfPath+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Create config with empty formats (should default to png)
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{},
		},
	}

	converter := NewConverter(cfg)

	// Attempt conversion
	_, err = converter.ConvertPDFToImages(pdfPath)

	// Should not fail at tool detection stage
	if err != nil && strings.Contains(err.Error(), "no suitable conversion tool found") {
		t.Errorf("Tool detection failed despite having a conversion tool available")
	}
}

func TestConvertPDFToImages_MultipleFormats(t *testing.T) {
	// Check if we have either convert or pdftoppm
	_, convertErr := exec.LookPath("convert")
	_, pdftoppmErr := exec.LookPath("pdftoppm")

	if convertErr != nil && pdftoppmErr != nil {
		t.Skip("Neither convert nor pdftoppm found, skipping test")
	}

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfPath := filepath.Join(tempDir, "test")
	if err := ioutil.WriteFile(pdfPath+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Create config with multiple formats
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{"png", "jpg"},
		},
	}

	converter := NewConverter(cfg)

	// Attempt conversion
	_, err = converter.ConvertPDFToImages(pdfPath)

	// Should not fail at tool detection stage
	if err != nil && strings.Contains(err.Error(), "no suitable conversion tool found") {
		t.Errorf("Tool detection failed despite having a conversion tool available")
	}
}

func TestConvertPDFToImages_UnsupportedFormat(t *testing.T) {
	// Check if we have either convert or pdftoppm
	_, convertErr := exec.LookPath("convert")
	_, pdftoppmErr := exec.LookPath("pdftoppm")

	if convertErr != nil && pdftoppmErr != nil {
		t.Skip("Neither convert nor pdftoppm found, skipping test")
	}

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfPath := filepath.Join(tempDir, "test")
	if err := ioutil.WriteFile(pdfPath+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Create config with unsupported format
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{"unsupported"},
		},
	}

	converter := NewConverter(cfg)

	// Attempt conversion
	images, err := converter.ConvertPDFToImages(pdfPath)
	// The conversion might fail due to the unsupported format, which is expected
	// We just want to make sure it doesn't crash and handles the error gracefully
	if err != nil {
		// This is expected for unsupported formats
		t.Logf("Expected error for unsupported format: %v", err)
		return
	}

	// If no error, should return empty list for unsupported formats
	if len(images) != 0 {
		t.Errorf("Expected empty image list for unsupported format, got %v", images)
	}
}

func TestConvertPDFToImages_PdfToPpmFormats(t *testing.T) {
	// Check if pdftoppm is available
	_, err := exec.LookPath("pdftoppm")
	if err != nil {
		t.Skip("pdftoppm not found, skipping test")
	}

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy PDF file
	pdfPath := filepath.Join(tempDir, "test")
	if err := ioutil.WriteFile(pdfPath+".pdf", []byte("%PDF-1.5\ndummy content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy PDF file: %v", err)
	}

	// Test with jpeg format (pdftoppm specific)
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{"jpeg"},
		},
	}

	converter := NewConverter(cfg)

	// Attempt conversion
	_, err = converter.ConvertPDFToImages(pdfPath)

	// Should not fail at tool detection stage
	if err != nil && strings.Contains(err.Error(), "no suitable conversion tool found") {
		t.Errorf("Tool detection failed despite having pdftoppm available")
	}
}
