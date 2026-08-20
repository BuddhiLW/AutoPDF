// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package api is the supported entry point for embedding AutoPDF in Go programs.
//
// Construct an Engine with NewEngine, then call Engine.Generate with the caller's
// context. Request and Result are the stable transport contracts. Generator is
// the primary extension seam; WithGenerator allows another implementation or a
// decorator to reuse the same Engine contract. Logger is optional and defaults
// to a no-op implementation.
//
// Subpackages under api/domain, api/application, api/adapters, api/builders,
// api/factories, and api/services are compatibility layers. New integrations
// should not depend on them.
package api
