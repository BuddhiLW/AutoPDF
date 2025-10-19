// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/converter"
)

// SimpleStructConverterAPI provides REST endpoints for struct conversion functionality using http.ServeMux
type SimpleStructConverterAPI struct {
	converter *converter.StructConverter
}

// NewSimpleStructConverterAPI creates a new SimpleStructConverterAPI instance
func NewSimpleStructConverterAPI() *SimpleStructConverterAPI {
	return &SimpleStructConverterAPI{
		converter: converter.BuildWithDefaults(),
	}
}

// NewSimpleStructConverterAPIWithConverter creates a new SimpleStructConverterAPI with a custom converter
func NewSimpleStructConverterAPIWithConverter(conv *converter.StructConverter) *SimpleStructConverterAPI {
	return &SimpleStructConverterAPI{
		converter: conv,
	}
}

// SimpleConvertStructRequest represents a request to convert a struct
type SimpleConvertStructRequest struct {
	Data     interface{}              `json:"data"`
	Options  *SimpleConversionOptions `json:"options,omitempty"`
	Metadata map[string]interface{}   `json:"metadata,omitempty"`
}

// SimpleConversionOptions represents conversion options
type SimpleConversionOptions struct {
	TagName        string `json:"tag_name,omitempty"`
	DefaultFlatten bool   `json:"default_flatten,omitempty"`
	OmitEmpty      bool   `json:"omit_empty,omitempty"`
	TimeFormat     string `json:"time_format,omitempty"`
	DurationFormat string `json:"duration_format,omitempty"`
	SliceSeparator string `json:"slice_separator,omitempty"`
}

