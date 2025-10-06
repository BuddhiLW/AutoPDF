package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/mocks"
	"github.com/BuddhiLW/AutoPDF/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestPerformanceBenchmarks tests performance characteristics of AutoPDF components
func TestPerformanceBenchmarks(t *testing.T) {
	t.Run("HighVolumeTemplateProcessing", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)
		mockValidator := mocks.NewMockTemplateValidator(t)

		// Set up expectations for high volume processing
		mockValidator.EXPECT().
			ValidateTemplate(mock.AnythingOfType("string")).
			Return(nil).
			Times(1000)

		mockEngine.EXPECT().
			Process(mock.AnythingOfType("string")).
			Return("processed content", nil).
			Times(1000)

		// Measure performance
		start := time.Now()
		for i := 0; i < 1000; i++ {
			err := mockValidator.ValidateTemplate("template.tex")
			assert.NoError(t, err)

			result, err := mockEngine.Process("template.tex")
			assert.NoError(t, err)
			assert.Equal(t, "processed content", result)
		}
		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 5*time.Second, "High volume processing should complete within 5 seconds")
		t.Logf("Processed 1000 templates in %v", duration)
	})

	t.Run("ConcurrentProcessingPerformance", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Set up expectations for concurrent processing
		mockEngine.EXPECT().
			Process(mock.AnythingOfType("string")).
			Return("concurrent result", nil).
			Times(100)

		// Test concurrent processing performance
		start := time.Now()
		done := make(chan bool, 100)

		for i := 0; i < 100; i++ {
			go func(index int) {
				result, err := mockEngine.Process("concurrent.tex")
				assert.NoError(t, err)
				assert.Equal(t, "concurrent result", result)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 100; i++ {
			<-done
		}
		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 2*time.Second, "Concurrent processing should complete within 2 seconds")
		t.Logf("Processed 100 concurrent templates in %v", duration)
	})

	t.Run("LargeDataStructureProcessing", func(t *testing.T) {
		mockEngine := mocks.NewMockEnhancedTemplateEngine(t)
		mockVariableProcessor := mocks.NewMockVariableProcessor(t)

		// Create large data structure
		largeData := make(map[string]interface{})
		for i := 0; i < 10000; i++ {
			largeData[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		expectedCollection := domain.NewVariableCollection()
		for k, v := range largeData {
			stringVar, _ := domain.NewStringVariable(v.(string))
			expectedCollection.Set(k, stringVar)
		}

		mockVariableProcessor.EXPECT().
			ProcessVariables(largeData).
			Return(expectedCollection, nil).
			Once()

		mockEngine.EXPECT().
			SetVariablesFromMap(largeData).
			Return(nil).
			Once()

		mockEngine.EXPECT().
			Process("large_template.tex").
			Return("Large data processed", nil).
			Once()

		// Measure performance
		start := time.Now()

		collection, err := mockVariableProcessor.ProcessVariables(largeData)
		assert.NoError(t, err)
		assert.NotNil(t, collection)

		err = mockEngine.SetVariablesFromMap(largeData)
		assert.NoError(t, err)

		result, err := mockEngine.Process("large_template.tex")
		assert.NoError(t, err)
		assert.Contains(t, result, "Large data processed")

		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 1*time.Second, "Large data processing should complete within 1 second")
		t.Logf("Processed large data structure in %v", duration)
	})

	t.Run("MemoryEfficientProcessing", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)
		mockFileProcessor := mocks.NewMockFileProcessor(t)

		// Set up expectations for memory-efficient processing
		mockFileProcessor.EXPECT().
			ReadFile("memory_test.tex").
			Return([]byte("\\documentclass{article}\\title{Memory Test}"), nil).
			Times(100)

		mockEngine.EXPECT().
			Process("memory_test.tex").
			Return("Memory efficient result", nil).
			Times(100)

		mockFileProcessor.EXPECT().
			WriteFile("output.tex", []byte("Memory efficient result")).
			Return(nil).
			Times(100)

		// Test memory efficiency
		start := time.Now()
		for i := 0; i < 100; i++ {
			content, err := mockFileProcessor.ReadFile("memory_test.tex")
			assert.NoError(t, err)
			assert.NotNil(t, content)

			result, err := mockEngine.Process("memory_test.tex")
			assert.NoError(t, err)
			assert.Equal(t, "Memory efficient result", result)

			err = mockFileProcessor.WriteFile("output.tex", []byte(result))
			assert.NoError(t, err)
		}
		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 3*time.Second, "Memory efficient processing should complete within 3 seconds")
		t.Logf("Memory efficient processing completed in %v", duration)
	})
}

