// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/api/rest"
	"github.com/BuddhiLW/AutoPDF/pkg/converter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// RESTAPIServerExample demonstrates how to set up a complete REST API server
// with the AutoPDF Struct Converter endpoints
func RESTAPIServerExample() {
	fmt.Println("=== AutoPDF Struct Converter REST API Server Example ===")

	// Create router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// Add CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Add compression middleware
	r.Use(middleware.Compress(5))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","service":"autopdf-struct-converter"}`,
			time.Now().Format(time.RFC3339))
	})

	// API documentation endpoint
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>AutoPDF Struct Converter API</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .method { font-weight: bold; color: #007bff; }
        .path { font-family: monospace; background: #e9ecef; padding: 2px 5px; }
    </style>
</head>
<body>
    <h1>AutoPDF Struct Converter REST API</h1>
    <p>This API provides endpoints for converting Go structs to AutoPDF variables.</p>
    
    <h2>Available Endpoints</h2>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/v1/struct-converter/convert</span>
        <p>Convert a struct to AutoPDF variables</p>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/v1/struct-converter/convert/flattened</span>
        <p>Convert a struct to flattened AutoPDF variables</p>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/v1/struct-converter/convert/template</span>
        <p>Convert a struct optimized for template usage</p>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/struct-converter/config</span>
        <p>Get current converter configuration</p>
    </div>
    
    <div class="endpoint">
        <span class="method">PUT</span> <span class="path">/api/v1/struct-converter/config</span>
        <p>Update converter configuration</p>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/struct-converter/converters</span>
        <p>List registered converters</p>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/v1/struct-converter/validate</span>
        <p>Validate a struct for conversion</p>
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/v1/struct-converter/preview</span>
        <p>Preview struct conversion with limited output</p>
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/struct-converter/health</span>
        <p>Health check endpoint</p>
    </div>
    
    <h2>Example Request</h2>
    <pre>
POST /api/v1/struct-converter/convert
Content-Type: application/json

{
    "data": {
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30,
        "active": true
    },
    "options": {
        "tag_name": "autopdf",
        "omit_empty": true
    }
}
    </pre>
    
    <h2>Example Response</h2>
    <pre>
{
    "variables": {
        "name": "John Doe",
        "email": "john@example.com",
        "age": "30",
        "active": "true"
    },
    "metadata": {
        "converted_at": "2025-01-07T15:30:00Z",
        "variable_count": 4
    },
    "success": true,
    "message": "Conversion completed successfully"
}
    </pre>
</body>
</html>
		`)
	})

	// Create struct converter API with custom configuration
	customConverter := converter.NewConverterBuilder().
		WithTagName("autopdf").
		WithOmitEmpty(true).
		WithTimeFormat("2006-01-02 15:04:05").
		WithDurationFormat("string").
		WithSliceSeparator(", ").
		WithBuiltinConverters().
		Build()

	api := rest.NewStructConverterAPIWithConverter(customConverter)

	// Register struct converter routes
	rest.RegisterStructConverterRoutes(r, api)

	// Add example endpoints
	r.Route("/examples", func(r chi.Router) {
		r.Get("/", apiDocumentationHandler)
		r.Get("/test-data", testDataHandler)
		r.Post("/convert-example", convertExampleHandler)
	})

	// Start server
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting AutoPDF Struct Converter API server on port %s\n", port)
		fmt.Printf("API Documentation: http://localhost:%s/\n", port)
		fmt.Printf("Health Check: http://localhost:%s/health\n", port)
		fmt.Printf("Struct Converter API: http://localhost:%s/api/v1/struct-converter/\n", port)
		fmt.Printf("Examples: http://localhost:%s/examples/\n", port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}

// API documentation handler
func apiDocumentationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	docs := map[string]interface{}{
		"service": "AutoPDF Struct Converter API",
		"version": "1.0.0",
		"endpoints": map[string]interface{}{
			"convert": map[string]interface{}{
				"method":      "POST",
				"path":        "/api/v1/struct-converter/convert",
				"description": "Convert a struct to AutoPDF variables",
				"request_body": map[string]interface{}{
					"data":     "interface{} - The struct to convert",
					"options":  "ConversionOptions - Optional conversion settings",
					"metadata": "map[string]interface{} - Optional metadata",
				},
			},
			"convert_flattened": map[string]interface{}{
				"method":      "POST",
				"path":        "/api/v1/struct-converter/convert/flattened",
				"description": "Convert a struct to flattened AutoPDF variables",
			},
			"convert_template": map[string]interface{}{
				"method":      "POST",
				"path":        "/api/v1/struct-converter/convert/template",
				"description": "Convert a struct optimized for template usage",
			},
			"config": map[string]interface{}{
				"method":      "GET",
				"path":        "/api/v1/struct-converter/config",
				"description": "Get current converter configuration",
			},
			"validate": map[string]interface{}{
				"method":      "POST",
				"path":        "/api/v1/struct-converter/validate",
				"description": "Validate a struct for conversion",
			},
			"preview": map[string]interface{}{
				"method":      "POST",
				"path":        "/api/v1/struct-converter/preview",
				"description": "Preview struct conversion with limited output",
			},
		},
		"examples": map[string]interface{}{
			"test_data":       "/examples/test-data",
			"convert_example": "/examples/convert-example",
		},
	}

	fmt.Fprintf(w, `%s`, formatJSON(docs))
}

// Test data handler
func testDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	testData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":     123,
			"name":   "John Doe",
			"email":  "john.doe@example.com",
			"active": true,
			"profile": map[string]interface{}{
				"bio":     "Software engineer",
				"website": "https://johndoe.dev",
			},
			"tags": []string{"admin", "user", "premium"},
		},
		"document": map[string]interface{}{
			"title":   "AutoPDF Integration Guide",
			"content": "This document explains how to use AutoPDF with struct conversion.",
			"author": map[string]interface{}{
				"name":  "Jane Smith",
				"email": "jane.smith@example.com",
			},
			"tags":      []string{"documentation", "integration", "autopdf"},
			"published": true,
			"version":   1,
		},
		"report": map[string]interface{}{
			"report_id":    "RPT-2025-001",
			"title":        "Monthly Performance Report",
			"generated_at": time.Now().Format(time.RFC3339),
			"summary": map[string]interface{}{
				"total_documents": 1250,
				"total_users":     89,
				"average_score":   94.7,
			},
			"tags": []string{"monthly", "performance", "report"},
		},
	}

	fmt.Fprintf(w, `%s`, formatJSON(testData))
}

// Convert example handler
func convertExampleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Example struct for demonstration
	type ExampleStruct struct {
		Name      string    `autopdf:"name"`
		Email     string    `autopdf:"email"`
		Age       int       `autopdf:"age"`
		Active    bool      `autopdf:"active"`
		CreatedAt time.Time `autopdf:"created_at"`
		Tags      []string  `autopdf:"tags,flatten"`
		Profile   struct {
			Bio     string `autopdf:"bio"`
			Website string `autopdf:"website"`
		} `autopdf:"profile"`
	}

	exampleData := ExampleStruct{
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Now(),
		Tags:      []string{"admin", "user", "premium"},
		Profile: struct {
			Bio     string `autopdf:"bio"`
			Website string `autopdf:"website"`
		}{
			Bio:     "Software engineer with 10+ years experience",
			Website: "https://johndoe.dev",
		},
	}

	// Convert using the converter
	converter := converter.BuildWithDefaults()
	variables, err := converter.ConvertStruct(exampleData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Conversion failed: %s"}`, err.Error())
		return
	}

	// Convert to map for JSON response
	result := make(map[string]interface{})
	variables.Range(func(name string, value interface{}) bool {
		result[name] = value
		return true
	})

	response := map[string]interface{}{
		"original_struct":     exampleData,
		"converted_variables": result,
		"conversion_metadata": map[string]interface{}{
			"converted_at":   time.Now().Format(time.RFC3339),
			"variable_count": len(result),
			"converter_type": "default",
		},
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `%s`, formatJSON(response))
}

// Helper function to format JSON
func formatJSON(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "Failed to format JSON: %s"}`, err.Error())
	}
	return string(jsonBytes)
}

// Example of how to run the server
func RunRESTAPIServer() {
	// This function can be called from main() to start the server
	RESTAPIServerExample()
}
