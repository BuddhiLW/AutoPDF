// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package domain

// BuildOptions represents the domain model for build command options
type BuildOptions struct {
	Clean   CleanOption
	Verbose VerboseOption
	Debug   DebugOption
	Force   ForceOption
}

// CleanOption represents the clean auxiliary files option
type CleanOption struct {
	Enabled bool
	Target  string // Directory to clean, defaults to current directory
}

// VerboseOption represents the verbose logging option
type VerboseOption struct {
	Enabled bool
	Level   int // Verbosity level (0=quiet, 1=normal, 2=verbose, 3=debug)
}

// DebugOption represents the debug information option
type DebugOption struct {
	Enabled bool
	Output  string // Debug output destination (stdout, file, etc.)
}

// ForceOption represents the force operations option
type ForceOption struct {
	Enabled   bool
	Overwrite bool // Whether to overwrite existing files
}

// NewBuildOptions creates a new BuildOptions with default values
func NewBuildOptions() BuildOptions {
	return BuildOptions{
		Clean:   CleanOption{Enabled: false, Target: "."},
		Verbose: VerboseOption{Enabled: false, Level: 1},
		Debug:   DebugOption{Enabled: false, Output: "stdout"},
		Force:   ForceOption{Enabled: false, Overwrite: false},
	}
}

// EnableClean enables the clean option with the specified target
func (bo *BuildOptions) EnableClean(target string) {
	bo.Clean.Enabled = true
	bo.Clean.Target = target
}

// EnableVerbose enables verbose logging with the specified level
func (bo *BuildOptions) EnableVerbose(level int) {
	bo.Verbose.Enabled = true
	bo.Verbose.Level = level
}

// EnableDebug enables debug information with the specified output
func (bo *BuildOptions) EnableDebug(output string) {
	bo.Debug.Enabled = true
	bo.Debug.Output = output
}

// EnableForce enables force operations with overwrite setting
func (bo *BuildOptions) EnableForce(overwrite bool) {
	bo.Force.Enabled = true
	bo.Force.Overwrite = overwrite
}

// HasAnyEnabled returns true if any option is enabled
func (bo *BuildOptions) HasAnyEnabled() bool {
	return bo.Clean.Enabled || bo.Verbose.Enabled || bo.Debug.Enabled || bo.Force.Enabled
}

// GetEnabledOptions returns a list of enabled option names
func (bo *BuildOptions) GetEnabledOptions() []string {
	var enabled []string
	if bo.Clean.Enabled {
		enabled = append(enabled, "clean")
	}
	if bo.Verbose.Enabled {
		enabled = append(enabled, "verbose")
	}
	if bo.Debug.Enabled {
		enabled = append(enabled, "debug")
	}
	if bo.Force.Enabled {
		enabled = append(enabled, "force")
	}
	return enabled
}
