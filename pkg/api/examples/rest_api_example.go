// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// REST API Example demonstrating struct conversion via HTTP endpoints

// Example structs for REST API demonstration
type RESTUser struct {
	ID        int       `autopdf:"id"`
	Name      string    `autopdf:"name"`
	Email     string    `autopdf:"email"`
	CreatedAt time.Time `autopdf:"created_at"`
	Profile   struct {
		Bio     string `autopdf:"bio"`
		Website string `autopdf:"website"`
		Avatar  string `autopdf:"avatar"`
	} `autopdf:"profile"`
	Settings struct {
		Theme    string `autopdf:"theme"`
		Language string `autopdf:"language"`
		Privacy  string `autopdf:"privacy"`
	} `autopdf:"settings"`
	Tags []string `autopdf:"tags,flatten"`
}

type RESTDocument struct {
	Title       string     `autopdf:"title"`
	Author      RESTUser   `autopdf:"author"`
	Content     string     `autopdf:"content"`
	CreatedAt   time.Time  `autopdf:"created_at"`
	UpdatedAt   *time.Time `autopdf:"updated_at"`
	URL         url.URL    `autopdf:"url"`
	Tags        []string   `autopdf:"tags,flatten"`
	IsPublished bool       `autopdf:"is_published"`
	Version     int        `autopdf:"version"`
}

