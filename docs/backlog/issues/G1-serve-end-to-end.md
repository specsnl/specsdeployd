---
id: G1
title: "cmd: wire serve end to end and drive it with signed webhooks"
type: Task
milestone: "M5 · Ship"
epic: G
depends_on: [D1, D3, D4, E4, F2]
repo: specsnl/specsdeployd
---

## Goal

Put the tested pieces together inside `serve`:
config → secrets → GitHub client → reporter → executor → queue → handler → server.

Then prove the whole tree works through the same seam labelsync's `harness_test.go` uses. `App` fields replace the runner, the prober and
GitHub's base URL, and a test sends signed deliveries and watches what comes out.

## Design reference

[`docs/design.md` § Package structure](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#package-structure),
[§ Testing strategy](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#testing-strategy) (`cmd` row)

## Scope

- [ ] `serve` builds the graph in one place, `internal/cmd/serve.go`. No package reaches into another's globals
- [ ] `App` gets `Runner runner.Runner`, `Prober probe.Prober` and `GitHub []github.Option`, each nil by default, in which case the real one is used
- [ ] `internal/cmd/harness_test.go` provides:
      - `startServe(t, cfg)`, which uses a generated age identity and secrets file and a free loopback port;
      - `fakeGitHub(t)`, which records statuses behind a mutex;
      - `deliver(t, event, body)`, which signs the delivery;
      - `waitForStatus(t, id, state)`, which polls a channel and never sleeps
- [ ] Scenarios:
      - a tag deploy succeeds, with statuses `pending`, then `in_progress`, then `success`;
      - a readyz failure rolls back, ending in `failure` with "rolled back";
      - an environment not hosted here posts no status;
      - a repository mismatch ends in `error`;
      - a supersede ends in `inactive`;
      - SIGTERM during a deployment ends in `error`
- [ ] `README.md` § Status: the daemon now deploys

## Tests

This issue is the tests: the harness and the scenarios above, run with `-race`.

## Depends on

Blocked by #D1, #D3, #D4, #E4, #F2

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
