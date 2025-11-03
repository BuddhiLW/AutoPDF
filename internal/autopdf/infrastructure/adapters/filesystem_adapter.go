// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"io/fs"
	"os"
	"time"

	ports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
)

// OSFileSystem implements ports.FileSystem using the standard library's os package
type OSFileSystem struct{}

// NewOSFileSystem creates a new OSFileSystem adapter
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// WriteFile writes data to a file
func (f *OSFileSystem) WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// ReadFile reads data from a file
func (f *OSFileSystem) ReadFile(ctx context.Context, path string) ([]byte, error) {
	return os.ReadFile(path)
}

// Remove removes a file or directory
func (f *OSFileSystem) Remove(ctx context.Context, path string) error {
	return os.Remove(path)
}

// Stat returns file information
func (f *OSFileSystem) Stat(ctx context.Context, path string) (ports.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &OSFileInfo{info: info}, nil
}

// IsNotExist checks if an error indicates a file doesn't exist
func (f *OSFileSystem) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// MkdirAll creates a directory and all parent directories
func (f *OSFileSystem) MkdirAll(ctx context.Context, path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Symlink creates a symbolic link
func (f *OSFileSystem) Symlink(ctx context.Context, oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

// Sync synchronizes file system
func (f *OSFileSystem) Sync(ctx context.Context) error {
	// For now, this is a no-op as os package doesn't have a global sync
	// In production, you might want to implement this differently
	return nil
}

// OSFileInfo implements ports.FileInfo using os.FileInfo
type OSFileInfo struct {
	info os.FileInfo
}

func (f *OSFileInfo) Name() string {
	return f.info.Name()
}

func (f *OSFileInfo) Size() int64 {
	return f.info.Size()
}

func (f *OSFileInfo) Mode() fs.FileMode {
	return f.info.Mode()
}

func (f *OSFileInfo) ModTime() time.Time {
	return f.info.ModTime()
}

func (f *OSFileInfo) IsDir() bool {
	return f.info.IsDir()
}

func (f *OSFileInfo) Sys() interface{} {
	return f.info.Sys()
}
