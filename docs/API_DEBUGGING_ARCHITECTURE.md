# API Debugging Architecture: A Complete Guide for REST Server Development

## 🎯 Overview

This document describes a comprehensive API debugging architecture implemented for the AutoPDF REST server, demonstrating how to build production-ready debugging capabilities that scale from development to production environments.

## 🏗️ Architecture Principles

### CLARITY Principles Applied

- **C - Compose**: Debug features are composed via dependency injection, not embedded
- **L - Layer Purity**: Logging concerns stay in adapters, domain remains clean
- **A - Architectural Performance**: Debug logging is opt-in with zero overhead when disabled
- **R - Represent Intent**: Debug options clearly express debugging intent through types
- **I - Input Guarded**: Request headers and environment variables are validated
- **T - Telemetry First**: Structured logging is first-class throughout the system
- **Y - Yield Safe Failure**: Debug failures don't crash the API, graceful degradation

## 🔧 Core Components

### 1. Request-Scoped Logging Factory

**Purpose**: Create isolated loggers for each API request to prevent log contamination.

```go
type APILoggerFactory struct {
    baseLogger *logger.LoggerAdapter
    logDir     string
}

func (f *APILoggerFactory) CreateRequestLogger(requestID string, debugEnabled bool) *logger.LoggerAdapter {
    if !debugEnabled {
        return f.baseLogger
    }
    
    // Create request-specific log file
    logFile := filepath.Join(f.logDir, fmt.Sprintf("autopdf-api-%s.log", requestID))
    
    // Return logger with request context
    return logger.NewLoggerAdapter(logger.Debug, "stdout").
        WithFields(zap.String("request_id", requestID))
}
```

**Benefits**:
- Isolated debugging per request
- Easy correlation of logs across services
- No log contamination between concurrent requests

### 2. Environment-Based Configuration

**Purpose**: Global debug settings that can be toggled without code changes.

```go
type APIDebugConfig struct {
    Enabled         bool   // AUTOPDF_API_DEBUG=true
    LogDirectory    string // AUTOPDF_API_LOG_DIR=/var/log/autopdf
    ConcreteFileDir string // AUTOPDF_API_CONCRETE_DIR=/tmp/autopdf
}

func LoadDebugConfigFromEnv() *APIDebugConfig {
    enabled, _ := strconv.ParseBool(os.Getenv("AUTOPDF_API_DEBUG"))
    logDir := getEnvOrDefault("AUTOPDF_API_LOG_DIR", "/tmp/autopdf/logs")
    concreteFileDir := getEnvOrDefault("AUTOPDF_API_CONCRETE_DIR", "/tmp/autopdf/concrete")
    
    return &APIDebugConfig{
        Enabled:         enabled,
        LogDirectory:    logDir,
        ConcreteFileDir: concreteFileDir,
    }
}
```

**Benefits**:
- Zero-code configuration changes
- Environment-specific debug levels
- Production-safe defaults

### 3. HTTP Middleware for Per-Request Debug

**Purpose**: Extract debug options from request headers and inject into context.

```go
func DebugMiddleware(baseConfig *config.APIDebugConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()
            requestID := generateRequestID()
            
            // Determine debug options from environment and headers
            debugEnabled := baseConfig.Enabled || parseBoolHeader(r.Header.Get("X-AutoPDF-Debug"))
            
            debugOptions := generation.DebugOptions{
                Enabled:            debugEnabled,
                LogToFile:          debugEnabled,
                CreateConcreteFile: debugEnabled,
                RequestID:          requestID,
            }
            
            // Parse other option headers
            options := generation.PDFGenerationOptions{
                Debug:      debugOptions,
                DoClean:    parseBoolHeader(r.Header.Get("X-AutoPDF-Clean")),
                Verbose:    parseVerboseLevel(r.Header.Get("X-AutoPDF-Verbose")),
                Force:      parseBoolHeader(r.Header.Get("X-AutoPDF-Force")),
                RequestID:  requestID,
            }
            
            // Inject into context
            ctx = context.WithValue(ctx, ContextKeyAutoPDFOptions, options)
            ctx = context.WithValue(ctx, ContextKeyRequestID, requestID)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

**Benefits**:
- Per-request debug control
- No API contract changes
- Backward compatible

### 4. Debug-Aware Template Processing

**Purpose**: Create persistent intermediate files for debugging complex transformations.

```go
type DebugTemplateProcessorDecorator struct {
    wrappedService  generation.TemplateProcessingService
    logger         *logger.LoggerAdapter
    concreteFileDir string
}

