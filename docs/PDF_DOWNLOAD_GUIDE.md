# AutoPDF Client Download Guide

This guide explains how clients can download generated PDFs and images from the AutoPDF REST API.

## Overview

AutoPDF provides multiple ways for clients to download generated files:

1. **Synchronous Generation**: Generate and download immediately
2. **Asynchronous Generation**: Generate in background, download when ready
3. **Direct Download**: Download previously generated files
4. **Multiple Formats**: PDF, PNG, JPEG, SVG

## API Endpoints

### 1. Generate PDF (Synchronous)

**Endpoint**: `POST /api/v1/pdf/generate`

**Request**:
```json
{
  "template_path": "templates/report.tex",
  "variables": {
    "title": "Monthly Report",
    "author": "John Doe",
    "date": "2025-01-07",
    "content": "Report content here..."
  },
  "options": {
    "engine": "xelatex",
    "debug": true,
    "cleanup": true,
    "timeout": 30,
    "conversion": {
      "do_convert": true,
      "format": "png",
      "quality": 95,
      "dpi": 300
    }
  }
}
```

**Response**:
```json
{
  "success": true,
  "request_id": "req-12345",
  "message": "PDF generated successfully",
  "files": [
    {
      "type": "pdf",
      "size": 1024000,
      "download_url": "/api/v1/pdf/download/req-12345",
      "expires_at": "2025-01-08T12:00:00Z"
    },
    {
      "type": "png",
      "size": 512000,
      "download_url": "/api/v1/pdf/download/req-12345/png",
      "expires_at": "2025-01-08T12:00:00Z"
    }
  ],
  "metadata": {
    "pages": "5",
    "generation_time": "2.5s"
  }
}
```

### 2. Generate PDF (Asynchronous)

**Endpoint**: `POST /api/v1/pdf/generate/async`

**Request**: Same as synchronous

**Response**:
```json
{
  "success": true,
  "request_id": "req-12345",
  "message": "PDF generation started",
  "status_url": "/api/v1/pdf/status/req-12345"
}
```

### 3. Check Generation Status

**Endpoint**: `GET /api/v1/pdf/status/{requestId}`

**Response**:
```json
{
  "request_id": "req-12345",
  "status": "completed",
  "progress": 100,
  "message": "Generation completed",
  "files": [
    {
      "type": "pdf",
      "size": 1024000,
      "download_url": "/api/v1/pdf/download/req-12345",
      "expires_at": "2025-01-08T12:00:00Z"
    }
  ]
}
```

### 4. Download Files

**Endpoint**: `GET /api/v1/pdf/download/{requestId}`

**Headers**:
```
Content-Type: application/pdf
Content-Disposition: attachment; filename="document_req-12345.pdf"
Content-Length: 1024000
X-Request-ID: req-12345
Cache-Control: private, max-age=3600
```

**Endpoint**: `GET /api/v1/pdf/download/{requestId}/{format}`

**Supported formats**: `png`, `jpeg`, `svg`

## Client Implementation Examples

### 1. Go Client

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

type PDFClient struct {
    BaseURL string
    Client  *http.Client
}

func NewPDFClient(baseURL string) *PDFClient {
    return &PDFClient{
        BaseURL: baseURL,
        Client: &http.Client{
            Timeout: 60 * time.Second,
        },
    }
}

func (c *PDFClient) GenerateAndDownload(templatePath string, variables map[string]interface{}, outputPath string) error {
    // 1. Generate PDF
    req := map[string]interface{}{
        "template_path": templatePath,
        "variables":     variables,
        "options": map[string]interface{}{
            "engine": "xelatex",
            "debug":  true,
        },
    }
    
    resp, err := c.makeRequest("POST", "/api/v1/pdf/generate", req)
    if err != nil {
        return err
    }
    
    // 2. Extract request ID
    requestID := resp["request_id"].(string)
    
    // 3. Download file
    return c.downloadFile(requestID, "", outputPath)
}

func (c *PDFClient) downloadFile(requestID, format, outputPath string) error {
    url := c.BaseURL + "/api/v1/pdf/download/" + requestID
    if format != "" {
        url += "/" + format
    }
    
    resp, err := c.Client.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("download failed: %d", resp.StatusCode)
    }
    
    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    _, err = io.Copy(file, resp.Body)
    return err
}
```

### 2. JavaScript/Node.js Client

```javascript
const axios = require('axios');
const fs = require('fs');

class PDFClient {
    constructor(baseURL) {
        this.baseURL = baseURL;
        this.client = axios.create({
            timeout: 60000,
        });
    }

