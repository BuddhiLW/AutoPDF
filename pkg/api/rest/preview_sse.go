// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrStreamingUnsupported reports a ResponseWriter that cannot flush, so
// server-sent events cannot be delivered over it.
var ErrStreamingUnsupported = errors.New("rest: streaming unsupported by this ResponseWriter")

// sseSink writes the event feed as a text/event-stream.
type sseSink struct {
	writer  http.ResponseWriter
	flusher http.Flusher
}

// newSSESink prepares the response for streaming and emits the preamble.
func newSSESink(writer http.ResponseWriter, retry time.Duration) (*sseSink, error) {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return nil, ErrStreamingUnsupported
	}

	header := writer.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	// Nginx buffers proxied responses by default, which holds events until the
	// buffer fills and defeats streaming entirely.
	header.Set("X-Accel-Buffering", "no")
	writer.WriteHeader(http.StatusOK)

	if retry > 0 {
		fmt.Fprintf(writer, "retry: %d\n\n", retry.Milliseconds())
	}
	flusher.Flush()

	return &sseSink{writer: writer, flusher: flusher}, nil
}

// Send writes one event frame and flushes it.
func (sink *sseSink) Send(_ context.Context, event PreviewEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(sink.writer, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, event.Type, data); err != nil {
		return err
	}
	sink.flusher.Flush()
	return nil
}

// Heartbeat writes an SSE comment, which clients ignore and proxies count as
// traffic.
func (sink *sseSink) Heartbeat(_ context.Context) error {
	if _, err := fmt.Fprint(sink.writer, ": ping\n\n"); err != nil {
		return err
	}
	sink.flusher.Flush()
	return nil
}

// Close is a no-op: ending the handler closes the response.
func (sink *sseSink) Close(_ context.Context, _ error) error { return nil }
