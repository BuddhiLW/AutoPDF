// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// PDFDownloadClient provides a client for downloading PDFs and images from AutoPDF API
type PDFDownloadClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewPDFDownloadClient creates a new PDF download client
func NewPDFDownloadClient(baseURL string) *PDFDownloadClient {
	return &PDFDownloadClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for PDF generation
		},
	}
}

// PDFGenerationRequest represents a request to generate a PDF
type PDFGenerationRequest struct {
	TemplatePath string                 `json:"template_path"`
	Variables    map[string]interface{} `json:"variables"`
	Options      *PDFGenerationOptions  `json:"options,omitempty"`
}

// PDFGenerationOptions represents PDF generation options
type PDFGenerationOptions struct {
	Engine       string                   `json:"engine,omitempty"`
	OutputFormat string                   `json:"output_format,omitempty"`
	Conversion   ExampleConversionOptions `json:"conversion,omitempty"`
	Debug        bool                     `json:"debug,omitempty"`
	Cleanup      bool                     `json:"cleanup,omitempty"`
	Timeout      int                      `json:"timeout,omitempty"`
}

// ExampleConversionOptions represents conversion options for images in examples
type ExampleConversionOptions struct {
	DoConvert bool    `json:"do_convert,omitempty"`
	Format    string  `json:"format,omitempty"`
	Quality   int     `json:"quality,omitempty"`
	DPI       int     `json:"dpi,omitempty"`
	Scale     float64 `json:"scale,omitempty"`
}

