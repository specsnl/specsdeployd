---
id: G2
title: "test: e2e suite on a GitHub runner with real Podman, Quadlet and systemd"
type: Task
milestone: "M5 · Ship"
epic: G
depends_on: [G1, B2]
repo: specsnl/specsdeployd
---

## Goal

`runner.Fake` proves that the daemon *asks* for the right commands. This suite proves that those commands *work*: a real
`podman pull` from a local registry, a real `podman tag`, a real Quadlet unit restarted by real systemd, a real
readyz, all under a sudoers file equivalent to the role's. It runs on GitHub's `ubuntu-24.04` runner, which has
Podman, systemd and passwordless sudo, and it becomes a required check.

## Design reference

[`docs/design.md` § Testing strategy](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#testing-strategy) (`test/e2e.bats` row),
[§ Local image refs](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#local-image-refs)

## Scope

- [ ] `test/e2e/app/`: a tiny Go HTTP app image that answers `/readyz` with `200`, or with `500` when built with `--build-arg READY=false`.
      It is built with `podman build` in the job and pushed as `:good` and `:bad` to a `registry:2` container at `localhost:5000`,
      which `registries.conf` marks as insecure
- [ ] `test/e2e/fakegithub/`: a small Go program serving the installation, token and status endpoints and recording the statuses to
      a file. The daemon points at it with `github.api_url`
- [ ] The job's setup:
      - create the `specsdeployd` user;
      - install `testdata/sudoers` from #E1, then check it with `visudo -cf`;
      - install the Quadlet template from #B2, with `Image=localhost/app-e2e-test:deployed` and `Pull=never`;
      - run `daemon-reload`;
      - write a config and a `secrets.age` encrypted to a generated identity;
      - start the built binary as `specsdeployd`
- [ ] `test/e2e.bats`, with bats-support and bats-assert, sending deliveries signed with `openssl dgst -sha256 -hmac`:
      1. deploy `:good`: the unit runs it, and `success` is recorded;
      2. deploy `:bad`: the deployment rolls back, the unit runs `:good` again, and `failure` is recorded;
      3. a delivery for an environment not hosted here: `202`, and no status;
      4. a bad signature: `401`
- [ ] A `task e2e` that runs locally only on Linux with systemd; on other systems it prints why it was skipped. A CI job `E2E`,
      added to the ruleset's required checks and to `ci_test.go`

## Tests

This issue is the suite. It must also fail when it should, so show in the PR a run with the `podman tag` line removed from the sudoers file.

## Depends on

Blocked by #G1, #B2

## Done when

The `E2E` job is green and required. `task checkall` passes, and the architecture page `overview.md` links to the suite.
