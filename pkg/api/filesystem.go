// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"io"
	"io/fs"
	"os"
)

// File is a readable file handed out by a FileSystem.
type File interface {
	io.Reader
	io.Closer
}

// TempFile is a writable temporary file created by a FileSystem.
type TempFile interface {
	io.Writer
	io.Closer
	Name() string
}

// FileSystem is the filesystem boundary of the pkg/api pipeline. Missing
// paths must yield errors that satisfy errors.Is(err, fs.ErrNotExist).
type FileSystem interface {
	Stat(path string) (fs.FileInfo, error)
	Open(path string) (File, error)
	MkdirAll(path string, perm fs.FileMode) error
	MkdirTemp(dir, pattern string) (string, error)
	CreateTemp(dir, pattern string) (TempFile, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	Rename(oldPath, newPath string) error
	Remove(path string) error
	RemoveAll(path string) error
}

// OSFileSystem is the default FileSystem, backed by the os package.
type OSFileSystem struct{}

// NewOSFileSystem returns the os-backed FileSystem.
func NewOSFileSystem() FileSystem { return OSFileSystem{} }

func (OSFileSystem) Stat(path string) (fs.FileInfo, error) { return os.Stat(path) }

func (OSFileSystem) Open(path string) (File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (OSFileSystem) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }

func (OSFileSystem) MkdirTemp(dir, pattern string) (string, error) { return os.MkdirTemp(dir, pattern) }

func (OSFileSystem) CreateTemp(dir, pattern string) (TempFile, error) {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (OSFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (OSFileSystem) Rename(oldPath, newPath string) error { return os.Rename(oldPath, newPath) }

func (OSFileSystem) Remove(path string) error { return os.Remove(path) }

func (OSFileSystem) RemoveAll(path string) error { return os.RemoveAll(path) }

func fileSystemOrOS(fileSystem FileSystem) FileSystem {
	if fileSystem == nil {
		return OSFileSystem{}
	}
	return fileSystem
}
