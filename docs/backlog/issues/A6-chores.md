---
id: A6
title: "chore: align the Go versions and add specsdeployd to the go label group"
type: Task
milestone: "M0 · Foundation"
epic: A
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

Two small leftovers that are worth fixing before real code arrives.

1. **The Go versions disagree.** `go.mod` says `go 1.26.5`, the `Dockerfile` builds with `golang:1.27.1-trixie`, and CI uses
   `go-version-file`, so it gets 1.26.5. A local `task build` and a CI build therefore use different toolchains.
2. **A box left unticked on #1:** specsnl/specsops `labels.yml` lists only `specs-cli` and `labelsync` in the `go` group,
   so this repo never receives the `go` label that dependabot's gomod updates use.

## Design reference

[`docs/design.md` § Distribution](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#distribution)

## Scope

- [ ] Choose one Go version and use it in `go.mod` (`go` plus `toolchain`, if needed), the `Dockerfile` `FROM`, and the
      compose image pins. Record the choice in `architecture/versioning.md`, which is labelsync's go.mod toolchain-pin paragraph
- [ ] Open a PR in specsnl/specsops adding `specsnl/specsdeployd` to the `go` group, and link it here

## Tests

`task checkall` and `task build` pass. CI and local report the same `go version`.

## Depends on

Nothing. **Parallel-safe.**

## Done when

Both boxes are ticked, and the specsops PR has merged.
