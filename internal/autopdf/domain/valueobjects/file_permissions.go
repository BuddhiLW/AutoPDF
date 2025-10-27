// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package valueobjects

import "os"

// File permission constants to replace magic numbers
const (
	DirectoryPermission = 0755
	FilePermission      = 0644
)

// GetDirectoryPermission returns the standard directory permission
func GetDirectoryPermission() os.FileMode {
	return DirectoryPermission
}

// GetFilePermission returns the standard file permission
func GetFilePermission() os.FileMode {
	return FilePermission
}
