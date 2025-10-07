// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package parallel

import (
	"context"
	"time"
)

// ParallelCompiler represents the core domain concept of parallel compilation
type ParallelCompiler interface {
	CompileTemplates(ctx context.Context, request ParallelCompilationRequest) (*ParallelCompilationResult, error)
}

// ParallelCompilationRequest represents a request for parallel compilation
type ParallelCompilationRequest struct {
	ConfigurationFile string
	TemplateFiles     []string
	MaxConcurrency    int
	Timeout           time.Duration
}

// ParallelCompilationResult represents the result of parallel compilation
type ParallelCompilationResult struct {
	SuccessfulBuilds []BuildResult
	FailedBuilds     []BuildFailure
	TotalDuration    time.Duration
	SuccessCount     int
	FailureCount     int
}

// BuildResult represents a successful build
type BuildResult struct {
	TemplateFile string
	PDFPath      string
	Duration     time.Duration
	Timestamp    time.Time
}

// BuildFailure represents a failed build
type BuildFailure struct {
	TemplateFile string
	Error        error
	Duration     time.Duration
	Timestamp    time.Time
}

// CompilationStrategy defines the contract for compilation strategies
type CompilationStrategy interface {
	Compile(ctx context.Context, template string, config string) (*BuildResult, error)
	CanHandle(template string) bool
}

// ParallelExecutionOrchestrator coordinates parallel execution
type ParallelExecutionOrchestrator interface {
	ExecuteParallel(ctx context.Context, tasks []CompilationTask) (*ParallelCompilationResult, error)
	ConfigureConcurrency(maxWorkers int) error
	ConfigureTimeout(timeout time.Duration) error
}

// CompilationTask represents a single compilation task
type CompilationTask struct {
	TemplateFile string
	ConfigFile   string
	Priority     int
	Timeout      time.Duration
}

// CompilationResultCollector aggregates compilation results
type CompilationResultCollector interface {
	AddSuccess(result BuildResult)
	AddFailure(failure BuildFailure)
	GetResults() *ParallelCompilationResult
	Reset()
}
