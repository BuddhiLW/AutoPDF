// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
)

// RealClockAdapter implements Clock using the standard library
type RealClockAdapter struct{}

// NewRealClockAdapter creates a new real clock adapter
func NewRealClockAdapter() ports.Clock {
	return &RealClockAdapter{}
}

// Now returns the current time
func (c *RealClockAdapter) Now() time.Time {
	return time.Now()
}

// Format formats a time using the given layout
func (c *RealClockAdapter) Format(t time.Time, layout string) string {
	return t.Format(layout)
}
