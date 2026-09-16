// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package result

import (
	"errors"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/v2/configs"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/strict"
)

// BuildFailure names the phase a failed build died in and keeps the cause
// attached, so whoever prints the error reads what actually went wrong.
//
// It is a pure function of err: it reads no files and writes no logs.
//
// The returned error satisfies errors.Is against the phase sentinel and
// errors.As against the original cause, and its message carries the cause in
// full, which for a strict.MissingVariablesError is every offending key with
// its location.
func BuildFailure(err error) error {
	if err == nil {
		return nil
	}

	var missingVariables *strict.MissingVariablesError
	if errors.As(err, &missingVariables) {
		return fmt.Errorf("%w: %w", configs.TemplateError, err)
	}

	return fmt.Errorf("%w: %w", configs.BuildError, err)
}