// RESTAPIClient provides a client for the struct converter REST API
type RESTAPIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewRESTAPIClient creates a new REST API client
func NewRESTAPIClient(baseURL string) *RESTAPIClient {
	return &RESTAPIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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

// ConvertStruct converts a struct using the REST API
func (client *RESTAPIClient) ConvertStruct(data interface{}, options *ConversionOptions) (*ConvertStructResponse, error) {
	req := ConvertStructRequest{
		Data:    data,
		Options: options,
		Metadata: map[string]interface{}{
			"client":  "rest-api-example",
			"version": "1.0.0",
		},
	}

	return client.makeRequest("POST", "/convert", req, &ConvertStructResponse{})
}

// ConvertStructFlattened converts a struct to flattened variables
func (client *RESTAPIClient) ConvertStructFlattened(data interface{}, options *ConversionOptions) (*ConvertStructResponse, error) {
	req := ConvertStructRequest{
		Data:    data,
		Options: options,
		Metadata: map[string]interface{}{
			"client":  "rest-api-example",
			"version": "1.0.0",
		},
	}

	return client.makeRequest("POST", "/convert/flattened", req, &ConvertStructResponse{})
}

// ConvertStructForTemplate converts a struct optimized for templates
func (client *RESTAPIClient) ConvertStructForTemplate(data interface{}, options *ConversionOptions) (*ConvertStructResponse, error) {
	req := ConvertStructRequest{
		Data:    data,
		Options: options,
		Metadata: map[string]interface{}{
			"client":  "rest-api-example",
			"version": "1.0.0",
		},
	}

	return client.makeRequest("POST", "/convert/template", req, &ConvertStructResponse{})
}

// GetConverterConfig retrieves the current converter configuration
func (client *RESTAPIClient) GetConverterConfig() (*ConverterConfigResponse, error) {
	return client.makeRequest("GET", "/config", nil, &ConverterConfigResponse{})
}

// ValidateStruct validates a struct for conversion
func (client *RESTAPIClient) ValidateStruct(data interface{}) (*ValidationResponse, error) {
	req := ValidationRequest{Data: data}
	return client.makeRequest("POST", "/validate", req, &ValidationResponse{})
}

// PreviewConversion provides a preview of struct conversion
func (client *RESTAPIClient) PreviewConversion(data interface{}, options *ConversionOptions, limit int) (*PreviewResponse, error) {
	req := PreviewRequest{
		Data:    data,
		Options: options,
		Limit:   limit,
	}

	return client.makeRequest("POST", "/preview", req, &PreviewResponse{})
}

// HealthCheck checks the health of the API
func (client *RESTAPIClient) HealthCheck() (*HealthResponse, error) {
	return client.makeRequest("GET", "/health", nil, &HealthResponse{})
}

// makeRequest makes an HTTP request to the API
func (client *RESTAPIClient) makeRequest(method, endpoint string, body interface{}, result interface{}) error {
	url := client.BaseURL + "/api/v1/struct-converter" + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// ExampleRESTAPIUsage demonstrates how to use the REST API
func ExampleRESTAPIUsage() {
	fmt.Println("=== AutoPDF Struct Converter REST API Example ===")

	// Create API client
	client := NewRESTAPIClient("http://localhost:8080")

	// 1. Health Check
	fmt.Println("\n1. Health Check")
	health, err := client.HealthCheck()
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		return
	}
	fmt.Printf("Status: %s, Version: %s, Timestamp: %s\n", health.Status, health.Version, health.Timestamp)

	// 2. Get Converter Configuration
	fmt.Println("\n2. Converter Configuration")
	config, err := client.GetConverterConfig()
	if err != nil {
		fmt.Printf("Failed to get config: %v\n", err)
		return
	}
	fmt.Printf("Tag Name: %s, Omit Empty: %t, Converter Count: %d\n",
		config.TagName, config.OmitEmpty, config.ConverterCount)

	// 3. Create test data
	fmt.Println("\n3. Creating Test Data")
	docURL, _ := url.Parse("https://example.com/document/123")
	now := time.Now()

	user := RESTUser{
		ID:        123,
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		CreatedAt: time.Date(2025, 1, 7, 12, 0, 0, 0, time.UTC),
		Profile: struct {
			Bio     string `autopdf:"bio"`
			Website string `autopdf:"website"`
			Avatar  string `autopdf:"avatar"`
		}{
			Bio:     "Software engineer with 10+ years experience",
			Website: "https://johndoe.dev",
			Avatar:  "https://example.com/avatar.jpg",
		},
		Settings: struct {
			Theme    string `autopdf:"theme"`
			Language string `autopdf:"language"`
			Privacy  string `autopdf:"privacy"`
		}{
			Theme:    "dark",
			Language: "en",
			Privacy:  "public",
		},
		Tags: []string{"admin", "user", "premium", "verified"},
	}

	document := RESTDocument{
		Title:       "AutoPDF REST API Integration",
		Author:      user,
		Content:     "This document demonstrates AutoPDF integration via REST API.",
		CreatedAt:   now,
		UpdatedAt:   &now,
		URL:         *docURL,
		Tags:        []string{"integration", "rest", "api", "autopdf"},
		IsPublished: true,
		Version:     1,
	}

	// 4. Validate Struct
	fmt.Println("\n4. Validating Struct")
	validation, err := client.ValidateStruct(document)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}
	fmt.Printf("Valid: %t", validation.Valid)
	if len(validation.Errors) > 0 {
		fmt.Printf(", Errors: %v", validation.Errors)
	}
	if len(validation.Warnings) > 0 {
		fmt.Printf(", Warnings: %v", validation.Warnings)
	}
	fmt.Println()

	// 5. Preview Conversion
	fmt.Println("\n5. Preview Conversion (first 5 variables)")
	preview, err := client.PreviewConversion(document, nil, 5)
	if err != nil {
		fmt.Printf("Preview failed: %v\n", err)
		return
	}
	fmt.Printf("Preview (showing %d of %d variables):\n", preview.Count, len(preview.Variables))
	for key, value := range preview.Variables {
		fmt.Printf("  %s: %v\n", key, value)
	}
	if preview.Truncated {
		fmt.Println("  ... (truncated)")
	}

	// 6. Convert Struct (Default)
	fmt.Println("\n6. Convert Struct (Default)")
	response, err := client.ConvertStruct(document, nil)
	if err != nil {
		fmt.Printf("Conversion failed: %v\n", err)
		return
	}
	fmt.Printf("Success: %t, Message: %s\n", response.Success, response.Message)
	fmt.Printf("Variables generated: %d\n", len(response.Variables))
	fmt.Printf("Converted at: %v\n", response.Metadata["converted_at"])

	// 7. Convert Struct (Flattened)
	fmt.Println("\n7. Convert Struct (Flattened)")
	flattenedResponse, err := client.ConvertStructFlattened(document, nil)
	if err != nil {
		fmt.Printf("Flattened conversion failed: %v\n", err)
		return
	}
	fmt.Printf("Success: %t, Message: %s\n", flattenedResponse.Success, flattenedResponse.Message)
	fmt.Printf("Flattened variables generated: %d\n", len(flattenedResponse.Variables))

	// Show some flattened variables
	fmt.Println("Sample flattened variables:")
	count := 0
	for key, value := range flattenedResponse.Variables {
		if count >= 5 {
			break
		}
		fmt.Printf("  %s: %v\n", key, value)
		count++
	}

	// 8. Convert Struct (Template Optimized)
	fmt.Println("\n8. Convert Struct (Template Optimized)")
	templateResponse, err := client.ConvertStructForTemplate(document, nil)
	if err != nil {
		fmt.Printf("Template conversion failed: %v\n", err)
		return
	}
	fmt.Printf("Success: %t, Message: %s\n", templateResponse.Success, templateResponse.Message)
	fmt.Printf("Template variables generated: %d\n", len(templateResponse.Variables))
	fmt.Printf("Optimized for: %v\n", templateResponse.Metadata["optimized_for"])

	// 9. Convert with Custom Options
	fmt.Println("\n9. Convert with Custom Options")
	customOptions := &ConversionOptions{
		TagName:        "autopdf",
		DefaultFlatten: false,
		OmitEmpty:      true,
		TimeFormat:     "2006-01-02 15:04:05",
		SliceSeparator: " | ",
	}

	customResponse, err := client.ConvertStruct(document, customOptions)
	if err != nil {
		fmt.Printf("Custom conversion failed: %v\n", err)
		return
	}
	fmt.Printf("Success: %t, Message: %s\n", customResponse.Success, customResponse.Message)
	fmt.Printf("Custom variables generated: %d\n", len(customResponse.Variables))

	// Show some custom formatted variables
	fmt.Println("Sample custom formatted variables:")
	count = 0
	for key, value := range customResponse.Variables {
		if count >= 5 {
			break
		}
		fmt.Printf("  %s: %v\n", key, value)
		count++
	}

	// 10. Demonstrate Error Handling
	fmt.Println("\n10. Error Handling Example")
	invalidData := "This is not a struct"
	_, err = client.ConvertStruct(invalidData, nil)
	if err != nil {
		fmt.Printf("Expected error for invalid data: %v\n", err)
	}

	fmt.Println("\n=== REST API Example Complete ===")
}

