// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"sync"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newQueueOnlySession(depth int) *previewTransportSession {
	return &previewTransportSession{
		historyLimit: 8,
		notify:       make(chan struct{}),
		pending:      make(chan pendingRevision, depth),
	}
}

func revisionEntry(revision uint64) pendingRevision {
	return pendingRevision{revision: revision, results: make(chan preview.Result, 1)}
}

// TestEnqueueBoundsPendingRevisions pins the backpressure bound itself.
//
// Revision submission is client-driven — an editor can emit one per keystroke —
// so the queue must refuse work past its depth rather than grow. No publisher
// drains here, which is the worst case: a compilation that has not returned yet.
func TestEnqueueBoundsPendingRevisions(t *testing.T) {
	session := newQueueOnlySession(2)

	assert.True(t, session.enqueue(revisionEntry(2)))
	assert.True(t, session.enqueue(revisionEntry(3)))
	assert.False(t, session.enqueue(revisionEntry(4)),
		"a full queue must refuse, so the handler can answer 429 instead of growing without bound")

	// Draining one slot admits exactly one more.
	<-session.pending
	assert.True(t, session.enqueue(revisionEntry(5)))
	assert.False(t, session.enqueue(revisionEntry(6)))
}

// TestEnqueueAfterCloseIsRefusedNotPanicking pins the close race. The session
// publisher is ended by closing the pending channel when the preview session
// finishes; a submission arriving concurrently would panic on send-to-closed
// unless both paths agree under the mutex.
func TestEnqueueAfterCloseIsRefusedNotPanicking(t *testing.T) {
	session := newQueueOnlySession(4)
	session.closePending()

	assert.NotPanics(t, func() {
		assert.False(t, session.enqueue(revisionEntry(2)),
			"a closed session must refuse new revisions")
	})
}

// TestClosePendingIsIdempotent keeps a second close (Close racing session
// expiry) from panicking on an already-closed channel.
func TestClosePendingIsIdempotent(t *testing.T) {
	session := newQueueOnlySession(1)
	assert.NotPanics(t, func() {
		session.closePending()
		session.closePending()
	})
}

// TestEnqueueIsSafeUnderConcurrentClose drives the actual race: many
// submissions against a close, which is what the mutex in enqueue/closePending
// exists to make safe. Run under -race this fails loudly if the guard is
// removed.
func TestEnqueueIsSafeUnderConcurrentClose(t *testing.T) {
	session := newQueueOnlySession(8)

	// Drain continuously so the queue does not simply saturate.
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		for range session.pending {
		}
	}()

	var writers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		writers.Add(1)
		go func(base uint64) {
			defer writers.Done()
			for revision := base; revision < base+50; revision++ {
				session.enqueue(revisionEntry(revision))
			}
		}(uint64(worker) * 100)
	}

	session.closePending()
	writers.Wait()

	<-drained
	require.True(t, session.pendingClosed)
}
