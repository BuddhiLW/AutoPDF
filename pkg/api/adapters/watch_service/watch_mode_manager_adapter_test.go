// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package watch_service

import (
	"context"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/stretchr/testify/assert"
)

func TestAdapterDoesNotClaimWatchModeStarted(t *testing.T) {
	manager := NewWatchModeManagerAdapter(nil)
	err := manager.StartWatchMode(context.Background(), generation.PDFGenerationRequest{})
	assert.ErrorIs(t, err, ErrWatchModeUnavailable)
	assert.Empty(t, manager.GetActiveWatches())
}
