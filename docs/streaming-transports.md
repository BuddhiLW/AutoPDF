# Streaming transports: SSE, WebSocket, HTTP/2

The preview adapter (`pkg/api/rest`) publishes one event feed per preview
session. Transports are projections of that feed rather than separate
implementations, so replay, ordering and close semantics cannot drift between
them.

## The feed

Every session keeps a bounded history ring of `PreviewEvent` values, each with a
monotonically increasing `ID`. A subscriber names the last event it saw and
receives everything after it. `pumpEvents` is the single definition of how a
subscriber is driven; a transport supplies only an `EventSink`:

```go
type EventSink interface {
    Send(ctx context.Context, event PreviewEvent) error
    Heartbeat(ctx context.Context) error
    Close(ctx context.Context, err error) error
}
```

Adding a transport means implementing that interface. It does not mean touching
the feed.

## Routes

| Route | Transport | Replay cursor |
| --- | --- | --- |
| `GET /sessions/{id}/events` | `text/event-stream` | `Last-Event-ID` header, or `?after=` |
| `GET /sessions/{id}/ws` | WebSocket, JSON text frames | `?after=` |

Both are mounted by `RegisterPreviewRoutes` under `/api/v1/preview`.

### Server-sent events

The stream opens with a `retry:` hint telling the browser how long to wait
before reconnecting, and emits `: ping` comments while idle. Responses carry
`X-Accel-Buffering: no`, without which nginx buffers a proxied response and
holds events until its buffer fills.

On reconnect the browser sends `Last-Event-ID` automatically, so no events are
lost across a dropped connection.

### WebSocket

Frames are JSON-encoded `PreviewEvent` values. Liveness uses protocol-level
pings, which detect a peer that vanished without closing. Writes carry a
deadline so one stalled reader cannot pin a session goroutine.

By default only same-origin handshakes are accepted. Set `OriginPatterns` to
allow others:

```go
rest.NewPreviewAPI(rest.PreviewAPIOptions{
    Engine:          engine,
    CompilerFactory: factory,
    OriginPatterns:  []string{"editor.example.com"},
})
```

## HTTP/2

`rest.NewServer` returns an `http.Server` that speaks HTTP/2 over TLS and over
cleartext (h2c). The cleartext path matters behind a TLS-terminating proxy,
where the back-end hop should stay multiplexed.

```go
server, err := rest.NewServer(rest.ServerOptions{
    Addr:                 ":8080",
    Handler:              router,
    MaxConcurrentStreams: 250,
})
```

### Timeouts

`WriteTimeout` and `ReadTimeout` default to zero, deliberately. Both are
deadlines on the **whole** message, so any non-zero value truncates every SSE
and WebSocket stream once it elapses — the stream simply dies mid-session with
no error the client can distinguish from a network fault.

Slow-client protection comes from `ReadHeaderTimeout` instead, which stops
applying once the request headers are in and is therefore safe for streaming.
Set `WriteTimeout` only on a server that mounts no streaming route.

## Backpressure

Revision submission is client-driven — an editor can emit one per keystroke — so
each session has a bounded queue drained by exactly one publisher goroutine.
Goroutine count tracks sessions, not revisions.

When the queue is full, `PUT /sessions/{id}/revisions/{n}` answers
`429 Too Many Requests` rather than growing without bound or blocking the
handler. Depth is configurable:

```go
rest.NewPreviewAPI(rest.PreviewAPIOptions{
    Engine:             engine,
    CompilerFactory:    factory,
    RevisionQueueDepth: 64, // default
})
```

Serial publication also makes event ordering deterministic: concurrent
publishers would otherwise assign event IDs in completion order rather than
revision order, which the replay cursor then hands to clients as-is.

## Tuning

| Option | Default | Purpose |
| --- | --- | --- |
| `Heartbeat` | 15s | Idle liveness signal on both transports |
| `RetryHint` | 2s | Reconnect delay advertised to SSE clients |
| `HistoryLimit` | 32 | Events retained for replay |
| `RevisionQueueDepth` | 64 | Revisions awaiting publication per session |
| `OriginPatterns` | same-origin | WebSocket handshake origins |
