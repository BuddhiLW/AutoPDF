// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package compilertest is the shared contract suite every
// ports.DocumentCompiler must pass. A backend adapter calls Run from its own
// tests with a Factory that builds the adapter over the injected
// ports.FileSystem and ports.CommandExecutor; the suite supplies an in-memory
// file system (MemFS) and a fake toolchain (FakeExecutor), so no real binary
// and no global state is involved.
package compilertest

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

// Subject describes one compiler under test.
type Subject struct {
	// Compiler is the adapter, built over the ports handed to the Factory.
	Compiler ports.DocumentCompiler
	// Engine is put in CompileOptions.Engine.
	Engine string
	// Backend is put in CompileOptions.Backend (nil = backend defaults).
	Backend ports.BackendOptions
	// Content is a valid rendered document for this backend.
	Content string
	// Binaries are the executables the compiler needs; removing any one of
	// them must yield a *ports.ToolchainMissingError naming it.
	Binaries []string
	// TemplateExtensions are the source-template extensions of this backend
	// (e.g. ".tex"); a template named <job><ext> in the working directory must
	// survive every compile.
	TemplateExtensions []string
	// IsProbe recognises toolchain probe commands; nil = DefaultIsProbe.
	IsProbe func(ports.Command) bool
}

// Factory builds the compiler under test over the given ports.
type Factory func(t *testing.T, fs ports.FileSystem, exec ports.CommandExecutor) Subject

// Timeout bounds every single compile the suite performs.
const Timeout = 5 * time.Second

const jobName = "doc"

// env is one fresh fixture: ports, subject and the paths a compile uses.
type env struct {
	fs         *MemFS
	exec       *FakeExecutor
	subject    Subject
	workingDir string
	outputPath string
}

func newEnv(t *testing.T, factory Factory) *env {
	t.Helper()
	fs := NewMemFS()
	exec := NewFakeExecutor(fs)
	subject := factory(t, fs, exec)
	if subject.Compiler == nil {
		t.Fatal("compilertest: Factory returned a Subject without a Compiler")
	}
	if subject.IsProbe != nil {
		exec.SetIsProbe(subject.IsProbe)
	}
	workingDir := filepath.Join(string(filepath.Separator), "compilertest", "work")
	outputPath := filepath.Join(string(filepath.Separator), "compilertest", "out", jobName+".pdf")
	exec.SetPDFPath(outputPath)
	return &env{fs: fs, exec: exec, subject: subject, workingDir: workingDir, outputPath: outputPath}
}

func (e *env) options() ports.CompileOptions {
	return ports.NewCompileOptions(e.subject.Engine, e.outputPath, e.workingDir).
		WithJobName(jobName).
		WithBackendOptions(e.subject.Backend)
}

// compile runs one compile bounded by Timeout so a hung adapter fails the
// test instead of the whole run.
func (e *env) compile(t *testing.T, ctx context.Context, opts ports.CompileOptions) (string, error) {
	t.Helper()
	type outcome struct {
		path string
		err  error
	}
	done := make(chan outcome, 1)
	go func() {
		path, err := e.subject.Compiler.Compile(ctx, e.subject.Content, opts)
		done <- outcome{path, err}
	}()
	select {
	case o := <-done:
		return o.path, o.err
	case <-time.After(Timeout):
		t.Fatalf("compile did not return within %s", Timeout)
		return "", nil
	}
}

func (e *env) assertNoPDF(t *testing.T) {
	t.Helper()
	if e.fs.Exists(e.outputPath) {
		t.Errorf("a failed compile must leave no PDF at %s", e.outputPath)
	}
}

