---
id: G4
title: "release: cut v0.2.0-rc.1"
type: Task
milestone: "M5 · Ship"
epic: G
depends_on: [G2, G3]
repo: specsnl/specsdeployd
---

## Goal

This is the first release that deploys. It is a release candidate: it updates only the `specsdeployd@rc` cask, and it is what
the staging VM pins in #G6. `v0.2.0` itself is cut in #G6, after acceptance.

## Design reference

[`docs/design.md` § Distribution](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#distribution)

## Scope

- [ ] `task release:dry-run`: the asset names match the role's contract, which `release_test.go` (#A7) already enforces
- [ ] Tag `v0.2.0-rc.1` and let the release workflow publish it as a pre-release
- [ ] Download the `.deb` for amd64 and arm64, and check that `dpkg -c` shows only `/usr/bin/specsdeployd`, that
      `specsdeployd version` prints `0.2.0-rc.1`, and that `checksums.txt` matches
- [ ] Release notes: what changed since `v0.1.0`, and the **host prerequisites**: the collection version with the
      `podman tag` sudoers line (X1), and the config, secrets and Quadlet that ansible-pull must add (X2)

## Tests

The release tests from #A7, plus the manual checks above, recorded in a comment on this issue.

## Depends on

Blocked by #G2, #G3

## Done when

The pre-release exists, the checks above are recorded, and the rc cask is updated.
