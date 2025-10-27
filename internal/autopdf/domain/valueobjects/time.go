// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package valueobjects

import "time"

// Time represents a time value object
type Time struct {
	value time.Time
}

// NewTime creates a new time value object
func NewTime(t time.Time) Time {
	return Time{value: t}
}

// Value returns the underlying time value
func (t Time) Value() time.Time {
	return t.value
}
