// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/api/application"
	"github.com/BuddhiLW/AutoPDF/pkg/api/builders"
	apiconfig "github.com/BuddhiLW/AutoPDF/pkg/api/config"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/api/factories"
	"github.com/BuddhiLW/AutoPDF/pkg/api/middleware"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// PDFGenerationAPI provides REST endpoints for PDF generation functionality
type PDFGenerationAPI struct {
	appService *application.PDFGenerationApplicationService
	config     *config.Config
}

// NewPDFGenerationAPI creates a new PDFGenerationAPI instance
func NewPDFGenerationAPI(cfg *config.Config) *PDFGenerationAPI {
	// Create factory and application service
	// Note: In a real implementation, you'd inject a proper logger
	// Default debugEnabled to false for legacy REST API compatibility
	factory := factories.NewPDFGenerationServiceFactory(cfg, nil, false)
	appService := factory.CreateApplicationService()

	return &PDFGenerationAPI{
		appService: appService,
		config:     cfg,
	}
}

// Routes returns the chi router with all PDF generation endpoints
func (api *PDFGenerationAPI) Routes() chi.Router {
	r := chi.NewRouter()

	// Apply debug middleware
	debugConfig := &apiconfig.APIDebugConfig{
		Enabled: true,
	}
	r.Use(middleware.DebugMiddleware(debugConfig))

	// PDF generation endpoints
	r.Post("/generate", api.GeneratePDF)
	r.Post("/generate/from-struct", api.GeneratePDFFromStruct) // NEW: struct-based generation
	r.Post("/generate/async", api.GeneratePDFAsync)
	r.Get("/status/{requestId}", api.GetGenerationStatus)
	r.Get("/download/{requestId}", api.DownloadFile)
	r.Get("/download/{requestId}/{format}", api.DownloadFileFormat)

	// Template endpoints
	r.Post("/templates/validate", api.ValidateTemplate)
	r.Get("/templates/variables", api.GetTemplateVariables)

	// Health and info endpoints
	r.Get("/health", api.HealthCheck)
	r.Get("/engines", api.GetSupportedEngines)
	r.Get("/formats", api.GetSupportedFormats)

	// Watch mode management endpoints
	r.Get("/watch", api.GetActiveWatchModes)
	r.Delete("/watch/{watchId}", api.StopWatchMode)
	r.Delete("/watch", api.StopAllWatchModes)

	return r
}

// PDFGenerationRequest represents a request to generate a PDF
type PDFGenerationRequest struct {
	TemplatePath string                 `json:"template_path"`
	Variables    map[string]interface{} `json:"variables"`
	Options      *PDFGenerationOptions  `json:"options,omitempty"`
}

// PDFGenerationStructRequest represents a request to generate a PDF from a struct
type PDFGenerationStructRequest struct {
	TemplatePath string                `json:"template_path"`
	Data         interface{}           `json:"data"` // Struct to convert to variables
	Options      *PDFGenerationOptions `json:"options,omitempty"`
}

// PDFGenerationOptions represents PDF generation options
type PDFGenerationOptions struct {
	Engine       string                `json:"engine,omitempty"`        // pdflatex, xelatex, lualatex
	OutputFormat string                `json:"output_format,omitempty"` // pdf, png, jpeg, svg
	Conversion   RESTConversionOptions `json:"conversion,omitempty"`
	Debug        bool                  `json:"debug,omitempty"`
	Cleanup      bool                  `json:"cleanup,omitempty"`
	Timeout      int                   `json:"timeout,omitempty"`    // seconds
	WatchMode    bool                  `json:"watch_mode,omitempty"` // Enable file watching
}

// RESTConversionOptions represents conversion options for images in REST API
type RESTConversionOptions struct {
	DoConvert bool    `json:"do_convert,omitempty"`
	Format    string  `json:"format,omitempty"`  // png, jpeg, svg
	Quality   int     `json:"quality,omitempty"` // 1-100
	DPI       int     `json:"dpi,omitempty"`     // dots per inch
	Scale     float64 `json:"scale,omitempty"`   // scale factor
}

