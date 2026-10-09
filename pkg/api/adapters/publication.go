// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import "fmt"

// publication decides whether a staged PDF may cross the boundary. Pure.
//
// A compiler may exit non-zero after writing a finished document, so a build
// error alone does not refuse it. An unfinished document is always refused,
// and when the build also failed, the build error leads: it names the cause,
// the defect names the symptom.
func publication(buildErr, incomplete error) error {
	switch {
	case incomplete == nil:
		return nil
	case buildErr != nil:
		return fmt.Errorf("%w (%w)", buildErr, incomplete)
	default:
		return incomplete
	}
}
