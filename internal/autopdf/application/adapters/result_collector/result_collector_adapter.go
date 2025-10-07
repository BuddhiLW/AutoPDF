// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package result_collector

import (
	"sync"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/parallel"
)

// ResultCollectorAdapter implements the CompilationResultCollector interface
type ResultCollectorAdapter struct {
	successfulBuilds []parallel.BuildResult
	failedBuilds     []parallel.BuildFailure
	mutex            sync.RWMutex
	startTime        time.Time
}

// NewResultCollectorAdapter creates a new result collector adapter
func NewResultCollectorAdapter() *ResultCollectorAdapter {
	return &ResultCollectorAdapter{
		successfulBuilds: make([]parallel.BuildResult, 0),
		failedBuilds:     make([]parallel.BuildFailure, 0),
		startTime:        time.Now(),
	}
}

// AddSuccess adds a successful build result
func (r *ResultCollectorAdapter) AddSuccess(result parallel.BuildResult) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.successfulBuilds = append(r.successfulBuilds, result)
}

// AddFailure adds a failed build result
func (r *ResultCollectorAdapter) AddFailure(failure parallel.BuildFailure) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.failedBuilds = append(r.failedBuilds, failure)
}

// GetResults returns the aggregated compilation results
func (r *ResultCollectorAdapter) GetResults() *parallel.ParallelCompilationResult {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return &parallel.ParallelCompilationResult{
		SuccessfulBuilds: r.successfulBuilds,
		FailedBuilds:     r.failedBuilds,
		TotalDuration:    time.Since(r.startTime),
		SuccessCount:     len(r.successfulBuilds),
		FailureCount:     len(r.failedBuilds),
	}
}

// Reset clears all collected results
func (r *ResultCollectorAdapter) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.successfulBuilds = make([]parallel.BuildResult, 0)
	r.failedBuilds = make([]parallel.BuildFailure, 0)
	r.startTime = time.Now()
}

// GetSuccessCount returns the number of successful builds
func (r *ResultCollectorAdapter) GetSuccessCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.successfulBuilds)
}

// GetFailureCount returns the number of failed builds
func (r *ResultCollectorAdapter) GetFailureCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.failedBuilds)
}

// GetTotalCount returns the total number of builds
func (r *ResultCollectorAdapter) GetTotalCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.successfulBuilds) + len(r.failedBuilds)
}

// IsComplete checks if all expected builds are complete
func (r *ResultCollectorAdapter) IsComplete(expectedCount int) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.GetTotalCount() >= expectedCount
}
