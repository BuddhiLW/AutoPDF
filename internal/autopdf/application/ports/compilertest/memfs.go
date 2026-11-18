// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package compilertest

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"sort"
	"sync"
	"time"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

// MemFS is an in-memory ports.FileSystem. It records every path a compiler
// writes through WriteFile, so the suite can tell staged files apart from
// files the toolchain or the test seeded.
type MemFS struct {
	mu      sync.Mutex
	files   map[string][]byte
	dirs    map[string]bool
	written []string
}

var _ ports.FileSystem = (*MemFS)(nil)

// NewMemFS returns an empty in-memory file system.
func NewMemFS() *MemFS {
	return &MemFS{files: map[string][]byte{}, dirs: map[string]bool{}}
}

// Seed places a file without recording it as written by the compiler.
func (m *MemFS) Seed(path string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[filepath.Clean(path)] = append([]byte(nil), data...)
}

// Contents returns a file's bytes and whether it exists.
func (m *MemFS) Contents(path string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.files[filepath.Clean(path)]
	return append([]byte(nil), data...), ok
}

// Exists reports whether a file exists.
func (m *MemFS) Exists(path string) bool {
	_, ok := m.Contents(path)
	return ok
}

// WrittenPaths returns, sorted and deduplicated, every path written via WriteFile.
func (m *MemFS) WrittenPaths() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	seen := map[string]bool{}
	var out []string
	for _, p := range m.written {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// WriteFile implements ports.FileSystem.
func (m *MemFS) WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path = filepath.Clean(path)
	m.files[path] = append([]byte(nil), data...)
	m.written = append(m.written, path)
	return nil
}

// ReadFile implements ports.FileSystem.
func (m *MemFS) ReadFile(ctx context.Context, path string) ([]byte, error) {
	data, ok := m.Contents(path)
	if !ok {
		return nil, notExist("read", path)
	}
	return data, nil
}

// Remove implements ports.FileSystem.
func (m *MemFS) Remove(ctx context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path = filepath.Clean(path)
	if _, ok := m.files[path]; !ok {
		return notExist("remove", path)
	}
	delete(m.files, path)
	return nil
}

// Rename implements ports.FileSystem.
func (m *MemFS) Rename(ctx context.Context, oldpath, newpath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	oldpath, newpath = filepath.Clean(oldpath), filepath.Clean(newpath)
	data, ok := m.files[oldpath]
	if !ok {
		return notExist("rename", oldpath)
	}
	m.files[newpath] = data
	delete(m.files, oldpath)
	return nil
}

// Stat implements ports.FileSystem.
func (m *MemFS) Stat(ctx context.Context, path string) (ports.FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	path = filepath.Clean(path)
	if data, ok := m.files[path]; ok {
		return memFileInfo{name: filepath.Base(path), size: int64(len(data))}, nil
	}
	if m.dirs[path] {
		return memFileInfo{name: filepath.Base(path), dir: true}, nil
	}
	return nil, notExist("stat", path)
}

// IsNotExist implements ports.FileSystem.
func (m *MemFS) IsNotExist(err error) bool { return errors.Is(err, fs.ErrNotExist) }

// MkdirAll implements ports.FileSystem.
func (m *MemFS) MkdirAll(ctx context.Context, path string, perm fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dirs[filepath.Clean(path)] = true
	return nil
}

// Symlink implements ports.FileSystem by copying the target's bytes.
func (m *MemFS) Symlink(ctx context.Context, oldname, newname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.files[filepath.Clean(oldname)]
	if !ok {
		return notExist("symlink", oldname)
	}
	m.files[filepath.Clean(newname)] = data
	return nil
}

// Sync implements ports.FileSystem.
func (m *MemFS) Sync(ctx context.Context) error { return nil }

func notExist(op, path string) error {
	return &fs.PathError{Op: op, Path: path, Err: fs.ErrNotExist}
}

type memFileInfo struct {
	name string
	size int64
	dir  bool
}

func (i memFileInfo) Name() string { return i.name }
func (i memFileInfo) Size() int64  { return i.size }
func (i memFileInfo) Mode() fs.FileMode {
	if i.dir {
		return fs.ModeDir | 0o755
	}
	return 0o644
}
func (i memFileInfo) ModTime() time.Time { return time.Time{} }
func (i memFileInfo) IsDir() bool        { return i.dir }
func (i memFileInfo) Sys() interface{}   { return nil }
