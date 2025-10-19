// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Struct-to-PDF Example demonstrating direct struct conversion to PDF

// Invoice represents a simple invoice structure
type Invoice struct {
	InvoiceNumber string     `autopdf:"invoice_number"`
	IssueDate     time.Time  `autopdf:"issue_date"`
	DueDate       time.Time  `autopdf:"due_date"`
	Vendor        Vendor     `autopdf:"vendor"`
	Customer      Customer   `autopdf:"customer"`
	Items         []LineItem `autopdf:"items"`
	Subtotal      float64    `autopdf:"subtotal"`
	Tax           float64    `autopdf:"tax"`
	Total         float64    `autopdf:"total"`
	Notes         string     `autopdf:"notes"`
}

// Vendor represents vendor information
type Vendor struct {
	Name    string `autopdf:"name"`
	Address string `autopdf:"address"`
	Phone   string `autopdf:"phone"`
	Email   string `autopdf:"email"`
}

// Customer represents customer information
type Customer struct {
	Name    string `autopdf:"name"`
	Address string `autopdf:"address"`
	Phone   string `autopdf:"phone"`
	Email   string `autopdf:"email"`
}

// LineItem represents an invoice line item
type LineItem struct {
	Description string  `autopdf:"description"`
	Quantity    int     `autopdf:"quantity"`
	UnitPrice   float64 `autopdf:"unit_price"`
	Total       float64 `autopdf:"total"`
}

// StructToPDFClient provides methods for generating PDFs from structs
type StructToPDFClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewStructToPDFClient creates a new struct-to-PDF client
func NewStructToPDFClient(baseURL string) *StructToPDFClient {
	return &StructToPDFClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for PDF generation
		},
	}
}

