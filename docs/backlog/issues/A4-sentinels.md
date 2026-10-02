---
id: A4
title: "specsdeployd: sentinel errors, KindOf, AppName and default paths"
type: Task
milestone: "M0 · Foundation"
epic: A
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

`internal/specsdeployd`: the package with no behaviour of its own that every other package may import.
It is labelsync's `internal/labelsync` pattern. Every failure the daemon can report has a sentinel here and a
stable `error_kind` string. Those strings appear in logs, in HTTP response bodies and in GitHub status
descriptions, so they are a **public contract**: they may be added to, never renamed.

## Design reference

[`docs/design.md` § Error handling](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#error-handling)

## Scope

- [ ] `configuration.go`: move `AppName` here from `internal/cmd`, and add `DefaultConfigPath`
      (`/etc/specsdeployd/config.json`), `DefaultSecretsPath` and `DefaultIdentityPath`
- [ ] `errors.go`: every sentinel in the design.md table, each with a one-line doc comment saying when it occurs
- [ ] `KindOf(err) string`: map wrapped errors to snake_case kinds, with `"internal"` as the fallback
- [ ] The architecture page `error-handling.md`, holding the table of sentinels and kinds, and the "three places" rule

## Tests

- A table-driven `KindOf` test over `allSentinels`, both bare and wrapped twice with `%w`
- **A source-parsing test** (`go/parser`): every exported `Err*` in the package is in `allSentinels`,
  and no two kinds collide. This is labelsync's `errors_test.go`

## Depends on

Nothing. **Parallel-safe.**

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
