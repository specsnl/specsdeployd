---
id: E2
title: "probe: poll readyz with a per-attempt timeout, a period and a deadline"
type: Task
milestone: "M3 · Deploy"
epic: E
depends_on: [A4]
repo: specsnl/specsdeployd
---

## Goal

`internal/probe`. Success is gated on `/readyz` (§10). This package waits for a `200` from the new container, or
gives up at the deadline. It uses Kubernetes semantics for `timeout_seconds` (per attempt), and design.md's `period_seconds` and
`deadline_seconds`.

## Design reference

[`docs/design.md` § Probes](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#probes)

## Scope

- [ ] `Prober` interface: `WaitReady(ctx, config.Probe) (Result, error)`. `Result` records the number of attempts, the last
      status code or error, and the elapsed time
- [ ] Only `200` counts as ready. **Redirects are not followed** (`CheckRedirect` returns `http.ErrUseLastResponse`), so a
      redirect to a login page cannot pass. TLS is verified normally
- [ ] Each attempt is bounded by `timeout_seconds`, and attempts are `period_seconds` apart. The deadline runs from when `WaitReady` was
      called, and missing it is `ErrNotReady`, wrapping the last observation
- [ ] `ctx` cancellation, from shutdown, returns immediately with `ctx.Err()`
- [ ] Response bodies are drained and closed, capped at 4 KiB
- [ ] The clock is injected: `Clock{Now, Sleep}`, as in labelsync's `ratelimit.Clock`
- [ ] The architecture page `probes.md`

## Tests

- `httptest` servers:
  - ready at once;
  - ready after N attempts;
  - never ready;
  - a `302`;
  - a `500` and then a `200`;
  - an attempt that hangs past the per-attempt timeout.
- A fake clock, so no test sleeps for real.
- The deadline is honoured exactly.

## Depends on

Blocked by #A4

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
