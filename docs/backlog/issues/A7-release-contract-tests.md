---
id: A7
title: "release_test.go: pin the artifacts the specsdeployd role depends on"
type: Task
milestone: "M0 · Foundation"
epic: A
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

The `specsdeployd` role downloads `v{version}/specsdeployd_{version}_linux_{arch}.deb`, checks it against
`checksums.txt` **by basename**, installs it with `apt`, and guards the unit on `/usr/bin/specsdeployd`. Any
drift in `.goreleaser.yml` breaks the install. Worse, a wrong `bindir` produces a unit that stays dormant, which **looks exactly like
working as intended** (architecture §9).

Pin all of it in top-level tests that parse the config files, in the style of labelsync's `release_test.go`. Each test decodes into a
partial struct ("a view of the promises rather than a second copy of the file"), and each has a doc comment naming the
silent failure it guards against.

## Design reference

[`docs/design.md` § Where it fits](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#where-it-fits),
[§ Distribution](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#distribution)

## Scope

- [ ] `release_test.go` (package `main`) on `.goreleaser.yml`:
      - linux and darwin × amd64 and arm64, with `CGO_ENABLED=0`;
      - `nfpms` produces `deb` only, with `bindir: /usr/bin`, **no `contents`** and **no `scripts`**;
      - the file name template yields `specsdeployd_{{ .Version }}_linux_{{ .Arch }}.deb`;
      - `checksum.name_template` is `checksums.txt`;
      - the ldflags `-X` path is `…/internal/cmd.Version`, and it matches the `Dockerfile`. Move today's check out of `version_test.go`
- [ ] The casks:
      - the `specsdeployd@rc` and `specsdeployd` pair shares an anchor;
      - the stable cask has `skip_upload: auto`;
      - both are pushed to `specsnl/homebrew-tap` with `HOMEBREW_TAP_GITHUB_TOKEN`, which `release.yml` passes;
      - the post-install hook removes the quarantine attribute with `xattr`
- [ ] `release.yml`: runs on `v*` tags, with `contents: write`
- [ ] `ci_test.go`: the required job names exist, matching the ruleset from A3

## Tests

This issue is the tests. To prove they work, make one broken change locally, such as `bindir: /usr/local/bin`, and paste the failure
into the PR.

## Depends on

Nothing. **Parallel-safe.** If A3 lands first, extend `ci_test.go` to cover the new job names.

## Done when

`task checkall` passes, every box above is ticked, and `task release:dry-run` still produces the same `dist/` asset names.
