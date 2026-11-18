// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"net/http"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	"github.com/go-chi/render"
)

// WithRegistry makes the API advertise the engines of registry.
func (api *PDFGenerationAPI) WithRegistry(registry *backend.Registry) *PDFGenerationAPI {
	api.registry = registry
	return api
}

func (api *PDFGenerationAPI) backends() *backend.Registry {
	if api.registry == nil {
		return backend.Default()
	}
	return api.registry
}

// GetSupportedEngines returns the engines of the registered backends
// GET /api/v1/pdf/engines
func (api *PDFGenerationAPI) GetSupportedEngines(w http.ResponseWriter, r *http.Request) {
	registry := api.backends()
	render.JSON(w, r, EnginesResponse{
		Engines: registry.EngineNames(),
		Default: registry.DefaultEngine(),
	})
}
