// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/converter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// StructConverterAPI provides REST endpoints for struct conversion functionality
type StructConverterAPI struct {
	converter *converter.StructConverter
}

// NewStructConverterAPI creates a new StructConverterAPI instance
func NewStructConverterAPI() *StructConverterAPI {
	return &StructConverterAPI{
		converter: converter.BuildWithDefaults(),
	}
}

// NewStructConverterAPIWithConverter creates a new StructConverterAPI with a custom converter
func NewStructConverterAPIWithConverter(conv *converter.StructConverter) *StructConverterAPI {
	return &StructConverterAPI{
		converter: conv,
	}
}

// Routes returns the chi router with all struct converter endpoints
func (api *StructConverterAPI) Routes() chi.Router {
	r := chi.NewRouter()

	// Struct conversion endpoints
	r.Post("/convert", api.ConvertStruct)
	r.Post("/convert/flattened", api.ConvertStructFlattened)
	r.Post("/convert/template", api.ConvertStructForTemplate)

	// Converter configuration endpoints
	r.Get("/config", api.GetConverterConfig)
	r.Put("/config", api.UpdateConverterConfig)

	// Converter registry endpoints
	r.Get("/converters", api.ListConverters)
	r.Post("/converters", api.RegisterConverter)
	r.Delete("/converters/{type}", api.UnregisterConverter)

	// Utility endpoints
	r.Post("/validate", api.ValidateStruct)
	r.Post("/preview", api.PreviewConversion)
	r.Get("/health", api.HealthCheck)

	return r
}

// ConvertStructRequest represents a request to convert a struct
type ConvertStructRequest struct {
	Data     interface{}            `json:"data"`
	Options  *ConversionOptions     `json:"options,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ConversionOptions represents conversion options
type ConversionOptions struct {
	TagName        string `json:"tag_name,omitempty"`
	DefaultFlatten bool   `json:"default_flatten,omitempty"`
	OmitEmpty      bool   `json:"omit_empty,omitempty"`
	TimeFormat     string `json:"time_format,omitempty"`
	DurationFormat string `json:"duration_format,omitempty"`
	SliceSeparator string `json:"slice_separator,omitempty"`
}

// ConvertStructResponse represents the response from struct conversion
type ConvertStructResponse struct {
	Variables map[string]interface{} `json:"variables"`
	Metadata  map[string]interface{} `json:"metadata"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
}

// ConverterConfigResponse represents the current converter configuration
type ConverterConfigResponse struct {
	TagName         string   `json:"tag_name"`
	DefaultFlatten  bool     `json:"default_flatten"`
	OmitEmpty       bool     `json:"omit_empty"`
	TimeFormat      string   `json:"time_format"`
	DurationFormat  string   `json:"duration_format"`
	SliceSeparator  string   `json:"slice_separator"`
	RegisteredTypes []string `json:"registered_types"`
	ConverterCount  int      `json:"converter_count"`
}

// ConverterInfo represents information about a registered converter
type ConverterInfo struct {
	Type        string `json:"type"`
	CanConvert  bool   `json:"can_convert"`
	Description string `json:"description,omitempty"`
}

// ConvertersListResponse represents the list of registered converters
type ConvertersListResponse struct {
	Converters []ConverterInfo `json:"converters"`
	Count      int             `json:"count"`
}

// ValidationRequest represents a request to validate a struct
type ValidationRequest struct {
	Data interface{} `json:"data"`
}

// ValidationResponse represents the response from struct validation
type ValidationResponse struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// PreviewRequest represents a request to preview conversion
type PreviewRequest struct {
	Data    interface{}        `json:"data"`
	Options *ConversionOptions `json:"options,omitempty"`
	Limit   int                `json:"limit,omitempty"`
}

