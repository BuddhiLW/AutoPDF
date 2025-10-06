package test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/BuddhiLW/AutoPDF/mocks"
	"github.com/BuddhiLW/AutoPDF/pkg/domain"
	"github.com/stretchr/testify/mock"
)

// BenchmarkTemplateEngine benchmarks the template engine performance
func BenchmarkTemplateEngine(b *testing.B) {
	mockEngine := mocks.NewMockTemplateEngine(b)

	// Set up expectations for all iterations
	mockEngine.EXPECT().
		Process(mock.AnythingOfType("string")).
		Return("benchmark result", nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := mockEngine.Process("benchmark.tex")
		if err != nil {
			b.Fatal(err)
		}
		if result != "benchmark result" {
			b.Fatal("unexpected result")
		}
	}
}

// BenchmarkEnhancedTemplateEngine benchmarks the enhanced template engine performance
func BenchmarkEnhancedTemplateEngine(b *testing.B) {
	mockEngine := mocks.NewMockEnhancedTemplateEngine(b)

	// Set up expectations for all iterations
	mockEngine.EXPECT().
		SetVariablesFromMap(mock.AnythingOfType("map[string]interface {}")).
		Return(nil).
		Times(b.N)

	mockEngine.EXPECT().
		Process(mock.AnythingOfType("string")).
		Return("enhanced benchmark result", nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		variables := map[string]interface{}{
			"title":   "Benchmark Test",
			"content": "Benchmark content",
		}

		err := mockEngine.SetVariablesFromMap(variables)
		if err != nil {
			b.Fatal(err)
		}

		result, err := mockEngine.Process("enhanced_benchmark.tex")
		if err != nil {
			b.Fatal(err)
		}
		if result != "enhanced benchmark result" {
			b.Fatal("unexpected result")
		}
	}
}

// BenchmarkVariableProcessing benchmarks variable processing performance
func BenchmarkVariableProcessing(b *testing.B) {
	mockProcessor := mocks.NewMockVariableProcessor(b)

	// Set up expectations for all iterations
	mockProcessor.EXPECT().
		ProcessVariables(mock.AnythingOfType("map[string]interface {}")).
		Return(domain.NewVariableCollection(), nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		variables := map[string]interface{}{
			"title":   "Benchmark Test",
			"author":  "Benchmark Author",
			"content": "Benchmark content for testing",
		}

		collection, err := mockProcessor.ProcessVariables(variables)
		if err != nil {
			b.Fatal(err)
		}
		if collection == nil {
			b.Fatal("unexpected nil collection")
		}
	}
}

// BenchmarkFileOperations benchmarks file operation performance
func BenchmarkFileOperations(b *testing.B) {
	mockFileProcessor := mocks.NewMockFileProcessor(b)

	// Set up expectations for all iterations
	mockFileProcessor.EXPECT().
		ReadFile(mock.AnythingOfType("string")).
		Return([]byte("benchmark file content"), nil).
		Times(b.N)

	mockFileProcessor.EXPECT().
		WriteFile(mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).
		Return(nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		content, err := mockFileProcessor.ReadFile("benchmark.tex")
		if err != nil {
			b.Fatal(err)
		}
		if len(content) == 0 {
			b.Fatal("unexpected empty content")
		}

		err = mockFileProcessor.WriteFile("output.tex", content)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentProcessing benchmarks concurrent processing performance
func BenchmarkConcurrentProcessing(b *testing.B) {
	mockEngine := mocks.NewMockTemplateEngine(b)

	// Set up expectations for all iterations
	mockEngine.EXPECT().
		Process(mock.AnythingOfType("string")).
		Return("concurrent benchmark result", nil).
		Times(b.N)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := mockEngine.Process("concurrent_benchmark.tex")
			if err != nil {
				b.Fatal(err)
			}
			if result != "concurrent benchmark result" {
				b.Fatal("unexpected result")
			}
		}
	})
}