// Run executes the DocumentCompiler contract against the compiler factory builds.
func Run(t *testing.T, factory Factory) {
	t.Helper()

	t.Run("SuccessWritesNonEmptyPDFAtOutputPath", func(t *testing.T) {
		e := newEnv(t, factory)
		path, err := e.compile(t, context.Background(), e.options())
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		if path != e.outputPath {
			t.Errorf("returned path = %q, want OutputPath %q", path, e.outputPath)
		}
		data, ok := e.fs.Contents(e.outputPath)
		if !ok {
			t.Fatalf("no PDF at OutputPath %s", e.outputPath)
		}
		if len(data) == 0 {
			t.Errorf("PDF at %s is empty", e.outputPath)
		}
		if len(e.exec.CompileCommands()) == 0 {
			t.Errorf("success without running the toolchain")
		}
	})

	t.Run("FailureLeavesNoPDF", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			mode Mode
		}{
			{"ToolchainExitsNonZero", ModeFail},
			{"ToolchainLeavesEmptyPDF", ModeFailEmptyPDF},
		} {
			t.Run(tc.name, func(t *testing.T) {
				e := newEnv(t, factory)
				e.exec.SetMode(tc.mode)
				e.fs.Seed(e.outputPath, []byte("%PDF stale output of an earlier run"))
				path, err := e.compile(t, context.Background(), e.options())
				if err == nil {
					t.Fatalf("compile succeeded (path %q), want an error", path)
				}
				if path != "" {
					t.Errorf("a failed compile returned path %q, want empty", path)
				}
				e.assertNoPDF(t)
			})
		}
	})

	t.Run("CancelledContextIsHonoured", func(t *testing.T) {
		t.Run("BeforeCompile", func(t *testing.T) {
			e := newEnv(t, factory)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_, err := e.compile(t, ctx, e.options())
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("err = %v, want context.Canceled", err)
			}
			if n := len(e.exec.CompileCommands()); n != 0 {
				t.Errorf("a cancelled compile ran %d toolchain commands", n)
			}
			e.assertNoPDF(t)
		})
		t.Run("DuringCompile", func(t *testing.T) {
			e := newEnv(t, factory)
			e.exec.SetMode(ModeBlock)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				select {
				case <-e.exec.Started():
					cancel()
				case <-time.After(Timeout):
				}
			}()
			_, err := e.compile(t, ctx, e.options())
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("err = %v, want context.Canceled", err)
			}
			e.assertNoPDF(t)
		})
	})

	t.Run("StagedFile", func(t *testing.T) {
		t.Run("NeverEqualsTheSourceTemplate", func(t *testing.T) {
			e := newEnv(t, factory)
			original := []byte("source template, must survive the compile")
			var templates []string
			for _, ext := range e.subject.TemplateExtensions {
				tpl := filepath.Join(e.workingDir, jobName+ext)
				e.fs.Seed(tpl, original)
				templates = append(templates, tpl)
			}
			if len(templates) == 0 {
				t.Fatal("compilertest: Subject.TemplateExtensions is empty")
			}
			if _, err := e.compile(t, context.Background(), e.options().WithDebug(true)); err != nil {
				t.Fatalf("compile: %v", err)
			}
			written := e.fs.WrittenPaths()
			if len(written) == 0 {
				t.Fatal("the compiler staged no source file")
			}
			for _, tpl := range templates {
				for _, w := range written {
					if w == filepath.Clean(tpl) {
						t.Errorf("staged file %s is the source template", w)
					}
				}
				got, ok := e.fs.Contents(tpl)
				if !ok || !bytes.Equal(got, original) {
					t.Errorf("source template %s was changed or removed", tpl)
				}
			}
		})
		t.Run("RemovedUnlessDebug", func(t *testing.T) {
			for _, tc := range []struct {
				name string
				mode Mode
			}{
				{"AfterSuccess", ModeSucceed},
				{"AfterFailure", ModeFail},
			} {
				t.Run(tc.name, func(t *testing.T) {
					e := newEnv(t, factory)
					e.exec.SetMode(tc.mode)
					_, _ = e.compile(t, context.Background(), e.options().WithDebug(false))
					written := e.fs.WrittenPaths()
					if len(written) == 0 {
						t.Fatal("the compiler staged no source file")
					}
					for _, w := range written {
						if e.fs.Exists(w) {
							t.Errorf("staged file %s survived a non-debug compile", w)
						}
					}
				})
			}
		})
		t.Run("KeptInDebug", func(t *testing.T) {
			e := newEnv(t, factory)
			if _, err := e.compile(t, context.Background(), e.options().WithDebug(true)); err != nil {
				t.Fatalf("compile: %v", err)
			}
			kept := false
			for _, w := range e.fs.WrittenPaths() {
				if data, ok := e.fs.Contents(w); ok && string(data) == e.subject.Content {
					kept = true
				}
			}
			if !kept {
				t.Error("debug mode must keep the staged file holding the rendered content")
			}
		})
	})

	t.Run("MissingToolchainIsTypedErrorNamingTheBinary", func(t *testing.T) {
		probe := newEnv(t, factory)
		if len(probe.subject.Binaries) == 0 {
			t.Fatal("compilertest: Subject.Binaries is empty")
		}
		for _, binary := range probe.subject.Binaries {
			t.Run(binary, func(t *testing.T) {
				e := newEnv(t, factory)
				e.exec.Uninstall(binary)
				_, err := e.compile(t, context.Background(), e.options())
				var missing *ports.ToolchainMissingError
				if !errors.As(err, &missing) {
					t.Fatalf("err = %v, want *ports.ToolchainMissingError", err)
				}
				if missing.Binary != binary {
					t.Errorf("ToolchainMissingError.Binary = %q, want %q", missing.Binary, binary)
				}
				if !errors.Is(err, ports.ErrToolchainMissing) {
					t.Errorf("errors.Is(err, ports.ErrToolchainMissing) = false")
				}
				e.assertNoPDF(t)
			})
		}
	})

	t.Run("ForeignBackendOptionsAreRefused", func(t *testing.T) {
		e := newEnv(t, factory)
		_, err := e.compile(t, context.Background(), e.options().WithBackendOptions(foreignOptions{}))
		if !errors.Is(err, ports.ErrBackendOptionsMismatch) {
			t.Fatalf("err = %v, want ports.ErrBackendOptionsMismatch", err)
		}
		if n := len(e.exec.CompileCommands()); n != 0 {
			t.Errorf("refused options still ran %d toolchain commands", n)
		}
		e.assertNoPDF(t)
	})
}

// foreignOptions belong to no real backend.
type foreignOptions struct{}

func (foreignOptions) BackendName() string { return "compilertest-foreign" }
