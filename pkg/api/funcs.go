package api

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/internal/tex"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/rwxrob/bonzai/futil"
	"gopkg.in/yaml.v3"
)

func GeneratePDF(cfg *config.Config, template config.Template) ([]byte, map[string]string, error) {
	defaultCfg := config.GetDefaultConfig()
	if cfg.Template == "" {
		cfg.Template = template
	}
	if cfg.Variables == nil {
		cfg.Variables = defaultCfg.Variables
	}
	if cfg.Engine == "" {
		cfg.Engine = defaultCfg.Engine
	}
	tmpDir := os.TempDir()
	if cfg.Output == "" {
		// No output file provided, generate a temporary one.
		tmpOutDir := filepath.Join(tmpDir, "out")
		err := os.MkdirAll(tmpOutDir, 0755)
		if err != nil {
			return nil, nil, err
		}
		cfg.Output = config.Output(filepath.Join(tmpOutDir, "output.pdf"))
	}
	log.Println("Output file:", cfg.Output)
	log.Println("Final merged config:", cfg)

	// Create a temporary config.yaml file, with the merged config
	writer, err := os.Create(filepath.Join(tmpDir, "config.yaml"))
	if err != nil {
		return nil, nil, err
	}
	defer writer.Close()

	log.Println("cfg:", cfg)
	// Encode the merged config to the temporary config.yaml file
	err = yaml.NewEncoder(writer).Encode(cfg)
	if err != nil {
		return nil, nil, err
	}

	// Build the pdf using the merged config
	err = tex.BuildCmd.Do(nil, cfg.Template.String(), writer.Name())
	if err != nil {
		// First check if output file exists before trying to read it
		if _, statErr := os.Stat(cfg.Output.String()); os.IsNotExist(statErr) {
			return nil, nil, err
		}

		// Now check if file is empty
		if futil.FileIsEmpty(cfg.Output.String()) {
			return nil, nil, err
		}

		// If we get here, file exists and has content despite LaTeX errors
		log.Printf("Warning: LaTeX reported errors but a PDF was produced: %v", err)
	}


	// Verify the file exists before attempting to read it
	if _, statErr := os.Stat(cfg.Output.String()); os.IsNotExist(statErr) {
		return nil, nil, os.ErrNotExist
	}

	// Read the generated pdf
	pdfBytes, err := os.ReadFile(cfg.Output.String())
	if err != nil {
		return nil, nil, err
	}

	// Verify that the file is not empty and contains valid PDF data
	if len(pdfBytes) == 0 {
		return nil, nil, os.ErrInvalid
	}

	// Basic check for PDF header signature
	if len(pdfBytes) < 5 || string(pdfBytes[0:5]) != "%PDF-" {
		return nil, nil, os.ErrInvalid
	}

	paths := make(map[string]string)
	if cfg.Conversion.Enabled {
		for _, format := range cfg.Conversion.Formats {
			file := convertToFormat(cfg.Output.String(), format)
			bts, err := os.ReadFile(file)
			if err != nil || len(bts) == 0 || len(bts) < 5 {
				return nil, nil, err
			}
			paths[format] = file
		}
	}

	// Return the generated pdf
	return pdfBytes, paths, nil
}

func convertToFormat(file string, format string) string {
	dir := filepath.Dir(file)
	filename := filepath.Base(file)
	ext := filepath.Ext(filename)
	newFilename := strings.TrimSuffix(filename, ext) + "." + format
	return filepath.Join(dir, newFilename)
}
