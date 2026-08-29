// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package parallel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/parallel"
)

// ParallelCompilationService implements the ParallelCompiler interface
type ParallelCompilationService struct {
	orchestrator    parallel.ParallelExecutionOrchestrator
	resultCollector parallel.CompilationResultCollector
	strategies      []parallel.CompilationStrategy
	maxConcurrency  int
	timeout         time.Duration
}

// NewParallelCompilationService creates a new parallel compilation service
func NewParallelCompilationService(
	orchestrator parallel.ParallelExecutionOrchestrator,
	resultCollector parallel.CompilationResultCollector,
	strategies []parallel.CompilationStrategy,
) *ParallelCompilationService {
	return &ParallelCompilationService{
		orchestrator:    orchestrator,
		resultCollector: resultCollector,
		strategies:      strategies,
		maxConcurrency:  4, // Default concurrency
		timeout:         30 * time.Second,
	}
}

// CompileTemplates compiles multiple templates in parallel
func (p *ParallelCompilationService) CompileTemplates(
	ctx context.Context,
	request parallel.ParallelCompilationRequest,
) (*parallel.ParallelCompilationResult, error) {
	startTime := time.Now()

	// Configure orchestrator
	if err := p.orchestrator.ConfigureConcurrency(request.MaxConcurrency); err != nil {
		return nil, fmt.Errorf("failed to configure concurrency: %w", err)
	}

	if err := p.orchestrator.ConfigureTimeout(request.Timeout); err != nil {
		return nil, fmt.Errorf("failed to configure timeout: %w", err)
	}

	// Create compilation tasks
	tasks := p.createCompilationTasks(request)

	// Execute parallel compilation
	result, err := p.orchestrator.ExecuteParallel(ctx, tasks)
	if err != nil {
		return nil, fmt.Errorf("parallel execution failed: %w", err)
	}

	// Calculate total duration
	result.TotalDuration = time.Since(startTime)

	return result, nil
}

// createCompilationTasks creates compilation tasks from the request
func (p *ParallelCompilationService) createCompilationTasks(
	request parallel.ParallelCompilationRequest,
) []parallel.CompilationTask {
	tasks := make([]parallel.CompilationTask, len(request.TemplateFiles))

	for i, templateFile := range request.TemplateFiles {
		tasks[i] = parallel.CompilationTask{
			TemplateFile: templateFile,
			ConfigFile:   request.ConfigurationFile,
			Priority:     i, // Simple priority based on order
			Timeout:      request.Timeout,
		}
	}

	return tasks
}

const (
	// DefaultMaxWorkers is the compilation concurrency used until
	// ConfigureConcurrency says otherwise.
	DefaultMaxWorkers = 4
	// DefaultTaskTimeout bounds a single compilation that carries no timeout
	// of its own.
	DefaultTaskTimeout = 30 * time.Second
)

// ParallelExecutionOrchestratorImpl implements the ParallelExecutionOrchestrator interface
type ParallelExecutionOrchestratorImpl struct {
	strategies []parallel.CompilationStrategy
	maxWorkers int
	timeout    time.Duration
}

// NewParallelExecutionOrchestrator creates a new parallel execution orchestrator.
// Strategies are the abstractions compilation is dispatched through; an
// orchestrator built without any fails every task as unhandled.
func NewParallelExecutionOrchestrator(strategies ...parallel.CompilationStrategy) *ParallelExecutionOrchestratorImpl {
	return &ParallelExecutionOrchestratorImpl{
		strategies: strategies,
		maxWorkers: DefaultMaxWorkers,
		timeout:    DefaultTaskTimeout,
	}
}

// compilationJob pairs a task with its position, so outcomes can be written to
// a pre-sized slice rather than collected through channels sized by the input.
type compilationJob struct {
	index int
	task  parallel.CompilationTask
}

// compilationOutcome is the result of one task: exactly one of result/failure
// is meaningful, selected by failed.
type compilationOutcome struct {
	result  parallel.BuildResult
	failure parallel.BuildFailure
	failed  bool
}

// workerCount derives how many workers to start. Pure: never more workers than
// there is work for, and never fewer than one when there is work.
func workerCount(maxWorkers, taskCount int) int {
	if taskCount <= 0 {
		return 0
	}
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	if maxWorkers > taskCount {
		return taskCount
	}
	return maxWorkers
}

// effectiveTimeout derives the deadline for one task. Pure. A task timeout of
// zero means "unset", not "already due": context.WithTimeout(ctx, 0) returns an
// expired context, so an unset timeout falls back to the configured one.
func effectiveTimeout(taskTimeout, configured time.Duration) time.Duration {
	if taskTimeout > 0 {
		return taskTimeout
	}
	if configured > 0 {
		return configured
	}
	return DefaultTaskTimeout
}

