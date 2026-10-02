---
id: E5
title: "cmd: deploy --environment --image runs the pipeline without GitHub"
type: Task
milestone: "M3 · Deploy"
epic: E
depends_on: [A5, C2, E4]
repo: specsnl/specsdeployd
---

## Goal

`specsdeployd deploy --environment specs-production --image ghcr.io/specsnl/specs:v2026.02.01-1`. The operator
runs the **same** pipeline as the daemon, including rollback, on the host, with no webhook and no GitHub. It is useful for
debugging, for the first deploy onto a new server, and when GitHub is down.

## Design reference

[`docs/design.md` § CLI](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#cli)

## Scope

- [ ] Flags `--environment` (required, must be hosted: `ErrUnknownEnvironment`) and `--image` (required). The same
      `event.Check`-style image rules apply, against a synthetic `tag` or `branch` deployment derived from the image's tag
- [ ] It refuses while `specsdeployd.service` is active (`SystemctlIsActive`, no sudo needed) with `ErrServiceActive`,
      unless given `--force`. The message explains why: the two could race on the same unit
- [ ] It needs **no** secrets, so it runs without an identity
- [ ] Pretty output shows each step as it finishes. `--output json` gives NDJSON step events, then an outcome record. Exit `0` only on
      `Succeeded`
- [ ] `--dry-run` prints the plan and runs nothing
- [ ] The usage page `commands.md`, and a runbook entry "deploy by hand"

## Tests

Command tests with `runner.Fake` and a fake prober injected through `App`:

- success, failure with rollback, an unknown environment, the service active with and without `--force`;
- `--dry-run` runs no commands;
- the NDJSON shape.

## Depends on

Blocked by #A5, #C2, #E4

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
