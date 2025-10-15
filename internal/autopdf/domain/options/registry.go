// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package options

// OptionRegistry provides a centralized registry for all available options
type OptionRegistry struct {
	knownOptions     map[string]bool
	knownSubcommands map[string]bool
}

// NewGlobalRegistry creates a new registry with all known options and subcommands
func NewGlobalRegistry() *OptionRegistry {
	return &OptionRegistry{
		knownOptions: map[string]bool{
			"clean":   true,
			"verbose": true,
			"debug":   true,
			"force":   true,
			"watch":   true,
		},
		knownSubcommands: map[string]bool{
			"convert": true,
			"config":  true,
			"watch":   true,
		},
	}
}

// IsOption checks if the given argument is a known option
func (r *OptionRegistry) IsOption(arg string) bool {
	return r.knownOptions[arg]
}

// IsSubcommand checks if the given argument is a known subcommand
func (r *OptionRegistry) IsSubcommand(arg string) bool {
	return r.knownSubcommands[arg]
}

// GetKnownOptions returns a list of all known option names
func (r *OptionRegistry) GetKnownOptions() []string {
	options := make([]string, 0, len(r.knownOptions))
	for option := range r.knownOptions {
		options = append(options, option)
	}
	return options
}

// GetKnownSubcommands returns a list of all known subcommand names
func (r *OptionRegistry) GetKnownSubcommands() []string {
	subcommands := make([]string, 0, len(r.knownSubcommands))
	for subcommand := range r.knownSubcommands {
		subcommands = append(subcommands, subcommand)
	}
	return subcommands
}
