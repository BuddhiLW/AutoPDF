// Copyright 2022 AutoPDF Pedro Branquinho
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/internal/config"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/futil"
)

// Compiler handles LaTeX compilation
type Compiler struct {
	Config *config.Config
}

// NewCompiler creates a new LaTeX compiler
func NewCompiler(cfg *config.Config) *Compiler {
	return &Compiler{Config: cfg}
}

// Compile processes a LaTeX file into PDF using the specified LaTeX engine
func (c *Compiler) Compile(texFile string) (string, error) {
	if texFile == "" {
		return "", errors.New("no LaTeX file specified")
	}

	// Check if the file exists
	if _, err := os.Stat(texFile); os.IsNotExist(err) {
		return "", fmt.Errorf("LaTeX file does not exist: %s", texFile)
	}

	// Determine the engine to use
	engine := c.Config.Engine
	if engine == "" {
		engine = "pdflatex" // Default engine
	}

	// Verify that the engine is installed
	if _, err := exec.LookPath(engine.String()); err != nil {
		return "", fmt.Errorf("LaTeX engine not found: %s", engine)
	}

	// Get the directory of the input file
	dir := filepath.Dir(texFile)
	baseName := filepath.Base(texFile)

	// Determine output PDF path
	outputPDF := filepath.Join(dir, replaceExt(baseName, ".pdf"))
	if c.Config.Output.String() != "" {
		outputPDF = c.Config.Output.String()
	}
	dirOutput := filepath.Dir(outputPDF)
	baseNameOutput := filepath.Base(outputPDF)

	// if dirOutput doesn't exist, create it
	err := futil.CreateDir(dirOutput)
	if err != nil {
		return "", fmt.Errorf("failed to create output directory: %s", err)
	}

	// Create command to run
	cmdStr := fmt.Sprintf("%s -interaction=nonstopmode -jobname=%s -output-directory=%s %s", engine, baseNameOutput, dirOutput, texFile)
	cmd := exec.Command("sh", "-c", cmdStr)
	log.Printf("Running command: %s", cmd.String())

	// Check if output PDF exists
	if _, err := os.Stat(fmt.Sprintf("%s.pdf", outputPDF)); os.IsNotExist(err) {
		return "", errors.New("PDF output file was not created")
	}

	return outputPDF, nil
}

// CompileWithBibtex runs LaTeX with BibTeX for bibliography processing
func (c *Compiler) CompileWithBibtex(texFile string) (string, error) {
	if texFile == "" {
		return "", errors.New("no LaTeX file specified")
	}

	dir := filepath.Dir(texFile)
	baseNameWithExt := filepath.Base(texFile)
	baseName := replaceExt(baseNameWithExt, "")

	// First LaTeX run
	cmd := exec.Command(c.Config.Engine.String(),
		"-interaction=nonstopmode",
		"-output-directory="+dir,
		texFile)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("first LaTeX run failed: %w", err)
	}

	// BibTeX run
	bibCmd := exec.Command("bibtex", filepath.Join(dir, baseName))
	if err := bibCmd.Run(); err != nil {
		return "", fmt.Errorf("BibTeX run failed: %w", err)
	}

	// Second LaTeX run
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("second LaTeX run failed: %w", err)
	}

	// Third LaTeX run
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("third LaTeX run failed: %w", err)
	}

	// Default output path
	outputPDF := filepath.Join(dir, baseName+".pdf")

	// If output path is set in config, use it instead of default
	if c.Config.Output.String() != "" {
		outputPDF = c.Config.Output.String()
	}

	// Check if output PDF exists
	if _, err := os.Stat(outputPDF); os.IsNotExist(err) {
		return "", errors.New("PDF output file was not created")
	}

	return outputPDF, nil
}

// Helper function to replace file extension
func replaceExt(filename, newExt string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return filename + newExt
	}
	return filename[:len(filename)-len(ext)] + newExt
}

// CompileCmd is the bonzai command for LaTeX compilation
var CompileCmd = &bonzai.Cmd{
	Name: `compile`,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		if len(args) == 0 {
			return errors.New("no LaTeX file specified")
		}

		texFile := args[0]

		// Create a default config for standalone use
		cfg := &config.Config{
			Engine: "pdflatex",
		}

		compiler := NewCompiler(cfg)
		outputPDF, err := compiler.Compile(texFile)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully compiled: %s\n", outputPDF)
		return nil
	},
}
