// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package watch

import (
	"time"
)

// FileWatcher represents the core domain concept of monitoring file changes
type FileWatcher interface {
	Watch() error
	Stop() error
	OnChange(callback FileChangeCallback) error
}

// FileChangeCallback defines the contract for handling file changes
type FileChangeCallback func(event FileChangeEvent) error

// FileChangeEvent represents a file system change event
type FileChangeEvent struct {
	FilePath      string
	Operation     FileOperation
	Timestamp     time.Time
	ShouldRebuild bool
}

// FileOperation represents the type of file system operation
type FileOperation string

const (
	WriteOp  FileOperation = "write"
	CreateOp FileOperation = "create"
	RemoveOp FileOperation = "remove"
	RenameOp FileOperation = "rename"
)

// WatchConfiguration represents the configuration for file watching
type WatchConfiguration struct {
	TemplateFile      string
	ConfigFile        string
	DebounceInterval  time.Duration
	ExclusionPatterns []string
	InclusionPatterns []string
}

// FilePatternMatcher defines the contract for pattern matching
type FilePatternMatcher interface {
	ShouldExclude(filePath string) bool
	ShouldInclude(filePath string) bool
	Matches(filePath string, patterns []string) bool
}

// DebounceStrategy defines the contract for debouncing file changes
type DebounceStrategy interface {
	ShouldTrigger(event FileChangeEvent) bool
	Reset()
}

// WatchService represents the application service for file watching
type WatchService interface {
	StartWatching(config WatchConfiguration) error
	StopWatching() error
	ConfigureExclusions(patterns []string) error
	ConfigureInterval(interval time.Duration) error
}

// FileChangeProcessor defines the contract for processing file changes
type FileChangeProcessor interface {
	ProcessChange(event FileChangeEvent) error
	CanProcess(event FileChangeEvent) bool
}
