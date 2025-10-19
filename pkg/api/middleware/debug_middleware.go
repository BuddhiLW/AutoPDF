package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/api/config"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/google/uuid"
)

// Context keys for storing debug information
type contextKey string

const (
	OptionsContextKey   contextKey = "autopdf_options"
	RequestIDContextKey contextKey = "request_id"
)

// DebugMiddleware extracts debug options from request headers
func DebugMiddleware(baseConfig *config.APIDebugConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Generate unique request ID
			requestID := generateRequestID()

			// Check for per-request debug header
			debugHeader := r.Header.Get("X-AutoPDF-Debug")

			debugOptions := generation.DebugOptions{
				Enabled:            baseConfig.Enabled || debugHeader == "true",
				LogToFile:          baseConfig.Enabled || debugHeader == "true",
				CreateConcreteFile: baseConfig.Enabled || debugHeader == "true",
				RequestID:          requestID,
			}

			// Also check for other option headers
			options := generation.PDFGenerationOptions{
				Debug:     debugOptions,
				DoClean:   parseBoolHeader(r.Header.Get("X-AutoPDF-Clean")),
				Verbose:   parseVerboseLevel(r.Header.Get("X-AutoPDF-Verbose")),
				Force:     parseBoolHeader(r.Header.Get("X-AutoPDF-Force")),
				RequestID: requestID,
				DoConvert: parseBoolHeader(r.Header.Get("X-AutoPDF-Convert")),
				Timeout:   parseTimeout(r.Header.Get("X-AutoPDF-Timeout")),
				Conversion: generation.ConversionOptions{
					// Formats will be handled by the API handler based on request body
					Enabled: parseBoolHeader(r.Header.Get("X-AutoPDF-Convert")),
				},
			}

			// Add to context
			ctx = context.WithValue(ctx, OptionsContextKey, options)
			ctx = context.WithValue(ctx, RequestIDContextKey, requestID)

			// Add request ID to response headers for debugging
			w.Header().Set("X-AutoPDF-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetOptionsFromContext extracts PDF generation options from context
func GetOptionsFromContext(ctx context.Context) (generation.PDFGenerationOptions, bool) {
	options, ok := ctx.Value(OptionsContextKey).(generation.PDFGenerationOptions)
	return options, ok
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDContextKey).(string)
	return requestID, ok
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return uuid.New().String()[:8] // Use first 8 characters for brevity
}

// parseVerboseLevel parses verbose level from header value
func parseVerboseLevel(verboseHeader string) int {
	if verboseHeader == "" {
		return 0
	}

	level, err := strconv.Atoi(verboseHeader)
	if err != nil {
		return 0
	}

	// Clamp to reasonable range
	if level < 0 {
		return 0
	}
	if level > 5 {
		return 5
	}

	return level
}

// RequestLoggingMiddleware adds request logging with timing
func RequestLoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			// Log request details
			// Note: In a real implementation, you'd inject a logger here
			// For now, we'll just add timing to response headers
			wrapped.Header().Set("X-Response-Time", duration.String())
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// parseBoolHeader parses a boolean from a header string
func parseBoolHeader(headerValue string) bool {
	val, _ := strconv.ParseBool(headerValue)
	return val
}

// parseTimeout parses a duration from a header string
func parseTimeout(headerValue string) time.Duration {
	if headerValue == "" {
		return 0 // No timeout specified
	}
	duration, err := time.ParseDuration(headerValue)
	if err != nil {
		return 0 // Invalid duration, default to no timeout
	}
	return duration
}