    async generateAndDownload(templatePath, variables, outputPath) {
        try {
            // 1. Generate PDF
            const response = await this.client.post(`${this.baseURL}/api/v1/pdf/generate`, {
                template_path: templatePath,
                variables: variables,
                options: {
                    engine: 'xelatex',
                    debug: true,
                }
            });

            const requestId = response.data.request_id;

            // 2. Download file
            await this.downloadFile(requestId, '', outputPath);
            
            return requestId;
        } catch (error) {
            throw new Error(`PDF generation failed: ${error.message}`);
        }
    }

    async downloadFile(requestId, format, outputPath) {
        const url = `${this.baseURL}/api/v1/pdf/download/${requestId}`;
        const downloadUrl = format ? `${url}/${format}` : url;

        const response = await this.client.get(downloadUrl, {
            responseType: 'stream'
        });

        const writer = fs.createWriteStream(outputPath);
        response.data.pipe(writer);

        return new Promise((resolve, reject) => {
            writer.on('finish', resolve);
            writer.on('error', reject);
        });
    }
}

// Usage
const client = new PDFClient('http://localhost:8080');
client.generateAndDownload('templates/report.tex', {
    title: 'Monthly Report',
    author: 'John Doe'
}, './output.pdf');
```

### 3. Python Client

```python
import requests
import json
import time

class PDFClient:
    def __init__(self, base_url):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.timeout = 60

    def generate_and_download(self, template_path, variables, output_path):
        # 1. Generate PDF
        payload = {
            "template_path": template_path,
            "variables": variables,
            "options": {
                "engine": "xelatex",
                "debug": True
            }
        }
        
        response = self.session.post(
            f"{self.base_url}/api/v1/pdf/generate",
            json=payload
        )
        response.raise_for_status()
        
        data = response.json()
        request_id = data["request_id"]
        
        # 2. Download file
        self.download_file(request_id, "", output_path)
        return request_id

    def download_file(self, request_id, format, output_path):
        url = f"{self.base_url}/api/v1/pdf/download/{request_id}"
        if format:
            url += f"/{format}"
        
        response = self.session.get(url)
        response.raise_for_status()
        
        with open(output_path, 'wb') as f:
            f.write(response.content)

# Usage
client = PDFClient('http://localhost:8080')
client.generate_and_download('templates/report.tex', {
    'title': 'Monthly Report',
    'author': 'John Doe'
}, './output.pdf')
```

### 4. cURL Examples

```bash
# Generate PDF
curl -X POST http://localhost:8080/api/v1/pdf/generate \
  -H "Content-Type: application/json" \
  -d '{
    "template_path": "templates/report.tex",
    "variables": {
      "title": "Monthly Report",
      "author": "John Doe"
    },
    "options": {
      "engine": "xelatex",
      "debug": true
    }
  }' \
  -o response.json

# Extract request ID
REQUEST_ID=$(jq -r '.request_id' response.json)

# Download PDF
curl -X GET "http://localhost:8080/api/v1/pdf/download/$REQUEST_ID" \
  -o "document.pdf"

# Download PNG image
curl -X GET "http://localhost:8080/api/v1/pdf/download/$REQUEST_ID/png" \
  -o "document.png"
