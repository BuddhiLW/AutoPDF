# Component composition and fast previews

AutoPDF now models a document as an immutable component tree. A component is
promoted through an exact `(kind, variant, target)` catalog definition, rendered
to a content-addressed fragment, and assembled into a deterministic render
manifest. The manifest records child/dependency edges, source assets, hashes,
and editor source locations.

The LaTeX target maps layout coupling—not user-authored commands—to files:

- `flow` becomes a `\input` fragment.
- top-level `section` becomes an `\include` unit and may participate in
  `\includeonly` focused builds.
- `artifact` embeds a separately resolved PDF or image.

`\input` by itself does not make TeX incrementally compile only that file. Fast
feedback comes from the whole pipeline: component hash caching avoids rerendering
unchanged fragments, bounded workers render independent components concurrently,
only changed workspace files are written, the preview session retains TeX aux
files, `\includeonly` skips unrelated section units, superseded builds are
cancelled, and only changed page payloads are returned to the UI.

Production PDF generation remains behind the existing stable `api.Result`
boundary. Preview is deliberately a separate, revisioned lifecycle so stale
output can never replace a newer browser revision.

## Public flow

1. Create immutable component definitions and freeze a `component.Catalog`, or
   use `api.NewDefaultDocumentEngine("latex", config)`.
2. Call `DocumentEngine.Project` for a pure manifest/projection, or
   `DocumentEngine.Generate` with a production `ProjectionGenerator`.
3. For interactive use, retain one `preview.Session` per editor document and
   submit edits through `DocumentEngine.Preview`.
4. Give mutable external assets a new URI or `AssetRef.Digest` whenever their
   bytes change; this intentionally invalidates the relevant fragment/source.

The compiler and asset resolver are explicit effect ports. This keeps component
renderers deterministic and makes caches safe to share across requests.

## Concrete adapters

- `api.NewProjectionGenerator` materializes a projection in an isolated
  production workspace, resolves versioned assets, and calls the canonical
  `Engine.Generate` pipeline while preserving `api.Result`.
- `preview.NewLaTeXCompiler` retains TeX auxiliary state in each session,
  terminates superseded compiler processes through context cancellation, and
  performs full-resolution rasterization only for changed page fingerprints.
- `rest.NewPreviewAPI` exposes monotonic revisions and replayable SSE events.
  A browser reducer replaces `changedPages`, deletes `removedPages`, and ignores
  older revisions.

Layered budgets and the browser paint hook are documented in
[Preview performance](preview-performance.md).
