// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"context"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/factories"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/valueobjects"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// ClarityMigrationExample demonstrates how to use the new CLARITY-compliant architecture
func ClarityMigrationExample() {
	// 1. Create configuration
	cfg := &config.Config{
		// ... configuration setup
	}

	// 2. Create debug configuration from environment
	debugConfig, err := valueobjects.NewDebugConfig(
		true,                    // enabled
		"/tmp/autopdf/concrete", // concrete directory
		"/tmp/autopdf/logs",     // log directory
	)
	if err != nil {
		fmt.Printf("Failed to create debug config: %v\n", err)
		return
	}

	// 3. Create factory with dependencies
	factory := factories.NewLaTeXCompilerFactory(cfg, debugConfig)

	// 4. Create compiler with debug decorators
	compiler := factory.CreateCompiler()

	// 5. Create compilation context
	compCtx, err := valueobjects.NewCompilationContext(
		"\\documentclass{article}\\begin{document}Hello World\\end{document}",
		"pdflatex",
		"/tmp/output.pdf",
		true, // debug mode
	)
	if err != nil {
		fmt.Printf("Failed to create compilation context: %v\n", err)
		return
	}

	// 6. Compile with full debug instrumentation
	ctx := context.Background()
	result, err := compiler.Compile(ctx, compCtx)
	if err != nil {
		fmt.Printf("Compilation failed: %v\n", err)
		return
	}

	fmt.Printf("Compilation successful: %s\n", result)
}

// TestableExample demonstrates how the new architecture enables easy testing
func TestableExample() {
	// This example shows how you can now easily test the LaTeX compiler
	// by injecting mock dependencies instead of relying on real filesystem/time/exec

	// In a real test, you would:
	// 1. Create mock implementations of FileSystem, Clock, DebugLogger, CommandExecutor
	// 2. Inject them into the factory
	// 3. Test the compiler without any real I/O operations
	// 4. Verify debug files are created with expected content
	// 5. Assert on log messages and file operations

	fmt.Println("This architecture enables 100% unit testable code with mocks")
}