// PDFGenerationResponse represents the response from PDF generation
type PDFGenerationResponse struct {
	Success     bool              `json:"success"`
	RequestID   string            `json:"request_id"`
	Message     string            `json:"message,omitempty"`
	Files       []GeneratedFile   `json:"files,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	DownloadURL string            `json:"download_url,omitempty"`
	WatchMode   bool              `json:"watch_mode,omitempty"` // Indicates if watch mode is active
}

// GeneratedFile represents a generated file
type GeneratedFile struct {
	Type        string `json:"type"` // pdf, png, jpeg, svg
	Size        int64  `json:"size"`
	DownloadURL string `json:"download_url"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

// AsyncPDFGenerationResponse represents the response for async PDF generation
type AsyncPDFGenerationResponse struct {
	Success   bool   `json:"success"`
	RequestID string `json:"request_id"`
	Message   string `json:"message,omitempty"`
	StatusURL string `json:"status_url"`
	WatchMode bool   `json:"watch_mode,omitempty"` // Indicates if watch mode is active
}

// GenerationStatusResponse represents the status of an async generation
type GenerationStatusResponse struct {
	RequestID string          `json:"request_id"`
	Status    string          `json:"status"`             // pending, processing, completed, failed
	Progress  int             `json:"progress,omitempty"` // 0-100
	Message   string          `json:"message,omitempty"`
	Files     []GeneratedFile `json:"files,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// TemplateValidationRequest represents a request to validate a template
type TemplateValidationRequest struct {
	TemplatePath string `json:"template_path"`
}

// TemplateValidationResponse represents the response from template validation
type TemplateValidationResponse struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// TemplateVariablesResponse represents template variables
type TemplateVariablesResponse struct {
	Variables []string `json:"variables"`
	Count     int      `json:"count"`
}

// PDFHealthResponse represents the health check response for PDF API
type PDFHealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// EnginesResponse represents supported LaTeX engines
type EnginesResponse struct {
	Engines []string `json:"engines"`
	Default string   `json:"default"`
}

// FormatsResponse represents supported output formats
type FormatsResponse struct {
	Formats []string `json:"formats"`
	Default string   `json:"default"`
}

// GeneratePDF generates a PDF synchronously
// POST /api/v1/pdf/generate
func (api *PDFGenerationAPI) GeneratePDF(w http.ResponseWriter, r *http.Request) {
	var req PDFGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, PDFGenerationResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Validate required fields
	if req.TemplatePath == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, PDFGenerationResponse{
			Success: false,
			Message: "template_path is required",
		})
		return
	}

	// Get request ID from context (set by middleware)
	requestID := r.Context().Value(middleware.RequestIDContextKey).(string)

	// Build PDF generation request using builder pattern
	builder := builders.NewPDFGenerationRequestBuilder().
		WithTemplate(req.TemplatePath).
		WithVariables(req.Variables)

	// Apply options if provided
	if req.Options != nil {
		if req.Options.Engine != "" {
			builder = builder.WithEngine(req.Options.Engine)
		}
		if req.Options.Debug {
			builder = builder.WithDebug(generation.DebugOptions{
				Enabled:            true,
				LogToFile:          true,
				CreateConcreteFile: true,
				RequestID:          requestID,
			})
		}
		if req.Options.Timeout > 0 {
			builder = builder.WithTimeout(time.Duration(req.Options.Timeout) * time.Second)
		}
		if req.Options.Conversion.DoConvert {
			builder = builder.WithConversion(true, req.Options.Conversion.Format)
		}
		if req.Options.WatchMode {
			builder = builder.WithWatchMode(true)
		}
	}

	pdfRequest := builder.Build()

	// Generate PDF
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := api.appService.GeneratePDF(ctx, pdfRequest)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, PDFGenerationResponse{
			Success:   false,
			RequestID: requestID,
			Message:   fmt.Sprintf("PDF generation failed: %v", err),
		})
		return
	}

	// Prepare response
	response := PDFGenerationResponse{
		Success:   true,
		RequestID: requestID,
		Message:   "PDF generated successfully",
		Files: []GeneratedFile{
			{
				Type:        "pdf",
				Size:        result.Metadata.FileSize,
				DownloadURL: fmt.Sprintf("/api/v1/pdf/download/%s", requestID),
				ExpiresAt:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
		},
		Metadata: map[string]string{
			"pages":        fmt.Sprintf("%d", result.Metadata.PageCount),
			"generated_at": result.Metadata.GeneratedAt.Format(time.RFC3339),
			"engine":       result.Metadata.Engine,
		},
		WatchMode: req.Options != nil && req.Options.WatchMode,
	}

	// Add conversion files if requested
	if req.Options != nil && req.Options.Conversion.DoConvert && len(result.ImagePaths) > 0 {
		for range result.ImagePaths {
			response.Files = append(response.Files, GeneratedFile{
				Type:        req.Options.Conversion.Format,
				Size:        1024, // Placeholder size
				DownloadURL: fmt.Sprintf("/api/v1/pdf/download/%s/%s", requestID, req.Options.Conversion.Format),
				ExpiresAt:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			})
		}
	}

	render.JSON(w, r, response)
}

// GeneratePDFFromStruct generates a PDF from a struct (converts struct to variables automatically)
// POST /api/v1/pdf/generate/from-struct
func (api *PDFGenerationAPI) GeneratePDFFromStruct(w http.ResponseWriter, r *http.Request) {
	var req PDFGenerationStructRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, PDFGenerationResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Validate required fields
	if req.TemplatePath == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, PDFGenerationResponse{
			Success: false,
			Message: "template_path is required",
		})
		return
	}

	if req.Data == nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, PDFGenerationResponse{
			Success: false,
			Message: "data is required",
		})
		return
	}

	// Get request ID from context (set by middleware)
	requestID := r.Context().Value(middleware.RequestIDContextKey).(string)

	// Build PDF generation request using builder pattern with struct conversion
	builder := builders.NewPDFGenerationRequestBuilder().
		WithTemplate(req.TemplatePath).
		WithVariablesFromStruct(req.Data) // Convert struct to TemplateVariables

	// Apply options if provided
	if req.Options != nil {
		if req.Options.Engine != "" {
			builder = builder.WithEngine(req.Options.Engine)
		}
		if req.Options.Debug {
			builder = builder.WithDebug(generation.DebugOptions{
				Enabled:            true,
				LogToFile:          true,
				CreateConcreteFile: true,
				RequestID:          requestID,
			})
		}
		if req.Options.Timeout > 0 {
			builder = builder.WithTimeout(time.Duration(req.Options.Timeout) * time.Second)
		}
		if req.Options.Conversion.DoConvert {
			builder = builder.WithConversion(true, req.Options.Conversion.Format)
		}
		if req.Options.WatchMode {
			builder = builder.WithWatchMode(true)
		}
	}

	pdfRequest := builder.Build()

	// Generate PDF
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := api.appService.GeneratePDF(ctx, pdfRequest)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, PDFGenerationResponse{
			Success:   false,
			RequestID: requestID,
			Message:   fmt.Sprintf("PDF generation failed: %v", err),
		})
		return
	}

	// Prepare response
	response := PDFGenerationResponse{
		Success:   true,
		RequestID: requestID,
		Message:   "PDF generated successfully from struct",
		Files: []GeneratedFile{
			{
				Type:        "pdf",
				Size:        result.Metadata.FileSize,
				DownloadURL: fmt.Sprintf("/api/v1/pdf/download/%s", requestID),
				ExpiresAt:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
		},
		Metadata: map[string]string{
			"template":     result.Metadata.Template,
			"engine":       result.Metadata.Engine,
			"generated_at": result.Metadata.GeneratedAt.Format(time.RFC3339),
			"struct_type":  fmt.Sprintf("%T", req.Data),
		},
		WatchMode: pdfRequest.Options.WatchMode,
	}

	// Add image files if conversion was requested
	if req.Options != nil && req.Options.Conversion.DoConvert {
		for range result.ImagePaths {
			response.Files = append(response.Files, GeneratedFile{
				Type:        req.Options.Conversion.Format,
				Size:        1024, // Placeholder size
				DownloadURL: fmt.Sprintf("/api/v1/pdf/download/%s/%s", requestID, req.Options.Conversion.Format),
				ExpiresAt:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			})
		}
	}

	render.JSON(w, r, response)
}

// GeneratePDFAsync generates a PDF asynchronously
// POST /api/v1/pdf/generate/async
func (api *PDFGenerationAPI) GeneratePDFAsync(w http.ResponseWriter, r *http.Request) {
	// Similar to GeneratePDF but returns immediately with status URL
	// Implementation would use a job queue or background processing

	requestID := r.Context().Value(middleware.RequestIDContextKey).(string)

	// TODO: Implement async processing
	// For now, return a placeholder response
	response := AsyncPDFGenerationResponse{
		Success:   true,
		RequestID: requestID,
		Message:   "PDF generation started",
		StatusURL: fmt.Sprintf("/api/v1/pdf/status/%s", requestID),
		WatchMode: false, // TODO: Implement watch mode for async requests
	}

	render.JSON(w, r, response)
}

// GetGenerationStatus gets the status of an async generation
// GET /api/v1/pdf/status/{requestId}
func (api *PDFGenerationAPI) GetGenerationStatus(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "requestId")
	if requestID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, GenerationStatusResponse{
			Status: "failed",
			Error:  "request_id is required",
		})
		return
	}

	// TODO: Implement status checking from job queue
	// For now, return a placeholder response
	response := GenerationStatusResponse{
		RequestID: requestID,
		Status:    "completed",
		Progress:  100,
		Message:   "Generation completed",
		Files: []GeneratedFile{
			{
				Type:        "pdf",
				Size:        1024,
				DownloadURL: fmt.Sprintf("/api/v1/pdf/download/%s", requestID),
				ExpiresAt:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
		},
	}

	render.JSON(w, r, response)
}

// DownloadFile downloads a generated file
// GET /api/v1/pdf/download/{requestId}
func (api *PDFGenerationAPI) DownloadFile(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "requestId")
	if requestID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}

	// TODO: Retrieve file from storage/cache
	// For now, return a placeholder PDF
	pdfBytes := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n/Contents 4 0 R\n>>\nendobj\n4 0 obj\n<<\n/Length 44\n>>\nstream\nBT\n/F1 12 Tf\n72 720 Td\n(Hello World) Tj\nET\nendstream\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000204 00000 n \ntrailer\n<<\n/Size 5\n/Root 1 0 R\n>>\nstartxref\n297\n%%EOF")

	// Set headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"document_%s.pdf\"", requestID))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.Header().Set("X-Request-ID", requestID)
	w.Header().Set("Cache-Control", "private, max-age=3600")

	w.Write(pdfBytes)
}

// DownloadFileFormat downloads a generated file in specific format
// GET /api/v1/pdf/download/{requestId}/{format}
func (api *PDFGenerationAPI) DownloadFileFormat(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "requestId")
	format := chi.URLParam(r, "format")

	if requestID == "" || format == "" {
		http.Error(w, "request_id and format are required", http.StatusBadRequest)
		return
	}

	// Validate format
	validFormats := map[string]string{
		"png":  "image/png",
		"jpeg": "image/jpeg",
		"jpg":  "image/jpeg",
		"svg":  "image/svg+xml",
	}

	contentType, exists := validFormats[format]
	if !exists {
		http.Error(w, "Invalid format. Supported: png, jpeg, svg", http.StatusBadRequest)
		return
	}

	// TODO: Retrieve converted file from storage/cache
	// For now, return a placeholder image
	imageBytes := []byte("placeholder image data")

	// Set headers for image download
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"document_%s.%s\"", requestID, format))
	w.Header().Set("Content-Length", strconv.Itoa(len(imageBytes)))
	w.Header().Set("X-Request-ID", requestID)
	w.Header().Set("Cache-Control", "private, max-age=3600")

	w.Write(imageBytes)
}

// ValidateTemplate validates a LaTeX template
// POST /api/v1/pdf/templates/validate
func (api *PDFGenerationAPI) ValidateTemplate(w http.ResponseWriter, r *http.Request) {
	var req TemplateValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, TemplateValidationResponse{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Invalid request body: %v", err)},
		})
		return
	}

	if req.TemplatePath == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, TemplateValidationResponse{
			Valid:  false,
			Errors: []string{"template_path is required"},
		})
		return
	}

	// TODO: Implement template validation
	// For now, return a placeholder response
	response := TemplateValidationResponse{
		Valid:    true,
		Warnings: []string{"Template validation not fully implemented"},
	}

	render.JSON(w, r, response)
}

// GetTemplateVariables extracts variables from a template
// GET /api/v1/pdf/templates/variables?template_path=...
func (api *PDFGenerationAPI) GetTemplateVariables(w http.ResponseWriter, r *http.Request) {
	templatePath := r.URL.Query().Get("template_path")
	if templatePath == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, TemplateVariablesResponse{
			Variables: []string{},
			Count:     0,
		})
		return
	}

	// TODO: Implement template variable extraction
	// For now, return a placeholder response
	response := TemplateVariablesResponse{
		Variables: []string{"title", "author", "date", "content"},
		Count:     4,
	}

	render.JSON(w, r, response)
}

// HealthCheck provides a health check endpoint
// GET /api/v1/pdf/health
func (api *PDFGenerationAPI) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := PDFHealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
	}

	render.JSON(w, r, response)
}

// GetSupportedEngines returns supported LaTeX engines
// GET /api/v1/pdf/engines
func (api *PDFGenerationAPI) GetSupportedEngines(w http.ResponseWriter, r *http.Request) {
	response := EnginesResponse{
		Engines: []string{"pdflatex", "xelatex", "lualatex"},
		Default: "pdflatex",
	}

	render.JSON(w, r, response)
}

// GetSupportedFormats returns supported output formats
// GET /api/v1/pdf/formats
func (api *PDFGenerationAPI) GetSupportedFormats(w http.ResponseWriter, r *http.Request) {
	response := FormatsResponse{
		Formats: []string{"pdf", "png", "jpeg", "svg"},
		Default: "pdf",
	}

	render.JSON(w, r, response)
}

// GetActiveWatchModes returns information about active watch modes
// GET /api/v1/pdf/watch
func (api *PDFGenerationAPI) GetActiveWatchModes(w http.ResponseWriter, r *http.Request) {
	activeWatches := api.appService.GetActiveWatchModes()

	response := map[string]interface{}{
		"active_watches": activeWatches,
		"count":          len(activeWatches),
	}

	render.JSON(w, r, response)
}

// StopWatchMode stops a specific watch mode
// DELETE /api/v1/pdf/watch/{watchId}
func (api *PDFGenerationAPI) StopWatchMode(w http.ResponseWriter, r *http.Request) {
	watchID := chi.URLParam(r, "watchId")
	if watchID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error": "watch_id is required",
		})
		return
	}

	err := api.appService.StopWatchMode(watchID)
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{
			"error": fmt.Sprintf("Failed to stop watch mode: %v", err),
		})
		return
	}

	render.JSON(w, r, map[string]string{
		"message": fmt.Sprintf("Watch mode %s stopped successfully", watchID),
	})
}

// StopAllWatchModes stops all active watch modes
// DELETE /api/v1/pdf/watch
func (api *PDFGenerationAPI) StopAllWatchModes(w http.ResponseWriter, r *http.Request) {
	err := api.appService.StopAllWatchModes()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{
			"error": fmt.Sprintf("Failed to stop all watch modes: %v", err),
		})
		return
	}

	render.JSON(w, r, map[string]string{
		"message": "All watch modes stopped successfully",
	})
}

// RegisterPDFGenerationRoutes registers all PDF generation routes with the main router
func RegisterPDFGenerationRoutes(r chi.Router, api *PDFGenerationAPI) {
	r.Route("/api/v1/pdf", func(r chi.Router) {
		r.Mount("/", api.Routes())
	})
}

// RegisterPDFGenerationRoutesWithDefaults registers PDF generation routes with default configuration
func RegisterPDFGenerationRoutesWithDefaults(r chi.Router) {
	cfg := &config.Config{} // Default config
	api := NewPDFGenerationAPI(cfg)
	RegisterPDFGenerationRoutes(r, api)
}
