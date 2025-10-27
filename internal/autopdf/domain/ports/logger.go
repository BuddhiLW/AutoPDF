// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package ports

// DebugLogger abstracts debug logging operations
type DebugLogger interface {
	Warn(msg string, args ...interface{})
	Info(msg string, args ...interface{})
}
