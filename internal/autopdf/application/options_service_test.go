// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"testing"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockCleaner struct {
	mock.Mock
}

func (m *MockCleaner) CleanAux(ctx context.Context, target string) error {
	args := m.Called(ctx, target)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
	verbosity int
}

func (m *MockLogger) SetVerbosity(level int) {
	m.verbosity = level
}

func (m *MockLogger) Log(level int, message string, args ...interface{}) {
	m.Called(level, message, args)
}

type MockDebugger struct {
	mock.Mock
}

func (m *MockDebugger) EnableDebug(output string) {
	m.Called(output)
}

func (m *MockDebugger) Debug(message string, args ...interface{}) {
	m.Called(message, args)
}

type MockForcer struct {
	mock.Mock
	forceMode bool
	overwrite bool
}

func (m *MockForcer) SetForceMode(overwrite bool) {
	m.forceMode = true
	m.overwrite = overwrite
	m.Called(overwrite)
}

func (m *MockForcer) ShouldOverwrite() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestOptionsService_ExecuteOptions_NoOptions(t *testing.T) {
	// Create mocks
	cleaner := &MockCleaner{}
	logger := &MockLogger{}
	debugger := &MockDebugger{}
	forcer := &MockForcer{}

	// Create service
	service := NewOptionsService(cleaner, logger, debugger, forcer)

	// Create options with no enabled options
	options := domain.NewBuildOptions()

	// Execute options
	err := service.ExecuteOptions(context.Background(), options)

	// Should succeed with no calls to mocks
	assert.NoError(t, err)
	cleaner.AssertNotCalled(t, "CleanAux")
	logger.AssertNotCalled(t, "SetVerbosity")
	debugger.AssertNotCalled(t, "EnableDebug")
	forcer.AssertNotCalled(t, "SetForceMode")
}

func TestOptionsService_ExecuteOptions_CleanOnly(t *testing.T) {
	// Create mocks
	cleaner := &MockCleaner{}
	logger := &MockLogger{}
	debugger := &MockDebugger{}
	forcer := &MockForcer{}

	// Setup expectations
	cleaner.On("CleanAux", mock.Anything, "/tmp").Return(nil)

	// Create service
	service := NewOptionsService(cleaner, logger, debugger, forcer)

	// Create options with clean enabled
	options := domain.NewBuildOptions()
	options.EnableClean("/tmp")

	// Execute options
	err := service.ExecuteOptions(context.Background(), options)

	// Should succeed and call cleaner
	assert.NoError(t, err)
	cleaner.AssertExpectations(t)
}

func TestOptionsService_ExecuteOptions_VerboseOnly(t *testing.T) {
	// Create mocks
	cleaner := &MockCleaner{}
	logger := &MockLogger{}
	debugger := &MockDebugger{}
	forcer := &MockForcer{}

	// Create service
	service := NewOptionsService(cleaner, logger, debugger, forcer)

	// Create options with verbose enabled
	options := domain.NewBuildOptions()
	options.EnableVerbose(3)

	// Execute options
	err := service.ExecuteOptions(context.Background(), options)

	// Should succeed and set verbosity
	assert.NoError(t, err)
	assert.Equal(t, 3, logger.verbosity)
}

func TestOptionsService_ExecuteOptions_DebugOnly(t *testing.T) {
	// Create mocks
	cleaner := &MockCleaner{}
	logger := &MockLogger{}
	debugger := &MockDebugger{}
	forcer := &MockForcer{}

	// Setup expectations
	debugger.On("EnableDebug", "debug.log").Return()

	// Create service
	service := NewOptionsService(cleaner, logger, debugger, forcer)

	// Create options with debug enabled
	options := domain.NewBuildOptions()
	options.EnableDebug("debug.log")

	// Execute options
	err := service.ExecuteOptions(context.Background(), options)

	// Should succeed and enable debug
	assert.NoError(t, err)
	debugger.AssertExpectations(t)
}

func TestOptionsService_ExecuteOptions_ForceOnly(t *testing.T) {
	// Create mocks
	cleaner := &MockCleaner{}
	logger := &MockLogger{}
	debugger := &MockDebugger{}
	forcer := &MockForcer{}

	// Setup expectations
	forcer.On("SetForceMode", true).Return()

	// Create service
	service := NewOptionsService(cleaner, logger, debugger, forcer)

	// Create options with force enabled
	options := domain.NewBuildOptions()
	options.EnableForce(true)

	// Execute options
	err := service.ExecuteOptions(context.Background(), options)

	// Should succeed and set force mode
	assert.NoError(t, err)
	forcer.AssertExpectations(t)
}

func TestOptionsService_ExecuteOptions_AllOptions(t *testing.T) {
	// Create mocks
	cleaner := &MockCleaner{}
	logger := &MockLogger{}
	debugger := &MockDebugger{}
	forcer := &MockForcer{}

	// Setup expectations
	cleaner.On("CleanAux", mock.Anything, "/tmp").Return(nil)
	debugger.On("EnableDebug", "debug.log").Return()
	forcer.On("SetForceMode", true).Return()

	// Create service
	service := NewOptionsService(cleaner, logger, debugger, forcer)

	// Create options with all options enabled
	options := domain.NewBuildOptions()
	options.EnableClean("/tmp")
	options.EnableVerbose(3)
	options.EnableDebug("debug.log")
	options.EnableForce(true)

	// Execute options
	err := service.ExecuteOptions(context.Background(), options)

	// Should succeed and call all services
	assert.NoError(t, err)
	cleaner.AssertExpectations(t)
	debugger.AssertExpectations(t)
	forcer.AssertExpectations(t)
	assert.Equal(t, 3, logger.verbosity)
}

func TestOptionsService_ExecuteOptions_CleanError(t *testing.T) {
	// Create mocks
	cleaner := &MockCleaner{}
	logger := &MockLogger{}
	debugger := &MockDebugger{}
	forcer := &MockForcer{}

	// Setup expectations - cleaner returns error
	cleaner.On("CleanAux", mock.Anything, "/tmp").Return(assert.AnError)

	// Create service
	service := NewOptionsService(cleaner, logger, debugger, forcer)

	// Create options with clean enabled
	options := domain.NewBuildOptions()
	options.EnableClean("/tmp")

	// Execute options
	err := service.ExecuteOptions(context.Background(), options)

	// Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clean auxiliary files")
	cleaner.AssertExpectations(t)
}