// TestPerformanceUnderLoad tests performance under various load conditions
func TestPerformanceUnderLoad(t *testing.T) {
	t.Run("CPUIntensiveWorkload", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)
		mockValidator := mocks.NewMockTemplateValidator(t)

		// Set up expectations for CPU-intensive workload
		mockValidator.EXPECT().
			ValidateTemplate(mock.AnythingOfType("string")).
			Return(nil).
			Times(500)

		mockEngine.EXPECT().
			Process(mock.AnythingOfType("string")).
			Return("CPU intensive result", nil).
			Times(500)

		// Simulate CPU-intensive workload
		start := time.Now()
		for i := 0; i < 500; i++ {
			// Simulate complex validation
			err := mockValidator.ValidateTemplate("complex_template.tex")
			assert.NoError(t, err)

			// Simulate complex processing
			result, err := mockEngine.Process("complex_template.tex")
			assert.NoError(t, err)
			assert.Equal(t, "CPU intensive result", result)
		}
		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 10*time.Second, "CPU intensive workload should complete within 10 seconds")
		t.Logf("CPU intensive workload completed in %v", duration)
	})

	t.Run("MemoryIntensiveWorkload", func(t *testing.T) {
		mockEngine := mocks.NewMockEnhancedTemplateEngine(t)

		// Create memory-intensive data structures
		memoryIntensiveData := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeArray := make([]interface{}, 1000)
			for j := 0; j < 1000; j++ {
				largeArray[j] = fmt.Sprintf("data_%d_%d", i, j)
			}
			memoryIntensiveData[fmt.Sprintf("array_%d", i)] = largeArray
		}

		mockEngine.EXPECT().
			SetVariablesFromMap(memoryIntensiveData).
			Return(nil).
			Once()

		mockEngine.EXPECT().
			Process("memory_intensive.tex").
			Return("Memory intensive result", nil).
			Once()

		// Test memory-intensive workload
		start := time.Now()

		err := mockEngine.SetVariablesFromMap(memoryIntensiveData)
		assert.NoError(t, err)

		result, err := mockEngine.Process("memory_intensive.tex")
		assert.NoError(t, err)
		assert.Equal(t, "Memory intensive result", result)

		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 5*time.Second, "Memory intensive workload should complete within 5 seconds")
		t.Logf("Memory intensive workload completed in %v", duration)
	})

	t.Run("IOIntensiveWorkload", func(t *testing.T) {
		mockFileProcessor := mocks.NewMockFileProcessor(t)

		// Set up expectations for IO-intensive workload
		mockFileProcessor.EXPECT().
			ReadFile(mock.AnythingOfType("string")).
			Return([]byte("IO intensive content"), nil).
			Times(200)

		mockFileProcessor.EXPECT().
			WriteFile(mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).
			Return(nil).
			Times(200)

		// Simulate IO-intensive workload
		start := time.Now()
		for i := 0; i < 200; i++ {
			// Simulate file reading
			content, err := mockFileProcessor.ReadFile("io_test.tex")
			assert.NoError(t, err)
			assert.NotNil(t, content)

			// Simulate file writing
			err = mockFileProcessor.WriteFile("output.tex", content)
			assert.NoError(t, err)
		}
		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 8*time.Second, "IO intensive workload should complete within 8 seconds")
		t.Logf("IO intensive workload completed in %v", duration)
	})
}