// ExampleRESTAPIWithPDFGeneration demonstrates using the REST API with PDF generation
func ExampleRESTAPIWithPDFGeneration() {
	fmt.Println("=== AutoPDF REST API with PDF Generation Example ===")

	client := NewRESTAPIClient("http://localhost:8080")

	// Create document data
	docURL, _ := url.Parse("https://example.com/report/2025-001")
	now := time.Now()

	reportData := struct {
		ReportID    string    `autopdf:"report_id"`
		Title       string    `autopdf:"title"`
		GeneratedAt time.Time `autopdf:"generated_at"`
		Author      struct {
			Name  string `autopdf:"name"`
			Email string `autopdf:"email"`
		} `autopdf:"author"`
		Summary struct {
			TotalDocuments int     `autopdf:"total_documents"`
			TotalUsers     int     `autopdf:"total_users"`
			AverageScore   float64 `autopdf:"average_score"`
		} `autopdf:"summary"`
		Tags []string `autopdf:"tags,flatten"`
	}{
		ReportID:    "RPT-2025-001",
		Title:       "Monthly Performance Report",
		GeneratedAt: now,
		Author: struct {
			Name  string `autopdf:"name"`
			Email string `autopdf:"email"`
		}{
			Name:  "System Reporter",
			Email: "reports@example.com",
		},
		Summary: struct {
			TotalDocuments int     `autopdf:"total_documents"`
			TotalUsers     int     `autopdf:"total_users"`
			AverageScore   float64 `autopdf:"average_score"`
		}{
			TotalDocuments: 1250,
			TotalUsers:     89,
			AverageScore:   94.7,
		},
		Tags: []string{"monthly", "performance", "report", "2025"},
	}

	// Convert struct to variables
	fmt.Println("Converting report data to variables...")
	response, err := client.ConvertStructForTemplate(reportData, nil)
	if err != nil {
		fmt.Printf("Conversion failed: %v\n", err)
		return
	}

	if !response.Success {
		fmt.Printf("Conversion failed: %s\n", response.Message)
		return
	}

	fmt.Printf("Successfully converted %d variables\n", len(response.Variables))

	// Display the variables that would be used in PDF generation
	fmt.Println("\nVariables for PDF generation:")
	for key, value := range response.Variables {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// In a real application, you would now use these variables with AutoPDF's PDF generation API
	fmt.Println("\nThese variables can now be used with AutoPDF's PDF generation system:")
	fmt.Println("1. Pass the variables to your PDF template")
	fmt.Println("2. Use AutoPDF's PDF generation API to create the final document")
	fmt.Println("3. The template will be populated with the converted struct data")

	fmt.Println("\n=== PDF Generation Example Complete ===")
}
