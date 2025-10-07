// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

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

// ParallelExecutionOrchestratorImpl implements the ParallelExecutionOrchestrator interface
type ParallelExecutionOrchestratorImpl struct {
	maxWorkers int
	timeout    time.Duration
}

// NewParallelExecutionOrchestrator creates a new parallel execution orchestrator
func NewParallelExecutionOrchestrator() *ParallelExecutionOrchestratorImpl {
	return &ParallelExecutionOrchestratorImpl{
		maxWorkers: 4,
		timeout:    30 * time.Second,
	}
}

// ExecuteParallel executes tasks in parallel
func (o *ParallelExecutionOrchestratorImpl) ExecuteParallel(
	ctx context.Context,
	tasks []parallel.CompilationTask,
) (*parallel.ParallelCompilationResult, error) {
	var wg sync.WaitGroup
	resultChan := make(chan parallel.BuildResult, len(tasks))
	errorChan := make(chan parallel.BuildFailure, len(tasks))

	// Create worker pool
	semaphore := make(chan struct{}, o.maxWorkers)

	for _, task := range tasks {
		wg.Add(1)
		go func(t parallel.CompilationTask) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			// Execute task with timeout
			taskCtx, cancel := context.WithTimeout(ctx, t.Timeout)
			defer cancel()

			// Find appropriate strategy
			strategy := o.findCompilationStrategy(t.TemplateFile)
			if strategy == nil {
				errorChan <- parallel.BuildFailure{
					TemplateFile: t.TemplateFile,
					Error:        fmt.Errorf("no compilation strategy found"),
					Timestamp:    time.Now(),
				}
				return
			}

			// Execute compilation
			result, err := strategy.Compile(taskCtx, t.TemplateFile, t.ConfigFile)
			if err != nil {
				errorChan <- parallel.BuildFailure{
					TemplateFile: t.TemplateFile,
					Error:        err,
					Timestamp:    time.Now(),
				}
			} else {
				resultChan <- *result
			}
		}(task)
	}

	// Wait for all tasks to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	var successfulBuilds []parallel.BuildResult
	var failedBuilds []parallel.BuildFailure

	for result := range resultChan {
		successfulBuilds = append(successfulBuilds, result)
	}

	for failure := range errorChan {
		failedBuilds = append(failedBuilds, failure)
	}

	return &parallel.ParallelCompilationResult{
		SuccessfulBuilds: successfulBuilds,
		FailedBuilds:     failedBuilds,
		SuccessCount:     len(successfulBuilds),
		FailureCount:     len(failedBuilds),
	}, nil
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

// findCompilationStrategy finds the appropriate compilation strategy
func (o *ParallelExecutionOrchestratorImpl) findCompilationStrategy(templateFile string) parallel.CompilationStrategy {
	// This would be injected with actual strategies
	// For now, return nil to indicate no strategy found
	return nil
}
