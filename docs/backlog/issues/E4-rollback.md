---
id: E4
title: "deploy: roll back to :previous when promote, restart or readyz fails"
type: Task
milestone: "M3 · Deploy"
epic: E
depends_on: [E3]
repo: specsnl/specsdeployd
---

## Goal

A failed readyz must not leave the server running an image that is not ready. When `Promote`, `Restart` or `Ready`
fails and a `:previous` image exists, put it back and prove it is ready again. The architecture asks only that failure be
reported. This extra step was agreed for v1.

## Design reference

[`docs/design.md` § Rollback](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#rollback)

## Scope

- [ ] Rollback steps:
      1. `PodmanTag(:previous → :deployed)`
      2. `SystemctlRestart(unit)`
      3. `WaitReady(readyz)`, with the same deadline
- [ ] The `Outcome` enum:
      - `Succeeded`;
      - `FailedNothingChanged` (pull or preserve failed);
      - `FailedRolledBack`;
      - `FailedNoPrevious` (the first deploy failed after promote);
      - `FailedRollbackFailed`;
      - `Interrupted`.
      Each outcome maps to one status state and a description template, a table shared with #F2
- [ ] `FailedRollbackFailed` logs at `ERROR`, with every step's captured output, and its description ends in "manual action
      needed"
- [ ] Rollback is **not** attempted on `Interrupted`, because shutdown must be fast and predictable
- [ ] The architecture page `deploy-pipeline.md` § Rollback, with the outcome table

## Tests

- One test per outcome, with `runner.Fake` and a fake prober, asserting the exact command sequence.
- **Two failures in a row:** deploy good image G, then fail with B1, then fail with B2. Both failures roll back to G.
  This works because a successful rollback points `:deployed` back at G, so the next `Preserve` copies G again. The test
  pins that invariant down.
- The one case where it breaks is a rollback whose **retag** fails. `:deployed` then still points at B1, and the next deploy
  would preserve B1 as `:previous`. That is exactly the "manual action needed" outcome, so document it in the runbook (#G3).
  Do not add any code for it.

## Depends on

Blocked by #E3

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
