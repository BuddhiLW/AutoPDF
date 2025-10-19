// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package options

import "time"

// BuildOptions represents the domain model for build command options
type BuildOptions struct {
	Clean   CleanOption
	Verbose VerboseOption
	Debug   DebugOption
	Force   ForceOption
	Watch   WatchOption
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

// WatchOption represents the watch mode option
type WatchOption struct {
	Enabled  bool
	Interval time.Duration // Watch interval, defaults to 500ms
}

// NewBuildOptions creates a new BuildOptions with default values
func NewBuildOptions() BuildOptions {
	return BuildOptions{
		Clean:   CleanOption{Enabled: false, Target: "."},
		Verbose: VerboseOption{Enabled: false, Level: 1},
		Debug:   DebugOption{Enabled: false, Output: "stdout"},
		Force:   ForceOption{Enabled: false, Overwrite: false},
		Watch:   WatchOption{Enabled: false, Interval: 500 * time.Millisecond},
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

// EnableWatch enables watch mode with the specified interval
func (bo *BuildOptions) EnableWatch(interval time.Duration) {
	bo.Watch.Enabled = true
	bo.Watch.Interval = interval
}

// HasAnyEnabled returns true if any option is enabled
func (bo *BuildOptions) HasAnyEnabled() bool {
	for _, option := range bo.GetEnabledOptions() {
		if option != "" {
			return true
		}
	}
	return false
	// equivalent to:
	// return bo.Clean.Enabled || bo.Verbose.Enabled || bo.Debug.Enabled || bo.Force.Enabled || bo.Watch.Enabled
}

// GetEnabledOptions returns a list of enabled option names as strings
func (bo *BuildOptions) GetEnabledOptions() []string {
	var enabled []string
	if bo.Debug.Enabled {
		enabled = append(enabled, "debug")
	}
	if bo.Force.Enabled {
		enabled = append(enabled, "force")
	}
	if bo.Watch.Enabled {
		enabled = append(enabled, "watch")
	}
	return enabled
}
