// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"path/filepath"
	"strings"
)

// PatternMatcherAdapter implements the FilePatternMatcher interface
type PatternMatcherAdapter struct {
	exclusionPatterns []string
	inclusionPatterns []string
}

// NewPatternMatcherAdapter creates a new pattern matcher adapter
func NewPatternMatcherAdapter() *PatternMatcherAdapter {
	return &PatternMatcherAdapter{
		exclusionPatterns: []string{
			"*.aux", "*.log", "*.out", "*.toc",
			"*.fdb_latexmk", "*.fls", "*.synctex.gz",
			"*.bbl", "*.blg", "*.idx", "*.ind",
			"*.lof", "*.lot", "*.nav", "*.snm",
			"*.vrb", "*.toc", "*.fls", "*.fdb_latexmk",
		},
		inclusionPatterns: []string{
			"*.tex", "*.yaml", "*.yml",
		},
	}
}

// ShouldExclude checks if a file should be excluded
func (p *PatternMatcherAdapter) ShouldExclude(filePath string) bool {
	fileName := filepath.Base(filePath)
	return p.Matches(fileName, p.exclusionPatterns)
}

// ShouldInclude checks if a file should be included
func (p *PatternMatcherAdapter) ShouldInclude(filePath string) bool {
	fileName := filepath.Base(filePath)
	return p.Matches(fileName, p.inclusionPatterns)
}

// Matches checks if a file matches any of the given patterns
func (p *PatternMatcherAdapter) Matches(filePath string, patterns []string) bool {
	fileName := filepath.Base(filePath)

	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}
	return false
}

// ConfigureExclusions updates the exclusion patterns
func (p *PatternMatcherAdapter) ConfigureExclusions(patterns []string) {
	p.exclusionPatterns = patterns
}

// ConfigureInclusions updates the inclusion patterns
func (p *PatternMatcherAdapter) ConfigureInclusions(patterns []string) {
	p.inclusionPatterns = patterns
}

// GetExclusionPatterns returns the current exclusion patterns
func (p *PatternMatcherAdapter) GetExclusionPatterns() []string {
	return p.exclusionPatterns
}

// GetInclusionPatterns returns the current inclusion patterns
func (p *PatternMatcherAdapter) GetInclusionPatterns() []string {
	return p.inclusionPatterns
}

// ValidatePattern validates a glob pattern
func (p *PatternMatcherAdapter) ValidatePattern(pattern string) bool {
	// Basic validation - check for common glob characters
	return strings.Contains(pattern, "*") || strings.Contains(pattern, "?") ||
		strings.Contains(pattern, "[") || strings.Contains(pattern, "]")
}
