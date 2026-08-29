// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
	"github.com/stretchr/testify/assert"
)

func TestWatchModeManagerReportsUnavailableCapability(t *testing.T) {
	manager := NewWatchModeManager(nil)
	err := manager.StartWatchMode(context.Background(), generation.PDFGenerationRequest{})
	assert.ErrorIs(t, err, ErrWatchModeUnavailable)
	assert.Empty(t, manager.GetActiveWatches())
}