// BenchmarkLargeDataStructures benchmarks large data structure processing
func BenchmarkLargeDataStructures(b *testing.B) {
	mockEngine := mocks.NewMockEnhancedTemplateEngine(b)

	// Set up expectations for all iterations
	mockEngine.EXPECT().
		SetVariablesFromMap(mock.AnythingOfType("map[string]interface {}")).
		Return(nil).
		Times(b.N)

	mockEngine.EXPECT().
		Process(mock.AnythingOfType("string")).
		Return("large data processed", nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create large data structure
		largeData := make(map[string]interface{})
		for j := 0; j < 1000; j++ {
			largeData[fmt.Sprintf("key_%d", j)] = fmt.Sprintf("value_%d", j)
		}

		err := mockEngine.SetVariablesFromMap(largeData)
		if err != nil {
			b.Fatal(err)
		}

		result, err := mockEngine.Process("large_data.tex")
		if err != nil {
			b.Fatal(err)
		}
		if result != "large data processed" {
			b.Fatal("unexpected result")
		}
	}
}

// BenchmarkNestedDataAccess benchmarks nested data access performance
func BenchmarkNestedDataAccess(b *testing.B) {
	mockProcessor := mocks.NewMockVariableProcessor(b)

	// Set up expectations for all iterations
	nestedVar, _ := domain.NewStringVariable("nested value")
	mockProcessor.EXPECT().
		GetNested(mock.AnythingOfType("string")).
		Return(nestedVar, nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		variable, err := mockProcessor.GetNested("document.title")
		if err != nil {
			b.Fatal(err)
		}
		if variable == nil {
			b.Fatal("unexpected nil variable")
		}
	}
}

// BenchmarkTemplateValidation benchmarks template validation performance
func BenchmarkTemplateValidation(b *testing.B) {
	mockValidator := mocks.NewMockTemplateValidator(b)

	// Set up expectations for all iterations
	mockValidator.EXPECT().
		ValidateTemplate(mock.AnythingOfType("string")).
		Return(nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := mockValidator.ValidateTemplate("benchmark_template.tex")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	mockEngine := mocks.NewMockTemplateEngine(b)

	// Set up expectations for all iterations
	mockEngine.EXPECT().
		Process(mock.AnythingOfType("string")).
		Return("memory allocation result", nil).
		Times(b.N)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Create new variables for each iteration to test allocation patterns
		variables := map[string]interface{}{
			"iteration": i,
			"data":      fmt.Sprintf("data_%d", i),
		}

		// Process template with new variables
		result, err := mockEngine.Process("allocation_test.tex")
		if err != nil {
			b.Fatal(err)
		}
		if result != "memory allocation result" {
			b.Fatal("unexpected result")
		}

		// Clear variables to test garbage collection
		_ = variables
	}
}

// BenchmarkStringOperations benchmarks string operation performance
func BenchmarkStringOperations(b *testing.B) {
	mockEngine := mocks.NewMockTemplateEngine(b)

	// Set up expectations for all iterations
	mockEngine.EXPECT().
		Process(mock.AnythingOfType("string")).
		Return("string operations result", nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test string operations
		template := fmt.Sprintf("\\documentclass{article}\\title{Test %d}\\begin{document}Content %d\\end{document}", i, i)

		result, err := mockEngine.Process(template)
		if err != nil {
			b.Fatal(err)
		}
		if result != "string operations result" {
			b.Fatal("unexpected result")
		}
	}
}

// BenchmarkErrorHandling benchmarks error handling performance
func BenchmarkErrorHandling(b *testing.B) {
	mockEngine := mocks.NewMockTemplateEngine(b)

	// Set up expectations for all iterations - simulate error handling
	mockEngine.EXPECT().
		Process(mock.AnythingOfType("string")).
		Return("", errors.New("benchmark error")).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := mockEngine.Process("error_handling.tex")
		// Always expect error in this benchmark
		if err == nil {
			b.Fatal("expected error")
		}
		if result != "" {
			b.Fatal("unexpected result")
		}
	}
}
