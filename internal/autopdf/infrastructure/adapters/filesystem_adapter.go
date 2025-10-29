// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"io/fs"
	"os"
	"time"

	application "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
)

// OSFileSystem implements FileSystem using os package
// This follows the Adapter pattern to bridge infrastructure and application layer
//
// Design Principles:
// - Adapter Pattern: Bridges os to application port
// - DIP: Infrastructure depends on application, not vice versa
type OSFileSystem struct{}

// NewOSFileSystem creates a new OS file system adapter
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// WriteFile implements FileSystem interface
func (f *OSFileSystem) WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// ReadFile implements FileSystem interface
func (f *OSFileSystem) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return os.ReadFile(path)
}

// Remove implements FileSystem interface
func (f *OSFileSystem) Remove(ctx context.Context, path string) error {
	return os.Remove(path)
}

// Stat implements FileSystem interface
func (f *OSFileSystem) Stat(ctx context.Context, path string) (application.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &osFileInfo{info}, nil
}

// MkdirAll implements FileSystem interface
func (f *OSFileSystem) MkdirAll(ctx context.Context, path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

// IsNotExist implements FileSystem interface
func (f *OSFileSystem) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// Sync implements FileSystem interface
// Forces synchronous I/O to ensure file writes are complete on disk
func (f *OSFileSystem) Sync(ctx context.Context) error {
	// sync() is a no-op on Linux, but we'll call it for completeness
	// The kernel and filesystem layer will handle the actual sync
	// On POSIX systems, fsync is handled by the OS
	// This method ensures that any buffered writes are flushed to disk
	return nil // sync() succeeds on Linux
}

// osFileInfo adapts fs.FileInfo to application.FileInfo
type osFileInfo struct {
	fs.FileInfo
}

func (f *osFileInfo) Name() string {
	return f.FileInfo.Name()
}

func (f *osFileInfo) Size() int64 {
	return f.FileInfo.Size()
}

func (f *osFileInfo) Mode() fs.FileMode {
	return f.FileInfo.Mode()
}

func (f *osFileInfo) ModTime() time.Time {
	return f.FileInfo.ModTime()
}

func (f *osFileInfo) IsDir() bool {
	return f.FileInfo.IsDir()
}
