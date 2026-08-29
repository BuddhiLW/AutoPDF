---
name: autopdf-preview
description: >
  Build a live LaTeX preview on AutoPDF — component documents (api.DocumentEngine,
  document.DocumentSpec), warm preview sessions, and the streaming transports that
  push changed pages to a browser over server-sent events, WebSocket or HTTP/2. Use
  when building an editor, playground or authoring UI that re-renders as the user
  types; when someone asks about AutoPDF preview sessions, SSE or WebSocket preview
  events, replay cursors, rest.NewPreviewAPI, rest.NewServer, h2c, or preview
  backpressure and 429s.
---

# Live preview with AutoPDF

Two layers. `api.DocumentEngine` turns a structured document into LaTeX and
keeps compilation warm; `pkg/api/rest` exposes that to a browser.

```go
import (
    "github.com/BuddhiLW/AutoPDF/v2/pkg/api"
    "github.com/BuddhiLW/AutoPDF/v2/pkg/api/rest"
    "github.com/BuddhiLW/AutoPDF/v2/pkg/document"
    "github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
)
```

## Documents are structured, not strings

```go
spec := document.DocumentSpec{
    SchemaVersion: document.CurrentSchemaVersion,
    ID:            "report",
    Blocks: []document.Component{{
        ID:      "intro",
        Kind:    "text",
        Variant: "default",
        Props:   document.Props{"content": "Hello."},
    }},
}
```

Built-in kinds are `text`, `section` and `artifact`, each with a `default`
variant. A `DocumentSpec` is immutable and validated: `engine.Validate(spec)`
returns `document.Problems`, and composing an invalid spec fails rather than
emitting broken LaTeX.

```go
engine, err := api.NewDefaultDocumentEngine("latex", api.DocumentEngineConfig{
    MaxWorkers: 4, // bounded fragment rendering
})
```

## Preview sessions

A session keeps TeX auxiliary state warm and supersedes work that a newer
revision has obsoleted.

```go
session, err := engine.NewPreviewSession(compiler, preview.Options{
    WorkspaceRoot: "/var/tmp/previews",
    IdleTTL:       10 * time.Minute,
})

results := engine.Preview(ctx, session, spec, api.ProjectionOptions{
    FocusSections: []string{"intro"}, // \includeonly-style focused build
})
result := <-results
```

`compiler` implements `preview.Compiler` — one method,
`Compile(context.Context, CompileRequest) (CompileOutput, error)` — so a test
substitutes a fake and never runs LaTeX.

`preview.Result` carries `ChangedPages` and `RemovedPages` rather than a whole
PDF, which is what makes repaint cheap. `Err` is `preview.ErrSuperseded` when a
newer revision cancelled this one; that is normal and must **not** repaint the
client.

## HTTP transport

```go
previewAPI, err := rest.NewPreviewAPI(rest.PreviewAPIOptions{
    Engine:          engine,
    CompilerFactory: rest.PreviewCompilerFactoryFunc(newCompiler),
})

router := chi.NewRouter()
rest.RegisterPreviewRoutes(router, previewAPI) // mounts /api/v1/preview
```

| Method | Route | Purpose |
| --- | --- | --- |
| `POST` | `/sessions` | Create a session |
| `PUT` | `/sessions/{id}/revisions/{n}` | Submit revision `n` |
| `GET` | `/sessions/{id}/events` | SSE stream |
| `GET` | `/sessions/{id}/ws` | WebSocket stream |
| `DELETE` | `/sessions/{id}` | Close |

Revisions must strictly increase; a stale one gets `409`. Both streams are
projections of one event feed and share a replay cursor, so a client picks
whichever transport suits it without changing semantics.

### SSE

```js
const events = new EventSource(`/api/v1/preview/sessions/${id}/events`);
events.onmessage = (e) => applyPatch(JSON.parse(e.data));
```

The browser sends `Last-Event-ID` automatically on reconnect, so no events are
lost across a dropped connection. The stream emits heartbeat comments while
idle and advertises a reconnect delay.

### WebSocket

```js
const socket = new WebSocket(`wss://host/api/v1/preview/sessions/${id}/ws?after=${lastID}`);
socket.onmessage = (e) => applyPatch(JSON.parse(e.data));
```

Same cursor, passed as `?after=`. Only same-origin handshakes are accepted
unless `OriginPatterns` is set.

## Serving it

```go
server, err := rest.NewServer(rest.ServerOptions{
    Addr:    ":8080",
    Handler: router,
})
server.ListenAndServe()
```

Speaks HTTP/2 over TLS and over cleartext (h2c), the latter mattering behind a
TLS-terminating proxy that should keep the back-end hop multiplexed.

**Do not set `WriteTimeout`.** It is a deadline on the whole response, so any
non-zero value silently truncates every SSE and WebSocket stream mid-session,
indistinguishable to the client from a network fault. `rest.NewServer` defaults
it and `ReadTimeout` to zero on purpose and carries slow-client protection in
`ReadHeaderTimeout`, which stops applying once headers are in. This is the most
common way to break a working streaming setup.

## Backpressure

Revision submission is client-driven — an editor can emit one per keystroke — so
each session has a bounded queue drained by one publisher goroutine. A full
queue answers `429 Too Many Requests`.

**A client must handle 429 by backing off, not retrying immediately.** Debounce
keystrokes (150–300ms is typical) rather than submitting per character.

```go
rest.NewPreviewAPI(rest.PreviewAPIOptions{
    Engine:             engine,
    CompilerFactory:    factory,
    Heartbeat:          15 * time.Second,
    RetryHint:          2 * time.Second,
    HistoryLimit:       32,   // events retained for replay
    RevisionQueueDepth: 64,
    OriginPatterns:     []string{"editor.example.com"},
})
```

`HistoryLimit` bounds replay: a client offline longer than that many events
cannot resume exactly and must resynchronise from a fresh session.

Call `previewAPI.Close()` on shutdown to terminate sessions and their
workspaces.

## Client reducer

Events arrive as `{id, type, revision, result}` with `type` one of `result`,
`error` or `closed`. Apply them as patches:

- `result.changedPages` — replace the page with that `number`
- `result.removedPages` — delete those page numbers
- track the highest `id` seen and pass it as the resume cursor

Reference: `docs/streaming-transports.md` in the AutoPDF repository.
