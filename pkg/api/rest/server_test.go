// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

// TestNewServerServesCleartextHTTP2 pins h2c support: a client that speaks
// prior-knowledge HTTP/2 over a plain TCP connection must be served as HTTP/2.
// This is the shape used behind a TLS-terminating proxy, where the back-end hop
// is cleartext but should stay multiplexed.
func TestNewServerServesCleartextHTTP2(t *testing.T) {
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, request.Proto)
	})

	server, err := NewServer(ServerOptions{Handler: handler})
	require.NoError(t, err)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	// An HTTP/2 client that skips TLS and the upgrade dance entirely.
	client := &http.Client{Transport: &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	}}

	response, err := client.Get("http://" + listener.Addr().String() + "/")
	require.NoError(t, err)
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/2.0", string(body), "server did not serve cleartext HTTP/2")
	assert.Equal(t, 2, response.ProtoMajor)
}

// TestNewServerDefaultsKeepStreamingAlive pins the timeout defaults. A non-zero
// WriteTimeout is a deadline on the whole response, so it silently truncates
// every SSE and WebSocket stream once it elapses; the guard against slow
// clients has to be ReadHeaderTimeout, which stops applying once headers land.
func TestNewServerDefaultsKeepStreamingAlive(t *testing.T) {
	server, err := NewServer(ServerOptions{Handler: http.NotFoundHandler()})
	require.NoError(t, err)

	assert.Zero(t, server.WriteTimeout,
		"a non-zero WriteTimeout truncates long-lived streams")
	assert.Zero(t, server.ReadTimeout,
		"a non-zero ReadTimeout closes WebSocket connections mid-session")
	assert.Equal(t, DefaultReadHeaderTimeout, server.ReadHeaderTimeout,
		"slowloris protection must still be on")
	assert.Equal(t, DefaultIdleTimeout, server.IdleTimeout)
}

// TestNewServerRequiresHandler keeps the constructor from producing a server
// that would 404 everything.
func TestNewServerRequiresHandler(t *testing.T) {
	_, err := NewServer(ServerOptions{})
	assert.ErrorIs(t, err, ErrServerHandlerRequired)
}