// PreviewResponse represents the response from preview conversion
type PreviewResponse struct {
	Variables map[string]interface{} `json:"variables"`
	Count     int                    `json:"count"`
	Truncated bool                   `json:"truncated"`
	Message   string                 `json:"message,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// ConvertStruct converts a struct to AutoPDF variables
// POST /api/v1/struct-converter/convert
func (api *StructConverterAPI) ConvertStruct(w http.ResponseWriter, r *http.Request) {
	var req ConvertStructRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ConvertStructResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.Data == nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ConvertStructResponse{
			Success: false,
			Message: "Data field is required",
		})
		return
	}

	// Apply options if provided
	converter := api.converter
	if req.Options != nil {
		converter = api.buildConverterWithOptions(req.Options)
	}

	// Convert the struct
	variables, err := converter.ConvertStruct(req.Data)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ConvertStructResponse{
			Success: false,
			Message: fmt.Sprintf("Conversion failed: %v", err),
		})
		return
	}

	// Convert Variables to map[string]interface{}
	result := make(map[string]interface{})
	variables.Range(func(name string, value config.Variable) bool {
		result[name] = value.String()
		return true
	})

	// Prepare response
	response := ConvertStructResponse{
		Variables: result,
		Metadata:  req.Metadata,
		Success:   true,
		Message:   "Conversion completed successfully",
	}

	// Add metadata
	if response.Metadata == nil {
		response.Metadata = make(map[string]interface{})
	}
	response.Metadata["converted_at"] = time.Now().Format(time.RFC3339)
	response.Metadata["variable_count"] = len(result)

	render.JSON(w, r, response)
}

// ConvertStructFlattened converts a struct to flattened AutoPDF variables
// POST /api/v1/struct-converter/convert/flattened
func (api *StructConverterAPI) ConvertStructFlattened(w http.ResponseWriter, r *http.Request) {
	var req ConvertStructRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ConvertStructResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.Data == nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ConvertStructResponse{
			Success: false,
			Message: "Data field is required",
		})
		return
	}

	// Use flattened converter
	converter := converter.BuildForFlattened()
	if req.Options != nil {
		converter = api.buildConverterWithOptions(req.Options)
	}

	// Convert the struct
	variables, err := converter.ConvertStruct(req.Data)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ConvertStructResponse{
			Success: false,
			Message: fmt.Sprintf("Conversion failed: %v", err),
		})
		return
	}

	// Get flattened representation
	flattened := variables.Flatten()

	// Convert map[string]string to map[string]interface{}
	variablesMap := make(map[string]interface{})
	for k, v := range flattened {
		variablesMap[k] = v
	}

	// Prepare response
	response := ConvertStructResponse{
		Variables: variablesMap,
		Metadata:  req.Metadata,
		Success:   true,
		Message:   "Flattened conversion completed successfully",
	}

	// Add metadata
	if response.Metadata == nil {
		response.Metadata = make(map[string]interface{})
	}
	response.Metadata["converted_at"] = time.Now().Format(time.RFC3339)
	response.Metadata["variable_count"] = len(flattened)
	response.Metadata["flattened"] = true

	render.JSON(w, r, response)
}

// ConvertStructForTemplate converts a struct optimized for template usage
// POST /api/v1/struct-converter/convert/template
func (api *StructConverterAPI) ConvertStructForTemplate(w http.ResponseWriter, r *http.Request) {
	var req ConvertStructRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ConvertStructResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.Data == nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ConvertStructResponse{
			Success: false,
			Message: "Data field is required",
		})
		return
	}

	// Use template-optimized converter
	converter := converter.BuildForTemplates()
	if req.Options != nil {
		converter = api.buildConverterWithOptions(req.Options)
	}

	// Convert the struct
	variables, err := converter.ConvertStruct(req.Data)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, ConvertStructResponse{
			Success: false,
			Message: fmt.Sprintf("Conversion failed: %v", err),
		})
		return
	}

	// Get flattened representation for templates
	flattened := variables.Flatten()

	// Convert map[string]string to map[string]interface{}
	variablesMap := make(map[string]interface{})
	for k, v := range flattened {
		variablesMap[k] = v
	}

	// Prepare response
	response := ConvertStructResponse{
		Variables: variablesMap,
		Metadata:  req.Metadata,
		Success:   true,
		Message:   "Template conversion completed successfully",
	}

	// Add metadata
	if response.Metadata == nil {
		response.Metadata = make(map[string]interface{})
	}
	response.Metadata["converted_at"] = time.Now().Format(time.RFC3339)
	response.Metadata["variable_count"] = len(flattened)
	response.Metadata["optimized_for"] = "templates"

	render.JSON(w, r, response)
}

// GetConverterConfig returns the current converter configuration
// GET /api/v1/struct-converter/config
func (api *StructConverterAPI) GetConverterConfig(w http.ResponseWriter, r *http.Request) {
	// Since we can't access unexported fields, return default configuration
	config := ConverterConfigResponse{
		TagName:         "autopdf",  // Default tag name
		DefaultFlatten:  false,      // Default flatten behavior
		OmitEmpty:       false,      // Default omit empty behavior
		TimeFormat:      "RFC3339",  // Default format
		DurationFormat:  "string",   // Default format
		SliceSeparator:  ", ",       // Default separator
		RegisteredTypes: []string{}, // Empty for now
		ConverterCount:  0,          // Empty for now
	}

	render.JSON(w, r, config)
}

// UpdateConverterConfig updates the converter configuration
// PUT /api/v1/struct-converter/config
func (api *StructConverterAPI) UpdateConverterConfig(w http.ResponseWriter, r *http.Request) {
	var req ConversionOptions
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Build new converter with updated options
	api.converter = api.buildConverterWithOptions(&req)

	render.JSON(w, r, map[string]string{
		"message": "Configuration updated successfully",
	})
}

// ListConverters returns the list of registered converters
// GET /api/v1/struct-converter/converters
func (api *StructConverterAPI) ListConverters(w http.ResponseWriter, r *http.Request) {
	// Since we can't access unexported registry, return empty list for now
	converters := make([]ConverterInfo, 0)

	response := ConvertersListResponse{
		Converters: converters,
		Count:      len(converters),
	}

	render.JSON(w, r, response)
}

// RegisterConverter registers a new converter (placeholder - would need custom converter types)
// POST /api/v1/struct-converter/converters
func (api *StructConverterAPI) RegisterConverter(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusNotImplemented)
	render.JSON(w, r, map[string]string{
		"error": "Custom converter registration not implemented in REST API",
	})
}

// UnregisterConverter unregisters a converter
// DELETE /api/v1/struct-converter/converters/{type}
func (api *StructConverterAPI) UnregisterConverter(w http.ResponseWriter, r *http.Request) {
	typeName := chi.URLParam(r, "type")
	if typeName == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{
			"error": "Type parameter is required",
		})
		return
	}

	// This is a simplified implementation - in practice, you'd need to parse the type
	render.Status(r, http.StatusNotImplemented)
	render.JSON(w, r, map[string]string{
		"error": "Converter unregistration not implemented in REST API",
	})
}

// ValidateStruct validates a struct for conversion
// POST /api/v1/struct-converter/validate
func (api *StructConverterAPI) ValidateStruct(w http.ResponseWriter, r *http.Request) {
	var req ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ValidationResponse{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Invalid request body: %v", err)},
		})
		return
	}

	if req.Data == nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ValidationResponse{
			Valid:  false,
			Errors: []string{"Data field is required"},
		})
		return
	}

	// Basic validation
	var errors []string
	var warnings []string

	// Check if it's a struct
	v := reflect.ValueOf(req.Data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		errors = append(errors, "Data must be a struct or pointer to struct")
	}

	// Check for circular references (basic check)
	if len(errors) == 0 {
		// This is a simplified check - in practice, you'd implement proper circular reference detection
		warnings = append(warnings, "Circular reference detection not implemented")
	}

	response := ValidationResponse{
		Valid:    len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}

	render.JSON(w, r, response)
}

// PreviewConversion provides a preview of struct conversion with limited output
// POST /api/v1/struct-converter/preview
func (api *StructConverterAPI) PreviewConversion(w http.ResponseWriter, r *http.Request) {
	var req PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, PreviewResponse{
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.Data == nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, PreviewResponse{
			Message: "Data field is required",
		})
		return
	}

	// Set default limit
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// Apply options if provided
	converter := api.converter
	if req.Options != nil {
		converter = api.buildConverterWithOptions(req.Options)
	}

	// Convert the struct
	variables, err := converter.ConvertStruct(req.Data)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, PreviewResponse{
			Message: fmt.Sprintf("Conversion failed: %v", err),
		})
		return
	}

	// Get flattened representation
	flattened := variables.Flatten()

	// Limit the output
	limited := make(map[string]interface{})
	count := 0
	truncated := false

	for key, value := range flattened {
		if count >= limit {
			truncated = true
			break
		}
		limited[key] = value
		count++
	}

	response := PreviewResponse{
		Variables: limited,
		Count:     count,
		Truncated: truncated,
		Message:   "Preview generated successfully",
	}

	render.JSON(w, r, response)
}

// HealthCheck provides a health check endpoint
// GET /api/v1/struct-converter/health
func (api *StructConverterAPI) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
	}

	render.JSON(w, r, response)
}

// buildConverterWithOptions builds a converter with the specified options
func (api *StructConverterAPI) buildConverterWithOptions(options *ConversionOptions) *converter.StructConverter {
	builder := converter.NewConverterBuilder()

	if options.TagName != "" {
		builder = builder.WithTagName(options.TagName)
	}

	if options.DefaultFlatten {
		builder = builder.WithDefaultFlatten(true)
	}

	if options.OmitEmpty {
		builder = builder.WithOmitEmpty(true)
	}

	if options.TimeFormat != "" {
		builder = builder.WithTimeFormat(options.TimeFormat)
	}

	if options.DurationFormat != "" {
		builder = builder.WithDurationFormat(options.DurationFormat)
	}

	if options.SliceSeparator != "" {
		builder = builder.WithSliceSeparator(options.SliceSeparator)
	}

	return builder.WithBuiltinConverters().Build()
}

// RegisterStructConverterRoutes registers all struct converter routes with the main router
func RegisterStructConverterRoutes(r chi.Router, api *StructConverterAPI) {
	r.Route("/api/v1/struct-converter", func(r chi.Router) {
		r.Mount("/", api.Routes())
	})
}

// RegisterStructConverterRoutesWithDefaults registers struct converter routes with default configuration
func RegisterStructConverterRoutesWithDefaults(r chi.Router) {
	api := NewStructConverterAPI()
	RegisterStructConverterRoutes(r, api)
}