// SimpleConvertStructResponse represents the response from struct conversion
type SimpleConvertStructResponse struct {
	Variables map[string]interface{} `json:"variables"`
	Metadata  map[string]interface{} `json:"metadata"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
}

// SimpleConverterConfigResponse represents the current converter configuration
type SimpleConverterConfigResponse struct {
	TagName         string   `json:"tag_name"`
	DefaultFlatten  bool     `json:"default_flatten"`
	OmitEmpty       bool     `json:"omit_empty"`
	TimeFormat      string   `json:"time_format"`
	DurationFormat  string   `json:"duration_format"`
	SliceSeparator  string   `json:"slice_separator"`
	RegisteredTypes []string `json:"registered_types"`
	ConverterCount  int      `json:"converter_count"`
}

// SimpleValidationRequest represents a request to validate a struct
type SimpleValidationRequest struct {
	Data interface{} `json:"data"`
}

// SimpleValidationResponse represents the response from struct validation
type SimpleValidationResponse struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// SimplePreviewRequest represents a request to preview conversion
type SimplePreviewRequest struct {
	Data    interface{}              `json:"data"`
	Options *SimpleConversionOptions `json:"options,omitempty"`
	Limit   int                      `json:"limit,omitempty"`
}

// SimplePreviewResponse represents the response from preview conversion
type SimplePreviewResponse struct {
	Variables map[string]interface{} `json:"variables"`
	Count     int                    `json:"count"`
	Truncated bool                   `json:"truncated"`
	Message   string                 `json:"message,omitempty"`
}

// SimpleHealthResponse represents the health check response
type SimpleHealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// ConvertStruct converts a struct to AutoPDF variables
// POST /api/v1/struct-converter/convert
func (api *SimpleStructConverterAPI) ConvertStruct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SimpleConvertStructRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		http.Error(w, "Data field is required", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Conversion failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert Variables to map[string]interface{}
	result := make(map[string]interface{})
	variables.Range(func(name string, value config.Variable) bool {
		result[name] = value.String()
		return true
	})

	// Prepare response
	response := SimpleConvertStructResponse{
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConvertStructFlattened converts a struct to flattened AutoPDF variables
// POST /api/v1/struct-converter/convert/flattened
func (api *SimpleStructConverterAPI) ConvertStructFlattened(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SimpleConvertStructRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		http.Error(w, "Data field is required", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Conversion failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Get flattened representation
	flattened := variables.Flatten()

	// Convert map[string]string to map[string]interface{}
	result := make(map[string]interface{})
	for k, v := range flattened {
		result[k] = v
	}

	// Prepare response
	response := SimpleConvertStructResponse{
		Variables: result,
		Metadata:  req.Metadata,
		Success:   true,
		Message:   "Flattened conversion completed successfully",
	}

	// Add metadata
	if response.Metadata == nil {
		response.Metadata = make(map[string]interface{})
	}
	response.Metadata["converted_at"] = time.Now().Format(time.RFC3339)
	response.Metadata["variable_count"] = len(result)
	response.Metadata["flattened"] = true

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConvertStructForTemplate converts a struct optimized for template usage
// POST /api/v1/struct-converter/convert/template
func (api *SimpleStructConverterAPI) ConvertStructForTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SimpleConvertStructRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		http.Error(w, "Data field is required", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Conversion failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Get flattened representation for templates
	flattened := variables.Flatten()

	// Convert map[string]string to map[string]interface{}
	result := make(map[string]interface{})
	for k, v := range flattened {
		result[k] = v
	}

	// Prepare response
	response := SimpleConvertStructResponse{
		Variables: result,
		Metadata:  req.Metadata,
		Success:   true,
		Message:   "Template conversion completed successfully",
	}

	// Add metadata
	if response.Metadata == nil {
		response.Metadata = make(map[string]interface{})
	}
	response.Metadata["converted_at"] = time.Now().Format(time.RFC3339)
	response.Metadata["variable_count"] = len(result)
	response.Metadata["optimized_for"] = "templates"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetConverterConfig returns the current converter configuration
// GET /api/v1/struct-converter/config
func (api *SimpleStructConverterAPI) GetConverterConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Since we can't access unexported fields, return default configuration
	config := SimpleConverterConfigResponse{
		TagName:         "autopdf",                                                 // Default tag name
		DefaultFlatten:  false,                                                     // Default flatten setting
		OmitEmpty:       false,                                                     // Default omit empty setting
		TimeFormat:      "RFC3339",                                                 // Default format
		DurationFormat:  "string",                                                  // Default format
		SliceSeparator:  ", ",                                                      // Default separator
		RegisteredTypes: []string{"string", "int", "float64", "bool", "time.Time"}, // Common types
		ConverterCount:  5,                                                         // Estimated count
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// ValidateStruct validates a struct for conversion
// POST /api/v1/struct-converter/validate
func (api *SimpleStructConverterAPI) ValidateStruct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SimpleValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		http.Error(w, "Data field is required", http.StatusBadRequest)
		return
	}

	// Basic validation
	var errors []string
	var warnings []string

	// Check if it's a struct
	if req.Data == nil {
		errors = append(errors, "Data must not be nil")
	}

	// This is a simplified check - in practice, you'd implement proper validation
	warnings = append(warnings, "Advanced validation not implemented")

	response := SimpleValidationResponse{
		Valid:    len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PreviewConversion provides a preview of struct conversion with limited output
// POST /api/v1/struct-converter/preview
func (api *SimpleStructConverterAPI) PreviewConversion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SimplePreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		http.Error(w, "Data field is required", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Conversion failed: %v", err), http.StatusInternalServerError)
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

	response := SimplePreviewResponse{
		Variables: limited,
		Count:     count,
		Truncated: truncated,
		Message:   "Preview generated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheck provides a health check endpoint
// GET /api/v1/struct-converter/health
func (api *SimpleStructConverterAPI) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := SimpleHealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// buildConverterWithOptions builds a converter with the specified options
func (api *SimpleStructConverterAPI) buildConverterWithOptions(options *SimpleConversionOptions) *converter.StructConverter {
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

// RegisterSimpleStructConverterRoutes registers all simple struct converter routes with the main router
func RegisterSimpleStructConverterRoutes(mux *http.ServeMux, api *SimpleStructConverterAPI) {
	mux.HandleFunc("/api/v1/struct-converter/convert", api.ConvertStruct)
	mux.HandleFunc("/api/v1/struct-converter/convert/flattened", api.ConvertStructFlattened)
	mux.HandleFunc("/api/v1/struct-converter/convert/template", api.ConvertStructForTemplate)
	mux.HandleFunc("/api/v1/struct-converter/config", api.GetConverterConfig)
	mux.HandleFunc("/api/v1/struct-converter/validate", api.ValidateStruct)
	mux.HandleFunc("/api/v1/struct-converter/preview", api.PreviewConversion)
	mux.HandleFunc("/api/v1/struct-converter/health", api.HealthCheck)
}

// RegisterSimpleStructConverterRoutesWithDefaults registers simple struct converter routes with default configuration
func RegisterSimpleStructConverterRoutesWithDefaults(mux *http.ServeMux) {
	api := NewSimpleStructConverterAPI()
	RegisterSimpleStructConverterRoutes(mux, api)
}
