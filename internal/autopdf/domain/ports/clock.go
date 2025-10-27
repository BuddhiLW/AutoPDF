// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package ports

import "time"

// Clock abstracts time operations
type Clock interface {
	Now() time.Time
	Format(t time.Time, layout string) string
}