func (d *DebugTemplateProcessorDecorator) Process(
    ctx context.Context,
    templatePath string,
    variables map[string]interface{},
    debugOptions generation.DebugOptions,
) (string, error) {
    // Process template normally
    processedContent, err := d.wrappedService.Process(ctx, templatePath, variables)
    if err != nil {
        return "", err
    }
    
    // Create persistent concrete file in debug mode
    if debugOptions.Enabled && debugOptions.CreateConcreteFile {
        concreteFile := d.createConcreteFile(templatePath, processedContent, debugOptions.RequestID)
        d.logger.InfoWithFields("Created concrete template file",
            "path", concreteFile,
            "request_id", debugOptions.RequestID,
            "original_template", templatePath,
        )
    }
    
    return processedContent, nil
}
```

**Benefits**:
- Visual debugging of template transformations
- Persistent intermediate files
- Request-specific file naming

### 5. Structured Error Logging

**Purpose**: Consistent error logging with full context and severity levels.

```go
func (ed *ErrorDetails) LogError(logger *logger.LoggerAdapter) *ErrorDetails {
    fields := []interface{}{
        "category", ed.Category,
        "severity", ed.Severity,
        "timestamp", ed.Timestamp,
    }
    
    if ed.FilePath != "" {
        fields = append(fields, "file_path", ed.FilePath)
    }
    if len(ed.Context) > 0 {
        fields = append(fields, "context", ed.Context)
    }
    
    // Log at appropriate level based on severity
    switch ed.Severity {
    case "high":
        logger.ErrorWithFields("PDF generation error", fields...)
    case "medium":
        logger.WarnWithFields("PDF generation warning", fields...)
    case "low":
        logger.InfoWithFields("PDF generation info", fields...)
    }
    
    return ed
}
```

**Benefits**:
- Consistent error formatting
- Severity-based log levels
- Rich context for debugging

## 🚀 Implementation Guide for REST Servers

### Step 1: Define Debug Domain Types

Create clear, typed structures for debug options:

```go
type DebugOptions struct {
    Enabled           bool
    LogToFile         bool
    LogFilePath       string
    CreateConcreteFile bool
    RequestID         string
}

type APIGenerationOptions struct {
    Debug      DebugOptions
    DoClean    bool
    Verbose    int
    Force      bool
    RequestID  string
}
```

### Step 2: Implement Logger Factory

Create a factory that can generate request-scoped loggers:

```go
type APILoggerFactory struct {
    baseLogger *logger.LoggerAdapter
    logDir     string
}

func NewAPILoggerFactory(debugEnabled bool, logDir string) *APILoggerFactory {
    level := logger.Detailed
    if debugEnabled {
        level = logger.Debug
    }
    
    baseLogger := logger.NewLoggerAdapter(level, "stdout")
    
    return &APILoggerFactory{
        baseLogger: baseLogger,
        logDir:     logDir,
    }
}
```

### Step 3: Create Environment Configuration

Support both environment variables and request headers:

```go
type APIDebugConfig struct {
    Enabled         bool
    LogDirectory    string
    ConcreteFileDir string
}

