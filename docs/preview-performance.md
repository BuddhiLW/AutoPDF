# Preview performance contract

AutoPDF measures preview latency by stage because one aggregate number hides
which feedback loop regressed. `preview.Result.Timings` reports source
materialization, TeX, page fingerprinting, full-resolution rasterization, and
total server work. The SSE transport sends changed and removed pages only; it
never sends the full PDF payload.

## Budgets

Deterministic, machine-local layers are CI gates:

| Layer | Representative input | Budget |
| --- | --- | ---: |
| warm preview orchestration | unchanged projection, 200 revisions | < 10 ms average |
| cold source materialization | 101 generated files | < 250 ms |
| warm source materialization | same 101 files | < 25 ms |

Environment-sensitive layers are benchmark reports, not pass/fail gates:

| Layer | Warm target | Focused-section target |
| --- | ---: | ---: |
| TeX pass | <= 750 ms | <= 500 ms |
| page fingerprint | <= 150 ms | <= 100 ms |
| changed-page raster | <= 250 ms/page | <= 200 ms/page |
| server transport serialization | <= 25 ms/event | <= 25 ms/event |
| browser reducer + paint | <= 16 ms/page update | <= 16 ms/page update |

Targets are alert thresholds, not claims about every host. Run
`make test-preview-performance` on the deployment-class machine and retain its
benchmark output when changing TeX, Poppler, document fixtures, or transport.

Observed baseline on 2026-08-29 (Linux amd64, Intel Core Ultra 9 185H,
`-benchtime=5x`, pdflatex + Poppler):

| Scenario | Observed |
| --- | ---: |
| cold composition, 100 components | 0.91 ms/op |
| warm composition, 100 components | 0.64 ms/op |
| warm two-page TeX | 133.9 ms/op |
| warm two-page fingerprint | 16.4 ms/op |
| unchanged-page full raster | 0.004 ms/op |
| focused-section TeX | 142.4 ms/op |
| focused-section fingerprint | 19.2 ms/op |
| changed focused-page raster | 28.1 ms/op |
| SSE event JSON, two 128 KiB changed pages | 0.48 ms/op |

## Browser measurement

Web editors live outside this repository. Call `rest.RegisterPreviewRoutes`,
consume its SSE `result` events, and key pages by
`Page.Number`. Apply `ChangedPages`, remove `RemovedPages`, then ignore every
event whose revision is older than the reducer's current revision.

Measure reducer plus paint around that state update with `performance.mark`,
then observe the next animation frame. Report it separately from SSE delivery;
network latency and browser paint are not server compile time.

```js
performance.mark("autopdf-preview-apply-start");
applyPreviewEvent(event);
requestAnimationFrame(() => {
  performance.measure(
    "autopdf-preview-apply",
    "autopdf-preview-apply-start"
  );
});
```

## Why unchanged pages still have hashes

The compiler makes a cheap low-DPI visual fingerprint pass over every page,
then performs full-DPI rasterization only for changed hashes. Complete page
identities let the session detect removals; omitted bytes keep unchanged pages
off the CPU-heavy raster path and off the wire.
