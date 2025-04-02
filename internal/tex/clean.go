// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/futil"
)

// Cleaner handles the removal of auxiliary LaTeX files
type Cleaner struct {
	Directory string
}

// NewCleaner creates a new cleaner for LaTeX auxiliary files
func NewCleaner(directory string) *Cleaner {
	return &Cleaner{Directory: directory}
}

// Extensions to be considered auxiliary files
var auxiliaryExtensions = []string{
	".aux", ".log", ".toc", ".lof", ".lot", ".out", ".nav", ".snm",
	".synctex.gz", ".fls", ".fdb_latexmk", ".bbl", ".blg", ".run.xml",
	".bcf", ".idx", ".ilg", ".ind", ".brf", ".vrb", ".xdv", ".dvi",
}

func isAuxFile(filename string) bool {
	for _, ext := range auxiliaryExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

func removeAuxFiles(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// Skip directories
	if info.IsDir() {
		return nil
	}

	// Remove file if it has an auxiliary extension
	if isAuxFile(info.Name()) {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
		log.Printf("Removed: %s", path)
	}

	return nil
}

// Clean removes all auxiliary files in the specified directory
func (c *Cleaner) Clean() error {
	// Check if directory exists
	if !futil.IsDir(c.Directory) {
		return fmt.Errorf("directory does not exist: %s", c.Directory)
	}

	// Walk through directory and remove auxiliary files
	err := filepath.Walk(c.Directory, removeAuxFiles)

	if err != nil {
		return fmt.Errorf("error cleaning auxiliary files: %w", err)
	}

	return nil
}

// CleanCmd is the bonzai command for cleaning auxiliary LaTeX files
var CleanCmd = &bonzai.Cmd{
	Name:  `clean`,
	Short: `remove LaTeX auxiliary files`,
	Long: `
The clean command removes auxiliary files created during LaTeX compilation.
These include .aux, .log, .toc, and other temporary files.

By default, it cleans the current directory. You can specify a different 
directory as an argument.
`,
	Do: func(caller *bonzai.Cmd, args ...string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		cleaner := NewCleaner(dir)
		if err := cleaner.Clean(); err != nil {
			return err
		}

		fmt.Printf("Successfully cleaned auxiliary files in: %s\n", dir)
		return nil
	},
}
