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

// WatchModeRESTExample demonstrates how to use watch mode via REST API
func WatchModeRESTExample() {
	fmt.Println("=== Watch Mode REST API Example ===")

	baseURL := "http://localhost:8080/api/v1/pdf"

	// Example 1: Generate PDF with watch mode enabled
	fmt.Println("1. Generating PDF with watch mode enabled...")

	requestBody := map[string]interface{}{
		"template_path": "example.tex",
		"variables": map[string]interface{}{
			"title":   "Watch Mode REST Test",
			"author":  "AutoPDF",
			"content": "This document was generated with watch mode enabled via REST API.",
		},
		"options": map[string]interface{}{
			"engine":     "pdflatex",
			"debug":      true,
			"watch_mode": true, // Enable watch mode
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}

	resp, err := http.Post(baseURL+"/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		return
	}

	fmt.Printf("Response: %+v\n", response)

	// Check if watch mode is active in the response
	if watchMode, ok := response["watch_mode"].(bool); ok && watchMode {
		fmt.Println("✅ Watch mode is active!")
	} else {
		fmt.Println("❌ Watch mode is not active")
	}

	// Example 2: Check active watch modes
	fmt.Println("\n2. Checking active watch modes...")

	resp, err = http.Get(baseURL + "/watch")
	if err != nil {
		fmt.Printf("Error checking watch modes: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading watch modes response: %v\n", err)
		return
	}

	var watchResponse map[string]interface{}
	err = json.Unmarshal(body, &watchResponse)
	if err != nil {
		fmt.Printf("Error unmarshaling watch response: %v\n", err)
		return
	}

	fmt.Printf("Active watch modes: %+v\n", watchResponse)

	// Example 3: Wait and check again
	fmt.Println("\n3. Waiting 5 seconds and checking again...")
	time.Sleep(5 * time.Second)

	resp, err = http.Get(baseURL + "/watch")
	if err != nil {
		fmt.Printf("Error checking watch modes: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading watch modes response: %v\n", err)
		return
	}

	err = json.Unmarshal(body, &watchResponse)
	if err != nil {
		fmt.Printf("Error unmarshaling watch response: %v\n", err)
		return
	}

	fmt.Printf("Active watch modes after 5 seconds: %+v\n", watchResponse)

	// Example 4: Stop all watch modes
	fmt.Println("\n4. Stopping all watch modes...")

	req, err := http.NewRequest("DELETE", baseURL+"/watch", nil)
	if err != nil {
		fmt.Printf("Error creating delete request: %v\n", err)
		return
	}

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Error stopping watch modes: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading stop response: %v\n", err)
		return
	}

	var stopResponse map[string]interface{}
	err = json.Unmarshal(body, &stopResponse)
	if err != nil {
		fmt.Printf("Error unmarshaling stop response: %v\n", err)
		return
	}

	fmt.Printf("Stop response: %+v\n", stopResponse)

	// Example 5: Verify no active watch modes
	fmt.Println("\n5. Verifying no active watch modes...")

	resp, err = http.Get(baseURL + "/watch")
	if err != nil {
		fmt.Printf("Error checking watch modes: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading watch modes response: %v\n", err)
		return
	}

	err = json.Unmarshal(body, &watchResponse)
	if err != nil {
		fmt.Printf("Error unmarshaling watch response: %v\n", err)
		return
	}

	fmt.Printf("Active watch modes after stopping: %+v\n", watchResponse)

	fmt.Println("\n=== Watch Mode REST API Example Complete ===")
}

// ExampleRequest shows the structure of a watch mode request
func ExampleRequest() {
	fmt.Println("=== Example Watch Mode Request ===")

	request := map[string]interface{}{
		"template_path": "document.tex",
		"variables": map[string]interface{}{
			"title":    "My Document",
			"author":   "John Doe",
			"date":     "2025-01-15",
			"content":  "This is the main content of the document.",
			"sections": []string{"Introduction", "Methodology", "Results", "Conclusion"},
		},
		"options": map[string]interface{}{
			"engine":     "pdflatex",
			"debug":      true,
			"cleanup":    true,
			"timeout":    30,
			"watch_mode": true, // This enables watch mode
			"conversion": map[string]interface{}{
				"do_convert": true,
				"format":     "png",
				"quality":    90,
				"dpi":        300,
			},
		},
	}

	jsonData, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling example: %v\n", err)
		return
	}

	fmt.Println("Example request body:")
	fmt.Println(string(jsonData))

	fmt.Println("\n=== Example Request Complete ===")
}