// GeneratePDFFromStruct generates a PDF from a struct using the REST API
func (client *StructToPDFClient) GeneratePDFFromStruct(templatePath string, data interface{}, options *StructPDFOptions) (*StructPDFGenerationResponse, error) {
	req := StructPDFGenerationRequest{
		TemplatePath: templatePath,
		Data:         data,
		Options:      options,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := client.BaseURL + "/api/v1/pdf/generate/from-struct"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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

	var result StructPDFGenerationResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// StructPDFGenerationRequest represents a struct-based PDF generation request
type StructPDFGenerationRequest struct {
	TemplatePath string            `json:"template_path"`
	Data         interface{}       `json:"data"`
	Options      *StructPDFOptions `json:"options,omitempty"`
}

// StructPDFOptions represents options for struct-based PDF generation
type StructPDFOptions struct {
	Engine    string `json:"engine,omitempty"`
	Debug     bool   `json:"debug,omitempty"`
	WatchMode bool   `json:"watch_mode,omitempty"`
}

// StructPDFGenerationResponse represents the API response (from rest/pdf_generation_api.go)
type StructPDFGenerationResponse struct {
	Success     bool                  `json:"success"`
	RequestID   string                `json:"request_id"`
	Message     string                `json:"message,omitempty"`
	Files       []StructGeneratedFile `json:"files,omitempty"`
	Metadata    map[string]string     `json:"metadata,omitempty"`
	DownloadURL string                `json:"download_url,omitempty"`
	WatchMode   bool                  `json:"watch_mode,omitempty"`
}

// StructGeneratedFile represents a generated file in the response
type StructGeneratedFile struct {
	Type        string `json:"type"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"download_url"`
	ExpiresAt   string `json:"expires_at"`
}

// ExampleStructToPDFWorkflow demonstrates the struct-to-PDF workflow
func ExampleStructToPDFWorkflow() {
	fmt.Println("=== Struct-to-PDF Workflow Example ===")
	fmt.Println("")

	// Create client
	client := NewStructToPDFClient("http://localhost:8080")

	// Create invoice data
	invoice := Invoice{
		InvoiceNumber: "INV-2025-001",
		IssueDate:     time.Date(2025, 10, 16, 0, 0, 0, 0, time.UTC),
		DueDate:       time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC),
		Vendor: Vendor{
			Name:    "AutoPDF Solutions Inc.",
			Address: "123 Tech Street, San Francisco, CA 94105",
			Phone:   "+1-555-PDF-AUTO",
			Email:   "billing@autopdf.com",
		},
		Customer: Customer{
			Name:    "Funerária Francana",
			Address: "456 Business Ave, Franca, SP 14400-000, Brazil",
			Phone:   "+55-16-3333-4444",
			Email:   "finance@funerariafrancana.com.br",
		},
		Items: []LineItem{
			{
				Description: "AutoPDF Enterprise License - Annual",
				Quantity:    1,
				UnitPrice:   999.00,
				Total:       999.00,
			},
			{
				Description: "Professional Support Package",
				Quantity:    12,
				UnitPrice:   99.00,
				Total:       1188.00,
			},
			{
				Description: "Custom Template Development",
				Quantity:    5,
				UnitPrice:   250.00,
				Total:       1250.00,
			},
		},
		Subtotal: 3437.00,
		Tax:      343.70,
		Total:    3780.70,
		Notes:    "Payment due within 30 days. Thank you for your business!",
	}

	// Generate PDF from struct
	fmt.Println("1. Generating PDF from Invoice struct...")
	response, err := client.GeneratePDFFromStruct(
		"templates/invoice.tex",
		invoice,
		&StructPDFOptions{
			Engine:    "xelatex",
			Debug:     true,
			WatchMode: false,
		},
	)

	if err != nil {
		fmt.Printf("   ❌ Failed to generate PDF: %v\n", err)
		return
	}

	if !response.Success {
		fmt.Printf("   ❌ Generation failed: %s\n", response.Message)
		return
	}

	fmt.Printf("   ✓ Success! Request ID: %s\n", response.RequestID)
	fmt.Printf("   ✓ Message: %s\n", response.Message)
	fmt.Printf("   ✓ Files generated: %d\n", len(response.Files))

	// Show metadata
	fmt.Println("\n2. Metadata:")
	for key, value := range response.Metadata {
		fmt.Printf("   %s: %s\n", key, value)
	}

	// Show files
	fmt.Println("\n3. Generated Files:")
	for i, file := range response.Files {
		fmt.Printf("   File %d:\n", i+1)
		fmt.Printf("      Type: %s\n", file.Type)
		fmt.Printf("      Size: %d bytes\n", file.Size)
		fmt.Printf("      Download: %s\n", file.DownloadURL)
		fmt.Printf("      Expires: %s\n", file.ExpiresAt)
	}

	fmt.Println("\n=== Workflow Complete ===")
}

// ExampleLocalStructToPDF demonstrates struct-to-PDF conversion using the local API
func ExampleLocalStructToPDF() {
	fmt.Println("=== Local Struct-to-PDF Example ===")
	fmt.Println("")

	// This example shows how to use the services package directly
	// without going through HTTP

	// Create invoice
	invoice := Invoice{
		InvoiceNumber: "INV-2025-002",
		IssueDate:     time.Now(),
		DueDate:       time.Now().Add(30 * 24 * time.Hour),
		Vendor: Vendor{
			Name:    "AutoPDF Solutions Inc.",
			Address: "123 Tech Street, San Francisco, CA 94105",
			Phone:   "+1-555-PDF-AUTO",
			Email:   "billing@autopdf.com",
		},
		Customer: Customer{
			Name:    "Example Corporation",
			Address: "789 Corporate Blvd, New York, NY 10001",
			Phone:   "+1-555-EXAMPLE",
			Email:   "ap@example.com",
		},
		Items: []LineItem{
			{
				Description: "Consulting Services",
				Quantity:    40,
				UnitPrice:   150.00,
				Total:       6000.00,
			},
		},
		Subtotal: 6000.00,
		Tax:      600.00,
		Total:    6600.00,
		Notes:    "Net 30 payment terms",
	}

	fmt.Printf("Invoice struct created:\n")
	fmt.Printf("   Number: %s\n", invoice.InvoiceNumber)
	fmt.Printf("   Customer: %s\n", invoice.Customer.Name)
	fmt.Printf("   Total: $%.2f\n", invoice.Total)
	fmt.Printf("   Items: %d\n", len(invoice.Items))

	fmt.Println("\n✓ This struct can be directly passed to GeneratePDFFromStruct()")
	fmt.Println("✓ The StructConverter will automatically convert all fields")
	fmt.Println("✓ Nested structs (Vendor, Customer) are flattened with dot notation")
	fmt.Println("✓ Arrays (Items) are indexed with bracket notation")

	fmt.Println("\nExample variable names generated:")
	fmt.Println("   invoice_number    -> INV-2025-002")
	fmt.Println("   vendor.name       -> AutoPDF Solutions Inc.")
	fmt.Println("   customer.email    -> ap@example.com")
	fmt.Println("   items[0].description -> Consulting Services")
	fmt.Println("   total             -> 6600.00")

	fmt.Println("\n=== Example Complete ===")
}