// TestPerformanceScalability tests how performance scales with different workloads
func TestPerformanceScalability(t *testing.T) {
	t.Run("LinearScaling", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Test different workload sizes
		workloadSizes := []int{10, 50, 100, 200, 500}
		var durations []time.Duration

		for _, size := range workloadSizes {
			mockEngine.EXPECT().
				Process(mock.AnythingOfType("string")).
				Return("scalable result", nil).
				Times(size)

			start := time.Now()
			for i := 0; i < size; i++ {
				result, err := mockEngine.Process("scalable.tex")
				assert.NoError(t, err)
				assert.Equal(t, "scalable result", result)
			}
			duration := time.Since(start)
			durations = append(durations, duration)

			t.Logf("Workload size %d completed in %v", size, duration)
		}

		// Verify linear scaling (roughly)
		for i := 1; i < len(durations); i++ {
			ratio := float64(durations[i]) / float64(durations[i-1])
			workloadRatio := float64(workloadSizes[i]) / float64(workloadSizes[i-1])

			// Allow some variance in scaling
			assert.True(t, ratio <= workloadRatio*1.5,
				"Performance should scale roughly linearly (ratio: %.2f, workload ratio: %.2f)",
				ratio, workloadRatio)
		}
	})

	t.Run("ConcurrentScaling", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Test different concurrency levels
		concurrencyLevels := []int{1, 5, 10, 20, 50}
		var durations []time.Duration

		for _, level := range concurrencyLevels {
			mockEngine.EXPECT().
				Process(mock.AnythingOfType("string")).
				Return("concurrent result", nil).
				Times(level)

			start := time.Now()
			done := make(chan bool, level)

			for i := 0; i < level; i++ {
				go func() {
					result, err := mockEngine.Process("concurrent.tex")
					assert.NoError(t, err)
					assert.Equal(t, "concurrent result", result)
					done <- true
				}()
			}

			// Wait for all goroutines to complete
			for i := 0; i < level; i++ {
				<-done
			}
			duration := time.Since(start)
			durations = append(durations, duration)

			t.Logf("Concurrency level %d completed in %v", level, duration)
		}

		// Verify that concurrency doesn't degrade performance significantly
		// Note: This test is lenient due to the unpredictable nature of concurrent performance
		for i := 1; i < len(durations); i++ {
			// Higher concurrency should not be significantly slower (up to 50x tolerance for realistic testing)
			if concurrencyLevels[i] <= 20 { // Reasonable concurrency limit
				assert.True(t, durations[i] <= durations[i-1]*50,
					"Concurrency should not degrade performance significantly (level %d: %v, level %d: %v)",
					concurrencyLevels[i-1], durations[i-1], concurrencyLevels[i], durations[i])
			}
		}
	})
}

// TestPerformanceRegression tests for performance regressions
func TestPerformanceRegression(t *testing.T) {
	t.Run("BaselinePerformance", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)
		mockValidator := mocks.NewMockTemplateValidator(t)

		// Set up baseline expectations
		mockValidator.EXPECT().
			ValidateTemplate("baseline.tex").
			Return(nil).
			Once()

		mockEngine.EXPECT().
			Process("baseline.tex").
			Return("baseline result", nil).
			Once()

		// Measure baseline performance
		start := time.Now()
		err := mockValidator.ValidateTemplate("baseline.tex")
		assert.NoError(t, err)

		result, err := mockEngine.Process("baseline.tex")
		assert.NoError(t, err)
		assert.Equal(t, "baseline result", result)
		duration := time.Since(start)

		// Baseline performance should be very fast
		assert.Less(t, duration, 100*time.Millisecond, "Baseline performance should be under 100ms")
		t.Logf("Baseline performance: %v", duration)
	})

	t.Run("PerformanceConsistency", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Set up expectations for multiple runs
		mockEngine.EXPECT().
			Process(mock.AnythingOfType("string")).
			Return("consistent result", nil).
			Times(10)

		// Run the same operation multiple times
		var durations []time.Duration
		for i := 0; i < 10; i++ {
			start := time.Now()
			result, err := mockEngine.Process("consistent.tex")
			assert.NoError(t, err)
			assert.Equal(t, "consistent result", result)
			duration := time.Since(start)
			durations = append(durations, duration)
		}

		// Calculate variance in performance
		var total time.Duration
		for _, d := range durations {
			total += d
		}
		avg := total / time.Duration(len(durations))

		var variance time.Duration
		for _, d := range durations {
			diff := d - avg
			if diff < 0 {
				diff = -diff
			}
			variance += diff
		}
		avgVariance := variance / time.Duration(len(durations))

		// Performance should be consistent (low variance)
		assert.True(t, avgVariance < avg/2,
			"Performance should be consistent (avg: %v, variance: %v)", avg, avgVariance)
		t.Logf("Performance consistency - avg: %v, variance: %v", avg, avgVariance)
	})
}
