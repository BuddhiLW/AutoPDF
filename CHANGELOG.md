# Changelog

Notable AutoPDF changes are documented here. Versions follow Semantic
Versioning.

## [2.1.0] - 2026-08-29

### Added

- **A `beamer` render target**, compiling a `DocumentSpec` into a Beamer
  presentation. `pkg/render/beamer` registers definitions for `section`,
  `text`, `span`, `link`, `heading`, `bullets`, `code`, `quote`, `table`,
  `image`, `columns`, `cards`, `card`, `callout`, `kicker`,
  `media-placeholder`, `notes`, `rule` and `scene`, and projects them to one
  `\include`d file per frame.
- **`api.ManifestProjector`**, the seam a render target implements to choose
  its document shape. A target supplies its own catalog and projector; `pkg/api`
  names none of them, which `test/architecture/render_target_openness_test.go`
  enforces. `DocumentEngineConfig.Projector` selects one, defaulting to the
  existing article-shaped LaTeX projection.
- **`autopdf deck`**, building a PDF presentation from a DocumentSpec, and
  **`autopdf deck watch`**, rebuilding on every save with a debounce and an
  optional regeneration command. Paired with
  [plato](https://github.com/BuddhiLW/plato), one Markdown or Org source
  produces both a Reveal deck and a Beamer PDF — see
  [docs/plato-integration.md](docs/plato-integration.md).
- `beamer.Options.Files` carries auxiliary files, such as a generated `.sty`,
  into the private compile workspace.
- An `autopdf-plato` skill, and `test/plato_integration`, a shared fixture
  corpus whose gate reports drift between the two document IRs.

### Fixed

- Unprintable graphics no longer abort a build. pdflatex treats an unknown
  graphics extension as fatal, so a single GIF or SVG produced no PDF at all;
  they now render as a visible placeholder, as video, audio and embeds already
  did.

## [2.0.0] - 2026-08-29

### Removed

- `rest.SimpleStructConverterAPI` and its constructors, route registrars, and
  nine `Simple*` request/response types. They reimplemented every handler and
  type of `StructConverterAPI`, differing only in transport. Use
  `StructConverterAPI`.

### Changed

- **Module path is now `github.com/BuddhiLW/AutoPDF/v2`.** Update imports and
  `go install github.com/BuddhiLW/AutoPDF/v2/cmd/autopdf@latest`.
- `NewPDFGenerationApplicationService`, `NewPDFOrchestrationService`, and
  `NewWatchModeManager` take the `ports.Logger` port instead of the concrete
  `*logger.LoggerAdapter`. Bridge an existing adapter with
  `infrastructure/adapters.NewLoggerPortAdapter`.
- `adapters.NoOpLogger` moved to `ports.NoOpLogger`, so any layer can
  null-object a logger without importing infrastructure.

### Added

- WebSocket preview transport at `GET /sessions/{id}/ws`, sharing the replay
  cursor, history ring, and ordering of the existing SSE endpoint.
- `rest.NewServer`, an `http.Server` speaking HTTP/2 over TLS and cleartext
  (h2c) with bounded concurrent streams and timeouts that do not truncate
  long-lived streams.
- SSE keep-alive: periodic heartbeat comments, a `retry:` reconnect hint, and
  `X-Accel-Buffering: no`.
- `compilation.EngineStrategy`, the first implementation of
  `parallel.CompilationStrategy`, adapting `api.Generator`. Parallel
  compilation previously had no strategy and reported every template as
  unhandled.
- `PreviewAPIOptions.Heartbeat`, `RetryHint`, `OriginPatterns`, and
  `RevisionQueueDepth`.
- Documentation: [Streaming transports](docs/streaming-transports.md).

### Fixed

- Bound goroutine **creation** in the parallel compilation orchestrator. It
  spawned one goroutine per task and acquired a semaphore inside it, which
  bounds execution only: 300 tasks produced 301 goroutines. A worker pool now
  holds the delta to the configured concurrency.
- Bound preview revision publication to one goroutine per session. Revision
  submission is client-driven, so the previous goroutine-per-revision shape let
  a client choose the goroutine count. A saturated queue now answers `429`.
- Preview event ordering is deterministic. Concurrent publishers could assign
  event IDs in completion order rather than revision order, which the replay
  cursor then served to clients.
- The parallel orchestrator now honours context cancellation instead of
  dispatching the whole backlog.
- `ConfigureTimeout` set a field nothing read, so a task with no timeout of its
  own reached `context.WithTimeout(ctx, 0)` — an already-expired context — and
  failed instantly.
- Logger context key had three declarations of two different types. Because
  `context.Value` compares keys by dynamic type, both package-private readers
  always missed and silently fell back to a default logger, discarding the
  caller's configured verbosity and sink.
- `StopAllWatchModes` reported `total_stopped` from the already-cleared map, so
  the summary always logged zero.

## [1.5.0] - 2026-08-29

### Changed

- Raise the supported Go 1.25 toolchain from `1.25.0` to `1.25.14`.
- Update fsnotify, chi, Bonzai, testify, zap, and their active transitive
  dependencies to current compatible releases.
- Prune stale indirect modules and checksums from the module graph.

### Security

- Build releases with the patched Go 1.25.14 standard library.
- Verify the resolved graph has no reachable vulnerabilities with
  `govulncheck`.

## [1.4.0] - 2026-08-29

### Added

- Immutable, versioned component documents with strict validation.
- Frozen component catalogs and deterministic LaTeX projections for `flow`,
  `section`, and `artifact` composition modes.
- Content-addressed fragment caching and bounded parallel rendering.
- Production projection generation through the existing `api.Engine` and
  `api.Result` boundary.
- Warm LaTeX preview sessions with cancellation, focused section builds,
  component-aware diagnostics, and selective page rasterization.
- Browser preview transport with monotonic revisions, replayable SSE events,
  changed-page payloads, and removed-page notifications.
- Preview latency budgets, regression gates, and stage-level timings.

### Compatibility

- Existing CLI, `api.Engine.Generate`, `api.Request`, and `api.Result` contracts
  remain unchanged.
- Component composition and preview APIs are additive.

[2.1.0]: https://github.com/BuddhiLW/AutoPDF/compare/v2.0.0...v2.1.0
[2.0.0]: https://github.com/BuddhiLW/AutoPDF/compare/v1.5.0...v2.0.0
[1.5.0]: https://github.com/BuddhiLW/AutoPDF/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/BuddhiLW/AutoPDF/compare/v1.3.3...v1.4.0