func LoadDebugConfigFromEnv() *APIDebugConfig {
    enabled, _ := strconv.ParseBool(os.Getenv("API_DEBUG"))
    logDir := getEnvOrDefault("API_LOG_DIR", "/tmp/api/logs")
    concreteFileDir := getEnvOrDefault("API_CONCRETE_DIR", "/tmp/api/concrete")
    
    return &APIDebugConfig{
        Enabled:         enabled,
        LogDirectory:    logDir,
        ConcreteFileDir: concreteFileDir,
    }
}
```

### Step 4: Implement HTTP Middleware

Create middleware that extracts debug options from headers:

```go
func DebugMiddleware(baseConfig *APIDebugConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()
            requestID := generateRequestID()
            
            // Parse debug options from headers
            debugEnabled := baseConfig.Enabled || parseBoolHeader(r.Header.Get("X-API-Debug"))
            
            options := APIGenerationOptions{
                Debug: DebugOptions{
                    Enabled:            debugEnabled,
                    LogToFile:          debugEnabled,
                    CreateConcreteFile: debugEnabled,
                    RequestID:          requestID,
                },
                DoClean:   parseBoolHeader(r.Header.Get("X-API-Clean")),
                Verbose:   parseVerboseLevel(r.Header.Get("X-API-Verbose")),
                Force:     parseBoolHeader(r.Header.Get("X-API-Force")),
                RequestID: requestID,
            }
            
            // Inject into context
            ctx = context.WithValue(ctx, ContextKeyAPIOptions, options)
            ctx = context.WithValue(ctx, ContextKeyRequestID, requestID)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Step 5: Integrate Logging into Services

Inject loggers into application services:

```go
type ApplicationService struct {
    repository SomeRepository
    logger     *logger.LoggerAdapter
}

func NewApplicationService(repo SomeRepository, logger *logger.LoggerAdapter) *ApplicationService {
    return &ApplicationService{
        repository: repo,
        logger:     logger,
    }
}

func (s *ApplicationService) ProcessRequest(ctx context.Context, req Request) (Response, error) {
    s.logger.InfoWithFields("Processing request",
        "request_id", getRequestIDFromContext(ctx),
        "request_type", req.Type,
    )
    
    // Process with structured logging
    result, err := s.repository.Find(ctx, req.ID)
    if err != nil {
        s.logger.ErrorWithFields("Repository error",
            "request_id", getRequestIDFromContext(ctx),
            "error", err,
        )
        return Response{}, err
    }
    
    s.logger.InfoWithFields("Request processed successfully",
        "request_id", getRequestIDFromContext(ctx),
        "result_count", len(result),
    )
    
    return Response{Data: result}, nil
}
```

## 🧪 Testing Strategy

### 1. Integration Testing

Test the complete debug flow:

```go
func TestAPIDebugIntegration(t *testing.T) {
    // Test environment variable configuration
    t.Run("Environment Variable Configuration", func(t *testing.T) {
        os.Setenv("API_DEBUG", "true")
        os.Setenv("API_LOG_DIR", "/tmp/test-logs")
        defer func() {
            os.Unsetenv("API_DEBUG")
            os.Unsetenv("API_LOG_DIR")
        }()
        
        config := LoadDebugConfigFromEnv()
        assert.True(t, config.Enabled)
        assert.Equal(t, "/tmp/test-logs", config.LogDirectory)
    })
    
    // Test request header parsing
    t.Run("Request Header Parsing", func(t *testing.T) {
        req := httptest.NewRequest("POST", "/api/process", nil)
        req.Header.Set("X-API-Debug", "true")
        req.Header.Set("X-API-Verbose", "2")
        
        debugConfig := &APIDebugConfig{Enabled: false}
        testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            options, ok := GetOptionsFromContext(r.Context())
            require.True(t, ok)
            assert.True(t, options.Debug.Enabled)
            assert.Equal(t, 2, options.Verbose)
        })
        
        middlewareHandler := DebugMiddleware(debugConfig)(testHandler)
        rr := httptest.NewRecorder()
        middlewareHandler.ServeHTTP(rr, req)
        
        assert.Equal(t, http.StatusOK, rr.Code)
    })
}
```

### 2. Logger Factory Testing

Test logger creation and scoping:

```go
func TestAPILoggerFactory(t *testing.T) {
    factory := NewAPILoggerFactory(false, "/tmp/test-logs")
    
    // Test request logger creation
    requestLogger := factory.CreateRequestLogger("test-request-123", true)
    assert.NotNil(t, requestLogger)
    
    // Test with debug disabled
    normalLogger := factory.CreateRequestLogger("test-request-456", false)
    assert.NotNil(t, normalLogger)
}
```

### 3. Error Logging Testing

Test structured error logging:

```go
func TestErrorDetailsLogging(t *testing.T) {
    testLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")
    
    errorDetails := NewErrorDetails("processing", "high").
        WithFilePath("/test/file.txt").
        AddContext("operation", "test")
    
    // Test logging doesn't panic
    assert.NotPanics(t, func() {
        errorDetails.LogError(testLogger)
    })
}
```

## 🎯 Usage Patterns

### Development Environment

```bash
# Enable global debug logging
export API_DEBUG=true
export API_LOG_DIR=/var/log/api
export API_CONCRETE_DIR=/tmp/api/concrete

# Start server with debug enabled
./api-server
```

### Per-Request Debugging

```bash
# Single request with debug enabled
curl -X POST http://localhost:8080/api/v1/process \
  -H "X-API-Debug: true" \
  -H "X-API-Clean: true" \
  -H "X-API-Verbose: 2" \
  -d @request.json

# Creates:
# - /var/log/api/api-{request-id}.log
# - /tmp/api/concrete/concrete-{request-id}.txt
```

### Production Monitoring

```bash
# Enable error-level logging only
export API_DEBUG=false
export API_LOG_LEVEL=error

# Monitor specific request
curl -X POST http://localhost:8080/api/v1/process \
  -H "X-API-Debug: true" \
  -H "X-Request-ID: prod-issue-123" \
  -d @request.json
```

## 📊 Log Output Examples

### Request Processing Log

```
2025-10-15T09:30:00.123-0300    INFO    Processing request    {"request_id": "abc123", "request_type": "pdf_generation"}
2025-10-15T09:30:00.124-0300    DEBUG   Variable resolution starting    {"request_id": "abc123", "variable_count": 5}
2025-10-15T09:30:00.125-0300    DEBUG   Resolved variable       {"request_id": "abc123", "key": "title", "resolved_value": "My Document"}
2025-10-15T09:30:00.130-0300    INFO    Variable resolution complete    {"request_id": "abc123", "input_count": 5, "output_count": 5}
2025-10-15T09:30:00.135-0300    INFO    Created concrete template file  {"request_id": "abc123", "path": "/tmp/api/concrete/concrete-template-abc123.tex"}
2025-10-15T09:30:00.200-0300    INFO    Request processed successfully    {"request_id": "abc123", "duration": "77ms"}
```

### Error Logging

```
2025-10-15T09:30:00.123-0300    ERROR   PDF generation error    {"category": "template_processing", "severity": "high", "timestamp": "2025-10-15T09:30:00.123-0300", "file_path": "/test/template.tex", "context": {"error_type":"syntax_error","template_name":"test_template"}}
```

## 🔍 Debugging Workflows

### 1. Development Debugging

1. **Enable global debug**: `export API_DEBUG=true`
2. **Make request**: Use normal API calls
3. **Check logs**: Review structured logs in console
4. **Inspect files**: Check concrete files in `/tmp/api/concrete/`

### 2. Production Issue Investigation

1. **Identify request**: Get request ID from error logs
2. **Enable per-request debug**: Add `X-API-Debug: true` header
3. **Reproduce issue**: Make same request with debug enabled
4. **Analyze logs**: Review request-specific log file
5. **Check concrete files**: Inspect intermediate processing files

### 3. Performance Analysis

1. **Enable verbose logging**: `X-API-Verbose: 3`
2. **Monitor timing**: Look for duration fields in logs
3. **Identify bottlenecks**: Find slow operations
4. **Optimize**: Focus on high-duration operations

## 🛡️ Security Considerations

### 1. Log File Permissions

```go
// Ensure log directories have proper permissions
if err := os.MkdirAll(logDir, 0755); err != nil {
    return fmt.Errorf("failed to create log directory: %w", err)
}
```

### 2. Sensitive Data Filtering

```go
func (s *Service) ProcessRequest(ctx context.Context, req Request) (Response, error) {
    // Log request without sensitive data
    s.logger.InfoWithFields("Processing request",
        "request_id", getRequestIDFromContext(ctx),
        "request_type", req.Type,
        "user_id", req.UserID, // Safe to log
        // Don't log: req.Password, req.Token, etc.
    )
}
```

### 3. Debug Mode Restrictions

```go
// Only allow debug mode in development
func isDebugAllowed() bool {
    env := os.Getenv("ENVIRONMENT")
    return env == "development" || env == "staging"
}
```

## 📈 Performance Impact

### Zero Overhead When Disabled

- No debug logging when `API_DEBUG=false`
- No file I/O for concrete files
- No additional memory allocation
- Minimal CPU overhead

### Controlled Overhead When Enabled

- Structured logging adds ~1-2ms per request
- File I/O for concrete files: ~5-10ms
- Memory usage: ~1-2MB per concurrent request
- Disk usage: ~10-50KB per request (logs + concrete files)

## 🎓 Best Practices

### 1. Log Levels

- **ERROR**: System errors, failures
- **WARN**: Recoverable issues, degraded performance
- **INFO**: Important business events, request completion
- **DEBUG**: Detailed processing steps, variable values

### 2. Context Fields

Always include:
- `request_id`: For correlation
- `user_id`: For user-specific debugging
- `operation`: For understanding what's happening
- `duration`: For performance analysis

### 3. Error Handling

```go
// Always log errors with context
if err != nil {
    s.logger.ErrorWithFields("Operation failed",
        "request_id", getRequestIDFromContext(ctx),
        "operation", "template_processing",
        "error", err,
        "retry_count", retryCount,
    )
    return err
}
```

### 4. Concrete File Management

```go
// Clean up old concrete files periodically
func cleanupOldConcreteFiles(dir string, maxAge time.Duration) {
    files, _ := filepath.Glob(filepath.Join(dir, "concrete-*.txt"))
    for _, file := range files {
        if info, err := os.Stat(file); err == nil {
            if time.Since(info.ModTime()) > maxAge {
                os.Remove(file)
            }
        }
    }
}
```

## 🚀 Advanced Features

### 1. Distributed Tracing

```go
// Add trace ID to logs
func (s *Service) ProcessRequest(ctx context.Context, req Request) (Response, error) {
    traceID := getTraceIDFromContext(ctx)
    s.logger.InfoWithFields("Processing request",
        "request_id", getRequestIDFromContext(ctx),
        "trace_id", traceID,
    )
}
```

### 2. Metrics Integration

```go
// Add metrics to debug logs
func (s *Service) ProcessRequest(ctx context.Context, req Request) (Response, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        s.logger.InfoWithFields("Request completed",
            "request_id", getRequestIDFromContext(ctx),
            "duration_ms", duration.Milliseconds(),
            "memory_mb", getMemoryUsage(),
        )
    }()
}
```

### 3. Conditional Debug Features

```go
// Enable advanced debugging based on request complexity
func (s *Service) ProcessRequest(ctx context.Context, req Request) (Response, error) {
    options := getOptionsFromContext(ctx)
    
    // Enable concrete files for complex requests
    if len(req.Variables) > 10 && options.Debug.Enabled {
        options.Debug.CreateConcreteFile = true
    }
}
```

## 📚 Conclusion

This API debugging architecture provides:

1. **Comprehensive Observability**: Full request lifecycle logging
2. **Flexible Configuration**: Environment and per-request control
3. **Production Safety**: Zero overhead when disabled
4. **Developer Experience**: Easy debugging with concrete files
5. **Maintainability**: Clean separation of concerns
6. **Scalability**: Request-scoped logging prevents contamination

The architecture follows CLARITY principles and provides a solid foundation for building debuggable, maintainable REST APIs that can scale from development to production environments.

## 🔗 Related Resources

- [CLARITY Principles Documentation](./CLARITY_PRINCIPLES.md)
- [Structured Logging Best Practices](./LOGGING_BEST_PRACTICES.md)
- [API Testing Strategies](./API_TESTING_STRATEGIES.md)
- [Production Debugging Guide](./PRODUCTION_DEBUGGING.md)
