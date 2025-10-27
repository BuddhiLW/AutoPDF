// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"os"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
)

// OSFileSystemAdapter implements FileSystem using the standard library
type OSFileSystemAdapter struct{}

// NewOSFileSystemAdapter creates a new OS filesystem adapter
func NewOSFileSystemAdapter() ports.FileSystem {
	return &OSFileSystemAdapter{}
}

// WriteFile writes content to a file
func (a *OSFileSystemAdapter) WriteFile(path string, content []byte, perm os.FileMode) error {
	return os.WriteFile(path, content, perm)
}

// ReadFile reads content from a file
func (a *OSFileSystemAdapter) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// MkdirAll creates a directory and all parent directories
func (a *OSFileSystemAdapter) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Remove removes a file or directory
func (a *OSFileSystemAdapter) Remove(path string) error {
	return os.Remove(path)
}

// Stat returns file information
func (a *OSFileSystemAdapter) Stat(path string) (ports.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &OSFileInfo{info: info}, nil
}

// Getwd returns the current working directory
func (a *OSFileSystemAdapter) Getwd() (string, error) {
	return os.Getwd()
}

// OSFileInfo implements FileInfo using os.FileInfo
type OSFileInfo struct {
	info os.FileInfo
}

// Size returns the file size
func (f *OSFileInfo) Size() int64 {
	return f.info.Size()
}

// IsDir returns true if the file is a directory
func (f *OSFileInfo) IsDir() bool {
	return f.info.IsDir()
}
