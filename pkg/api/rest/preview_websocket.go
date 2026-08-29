// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
)

// websocketWriteTimeout bounds a single frame write, so one stalled reader
// cannot pin a session goroutine indefinitely.
const websocketWriteTimeout = 10 * time.Second

// websocketSink writes the event feed as JSON text frames.
type websocketSink struct {
	conn *websocket.Conn
}

// Send writes one event as a JSON text frame under a write deadline.
func (sink *websocketSink) Send(ctx context.Context, event PreviewEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	writeCtx, cancel := context.WithTimeout(ctx, websocketWriteTimeout)
	defer cancel()
	return sink.conn.Write(writeCtx, websocket.MessageText, data)
}

// Heartbeat sends a protocol-level ping and waits for the pong, which detects a
// peer that has gone away without closing.
func (sink *websocketSink) Heartbeat(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, websocketWriteTimeout)
	defer cancel()
	return sink.conn.Ping(pingCtx)
}

// Close sends a close frame naming why the stream ended.
func (sink *websocketSink) Close(_ context.Context, err error) error {
	status, reason := websocket.StatusNormalClosure, ""
	if err != nil && !errors.Is(err, context.Canceled) {
		status, reason = websocket.StatusInternalError, err.Error()
		if len(reason) > 100 { // close reasons are capped at 123 bytes
			reason = reason[:100]
		}
	}
	return sink.conn.Close(status, reason)
}

// streamWebSocket upgrades the request and drives the same event feed the SSE
// endpoint serves, honouring the same ?after= replay cursor.
func (api *PreviewAPI) streamWebSocket(writer http.ResponseWriter, request *http.Request) {
	transport, exists := api.lookup(chi.URLParam(request, "sessionID"))
	if !exists {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "preview session not found"})
		return
	}

	conn, err := websocket.Accept(writer, request, &websocket.AcceptOptions{
		OriginPatterns: api.originPatterns,
	})
	if err != nil {
		return // Accept has already written the failure
	}

	sink := &websocketSink{conn: conn}
	ctx := request.Context()

	// Drain incoming frames so pings and close frames are processed; the
	// preview protocol is server-to-client, so client payloads are ignored.
	go func() {
		for {
			if _, _, err := conn.Read(ctx); err != nil {
				return
			}
		}
	}()

	pumpErr := pumpEvents(ctx, transport, sink, streamOptions{
		after:     eventCursor(request),
		heartbeat: api.heartbeat,
	})
	_ = sink.Close(ctx, pumpErr)
}