// runTask executes one compilation. This is the boundary: the only step that
// reaches a compilation strategy, and the only place effects occur.
func (o *ParallelExecutionOrchestratorImpl) runTask(
	ctx context.Context,
	task parallel.CompilationTask,
) compilationOutcome {
	startedAt := time.Now()

	failure := func(err error) compilationOutcome {
		return compilationOutcome{
			failed: true,
			failure: parallel.BuildFailure{
				TemplateFile: task.TemplateFile,
				Error:        err,
				Duration:     time.Since(startedAt),
				Timestamp:    time.Now(),
			},
		}
	}

	strategy := o.findCompilationStrategy(task.TemplateFile)
	if strategy == nil {
		return failure(fmt.Errorf("no compilation strategy found for %q", task.TemplateFile))
	}

	taskCtx, cancel := context.WithTimeout(ctx, effectiveTimeout(task.Timeout, o.timeout))
	defer cancel()

	result, err := strategy.Compile(taskCtx, task.TemplateFile, task.ConfigFile)
	if err != nil {
		return failure(err)
	}
	if result == nil {
		return failure(fmt.Errorf("strategy returned no result for %q", task.TemplateFile))
	}

	build := *result
	if build.Duration == 0 {
		build.Duration = time.Since(startedAt)
	}
	if build.Timestamp.IsZero() {
		build.Timestamp = time.Now()
	}
	return compilationOutcome{result: build}
}

// ExecuteParallel runs tasks through a bounded worker pool.
//
// The pool bounds goroutine CREATION, not merely execution: exactly
// workerCount goroutines exist regardless of how many tasks are queued.
// Acquiring a semaphore inside a per-task goroutine would bound only progress,
// leaving one parked goroutine per task.
func (o *ParallelExecutionOrchestratorImpl) ExecuteParallel(
	ctx context.Context,
	tasks []parallel.CompilationTask,
) (*parallel.ParallelCompilationResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	outcomes := make([]compilationOutcome, len(tasks))
	workers := workerCount(o.maxWorkers, len(tasks))
	if workers == 0 {
		return &parallel.ParallelCompilationResult{}, nil
	}

	jobs := make(chan compilationJob)
	var pool sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		pool.Add(1)
		go func() {
			defer pool.Done()
			for job := range jobs {
				outcomes[job.index] = o.runTask(ctx, job.task)
			}
		}()
	}

	// Dispatch. An unbuffered channel applies backpressure: dispatch advances
	// only as fast as a worker takes work, so a cancelled run stops feeding
	// the pool instead of draining the whole backlog.
	cancelled := false
	for index, task := range tasks {
		select {
		case <-ctx.Done():
			cancelled = true
		case jobs <- compilationJob{index: index, task: task}:
		}
		if cancelled {
			break
		}
	}
	close(jobs)
	pool.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return collectOutcomes(outcomes), nil
}

// collectOutcomes folds per-task outcomes into the aggregate result, preserving
// task order. Pure.
func collectOutcomes(outcomes []compilationOutcome) *parallel.ParallelCompilationResult {
	successfulBuilds := make([]parallel.BuildResult, 0, len(outcomes))
	failedBuilds := make([]parallel.BuildFailure, 0, len(outcomes))

	for _, outcome := range outcomes {
		if outcome.failed {
			failedBuilds = append(failedBuilds, outcome.failure)
			continue
		}
		successfulBuilds = append(successfulBuilds, outcome.result)
	}

	return &parallel.ParallelCompilationResult{
		SuccessfulBuilds: successfulBuilds,
		FailedBuilds:     failedBuilds,
		SuccessCount:     len(successfulBuilds),
		FailureCount:     len(failedBuilds),
	}
}

// ConfigureConcurrency sets the maximum number of concurrent workers
func (o *ParallelExecutionOrchestratorImpl) ConfigureConcurrency(maxWorkers int) error {
	if maxWorkers <= 0 {
		return fmt.Errorf("maxWorkers must be positive")
	}
	o.maxWorkers = maxWorkers
	return nil
}

// ConfigureTimeout sets the timeout for individual tasks
func (o *ParallelExecutionOrchestratorImpl) ConfigureTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	o.timeout = timeout
	return nil
}

// findCompilationStrategy returns the first injected strategy that claims the
// template, or nil when none does. Pure over the strategy set.
func (o *ParallelExecutionOrchestratorImpl) findCompilationStrategy(templateFile string) parallel.CompilationStrategy {
	for _, strategy := range o.strategies {
		if strategy != nil && strategy.CanHandle(templateFile) {
			return strategy
		}
	}
	return nil
}
