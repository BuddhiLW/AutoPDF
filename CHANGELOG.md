# Changelog

Notable AutoPDF changes are documented here. Versions follow Semantic
Versioning.

## [2.2.0] - 2026-09-16

### Changed

- **A template that reads a variable it is not given now REFUSES to render.**
  This is a behaviour change and it is the point of the release. Previously a
  missing key rendered as empty and the document compiled anyway, which is
  convenient for a report and is the worst possible failure for a contract: the
  document compiles, looks right, is signed, and a clause that was supposed to
  be there simply is not. Six example fixtures under `test/` have been emitting
  `\title{<no value>}` and silently empty `range` blocks for some time, and
  nothing flagged it.

  A site that is deliberately allowed to be blank says so, at the point of use,
  with a reason:

  ```
  delim[[ optional .field "reason the blank is acceptable" ]]
  ```

  The reason is mandatory. The pipe spelling `.field | optional "why"` is
  rejected as malformed rather than accepted, because a pipe reverses the
  argument order and would print the reason where the value belongs.

  Waivers live in the template, not in the YAML, on purpose: the YAML is case
  data, generated per document, and a waiver list there could be injected by
  the same path that produced the blank. The template changes rarely and under
  review.

  Absent, `nil`, and empty-or-blank strings are all refused, and are reported as
  distinct causes. `0`, `0.0`, `false`, and empty slices and maps are stated
  facts and render normally.

- **Build failures now carry their cause and name their phase.** Every failure
  used to flatten to `failed to build config`, whatever had actually gone
  wrong. Errors now wrap the cause, so `errors.Is` against the package
  sentinels still works, `errors.As` reaches `*strict.MissingVariablesError`,
  and the operator gets every offending key with `file:line:column` instead of
  one sentence naming the wrong phase.

### Added

- **`pkg/template/strict`**, the pure core: `Policy` as a value, `Reference`,
  `Bindings`, `Verdict`, `Offence`, `*MissingVariablesError`, and the `Auditor`
  port with `StrictAuditor` implementing it. No I/O, no compilation, no
  logging. Renderers depend on the port, so a caller can inject its own policy
  and tests inject stubs instead of reaching for global state.
- **`optional`**, a template builtin taking a field and a mandatory reason.
- **`result.BuildFailure`**, the pure mapping from a cause to a phase-named
  wrapped error.
- **`ports.StagedTemplateSuffix`**, the suffix the rendered intermediate is
  written under.

### Fixed

- **A successful build could overwrite and then delete the source template.**
  The staged intermediate was written into the template's own directory under a
  job name derived from the OUTPUT file, so `template: "contrato.tex"` with
  `output: "contrato.pdf"` staged onto `contrato.tex` itself and removed it on
  cleanup. The PDF came out correct, exit code 0, no warning, and the source was
  gone. The repository's own fixtures never collided, because they name the
  output differently from the template; the destructive case was the natural
  one, naming the PDF after the document. The staged file now carries
  `.autopdf.tex` and cannot collide. Both the `pdflatex` and the `latexmk`
  adapters were affected.
- The scaffolded default configuration omitted `content`, which the shipped
  sample template and `configs/sample-config.yaml` both use.

### Known limitation

The static audit resolves reads against the root data. Inside a `range` or
`with` body the dot is rebound and the key is relative to a value the analysis
cannot know, so those sites fall through to `missingkey=error`, which fails
closed on the first miss and cannot be waived with `optional`. Documented in
`pkg/template/strict/doc.go` rather than papered over.

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
