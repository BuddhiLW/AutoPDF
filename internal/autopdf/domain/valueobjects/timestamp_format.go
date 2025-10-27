// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package valueobjects

// Timestamp format constants to replace magic strings
const (
	DebugFileTimestampFormat = "20060102-150405"
	LogTimestampFormat       = "2006-01-02 15:04:05"
)

// GetDebugFileTimestampFormat returns the timestamp format for debug files
func GetDebugFileTimestampFormat() string {
	return DebugFileTimestampFormat
}

// GetLogTimestampFormat returns the timestamp format for log entries
func GetLogTimestampFormat() string {
	return LogTimestampFormat
}
