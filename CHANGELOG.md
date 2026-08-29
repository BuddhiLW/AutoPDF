# Changelog

Notable AutoPDF changes are documented here. Versions follow Semantic
Versioning.

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

[1.5.0]: https://github.com/BuddhiLW/AutoPDF/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/BuddhiLW/AutoPDF/compare/v1.3.3...v1.4.0
