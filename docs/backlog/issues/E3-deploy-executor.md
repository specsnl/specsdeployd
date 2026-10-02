---
id: E3
title: "deploy: pure Plan() and an executor for pull → preserve → promote → restart → ready"
type: Task
milestone: "M3 · Deploy"
epic: E
depends_on: [C2, C6, E1, E2, B2]
repo: specsnl/specsdeployd
---

## Goal

`internal/deploy`. This splits *what* to do, a pure `Plan(env, deployment) []Step` that can be tested with golden files, from
*doing* it: an `Executor` that runs the steps through `Runner` and `Prober` and emits one event per step. It is the
equivalent of labelsync's plan/apply split, and it keeps the interesting logic free of fakes.

## Design reference

[`docs/design.md` § Deploy pipeline](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#deploy-pipeline),
[§ Local image refs](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#local-image-refs),
[§ Queue](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#queue) (point of no return)

## Scope

- [ ] `Plan(config.Environment, event.Deployment) Plan`: pure, with these steps:
      1. `Pull(image)`
      2. `Preserve(:deployed → :previous)`
      3. `Promote(image → :deployed)`, marked as the **point of no return**
      4. `Restart(unit)`
      5. `Ready(readyz)`
      Each step has its default timeout. The plan has a `String()` for logs and a JSON form for `deploy --output json`
- [ ] `Executor.Run(ctx, Plan) Outcome`:
      - it runs the steps in order and emits a `StepEvent{step, duration_ms, error_kind, output}` to an `Events` sink;
      - `Preserve` failing with "image not known" (the pattern from #B2) is **skipped**, meaning a first deploy. Any other failure is an
        error, and nothing has changed
- [ ] **Cancellation, which is how shutdown reaches a deployment:**
      - before `Promote`: stop, and return outcome `Interrupted`, with nothing changed;
      - after `Promote`: run `Restart` to completion with its own timeout under a detached context, skip `Ready`, and return
        `Interrupted`, with the image changed but readiness unconfirmed
- [ ] Sentinels `ErrImagePullFailed`, `ErrImageTagFailed`, `ErrRestartFailed` and `ErrNotReady` wrap the runner's or prober's error
- [ ] The rollback branch is a hook left for #E4: in this issue, a failure after `Promote` returns `FailedNoRollback`
- [ ] The plan half has pure imports, covered by the #C2 boundary test
- [ ] The architecture page `deploy-pipeline.md`

## Tests

- `Plan()` golden files for `tag`, `branch` and digest deployments.
- Executor tests with `runner.Fake` and a fake prober:
  - the happy path, with the exact command order asserted;
  - pull fails;
  - first deploy (preserve skipped);
  - promote fails;
  - restart fails;
  - not ready;
  - cancellation before and after the point of no return.

## Depends on

Blocked by #C2, #C6, #E1, #E2, #B2 (the "image not known" pattern)

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
