// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package ports

import "os"

// FileSystem abstracts all file system operations
type FileSystem interface {
	WriteFile(path string, content []byte, perm os.FileMode) error
	ReadFile(path string) ([]byte, error)
	MkdirAll(path string, perm os.FileMode) error
	Remove(path string) error
	Stat(path string) (FileInfo, error)
	Getwd() (string, error)
}

// FileInfo abstracts file information
type FileInfo interface {
	Size() int64
	IsDir() bool
}
