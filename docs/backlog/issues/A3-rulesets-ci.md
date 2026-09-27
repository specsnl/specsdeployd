---
id: A3
title: "ci: commit the rulesets, add dependabot, and split CI into the required checks"
type: Task
milestone: "M0 · Foundation"
epic: A
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

Make the repository enforce what labelsync's enforces, and commit the rules so they can be reviewed:
PRs are required, merge commits only, threads must be resolved, tags must start with `v`, and there is a fixed set of required checks.

## Design reference

[`docs/design.md` § Testing strategy](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#testing-strategy)

## Scope

- [ ] `.github/rulesets/Main.json`, copied from labelsync's:
      - no deletion and no non-fast-forward pushes;
      - PRs required, with 0 approvals and thread resolution required;
      - merge commits only;
      - required checks `Unit tests`, `Integration tests` and `Markdown gate`.
      G2 adds `E2E` once that job exists
- [ ] `.github/rulesets/Tags must have v-prefix.json`, the file labelsync added in `9bbab67`
- [ ] Apply both rulesets to the repository, and record in the PR that they were applied
- [ ] Split `ci.yml` into jobs:
      - `Unit tests`: tidy diff, golangci-lint, `go test -race -count=1 ./...`, a build smoke;
      - `Integration tests`: `go test -race -count=1 -tags=integration ./...`. It is real rather than labelsync's no-op, because the
        runner and secrets tests need it.
- [ ] `.github/workflows/md.yml`: a paths filter, markdownlint, and a `Markdown gate` job that always reports
- [ ] `.github/dependabot.yml`, the same shape as labelsync's: weekly updates for github-actions, `gomod /` (and `/docs` once A2 has landed) and docker, with grouped
      minor and patch updates, the `build` commit prefix with a scope, and the `github-actions`, `go` and `docker` labels

## Tests

The `ci_test.go` part is A7. For this issue, it is enough that CI goes green with the new job names.

## Depends on

Nothing. **Parallel-safe.**

## Done when

`task checkall` passes, every box above is ticked, the rulesets are active on GitHub, and a test PR shows the three
required checks.
