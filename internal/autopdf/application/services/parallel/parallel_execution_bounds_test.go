// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package parallel

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/domain/parallel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// blockingStrategy is a CompilationStrategy that parks every call on release
// until the test lets it go, so the test can observe how many compilations are
// in flight — and how many goroutines exist — at peak.
type blockingStrategy struct {
	entered  chan struct{}
	release  chan struct{}
	inFlight int64
	peak     int64
}

func newBlockingStrategy(capacity int) *blockingStrategy {
	return &blockingStrategy{
		entered: make(chan struct{}, capacity),
		release: make(chan struct{}),
	}
}

func (s *blockingStrategy) CanHandle(string) bool { return true }

func (s *blockingStrategy) Compile(ctx context.Context, template, _ string) (*parallel.BuildResult, error) {
	current := atomic.AddInt64(&s.inFlight, 1)
	for {
		peak := atomic.LoadInt64(&s.peak)
		if current <= peak || atomic.CompareAndSwapInt64(&s.peak, peak, current) {
			break
		}
	}
	defer atomic.AddInt64(&s.inFlight, -1)

	s.entered <- struct{}{}
	select {
	case <-s.release:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return &parallel.BuildResult{TemplateFile: template, PDFPath: template + ".pdf"}, nil
}

func tasksNamed(n int, timeout time.Duration) []parallel.CompilationTask {
	tasks := make([]parallel.CompilationTask, n)
	for i := range tasks {
		tasks[i] = parallel.CompilationTask{
			TemplateFile: fmt.Sprintf("doc-%d.tex", i),
			ConfigFile:   "autopdf.yaml",
			Priority:     i,
			Timeout:      timeout,
		}
	}
	return tasks
}

// TestExecuteParallelBoundsGoroutineCreation is the empirical gate on bounded
// concurrency: the orchestrator must cap the number of goroutines that EXIST,
// not merely how many make progress.
//
// A semaphore acquired inside a per-task goroutine bounds execution only —
// feeding it N tasks still creates N goroutines, all but maxWorkers of them
// parked. The measurement that separates the two shapes is the goroutine delta
// while every worker is blocked: it tracks the worker count under a real pool
// and the task count under a semaphore-inside-goroutine fan-out.
func TestExecuteParallelBoundsGoroutineCreation(t *testing.T) {
	const (
		taskCount  = 300
		maxWorkers = 4
	)

	strategy := newBlockingStrategy(taskCount)
	orchestrator := NewParallelExecutionOrchestrator(strategy)
	require.NoError(t, orchestrator.ConfigureConcurrency(maxWorkers))

	before := runtime.NumGoroutine()

	var (
		result *parallel.ParallelCompilationResult
		err    error
		done   = make(chan struct{})
	)
	go func() {
		defer close(done)
		result, err = orchestrator.ExecuteParallel(context.Background(), tasksNamed(taskCount, time.Minute))
	}()

	// Wait until every worker slot is occupied and blocked in Compile.
	for i := 0; i < maxWorkers; i++ {
		select {
		case <-strategy.entered:
		case <-time.After(5 * time.Second):
			t.Fatalf("only %d of %d workers entered Compile", i, maxWorkers)
		}
	}

	// Give a semaphore-style implementation the chance to have spawned the
	// rest of its goroutines before sampling.
	time.Sleep(100 * time.Millisecond)
	delta := runtime.NumGoroutine() - before

	close(strategy.release)
	<-done

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, taskCount, result.SuccessCount, "every task must still complete")

	// Workers + the dispatching goroutine + this test's driver, with headroom.
	assert.LessOrEqual(t, delta, maxWorkers+8,
		"goroutine delta %d suggests one goroutine per task (%d) rather than a bounded pool of %d; "+
			"a semaphore inside a per-task goroutine bounds execution, not creation",
		delta, taskCount, maxWorkers)

	assert.LessOrEqual(t, int(atomic.LoadInt64(&strategy.peak)), maxWorkers,
		"more compilations ran concurrently than maxWorkers allows")
}

// TestExecuteParallelUsesInjectedStrategies pins that the orchestrator
// dispatches through the CompilationStrategy abstraction it was given.
func TestExecuteParallelUsesInjectedStrategies(t *testing.T) {
	strategy := newBlockingStrategy(8)
	close(strategy.release) // never block

	orchestrator := NewParallelExecutionOrchestrator(strategy)
	require.NoError(t, orchestrator.ConfigureConcurrency(2))

	result, err := orchestrator.ExecuteParallel(context.Background(), tasksNamed(5, time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 5, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
}

// TestExecuteParallelWithNoStrategyFails keeps the previous behaviour for
// unhandled templates: a failure per task, not a panic or a silent success.
func TestExecuteParallelWithNoStrategyFails(t *testing.T) {
	orchestrator := NewParallelExecutionOrchestrator()
	require.NoError(t, orchestrator.ConfigureConcurrency(2))

	result, err := orchestrator.ExecuteParallel(context.Background(), tasksNamed(3, time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 0, result.SuccessCount)
	assert.Equal(t, 3, result.FailureCount)
}

// TestExecuteParallelHonoursCancellation pins that cancelling the caller's
// context stops the run instead of dispatching the whole backlog.
func TestExecuteParallelHonoursCancellation(t *testing.T) {
	strategy := newBlockingStrategy(64)
	orchestrator := NewParallelExecutionOrchestrator(strategy)
	require.NoError(t, orchestrator.ConfigureConcurrency(2))

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	var err error
	go func() {
		defer wg.Done()
		_, err = orchestrator.ExecuteParallel(ctx, tasksNamed(200, time.Minute))
	}()

	<-strategy.entered
	cancel()

	waitDone := make(chan struct{})
	go func() { wg.Wait(); close(waitDone) }()
	select {
	case <-waitDone:
	case <-time.After(5 * time.Second):
		t.Fatal("ExecuteParallel did not return after cancellation")
	}

	assert.True(t, errors.Is(err, context.Canceled),
		"expected context.Canceled, got %v", err)
}

// TestExecuteParallelFallsBackToConfiguredTimeout pins that a task carrying no
// timeout inherits the orchestrator's configured one. Passing a zero duration
// to context.WithTimeout yields an ALREADY-expired context, so treating an
// unset timeout as a real deadline fails every such task instantly.
func TestExecuteParallelFallsBackToConfiguredTimeout(t *testing.T) {
	strategy := newBlockingStrategy(8)
	close(strategy.release)

	orchestrator := NewParallelExecutionOrchestrator(strategy)
	require.NoError(t, orchestrator.ConfigureConcurrency(2))
	require.NoError(t, orchestrator.ConfigureTimeout(time.Minute))

	result, err := orchestrator.ExecuteParallel(context.Background(), tasksNamed(4, 0))
	require.NoError(t, err)
	assert.Equal(t, 4, result.SuccessCount,
		"tasks with an unset timeout must inherit the configured timeout, not expire immediately")
}
