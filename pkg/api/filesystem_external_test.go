package api_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

type memFileSystem struct {
	mu      sync.Mutex
	files   map[string][]byte
	dirs    map[string]bool
	ops     []string
	serial  int
	failOn  string
	failErr error
}

var _ api.FileSystem = (*memFileSystem)(nil)

func newMemFileSystem() *memFileSystem {
	return &memFileSystem{files: map[string][]byte{}, dirs: map[string]bool{"/": true}}
}

func (m *memFileSystem) record(op, path string) error {
	m.ops = append(m.ops, op+" "+path)
	if m.failOn != "" && m.failOn == op {
		return m.failErr
	}
	return nil
}

func (m *memFileSystem) Stat(path string) (fs.FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	path = filepath.Clean(path)
	if err := m.record("stat", path); err != nil {
		return nil, err
	}
	if data, ok := m.files[path]; ok {
		return memInfo{name: filepath.Base(path), size: int64(len(data))}, nil
	}
	if m.dirs[path] {
		return memInfo{name: filepath.Base(path), dir: true}, nil
	}
	return nil, &fs.PathError{Op: "stat", Path: path, Err: fs.ErrNotExist}
}

func (m *memFileSystem) Open(path string) (api.File, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	path = filepath.Clean(path)
	if err := m.record("open", path); err != nil {
		return nil, err
	}
	data, ok := m.files[path]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *memFileSystem) MkdirAll(path string, _ fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path = filepath.Clean(path)
	if err := m.record("mkdirall", path); err != nil {
		return err
	}
	for dir := path; ; dir = filepath.Dir(dir) {
		m.dirs[dir] = true
		if dir == filepath.Dir(dir) {
			return nil
		}
	}
}

func (m *memFileSystem) MkdirTemp(dir, pattern string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if dir == "" {
		dir = "/tmp"
	}
	m.serial++
	path := filepath.Join(dir, fmt.Sprintf("%s%d", pattern, m.serial))
	if err := m.record("mkdirtemp", path); err != nil {
		return "", err
	}
	m.dirs[path] = true
	return path, nil
}

func (m *memFileSystem) CreateTemp(dir, pattern string) (api.TempFile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if dir == "" {
		dir = "/tmp"
	}
	m.serial++
	path := filepath.Join(dir, strings.Replace(pattern, "*", fmt.Sprint(m.serial), 1))
	if err := m.record("createtemp", path); err != nil {
		return nil, err
	}
	m.files[path] = nil
	return &memTempFile{fs: m, name: path}, nil
}

func (m *memFileSystem) WriteFile(path string, data []byte, _ fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path = filepath.Clean(path)
	if err := m.record("writefile", path); err != nil {
		return err
	}
	if !m.dirs[filepath.Dir(path)] {
		return &fs.PathError{Op: "write", Path: path, Err: fs.ErrNotExist}
	}
	m.files[path] = append([]byte(nil), data...)
	return nil
}

func (m *memFileSystem) Rename(oldPath, newPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	oldPath, newPath = filepath.Clean(oldPath), filepath.Clean(newPath)
	if err := m.record("rename", oldPath); err != nil {
		return err
	}
	data, ok := m.files[oldPath]
	if !ok {
		return &fs.PathError{Op: "rename", Path: oldPath, Err: fs.ErrNotExist}
	}
	delete(m.files, oldPath)
	m.files[newPath] = data
	return nil
}

func (m *memFileSystem) Remove(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path = filepath.Clean(path)
	if err := m.record("remove", path); err != nil {
		return err
	}
	delete(m.files, path)
	delete(m.dirs, path)
	return nil
}

func (m *memFileSystem) RemoveAll(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path = filepath.Clean(path)
	if err := m.record("removeall", path); err != nil {
		return err
	}
	prefix := path + string(filepath.Separator)
	for name := range m.files {
		if name == path || strings.HasPrefix(name, prefix) {
			delete(m.files, name)
		}
	}
	for name := range m.dirs {
		if name == path || strings.HasPrefix(name, prefix) {
			delete(m.dirs, name)
		}
	}
	return nil
}

func (m *memFileSystem) put(path string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	path = filepath.Clean(path)
	m.files[path] = data
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		m.dirs[dir] = true
		if dir == filepath.Dir(dir) {
			break
		}
	}
}

func (m *memFileSystem) read(path string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.files[filepath.Clean(path)]
	return data, ok
}

func (m *memFileSystem) paths() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.files))
	for name := range m.files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type memTempFile struct {
	fs   *memFileSystem
	name string
}

func (f *memTempFile) Write(p []byte) (int, error) {
	f.fs.mu.Lock()
	defer f.fs.mu.Unlock()
	f.fs.files[f.name] = append(f.fs.files[f.name], p...)
	return len(p), nil
}

func (f *memTempFile) Close() error { return nil }
func (f *memTempFile) Name() string { return f.name }

type memInfo struct {
	name string
	size int64
	dir  bool
}

func (i memInfo) Name() string       { return i.name }
func (i memInfo) Size() int64        { return i.size }
func (i memInfo) Mode() fs.FileMode  { return 0o644 }
func (i memInfo) ModTime() time.Time { return time.Time{} }
func (i memInfo) IsDir() bool        { return i.dir }
func (i memInfo) Sys() any           { return nil }

