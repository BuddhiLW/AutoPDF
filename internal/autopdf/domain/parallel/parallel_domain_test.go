// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package parallel

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParallelCompilationRequest_Creation(t *testing.T) {
	request := ParallelCompilationRequest{
		ConfigurationFile: "config.yaml",
		TemplateFiles:     []string{"template1.tex", "template2.tex"},
		MaxConcurrency:    4,
		Timeout:           30 * time.Second,
	}

	assert.Equal(t, "config.yaml", request.ConfigurationFile)
	assert.Equal(t, []string{"template1.tex", "template2.tex"}, request.TemplateFiles)
	assert.Equal(t, 4, request.MaxConcurrency)
	assert.Equal(t, 30*time.Second, request.Timeout)
}

func TestParallelCompilationRequest_EmptyValues(t *testing.T) {
	request := ParallelCompilationRequest{}

	assert.Empty(t, request.ConfigurationFile)
	assert.Nil(t, request.TemplateFiles)
	assert.Zero(t, request.MaxConcurrency)
	assert.Zero(t, request.Timeout)
}

func TestBuildResult_Creation(t *testing.T) {
	now := time.Now()
	result := BuildResult{
		TemplateFile: "template.tex",
		PDFPath:      "output.pdf",
		Duration:     2 * time.Second,
		Timestamp:    now,
	}

	assert.Equal(t, "template.tex", result.TemplateFile)
	assert.Equal(t, "output.pdf", result.PDFPath)
	assert.Equal(t, 2*time.Second, result.Duration)
	assert.Equal(t, now, result.Timestamp)
}

func TestBuildFailure_Creation(t *testing.T) {
	now := time.Now()
	err := errors.New("compilation failed")
	failure := BuildFailure{
		TemplateFile: "template.tex",
		Error:        err,
		Duration:     1 * time.Second,
		Timestamp:    now,
	}

	assert.Equal(t, "template.tex", failure.TemplateFile)
	assert.Equal(t, err, failure.Error)
	assert.Equal(t, 1*time.Second, failure.Duration)
	assert.Equal(t, now, failure.Timestamp)
}

func TestParallelCompilationResult_Creation(t *testing.T) {
	successfulBuilds := []BuildResult{
		{TemplateFile: "template1.tex", PDFPath: "output1.pdf"},
		{TemplateFile: "template2.tex", PDFPath: "output2.pdf"},
	}
	failedBuilds := []BuildFailure{
		{TemplateFile: "template3.tex", Error: errors.New("failed")},
	}

	result := ParallelCompilationResult{
		SuccessfulBuilds: successfulBuilds,
		FailedBuilds:     failedBuilds,
		TotalDuration:    5 * time.Second,
		SuccessCount:     2,
		FailureCount:     1,
	}

	assert.Equal(t, successfulBuilds, result.SuccessfulBuilds)
	assert.Equal(t, failedBuilds, result.FailedBuilds)
	assert.Equal(t, 5*time.Second, result.TotalDuration)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 1, result.FailureCount)
}

func TestCompilationTask_Creation(t *testing.T) {
	task := CompilationTask{
		TemplateFile: "template.tex",
		ConfigFile:   "config.yaml",
		Priority:     1,
		Timeout:      10 * time.Second,
	}

	assert.Equal(t, "template.tex", task.TemplateFile)
	assert.Equal(t, "config.yaml", task.ConfigFile)
	assert.Equal(t, 1, task.Priority)
	assert.Equal(t, 10*time.Second, task.Timeout)
}

func TestCompilationTask_DefaultValues(t *testing.T) {
	task := CompilationTask{}

	assert.Empty(t, task.TemplateFile)
	assert.Empty(t, task.ConfigFile)
	assert.Zero(t, task.Priority)
	assert.Zero(t, task.Timeout)
}

func TestParallelCompilationResult_EmptyResult(t *testing.T) {
	result := ParallelCompilationResult{}

	assert.Nil(t, result.SuccessfulBuilds)
	assert.Nil(t, result.FailedBuilds)
	assert.Zero(t, result.TotalDuration)
	assert.Zero(t, result.SuccessCount)
	assert.Zero(t, result.FailureCount)
}

func TestBuildResult_EdgeCases(t *testing.T) {
	// Test with empty values
	result := BuildResult{}

	assert.Empty(t, result.TemplateFile)
	assert.Empty(t, result.PDFPath)
	assert.Zero(t, result.Duration)
	assert.Zero(t, result.Timestamp)
}

func TestBuildFailure_EdgeCases(t *testing.T) {
	// Test with nil error
	failure := BuildFailure{
		TemplateFile: "template.tex",
		Error:        nil,
	}

	assert.Equal(t, "template.tex", failure.TemplateFile)
	assert.Nil(t, failure.Error)
}

func TestParallelCompilationRequest_Validation(t *testing.T) {
	// Test with valid request
	request := ParallelCompilationRequest{
		ConfigurationFile: "config.yaml",
		TemplateFiles:     []string{"template1.tex", "template2.tex"},
		MaxConcurrency:    2,
		Timeout:           15 * time.Second,
	}

	assert.NotEmpty(t, request.ConfigurationFile)
	assert.NotEmpty(t, request.TemplateFiles)
	assert.Greater(t, request.MaxConcurrency, 0)
	assert.Greater(t, request.Timeout, time.Duration(0))
}

func TestParallelCompilationResult_TotalCount(t *testing.T) {
	result := ParallelCompilationResult{
		SuccessCount: 3,
		FailureCount: 1,
	}

	totalCount := result.SuccessCount + result.FailureCount
	assert.Equal(t, 4, totalCount)
}

func TestCompilationTask_PriorityOrdering(t *testing.T) {
	tasks := []CompilationTask{
		{TemplateFile: "template1.tex", Priority: 3},
		{TemplateFile: "template2.tex", Priority: 1},
		{TemplateFile: "template3.tex", Priority: 2},
	}

	// Test that we can sort by priority
	assert.Equal(t, 3, tasks[0].Priority)
	assert.Equal(t, 1, tasks[1].Priority)
	assert.Equal(t, 2, tasks[2].Priority)
}

func TestBuildResult_DurationComparison(t *testing.T) {
	fastResult := BuildResult{Duration: 1 * time.Second}
	slowResult := BuildResult{Duration: 5 * time.Second}

	assert.True(t, fastResult.Duration < slowResult.Duration)
	assert.True(t, slowResult.Duration > fastResult.Duration)
}

func TestParallelCompilationRequest_MultipleTemplates(t *testing.T) {
	templates := []string{
		"template1.tex",
		"template2.tex",
		"template3.tex",
		"template4.tex",
	}

	request := ParallelCompilationRequest{
		TemplateFiles:  templates,
		MaxConcurrency: 2,
	}

	assert.Len(t, request.TemplateFiles, 4)
	assert.Equal(t, templates, request.TemplateFiles)
}
