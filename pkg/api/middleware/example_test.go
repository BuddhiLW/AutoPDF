// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/config"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/middleware"
)

func ExampleDebugMiddleware() {
	handler := middleware.DebugMiddleware(&config.APIDebugConfig{})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			opts, _ := middleware.GetOptionsFromContext(r.Context())
			id, _ := middleware.GetRequestIDFromContext(r.Context())
			fmt.Println("debug:", opts.Debug.Enabled)
			fmt.Println("verbose:", opts.Verbose)
			fmt.Println("clean:", opts.DoClean)
			fmt.Println("convert:", opts.Conversion.Enabled)
			fmt.Println("timeout:", opts.Timeout)
			fmt.Println("request id in header:", w.Header().Get("X-AutoPDF-Request-ID") == id)
		}),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/pdf/generate", nil)
	req.Header.Set("X-AutoPDF-Debug", "true")
	req.Header.Set("X-AutoPDF-Verbose", "3")
	req.Header.Set("X-AutoPDF-Clean", "true")
	req.Header.Set("X-AutoPDF-Convert", "true")
	req.Header.Set("X-AutoPDF-Timeout", "45s")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	// Output:
	// debug: true
	// verbose: 3
	// clean: true
	// convert: true
	// timeout: 45s
	// request id in header: true
}
