// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"crypto/tls"
	"errors"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// ErrServerHandlerRequired reports a server built without a handler.
var ErrServerHandlerRequired = errors.New("rest: server handler is required")

const (
	// DefaultReadHeaderTimeout bounds how long a client may take to send its
	// request headers. This is the slowloris guard, and unlike WriteTimeout it
	// is safe for streaming endpoints because it stops applying once the
	// headers are in.
	DefaultReadHeaderTimeout = 10 * time.Second
	// DefaultIdleTimeout closes keep-alive connections that go silent.
	DefaultIdleTimeout = 120 * time.Second
	// DefaultMaxConcurrentStreams caps in-flight HTTP/2 streams per connection.
	// A single connection multiplexing unbounded streams is how one client
	// exhausts server concurrency.
	DefaultMaxConcurrentStreams = 250
)

// ServerOptions configures an HTTP server suited to this API's long-lived
// streaming endpoints.
type ServerOptions struct {
	Addr    string
	Handler http.Handler

	// ReadHeaderTimeout bounds request-header reads. Defaults to
	// DefaultReadHeaderTimeout.
	ReadHeaderTimeout time.Duration
	// IdleTimeout bounds idle keep-alive connections. Defaults to
	// DefaultIdleTimeout.
	IdleTimeout time.Duration
	// WriteTimeout is a deadline on the WHOLE response, so any non-zero value
	// truncates server-sent events and WebSocket streams once it elapses.
	// Zero, the default, is correct for this API; set it only on a server that
	// mounts no streaming route.
	WriteTimeout time.Duration
	// ReadTimeout is likewise a deadline on the whole request body. Zero, the
	// default, keeps WebSocket connections readable for their lifetime.
	ReadTimeout time.Duration

	// MaxConcurrentStreams caps concurrent HTTP/2 streams per connection.
	// Defaults to DefaultMaxConcurrentStreams.
	MaxConcurrentStreams uint32

	// TLSConfig enables HTTPS. When nil the server still speaks HTTP/2 over
	// cleartext (h2c), which is what a TLS-terminating proxy in front of it
	// needs in order to keep using HTTP/2 on the back end.
	TLSConfig *tls.Config
}

// NewServer builds an http.Server that speaks HTTP/2 over both TLS and
// cleartext, with timeouts chosen so that streaming endpoints survive.
//
// Serve it with ListenAndServeTLS when TLSConfig is set, ListenAndServe
// otherwise.
func NewServer(options ServerOptions) (*http.Server, error) {
	if options.Handler == nil {
		return nil, ErrServerHandlerRequired
	}
	if options.ReadHeaderTimeout <= 0 {
		options.ReadHeaderTimeout = DefaultReadHeaderTimeout
	}
	if options.IdleTimeout <= 0 {
		options.IdleTimeout = DefaultIdleTimeout
	}
	if options.MaxConcurrentStreams == 0 {
		options.MaxConcurrentStreams = DefaultMaxConcurrentStreams
	}

	http2Server := &http2.Server{
		MaxConcurrentStreams: options.MaxConcurrentStreams,
		IdleTimeout:          options.IdleTimeout,
	}

	server := &http.Server{
		Addr: options.Addr,
		// h2c upgrades cleartext connections that ask for HTTP/2 and passes
		// everything else through as HTTP/1.1.
		Handler:           h2c.NewHandler(options.Handler, http2Server),
		ReadHeaderTimeout: options.ReadHeaderTimeout,
		ReadTimeout:       options.ReadTimeout,
		WriteTimeout:      options.WriteTimeout,
		IdleTimeout:       options.IdleTimeout,
		TLSConfig:         options.TLSConfig,
	}

	// Registers the h2 ALPN protocol so TLS connections negotiate HTTP/2.
	if err := http2.ConfigureServer(server, http2Server); err != nil {
		return nil, err
	}

	return server, nil
}
