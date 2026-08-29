// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnsupportedAsyncEndpointsDoNotReportSuccess(t *testing.T) {
	api := NewPDFGenerationAPI(config.GetDefaultConfig())

	t.Run("generate", func(t *testing.T) {
		response := httptest.NewRecorder()
		api.GeneratePDFAsync(response, httptest.NewRequest(http.MethodPost, "/api/v1/pdf/generate/async", nil))
		assert.Equal(t, http.StatusNotImplemented, response.Code)
		assert.Contains(t, response.Body.String(), `"success":false`)
	})

	t.Run("status", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/v1/pdf/status/id", nil), "requestId", "id")
		api.GetGenerationStatus(response, request)
		assert.Equal(t, http.StatusNotImplemented, response.Code)
		assert.Contains(t, response.Body.String(), `"status":"unsupported"`)
	})

	t.Run("download", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/v1/pdf/download/id", nil), "requestId", "id")
		api.DownloadFile(response, request)
		assert.Equal(t, http.StatusNotImplemented, response.Code)
		assert.NotEqual(t, "application/pdf", response.Header().Get("Content-Type"))
	})
}

func withRouteParam(request *http.Request, key, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
}
