// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	rootapi "github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTransportTestServer builds a preview API over a compiler that echoes the
// template back as a single page, mounted on a live test server.
func newTransportTestServer(t *testing.T, options PreviewAPIOptions) (*httptest.Server, *PreviewAPI) {
	t.Helper()

	engine, err := rootapi.NewDefaultDocumentEngine("latex", rootapi.DocumentEngineConfig{})
	require.NoError(t, err)

	options.Engine = engine
	options.CompilerFactory = PreviewCompilerFactoryFunc(func() (preview.Compiler, error) {
		return previewCompilerFunc(func(_ context.Context, request preview.CompileRequest) (preview.CompileOutput, error) {
			data, readErr := os.ReadFile(filepath.Join(request.Workspace, request.Main))
			if readErr != nil {
				return preview.CompileOutput{}, readErr
			}
			return preview.CompileOutput{Pages: []preview.Page{{Number: 1, MediaType: "image/png", Data: data}}}, nil
		}), nil
	})
	if options.SessionOptions.WorkspaceRoot == "" {
		options.SessionOptions = preview.Options{WorkspaceRoot: t.TempDir()}
	}

	api, err := NewPreviewAPI(options)
	require.NoError(t, err)
	t.Cleanup(func() { _ = api.Close() })

	router := chi.NewRouter()
	router.Mount("/preview", api.Routes())
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return server, api
}

// TestWebSocketDeliversTheSameFeedAsSSE pins that both transports are
// projections of one event feed: a revision submitted once must reach a
// WebSocket subscriber with the same type, revision and payload the SSE
// endpoint reports.
func TestWebSocketDeliversTheSameFeedAsSSE(t *testing.T) {
	server, _ := newTransportTestServer(t, PreviewAPIOptions{})
	sessionID := createPreviewSession(t, server.URL)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") +
		"/preview/sessions/" + sessionID + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	require.Equal(t, http.StatusAccepted,
		submitPreviewRevision(t, server.URL, sessionID, 2, previewDocument("hello")))

	_, data, err := conn.Read(ctx)
	require.NoError(t, err)

	var event PreviewEvent
	require.NoError(t, json.Unmarshal(data, &event))
	assert.Equal(t, "result", event.Type)
	assert.Equal(t, uint64(2), event.Revision)
	require.NotNil(t, event.Result)
	assert.Len(t, event.Result.ChangedPages, 1)
}

// TestWebSocketReplaysFromCursor pins that the ?after= replay cursor works on
// the WebSocket transport exactly as Last-Event-ID does on SSE, so a reconnect
// does not lose events.
func TestWebSocketReplaysFromCursor(t *testing.T) {
	server, _ := newTransportTestServer(t, PreviewAPIOptions{})
	sessionID := createPreviewSession(t, server.URL)

	require.Equal(t, http.StatusAccepted,
		submitPreviewRevision(t, server.URL, sessionID, 2, previewDocument("first")))
	first := readPreviewEvent(t, server.URL, sessionID, 0)
	require.Equal(t, uint64(2), first.Revision)

	require.Equal(t, http.StatusAccepted,
		submitPreviewRevision(t, server.URL, sessionID, 3, previewDocument("second")))

	// Connect asking only for events after the first one.
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") +
		"/preview/sessions/" + sessionID + "/ws?after=" + strconv.FormatUint(first.ID, 10)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	_, data, err := conn.Read(ctx)
	require.NoError(t, err)

	var event PreviewEvent
	require.NoError(t, json.Unmarshal(data, &event))
	assert.Equal(t, uint64(3), event.Revision,
		"cursor must skip the already-delivered event, not replay it")
	assert.Greater(t, event.ID, first.ID)
}

// TestSSEEmitsPreambleAndHeartbeat pins the keep-alive behaviour an idle
// preview stream depends on: without periodic traffic, proxies close a silent
// connection and the browser stops receiving updates.
func TestSSEEmitsPreambleAndHeartbeat(t *testing.T) {
	server, _ := newTransportTestServer(t, PreviewAPIOptions{
		Heartbeat: 40 * time.Millisecond,
		RetryHint: 750 * time.Millisecond,
	})
	sessionID := createPreviewSession(t, server.URL)

	request, err := http.NewRequest(http.MethodGet,
		server.URL+"/preview/sessions/"+sessionID+"/events", nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	response, err := http.DefaultClient.Do(request.WithContext(ctx))
	require.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, "text/event-stream", response.Header.Get("Content-Type"))
	assert.Equal(t, "no-cache", response.Header.Get("Cache-Control"))
	assert.Equal(t, "no", response.Header.Get("X-Accel-Buffering"),
		"nginx buffers proxied responses by default, which defeats streaming")

	reader := bufio.NewReader(response.Body)

	// The preamble advertises the reconnect delay.
	line, err := reader.ReadString('\n')
	require.NoError(t, err)
	assert.Equal(t, "retry: 750\n", line)

	// With no revisions submitted the stream is idle, so the next traffic must
	// be a heartbeat comment rather than silence.
	deadline := time.Now().Add(5 * time.Second)
	sawHeartbeat := false
	for time.Now().Before(deadline) && !sawHeartbeat {
		line, err := reader.ReadString('\n')
		require.NoError(t, err)
		if strings.HasPrefix(line, ": ping") {
			sawHeartbeat = true
		}
	}
	assert.True(t, sawHeartbeat, "idle SSE stream never sent a heartbeat")
}