// PDFGenerationResponse represents the response from PDF generation
type PDFGenerationResponse struct {
	Success     bool              `json:"success"`
	RequestID   string            `json:"request_id"`
	Message     string            `json:"message,omitempty"`
	Files       []GeneratedFile   `json:"files,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	DownloadURL string            `json:"download_url,omitempty"`
}

// GeneratedFile represents a generated file
type GeneratedFile struct {
	Type        string `json:"type"`
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
}

// GenerationStatusResponse represents the status of an async generation
type GenerationStatusResponse struct {
	RequestID string          `json:"request_id"`
	Status    string          `json:"status"`
	Progress  int             `json:"progress,omitempty"`
	Message   string          `json:"message,omitempty"`
	Files     []GeneratedFile `json:"files,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// GeneratePDF generates a PDF synchronously
func (client *PDFDownloadClient) GeneratePDF(req PDFGenerationRequest) (*PDFGenerationResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", client.BaseURL+"/api/v1/pdf/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var response PDFGenerationResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GeneratePDFAsync generates a PDF asynchronously
func (client *PDFDownloadClient) GeneratePDFAsync(req PDFGenerationRequest) (*AsyncPDFGenerationResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", client.BaseURL+"/api/v1/pdf/generate/async", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var response AsyncPDFGenerationResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetGenerationStatus gets the status of an async generation
func (client *PDFDownloadClient) GetGenerationStatus(requestID string) (*GenerationStatusResponse, error) {
	httpReq, err := http.NewRequest("GET", client.BaseURL+"/api/v1/pdf/status/"+requestID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var response GenerationStatusResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// DownloadFile downloads a generated file
func (client *PDFDownloadClient) DownloadFile(requestID, format, outputPath string) error {
	url := client.BaseURL + "/api/v1/pdf/download/" + requestID
	if format != "" {
		url += "/" + format
	}

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Copy response body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DownloadAllFiles downloads all files from a generation response
func (client *PDFDownloadClient) DownloadAllFiles(response *PDFGenerationResponse, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, file := range response.Files {
		outputPath := filepath.Join(outputDir, fmt.Sprintf("document_%s.%s", response.RequestID, file.Type))

		fmt.Printf("Downloading %s file (%d bytes) to %s\n", file.Type, file.Size, outputPath)

		if err := client.DownloadFile(response.RequestID, file.Type, outputPath); err != nil {
			return fmt.Errorf("failed to download %s file: %w", file.Type, err)
		}

		fmt.Printf("Successfully downloaded %s\n", outputPath)
	}

	return nil
}

// ExamplePDFDownload demonstrates how to generate and download PDFs
func ExamplePDFDownload() {
	fmt.Println("=== AutoPDF PDF Download Example ===")

	// Create client
	client := NewPDFDownloadClient("http://localhost:8080")

	// 1. Generate PDF with conversion options
	fmt.Println("\n1. Generating PDF with image conversion...")

	req := PDFGenerationRequest{
		TemplatePath: "templates/report.tex",
		Variables: map[string]interface{}{
			"title":   "Monthly Report",
			"author":  "John Doe",
			"date":    time.Now().Format("2006-01-02"),
			"content": "This is a sample report content.",
			"summary": map[string]interface{}{
				"total_users":  150,
				"active_users": 120,
				"revenue":      25000.50,
			},
		},
		Options: &PDFGenerationOptions{
			Engine:  "xelatex",
			Debug:   true,
			Cleanup: true,
			Timeout: 30,
			Conversion: ExampleConversionOptions{
				DoConvert: true,
				Format:    "png",
				Quality:   95,
				DPI:       300,
				Scale:     1.0,
			},
		},
	}

	response, err := client.GeneratePDF(req)
	if err != nil {
		fmt.Printf("PDF generation failed: %v\n", err)
		return
	}

	if !response.Success {
		fmt.Printf("PDF generation failed: %s\n", response.Message)
		return
	}

	fmt.Printf("PDF generated successfully!\n")
	fmt.Printf("Request ID: %s\n", response.RequestID)
	fmt.Printf("Files available: %d\n", len(response.Files))

	// Display file information
	for _, file := range response.Files {
		fmt.Printf("  - %s file: %d bytes, expires at %s\n",
			file.Type, file.Size, file.ExpiresAt)
	}

	// 2. Download all files
	fmt.Println("\n2. Downloading all files...")
	outputDir := "./downloads"
	if err := client.DownloadAllFiles(response, outputDir); err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	// 3. Download specific format
	fmt.Println("\n3. Downloading specific format...")
	if err := client.DownloadFile(response.RequestID, "png", "./downloads/document.png"); err != nil {
		fmt.Printf("PNG download failed: %v\n", err)
	} else {
		fmt.Println("PNG file downloaded successfully")
	}

	// 4. Async generation example
	fmt.Println("\n4. Async PDF generation...")
	asyncReq := PDFGenerationRequest{
		TemplatePath: "templates/large-report.tex",
		Variables: map[string]interface{}{
			"title": "Large Report",
			"data":  "Large amount of data...",
		},
		Options: &PDFGenerationOptions{
			Engine:  "pdflatex",
			Timeout: 60,
		},
	}

	asyncResponse, err := client.GeneratePDFAsync(asyncReq)
	if err != nil {
		fmt.Printf("Async generation failed: %v\n", err)
		return
	}

	fmt.Printf("Async generation started: %s\n", asyncResponse.RequestID)
	fmt.Printf("Status URL: %s\n", asyncResponse.StatusURL)

	// 5. Poll for completion
	fmt.Println("\n5. Polling for completion...")
	for i := 0; i < 10; i++ {
		status, err := client.GetGenerationStatus(asyncResponse.RequestID)
		if err != nil {
			fmt.Printf("Status check failed: %v\n", err)
			break
		}

		fmt.Printf("Status: %s (Progress: %d%%)\n", status.Status, status.Progress)

		if status.Status == "completed" {
			fmt.Println("Async generation completed!")
			if len(status.Files) > 0 {
				fmt.Printf("Files available: %d\n", len(status.Files))
				for _, file := range status.Files {
					fmt.Printf("  - %s: %d bytes\n", file.Type, file.Size)
				}
			}
			break
		} else if status.Status == "failed" {
			fmt.Printf("Async generation failed: %s\n", status.Error)
			break
		}

		time.Sleep(2 * time.Second)
	}

	fmt.Println("\n=== PDF Download Example Complete ===")
}

// ExampleDirectDownload demonstrates direct file download
func ExampleDirectDownload() {
	fmt.Println("=== AutoPDF Direct Download Example ===")

	client := NewPDFDownloadClient("http://localhost:8080")

	// Download PDF directly
	fmt.Println("Downloading PDF directly...")
	if err := client.DownloadFile("request-123", "", "./downloads/direct.pdf"); err != nil {
		fmt.Printf("Direct download failed: %v\n", err)
	} else {
		fmt.Println("PDF downloaded successfully")
	}

	// Download PNG image
	fmt.Println("Downloading PNG image...")
	if err := client.DownloadFile("request-123", "png", "./downloads/direct.png"); err != nil {
		fmt.Printf("PNG download failed: %v\n", err)
	} else {
		fmt.Println("PNG downloaded successfully")
	}

	// Download JPEG image
	fmt.Println("Downloading JPEG image...")
	if err := client.DownloadFile("request-123", "jpeg", "./downloads/direct.jpg"); err != nil {
		fmt.Printf("JPEG download failed: %v\n", err)
	} else {
		fmt.Println("JPEG downloaded successfully")
	}

	fmt.Println("=== Direct Download Example Complete ===")
}

// ExampleWithErrorHandling demonstrates proper error handling
func ExampleWithErrorHandling() {
	fmt.Println("=== AutoPDF Error Handling Example ===")

	client := NewPDFDownloadClient("http://localhost:8080")

	// Test with invalid template
	fmt.Println("Testing with invalid template...")
	req := PDFGenerationRequest{
		TemplatePath: "nonexistent.tex",
		Variables: map[string]interface{}{
			"title": "Test",
		},
	}

	response, err := client.GeneratePDF(req)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else if !response.Success {
		fmt.Printf("Expected failure: %s\n", response.Message)
	}

	// Test with invalid request ID
	fmt.Println("Testing with invalid request ID...")
	if err := client.DownloadFile("invalid-id", "", "./downloads/test.pdf"); err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	fmt.Println("=== Error Handling Example Complete ===")
}