func TestProjectionGeneratorMaterializesThroughInjectedFileSystem(t *testing.T) {
	memfs := newMemFileSystem()
	var seen map[string]string
	engine, err := api.NewEngine(api.WithGenerator(api.GeneratorFunc(func(_ context.Context, request api.Request) (api.Result, error) {
		seen = map[string]string{}
		for _, name := range memfs.paths() {
			if strings.HasPrefix(name, request.WorkingDir+string(filepath.Separator)) {
				data, _ := memfs.read(name)
				relative, _ := filepath.Rel(request.WorkingDir, name)
				seen[filepath.ToSlash(relative)] = string(data)
			}
		}
		if request.TemplatePath != filepath.Join(request.WorkingDir, "main.tex") {
			return api.Result{}, fmt.Errorf("unexpected main path %q", request.TemplatePath)
		}
		return api.Result{PDF: []byte("%PDF-fake")}, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	assetData := []byte("png")
	digest := sha256.Sum256(assetData)
	generator, err := api.NewProjectionGenerator(api.ProjectionGeneratorOptions{
		Engine:        engine,
		FileSystem:    memfs,
		WorkspaceRoot: "/virtual",
		Resolver: assetResolverFunc(func(context.Context, latex.AssetBinding) ([]byte, error) {
			return assetData, nil
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := generator.GenerateProjection(context.Background(), latex.Projection{
		Main: "main.tex",
		Files: []latex.File{
			{Path: "main.tex", Content: []byte("main")},
			{Path: "fragments/body.tex", Content: []byte("body")},
		},
		Assets: []latex.AssetBinding{{AssetID: "logo", Path: "assets/logo.png", Digest: "sha256:" + hex.EncodeToString(digest[:])}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(result.PDF) != "%PDF-fake" {
		t.Fatalf("unexpected result %q", result.PDF)
	}
	want := map[string]string{"main.tex": "main", "fragments/body.tex": "body", "assets/logo.png": "png"}
	if fmt.Sprint(seen) != fmt.Sprint(want) {
		t.Fatalf("workspace seen by engine = %v, want %v", seen, want)
	}
	if leftover := memfs.paths(); len(leftover) != 0 {
		t.Fatalf("workspace not removed through the port: %v", leftover)
	}
	for _, op := range memfs.ops {
		if !strings.Contains(op, "/virtual/") {
			t.Fatalf("operation escaped the workspace root: %q", op)
		}
	}
}

func TestProjectionGeneratorReportsInjectedFileSystemFailure(t *testing.T) {
	memfs := newMemFileSystem()
	memfs.failOn, memfs.failErr = "rename", errors.New("disk full")
	engine, err := api.NewEngine(api.WithGenerator(api.GeneratorFunc(func(context.Context, api.Request) (api.Result, error) {
		t.Fatal("engine must not run when materialization fails")
		return api.Result{}, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	generator, err := api.NewProjectionGenerator(api.ProjectionGeneratorOptions{Engine: engine, FileSystem: memfs})
	if err != nil {
		t.Fatal(err)
	}
	_, err = generator.GenerateProjection(context.Background(), latex.Projection{Main: "main.tex", Files: []latex.File{{Path: "main.tex", Content: []byte("x")}}})
	if err == nil || !strings.Contains(err.Error(), "publish projection file") || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("expected publish failure, got %v", err)
	}
	if leftover := memfs.paths(); len(leftover) != 0 {
		t.Fatalf("staged file or workspace leaked: %v", leftover)
	}
}

func TestValidationUtilsInspectsInjectedFileSystem(t *testing.T) {
	memfs := newMemFileSystem()
	pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), api.MinPDFFileSize)...)
	memfs.put("/docs/out.pdf", pdf)
	memfs.put("/docs/bad.pdf", bytes.Repeat([]byte("x"), api.MinPDFFileSize+10))
	memfs.put("/docs/template.tex", bytes.Repeat([]byte("t"), api.MinTemplateFileSize+1))
	utils := api.NewValidationUtils().WithFileSystem(memfs)

	if details, err := utils.ValidatePDFPath("/docs/out.pdf"); details != nil || err != nil {
		t.Fatalf("valid in-memory PDF rejected: %+v %v", details, err)
	}
	if details, err := utils.ValidatePDFContent("/docs/out.pdf"); details != nil || err != nil {
		t.Fatalf("valid in-memory PDF content rejected: %+v %v", details, err)
	}
	if details, _ := utils.ValidatePDFContent("/docs/bad.pdf"); details == nil {
		t.Fatal("expected invalid header to be reported")
	}
	if details, err := utils.ValidatePDFPath("/docs/missing.pdf"); err != nil || details == nil || details.Context[api.ContextKeyError] != api.ErrPDFFileNotFound {
		t.Fatalf("expected not-found details, got %+v %v", details, err)
	}
	if details, err := utils.ValidateTemplatePath("/docs/template.tex"); details != nil || err != nil {
		t.Fatalf("valid in-memory template rejected: %+v %v", details, err)
	}
	if details, err := utils.ValidateOutputPath("/out/new/report.pdf"); details != nil || err != nil {
		t.Fatalf("output path rejected: %+v %v", details, err)
	}
	if info, err := memfs.Stat("/out/new"); err != nil || !info.IsDir() {
		t.Fatalf("output directory not created through the port: %v", err)
	}
	if details, _ := utils.ValidateOutputPath("/docs/out.pdf"); details == nil || details.Context[api.ContextKeyError] != api.ErrOutputPathExists {
		t.Fatalf("expected exists details, got %+v", details)
	}
	metadata, err := utils.GetFileMetadata("/docs/out.pdf")
	if err != nil || metadata["file_size"] != int64(len(pdf)) {
		t.Fatalf("metadata = %v, %v", metadata, err)
	}
}
