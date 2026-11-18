// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package compilertest

import (
	"context"
	"errors"
	"slices"
	"sync"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

// Mode selects how FakeExecutor answers compile (non-probe) commands.
type Mode int

const (
	// ModeSucceed writes a non-empty PDF at PDFPath and exits 0.
	ModeSucceed Mode = iota
	// ModeFail exits 1 with a diagnostic and writes nothing.
	ModeFail
	// ModeFailEmptyPDF leaves an empty PDF at PDFPath and exits 1.
	ModeFailEmptyPDF
	// ModeBlock blocks until ctx is done, then returns ctx.Err().
	ModeBlock
)

// FakePDF is the content ModeSucceed writes.
var FakePDF = []byte("%PDF-1.7\n% compilertest fake\n%%EOF\n")

// FakeExecutor is a ports.CommandExecutor standing in for a backend toolchain.
// Toolchain outputs go to the MemFS through Seed, so MemFS.WrittenPaths stays
// the set of files the compiler itself wrote.
type FakeExecutor struct {
	mu       sync.Mutex
	fs       *MemFS
	mode     Mode
	pdfPath  string
	missing  map[string]bool
	isProbe  func(ports.Command) bool
	started  chan struct{}
	once     sync.Once
	commands []ports.Command
}

var _ ports.CommandExecutor = (*FakeExecutor)(nil)

// NewFakeExecutor returns an executor in ModeSucceed with every binary present.
func NewFakeExecutor(fs *MemFS) *FakeExecutor {
	return &FakeExecutor{
		fs:      fs,
		missing: map[string]bool{},
		isProbe: DefaultIsProbe,
		started: make(chan struct{}),
	}
}

// DefaultIsProbe treats `which <bin>` and `<bin> --version` as toolchain probes.
func DefaultIsProbe(cmd ports.Command) bool {
	return cmd.Executable == "which" || slices.Contains(cmd.Args, "--version")
}

// SetMode selects the compile behaviour.
func (e *FakeExecutor) SetMode(mode Mode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.mode = mode
}

// SetPDFPath sets where compile commands write their PDF.
func (e *FakeExecutor) SetPDFPath(path string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pdfPath = path
}

// SetIsProbe overrides how probe commands are recognised.
func (e *FakeExecutor) SetIsProbe(isProbe func(ports.Command) bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.isProbe = isProbe
}

// Uninstall marks binaries as absent from the toolchain.
func (e *FakeExecutor) Uninstall(binaries ...string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, b := range binaries {
		e.missing[b] = true
	}
}

// Started is closed when the first compile command starts.
func (e *FakeExecutor) Started() <-chan struct{} { return e.started }

// Commands returns every command executed so far.
func (e *FakeExecutor) Commands() []ports.Command {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]ports.Command(nil), e.commands...)
}

// CompileCommands returns the executed commands that were not probes.
func (e *FakeExecutor) CompileCommands() []ports.Command {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []ports.Command
	for _, c := range e.commands {
		if !e.isProbe(c) {
			out = append(out, c)
		}
	}
	return out
}

// Execute implements ports.CommandExecutor.
func (e *FakeExecutor) Execute(ctx context.Context, cmd ports.Command) (ports.CommandResult, error) {
	e.mu.Lock()
	e.commands = append(e.commands, cmd)
	mode, pdfPath, probe := e.mode, e.pdfPath, e.isProbe(cmd)
	executableMissing := e.missing[cmd.Executable]
	probedMissing := cmd.Executable == "which" && len(cmd.Args) > 0 && e.missing[cmd.Args[0]]
	e.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return ports.NewCommandResult("", "", -1, 0), err
	}
	if executableMissing {
		return ports.NewCommandResult("", "", -1, 0),
			&ports.ToolchainMissingError{Binary: cmd.Executable, Err: errors.New("executable file not found in $PATH")}
	}
	if probe {
		if probedMissing {
			return ports.NewCommandResult("", "", 1, 0), errors.New("exit status 1")
		}
		return ports.NewCommandResult("compilertest fake toolchain 1.0\n", "", 0, 0), nil
	}

	e.once.Do(func() { close(e.started) })
	switch mode {
	case ModeSucceed:
		e.fs.Seed(pdfPath, FakePDF)
		return ports.NewCommandResult("", "", 0, 0), nil
	case ModeFail:
		return ports.NewCommandResult("", "! compilertest: forced failure\n", 1, 0), errors.New("exit status 1")
	case ModeFailEmptyPDF:
		e.fs.Seed(pdfPath, nil)
		return ports.NewCommandResult("", "! compilertest: forced failure\n", 1, 0), errors.New("exit status 1")
	case ModeBlock:
		<-ctx.Done()
		return ports.NewCommandResult("", "", -1, 0), ctx.Err()
	default:
		return ports.NewCommandResult("", "", 1, 0), errors.New("compilertest: unknown mode")
	}
}
