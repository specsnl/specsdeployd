---
id: D4
title: "queue: one worker per environment, supersede, monotonic IDs and shutdown"
type: Task
milestone: "M2 · Receive"
epic: D
depends_on: [A4, C6]
repo: specsnl/specsdeployd
---

## Goal

`internal/queue`. This decouples GitHub's 10-second delivery timeout from a deployment that can take minutes, and makes
concurrent deliveries safe. Each environment has **at most one running and one pending deployment**. The newest deployment wins, IDs only move
forward, and shutdown has exactly two branches, split at the point of no return.

## Design reference

[`docs/design.md` § Queue](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#queue),
[§ Status mapping](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#status-mapping)

## Scope

- [ ] `Queue.Enqueue(event.Deployment) error`: it never blocks, and returns `ErrShuttingDown` after drain has started
- [ ] For each environment:
      - if nothing is running, start the deployment;
      - otherwise replace the pending one, whose status is `inactive` "superseded by #N";
      - a deployment whose ID is lower than the running one or the last started one is `inactive` straight away, and never runs
- [ ] Workers call a `Deployer` interface, implemented by `deploy.Executor` in #E3, and report through `Reporter`.
      A deployment is `pending` when accepted and `in_progress` when its worker starts it
- [ ] Different environments run concurrently. The same environment never does
- [ ] `Drain(ctx)`:
      1. pending deployments get `error` "specsdeployd stopped before running this deployment";
      2. running deployments are told to stop. The executor decides between abort (before the promote step) and
         complete-restart-then-`error` (after it), per #E3;
      3. `Drain` waits until the running deployments finish or `ctx` expires
- [ ] The architecture page `queue.md`

## Tests

- Tests use a fake deployer (blocking on channels, so each test controls exactly when a deployment finishes) and a fake reporter:
  - supersede, including three quick arrivals, where only the first and the last run;
  - out-of-order IDs;
  - two environments running in parallel;
  - drain with a pending deployment;
  - drain while a deployment is running;
  - `Enqueue` after drain.
- Run with `-race`.

## Depends on

Blocked by #A4, #C6

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
