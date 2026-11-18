// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/rest"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/converter"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const baseURL = "http://localhost:8080"

func call(method, url string, body, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, data)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

func download(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = io.Copy(f, resp.Body)
	return err
}

func ExampleRegisterStructConverterRoutes() {
	r := chi.NewRouter()
	rest.RegisterStructConverterRoutesWithDefaults(r)
	srv := httptest.NewServer(r)
	defer srv.Close()

	var cfg rest.ConverterConfigResponse
	if err := call(http.MethodGet, srv.URL+"/api/v1/struct-converter/config", nil, &cfg); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("tag=%s separator=%q\n", cfg.TagName, cfg.SliceSeparator)
	// Output:
	// tag=autopdf separator=", "
}

func ExampleNewServer() {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID, chimiddleware.RealIP, chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	conv := converter.NewConverterBuilder().
		WithBuiltinConverters().
		WithOmitEmpty(true).
		WithTimeFormat("2006-01-02 15:04:05").
		Build()
	rest.RegisterStructConverterRoutes(r, rest.NewStructConverterAPIWithConverter(conv))
	rest.RegisterPDFGenerationRoutes(r, rest.NewPDFGenerationAPI(&config.Config{}))

	server, err := rest.NewServer(rest.ServerOptions{Addr: ":8080", Handler: r})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}

func ExamplePDFGenerationRequest() {
	req := rest.PDFGenerationRequest{
		TemplatePath: "document.tex",
		Variables: map[string]any{
			"title":  "My Document",
			"author": "John Doe",
		},
		Options: &rest.PDFGenerationOptions{
			Engine:    "pdflatex",
			Cleanup:   true,
			Timeout:   30,
			WatchMode: true,
			Conversion: rest.RESTConversionOptions{
				DoConvert: true,
				Format:    "png",
				DPI:       300,
			},
		},
	}

	body, _ := json.MarshalIndent(req, "", "  ")
	fmt.Println(string(body))
	// Output:
	// {
	//   "template_path": "document.tex",
	//   "variables": {
	//     "author": "John Doe",
	//     "title": "My Document"
	//   },
	//   "options": {
	//     "engine": "pdflatex",
	//     "conversion": {
	//       "do_convert": true,
	//       "format": "png",
	//       "dpi": 300
	//     },
	//     "cleanup": true,
	//     "timeout": 30,
	//     "watch_mode": true
	//   }
	// }
}

func ExamplePDFGenerationResponse() {
	var resp rest.PDFGenerationResponse
	err := call(http.MethodPost, baseURL+"/api/v1/pdf/generate", rest.PDFGenerationRequest{
		TemplatePath: "templates/report.tex",
		Variables: map[string]any{
			"title":   "Monthly Report",
			"summary": map[string]any{"total_users": 150, "revenue": 25000.50},
		},
		Options: &rest.PDFGenerationOptions{
			Engine:     "xelatex",
			Conversion: rest.RESTConversionOptions{DoConvert: true, Format: "png", DPI: 300},
		},
	}, &resp)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range resp.Files {
		path := filepath.Join("downloads", resp.RequestID+"."+file.Type)
		if err := download(baseURL+file.DownloadURL, path); err != nil {
			log.Fatal(err)
		}
		fmt.Println("downloaded", path, file.Size, "bytes")
	}
}

func ExampleAsyncPDFGenerationResponse() {
	var started rest.AsyncPDFGenerationResponse
	err := call(http.MethodPost, baseURL+"/api/v1/pdf/generate/async", rest.PDFGenerationRequest{
		TemplatePath: "templates/large-report.tex",
		Variables:    map[string]any{"title": "Large Report"},
		Options:      &rest.PDFGenerationOptions{Engine: "pdflatex", Timeout: 60},
	}, &started)
	if err != nil {
		log.Fatal(err)
	}

	for {
		var status rest.GenerationStatusResponse
		if err := call(http.MethodGet, baseURL+"/api/v1/pdf/status/"+started.RequestID, nil, &status); err != nil {
			log.Fatal(err)
		}
		switch status.Status {
		case "completed":
			fmt.Println("files:", len(status.Files))
			return
		case "failed":
			log.Fatal(status.Error)
		}
		time.Sleep(2 * time.Second)
	}
}

func ExamplePDFGenerationStructRequest() {
	type LineItem struct {
		Description string  `autopdf:"description"`
		Quantity    int     `autopdf:"quantity"`
		UnitPrice   float64 `autopdf:"unit_price"`
	}
	type Invoice struct {
		Number   string     `autopdf:"invoice_number"`
		Customer string     `autopdf:"customer"`
		Items    []LineItem `autopdf:"items"`
		Total    float64    `autopdf:"total"`
	}

	var resp rest.PDFGenerationResponse
	err := call(http.MethodPost, baseURL+"/api/v1/pdf/generate/from-struct", rest.PDFGenerationStructRequest{
		TemplatePath: "templates/invoice.tex",
		Data: Invoice{
			Number:   "INV-2025-001",
			Customer: "Example Corporation",
			Items:    []LineItem{{Description: "Consulting", Quantity: 40, UnitPrice: 150}},
			Total:    6000,
		},
		Options: &rest.PDFGenerationOptions{Engine: "xelatex"},
	}, &resp)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.RequestID, resp.DownloadURL)
}

func ExamplePDFGenerationAPI_watchMode() {
	var resp rest.PDFGenerationResponse
	err := call(http.MethodPost, baseURL+"/api/v1/pdf/generate", rest.PDFGenerationRequest{
		TemplatePath: "example.tex",
		Variables:    map[string]any{"title": "Watch Mode"},
		Options:      &rest.PDFGenerationOptions{Engine: "pdflatex", WatchMode: true},
	}, &resp)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("watching:", resp.WatchMode)

	var active map[string]any
	if err := call(http.MethodGet, baseURL+"/api/v1/pdf/watch", nil, &active); err != nil {
		log.Fatal(err)
	}
	fmt.Println("active:", active)

	if err := call(http.MethodDelete, baseURL+"/api/v1/pdf/watch", nil, nil); err != nil {
		log.Fatal(err)
	}
}
