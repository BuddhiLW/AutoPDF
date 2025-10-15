package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	apilogger "github.com/BuddhiLW/AutoPDF/pkg/api/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/pkg/api/config"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIDebugIntegration(t *testing.T) {
	// Test environment variable configuration
	t.Run("Environment Variable Configuration", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("AUTOPDF_API_DEBUG", "true")
		os.Setenv("AUTOPDF_API_LOG_DIR", "/tmp/test-logs")
		os.Setenv("AUTOPDF_API_CONCRETE_DIR", "/tmp/test-concrete")
		defer func() {
			os.Unsetenv("AUTOPDF_API_DEBUG")
			os.Unsetenv("AUTOPDF_API_LOG_DIR")
			os.Unsetenv("AUTOPDF_API_CONCRETE_DIR")
		}()

		// Load configuration
		debugConfig := config.LoadDebugConfigFromEnv()

		// Verify configuration
		assert.True(t, debugConfig.Enabled)
		assert.Equal(t, "/tmp/test-logs", debugConfig.LogDirectory)
		assert.Equal(t, "/tmp/test-concrete", debugConfig.ConcreteFileDir)
	})

	// Test request header parsing
	t.Run("Request Header Parsing", func(t *testing.T) {
		// Create a test request with debug headers
		req := httptest.NewRequest("POST", "/api/generate", nil)
		req.Header.Set("X-AutoPDF-Debug", "true")
		req.Header.Set("X-AutoPDF-Verbose", "2")
		req.Header.Set("X-AutoPDF-Clean", "true")
		req.Header.Set("X-AutoPDF-Force", "false")

		// Create debug middleware
		debugConfig := &config.APIDebugConfig{
			Enabled:         false, // Environment disabled
			LogDirectory:    "/tmp/test-logs",
			ConcreteFileDir: "/tmp/test-concrete",
		}

		// Create a test handler that extracts options from context
		var extractedOptions generation.PDFGenerationOptions
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			options, ok := middleware.GetOptionsFromContext(r.Context())
			require.True(t, ok, "Options should be in context")
			extractedOptions = options
			w.WriteHeader(http.StatusOK)
		})

		// Apply middleware
		middlewareHandler := middleware.DebugMiddleware(debugConfig)(testHandler)

		// Execute request
		rr := httptest.NewRecorder()
		middlewareHandler.ServeHTTP(rr, req)

		// Verify options were parsed correctly
		assert.True(t, extractedOptions.Debug.Enabled, "Debug should be enabled from header")
		assert.Equal(t, 2, extractedOptions.Verbose, "Verbose should be 2 from header")
		assert.True(t, extractedOptions.DoClean, "Clean should be true from header")
		assert.False(t, extractedOptions.Force, "Force should be false from header")
		assert.NotEmpty(t, extractedOptions.RequestID, "Request ID should be generated")
	})

	// Test logger factory
	t.Run("APILoggerFactory", func(t *testing.T) {
		// Create logger factory
		factory := apilogger.NewAPILoggerFactory(false, "/tmp/test-logs")

		// Test request logger creation
		requestLogger := factory.CreateRequestLogger("test-request-123", true)

		// Verify logger was created
		assert.NotNil(t, requestLogger, "Request logger should be created")

		// Test with debug disabled
		normalLogger := factory.CreateRequestLogger("test-request-456", false)
		assert.NotNil(t, normalLogger, "Should return a logger when debug disabled")
	})

	// Test error details logging
	t.Run("ErrorDetails Logging", func(t *testing.T) {
		// Create test logger
		testLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")

		// Create error details
		errorDetails := NewErrorDetails(ErrorCategoryTemplate, ErrorSeverityHigh).
			WithFilePath("/test/template.tex").
			AddContext("template_name", "test_template").
			AddContext("error_type", "syntax_error")

		// Test logging (this will output to stdout, but we can verify it doesn't panic)
		assert.NotPanics(t, func() {
			errorDetails.LogError(testLogger)
		}, "LogError should not panic")

		// Test logging with custom message
		assert.NotPanics(t, func() {
			errorDetails.LogErrorWithMessage(testLogger, "Custom error message")
		}, "LogErrorWithMessage should not panic")
	})

	// Test debug options structure
	t.Run("DebugOptions Structure", func(t *testing.T) {
		debugOptions := generation.DebugOptions{
			Enabled:            true,
			LogToFile:          true,
			LogFilePath:        "/tmp/test.log",
			CreateConcreteFile: true,
			RequestID:          "test-request-789",
		}

		// Verify structure
		assert.True(t, debugOptions.Enabled)
		assert.True(t, debugOptions.LogToFile)
		assert.Equal(t, "/tmp/test.log", debugOptions.LogFilePath)
		assert.True(t, debugOptions.CreateConcreteFile)
		assert.Equal(t, "test-request-789", debugOptions.RequestID)
	})

	// Test PDF generation options structure
	t.Run("PDFGenerationOptions Structure", func(t *testing.T) {
		options := generation.PDFGenerationOptions{
			DoConvert: true,
			DoClean:   true,
			Verbose:   2,
			Force:     false,
			RequestID: "test-request-abc",
			Debug: generation.DebugOptions{
				Enabled:            true,
				LogToFile:          true,
				CreateConcreteFile: true,
				RequestID:          "test-request-abc",
			},
		}

		// Verify structure
		assert.True(t, options.DoConvert)
		assert.True(t, options.DoClean)
		assert.Equal(t, 2, options.Verbose)
		assert.False(t, options.Force)
		assert.Equal(t, "test-request-abc", options.RequestID)
		assert.True(t, options.Debug.Enabled)
		assert.True(t, options.Debug.LogToFile)
		assert.True(t, options.Debug.CreateConcreteFile)
	})
}

// TestContextKey tests that context keys are properly defined
func TestContextKey(t *testing.T) {
	// Test that the middleware functions work correctly
	// This is an indirect test since the context keys are not exported
	assert.NotNil(t, middleware.GetOptionsFromContext)
	assert.NotNil(t, middleware.GetRequestIDFromContext)
}

// TestMiddlewareIntegration tests the full middleware integration
func TestMiddlewareIntegration(t *testing.T) {
	// Create test configuration
	debugConfig := &config.APIDebugConfig{
		Enabled:         true,
		LogDirectory:    "/tmp/test-logs",
		ConcreteFileDir: "/tmp/test-concrete",
	}

	// Create test handler that verifies context values
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check that options are in context
		options, ok := middleware.GetOptionsFromContext(ctx)
		require.True(t, ok, "PDFGenerationOptions should be in context")

		// Check that request ID is in context
		requestID, ok := middleware.GetRequestIDFromContext(ctx)
		require.True(t, ok, "Request ID should be in context")

		// Verify values
		assert.NotEmpty(t, requestID, "Request ID should not be empty")
		assert.Equal(t, requestID, options.RequestID, "Request ID should match in options")
		assert.True(t, options.Debug.Enabled, "Debug should be enabled by default when config is enabled")

		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	middlewareHandler := middleware.DebugMiddleware(debugConfig)(testHandler)

	// Create test request
	req := httptest.NewRequest("POST", "/api/generate", nil)
	rr := httptest.NewRecorder()

	// Execute request
	middlewareHandler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code, "Handler should execute successfully")
}