```

## Download Patterns

### 1. Direct Download (Recommended for Small Files)

```go
// Generate and download in one request
func (c *PDFClient) GenerateAndDownloadDirect(templatePath string, variables map[string]interface{}) ([]byte, error) {
    req := PDFGenerationRequest{
        TemplatePath: templatePath,
        Variables:    variables,
        Options: &PDFGenerationOptions{
            Engine: "xelatex",
            Debug:  true,
        },
    }
    
    resp, err := c.GeneratePDF(req)
    if err != nil {
        return nil, err
    }
    
    // Download immediately
    return c.DownloadFileBytes(resp.RequestID, "")
}
```

### 2. Async Download (Recommended for Large Files)

```go
// Generate asynchronously and poll for completion
func (c *PDFClient) GenerateAndDownloadAsync(templatePath string, variables map[string]interface{}) error {
    // Start async generation
    asyncResp, err := c.GeneratePDFAsync(PDFGenerationRequest{
        TemplatePath: templatePath,
        Variables:    variables,
    })
    if err != nil {
        return err
    }
    
    // Poll for completion
    for {
        status, err := c.GetGenerationStatus(asyncResp.RequestID)
        if err != nil {
            return err
        }
        
        switch status.Status {
        case "completed":
            // Download all files
            for _, file := range status.Files {
                err := c.DownloadFile(asyncResp.RequestID, file.Type, fmt.Sprintf("./output.%s", file.Type))
                if err != nil {
                    return err
                }
            }
            return nil
            
        case "failed":
            return fmt.Errorf("generation failed: %s", status.Error)
            
        case "pending", "processing":
            time.Sleep(2 * time.Second)
            continue
        }
    }
}
```

### 3. Batch Download

```go
// Download multiple formats
func (c *PDFClient) DownloadAllFormats(requestID string, outputDir string) error {
    formats := []string{"pdf", "png", "jpeg", "svg"}
    
    for _, format := range formats {
        outputPath := filepath.Join(outputDir, fmt.Sprintf("document.%s", format))
        err := c.DownloadFile(requestID, format, outputPath)
        if err != nil {
            log.Printf("Failed to download %s: %v", format, err)
            continue
        }
        log.Printf("Downloaded %s successfully", format)
    }
    
    return nil
}
```

## Error Handling

### Common Error Scenarios

1. **Template Not Found**
```json
{
  "success": false,
  "message": "Template not found: templates/missing.tex"
}
```

2. **Generation Timeout**
```json
{
  "success": false,
  "message": "PDF generation timeout after 30 seconds"
}
```

3. **Invalid Variables**
```json
{
  "success": false,
  "message": "Template variable validation failed: missing required variable 'title'"
}
```

4. **File Not Found**
```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{
  "error": "File not found for request ID: req-12345"
}
```

### Error Handling Best Practices

```go
func (c *PDFClient) GenerateWithRetry(req PDFGenerationRequest, maxRetries int) (*PDFGenerationResponse, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        resp, err := c.GeneratePDF(req)
        if err == nil {
            return resp, nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !isRetryableError(err) {
            return nil, err
        }
        
        // Exponential backoff
        time.Sleep(time.Duration(1<<i) * time.Second)
    }
    
    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func isRetryableError(err error) bool {
    // Check for timeout, network errors, 5xx status codes
    return strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "connection") ||
           strings.Contains(err.Error(), "500")
}
```

## Performance Considerations

### 1. File Size Limits

- **PDF**: Up to 100MB
- **Images**: Up to 50MB per image
- **Total request**: Up to 200MB

### 2. Timeout Settings

- **Synchronous generation**: 30-60 seconds
- **Asynchronous generation**: 5-10 minutes
- **Download timeout**: 60 seconds

### 3. Caching Strategy

```go
// Cache generated files for reuse
type CachedPDFClient struct {
    *PDFClient
    cache map[string][]byte
    mutex sync.RWMutex
}

func (c *CachedPDFClient) GenerateWithCache(templatePath string, variables map[string]interface{}) ([]byte, error) {
    // Create cache key
    key := fmt.Sprintf("%s:%x", templatePath, hashVariables(variables))
    
    // Check cache
    c.mutex.RLock()
    if cached, exists := c.cache[key]; exists {
        c.mutex.RUnlock()
        return cached, nil
    }
    c.mutex.RUnlock()
    
    // Generate and cache
    resp, err := c.GeneratePDF(PDFGenerationRequest{
        TemplatePath: templatePath,
        Variables:    variables,
    })
    if err != nil {
        return nil, err
    }
    
    pdfBytes, err := c.DownloadFileBytes(resp.RequestID, "")
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    c.mutex.Lock()
    c.cache[key] = pdfBytes
    c.mutex.Unlock()
    
    return pdfBytes, nil
}
```

## Security Considerations

### 1. Authentication

```go
// Add authentication headers
func (c *PDFClient) makeAuthenticatedRequest(method, url string, body interface{}) (*http.Response, error) {
    req, err := http.NewRequest(method, url, body)
    if err != nil {
        return nil, err
    }
    
    // Add JWT token
    req.Header.Set("Authorization", "Bearer "+c.authToken)
    
    return c.Client.Do(req)
}
```

### 2. File Validation

```go
// Validate downloaded files
func validatePDFFile(data []byte) error {
    if len(data) < 4 {
        return fmt.Errorf("file too small")
    }
    
    // Check PDF header
    if !bytes.HasPrefix(data, []byte("%PDF-")) {
        return fmt.Errorf("invalid PDF format")
    }
    
    return nil
}

func validateImageFile(data []byte, format string) error {
    switch format {
    case "png":
        if !bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
            return fmt.Errorf("invalid PNG format")
        }
    case "jpeg":
        if !bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
            return fmt.Errorf("invalid JPEG format")
        }
    }
    
    return nil
}
```

## Summary

AutoPDF provides a comprehensive REST API for PDF generation and download with the following key features:

- **Multiple generation modes**: Synchronous and asynchronous
- **Multiple output formats**: PDF, PNG, JPEG, SVG
- **Robust error handling**: Detailed error messages and retry logic
- **Performance optimization**: Caching, timeouts, and file size limits
- **Security**: Authentication and file validation
- **Client libraries**: Go, JavaScript, Python examples provided

The API follows RESTful principles and provides clear, consistent responses that make it easy for clients to integrate PDF generation into their applications.
