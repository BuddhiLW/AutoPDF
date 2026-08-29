// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DefaultStreamHeartbeat is how often an idle subscriber is pinged. Proxies and
// load balancers commonly drop a connection that has been silent for 30-60s,
// and a preview session is idle whenever the author stops typing.
const DefaultStreamHeartbeat = 15 * time.Second

// DefaultStreamRetry is the reconnect delay advertised to SSE clients.
const DefaultStreamRetry = 2 * time.Second

// EventSink is one subscriber's wire format. It is the seam that keeps the
// event feed independent of the transport carrying it: SSE and WebSocket are
// two implementations, and a third costs no change to the feed.
//
// The set of transports is open, so this is an interface rather than a switch
// over a transport enum.
type EventSink interface {
	// Send delivers one event. A returned error ends the subscription.
	Send(ctx context.Context, event PreviewEvent) error
	// Heartbeat signals liveness on an idle stream. Implementations for which
	// the protocol already keeps itself alive may return nil without writing.
	Heartbeat(ctx context.Context) error
	// Close releases the subscriber, reporting why the stream ended.
	Close(ctx context.Context, err error) error
}

// streamOptions tunes one subscription.
type streamOptions struct {
	after     uint64
	heartbeat time.Duration
}

// pumpEvents feeds one subscriber from the cursor `after` until the session
// closes, the sink errors, or ctx ends.
//
// This is the single definition of how a subscriber is driven. Every transport
// projects from here, so replay semantics, ordering and close behaviour cannot
// drift between them.
func pumpEvents(
	ctx context.Context,
	session *previewTransportSession,
	sink EventSink,
	options streamOptions,
) error {
	heartbeat := options.heartbeat
	if heartbeat <= 0 {
		heartbeat = DefaultStreamHeartbeat
	}
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()

	after := options.after
	for {
		events, wait, closed := session.eventsAfter(after)
		for _, event := range events {
			if err := sink.Send(ctx, event); err != nil {
				return err
			}
			after = event.ID
		}
		if closed {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-wait:
			// New events, or the session closed; loop and drain.
		case <-ticker.C:
			if err := sink.Heartbeat(ctx); err != nil {
				return err
			}
		}
	}
}

// eventCursor reads the replay position a client is resuming from. SSE clients
// send Last-Event-ID automatically on reconnect; other transports pass ?after=.
func eventCursor(request *http.Request) uint64 {
	value := request.Header.Get("Last-Event-ID")
	if value == "" {
		value = request.URL.Query().Get("after")
	}
	cursor, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return cursor
}
