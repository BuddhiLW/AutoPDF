// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
)

// StringFormatter abstracts string formatting operations (DIP compliance)
type StringFormatter interface {
	Format(format string, args ...interface{}) string
}

// DefaultStringFormatter uses fmt package (infrastructure adapter)
type DefaultStringFormatter struct{}

func (f *DefaultStringFormatter) Format(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
