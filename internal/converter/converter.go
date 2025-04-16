package converter

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// Converter handles PDF to image conversion
type Converter struct {
	Config *config.Config
}

// NewConverter creates a new converter
func NewConverter(cfg *config.Config) *Converter {
	return &Converter{Config: cfg}
}

// ConvertPDFToImages converts a PDF file to the specified image formats
func (c *Converter) ConvertPDFToImages(pdfFile string) ([]string, error) {
	if pdfFile == "" {
		return nil, errors.New("no PDF file specified")
	}

	pdfFile = fmt.Sprintf("%s.pdf", pdfFile)
	// Check if the file exists
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("PDF file does not exist: %s", pdfFile)
	}

	// Check if conversion is enabled in the config
	if !c.Config.Conversion.Enabled {
		return nil, nil // Conversion is disabled, return without error
	}

	// Get the formats for conversion
	formats := c.Config.Conversion.Formats
	if len(formats) == 0 {
		formats = []string{"png"} // Default format
	}

	outputFiles := []string{}
	dir := filepath.Dir(pdfFile)
	baseName := strings.TrimSuffix(filepath.Base(pdfFile), filepath.Ext(pdfFile))

	// Try to find ImageMagick's convert tool
	_, err := exec.LookPath("convert")
	if err == nil {
		// ImageMagick approach
		for _, format := range formats {
			outputPath := filepath.Join(dir, fmt.Sprintf("%s.%s", baseName, format))
			cmd := exec.Command("convert",
				"-density", "300",
				pdfFile,
				outputPath)

			if err := cmd.Run(); err != nil {
				return outputFiles, fmt.Errorf("image conversion failed for %s: %w", format, err)
			}

			outputFiles = append(outputFiles, outputPath)
		}

		return outputFiles, nil
	}

	// Try pdftoppm for PDF to image conversion (part of poppler-utils)
	_, err = exec.LookPath("pdftoppm")
	if err == nil {
		for _, format := range formats {
			// pdftoppm has specific flags for different formats
			var outputPrefix string
			var args []string

			switch format {
			case "png":
				outputPrefix = filepath.Join(dir, baseName)
				args = []string{"-png", pdfFile, outputPrefix}
			case "jpg", "jpeg":
				outputPrefix = filepath.Join(dir, baseName)
				args = []string{"-jpeg", pdfFile, outputPrefix}
			default:
				continue // Skip unsupported formats
			}

			cmd := exec.Command("pdftoppm", args...)

			if err := cmd.Run(); err != nil {
				return outputFiles, fmt.Errorf("image conversion failed for %s: %w", format, err)
			}

			// pdftoppm adds numeric suffixes to outputs for multi-page documents
			// For simplicity, we'll just note the base name
			outputFiles = append(outputFiles, fmt.Sprintf("%s-*.%s", outputPrefix, format))
		}

		return outputFiles, nil
	}

	return nil, errors.New("no suitable conversion tool found (tried 'convert' and 'pdftoppm')")
}
