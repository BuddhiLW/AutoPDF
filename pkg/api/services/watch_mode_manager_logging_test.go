// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	ports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordedEntry is one message captured by recordingLogger.
type recordedEntry struct {
	Level   string
	Message string
	Fields  map[string]interface{}
}

// recordingLogger is a test double implementing the ports.Logger port. It
// exists to make log output assertable; depending on the port rather than on
// *logger.LoggerAdapter is what allows it to be substituted here at all.
type recordingLogger struct {
	entries []recordedEntry
}

func (r *recordingLogger) record(level, msg string, fields []ports.LogField) {
	collected := make(map[string]interface{}, len(fields))
	for _, field := range fields {
		collected[field.Key] = field.Value
	}
	r.entries = append(r.entries, recordedEntry{Level: level, Message: msg, Fields: collected})
}

func (r *recordingLogger) Debug(_ context.Context, msg string, fields ...ports.LogField) {
	r.record("debug", msg, fields)
}

func (r *recordingLogger) Info(_ context.Context, msg string, fields ...ports.LogField) {
	r.record("info", msg, fields)
}

func (r *recordingLogger) Warn(_ context.Context, msg string, fields ...ports.LogField) {
	r.record("warn", msg, fields)
}

func (r *recordingLogger) Error(_ context.Context, msg string, fields ...ports.LogField) {
	r.record("error", msg, fields)
}

// find returns the first entry with the given message.
func (r *recordingLogger) find(msg string) (recordedEntry, bool) {
	for _, entry := range r.entries {
		if entry.Message == msg {
			return entry, true
		}
	}
	return recordedEntry{}, false
}

// withActiveWatches seeds the manager with n cancellable watch instances.
func withActiveWatches(t *testing.T, manager *WatchModeManager, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		id := fmt.Sprintf("watch-%d", i)
		manager.activeWatches[id] = &WatchInstance{
			ID:           id,
			TemplatePath: "seeded.tex",
			StartedAt:    time.Now(),
			Context:      ctx,
			Cancel:       cancel,
		}
	}
}

// TestStopAllWatchModesReportsCountStopped pins the reported total against the
// number of watches actually stopped. The count must be captured before the
// map is cleared; reading len() afterwards reports 0 for every input, which
// no test could catch while the manager depended on a concrete logger.
func TestStopAllWatchModesReportsCountStopped(t *testing.T) {
	recorder := &recordingLogger{}
	manager := NewWatchModeManager(recorder)
	withActiveWatches(t, manager, 3)
	require.Len(t, manager.GetActiveWatches(), 3)

	require.NoError(t, manager.StopAllWatchModes())

	entry, ok := recorder.find("All watch modes stopped")
	require.True(t, ok, "expected a summary log entry; got %+v", recorder.entries)
	assert.Equal(t, 3, entry.Fields["total_stopped"],
		"summary must report the number stopped, not the size of the cleared map")
	assert.Empty(t, manager.GetActiveWatches())
}

// TestNewWatchModeManagerToleratesNilLogger checks that a nil Logger is
// replaced by the null object, so an unconfigured manager logs nothing rather
// than panicking on a nil interface.
func TestNewWatchModeManagerToleratesNilLogger(t *testing.T) {
	manager := NewWatchModeManager(nil)
	withActiveWatches(t, manager, 1)

	assert.NotPanics(t, func() {
		_ = manager.StopAllWatchModes()
	})
}
