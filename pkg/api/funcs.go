package api

import (
	"log"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/internal/tex"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/rwxrob/bonzai/futil"
	"gopkg.in/yaml.v3"
)

func GeneratePDF(cfg *config.Config, template config.Template) ([]byte, error) {
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
			return nil, err
		}
		cfg.Output = config.Output(filepath.Join(tmpOutDir, "output.pdf"))
	}
	log.Println("Output file:", cfg.Output)
	log.Println("Final merged config:", cfg)

	// Create a temporary config.yaml file, with the merged config
	writer, err := os.Create(filepath.Join(tmpDir, "config.yaml"))
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	log.Println("cfg:", cfg)
	// Encode the merged config to the temporary config.yaml file
	err = yaml.NewEncoder(writer).Encode(cfg)
	if err != nil {
		return nil, err
	}

	// Build the pdf using the merged config
	err = tex.BuildCmd.Do(nil, cfg.Template.String(), writer.Name())
	if err != nil {
		if futil.FileIsEmpty(cfg.Output.String()) {
			return nil, err
		}
		// Normal to LaTeX to send build verbose info to stderr
		// so we need to return the error only if the output file is empty
		//
		// To print the verbose "error" info, we need to print the stderr
		// log.Println("Stderr while building pdf:", err)
	}

	// Read the generated pdf
	pdfBytes, err := os.ReadFile(cfg.Output.String())
	if err != nil {
		return nil, err
	}

	// Return the generated pdf
	return pdfBytes, nil
}
