---
id: A5
title: "cmd: App injection struct, output writer, exit codes and daemon logging"
type: Task
milestone: "M0 · Foundation"
epic: A
depends_on: [A4]
repo: specsnl/specsdeployd
---

## Goal

Turn today's `App{Out, Err}` into the injection container every later command needs. This is the labelsync and
specs-cli pattern. Tests replace fields on the struct, and never globals.

The one intended departure from labelsync is **logging**. A daemon logs by default: slog JSON to stdout, collected
by journald. labelsync's slog is silent unless `--debug` is given.

## Design reference

[`docs/design.md` § Logging](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#logging),
[§ Exit codes](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#exit-codes),
[§ Boundary rule](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#boundary-rule) (seams)

## Scope

- [ ] `internal/cmd/app.go`: `App` gets `Stdout`, `Stderr`, `Out output.Writer`, `Logger *slog.Logger`, and
      `Now func() time.Time`. The fields for later seams (`Runner`, `Prober`, `GitHub []github.Option`) are
      added by the issues that introduce those packages
- [ ] Global flags `--config`, `--log-level`, `--log-format json|text` and `-o/--output pretty|json`. Flag names are constants.
      Rebuild the writer and logger in `PersistentPreRunE` from the command's streams, so tests capture everything.
      Reject an unknown `--output` or `--log-format` value (specs-cli #115)
- [ ] `internal/util/output`:
      - a `Writer` interface with pretty and NDJSON implementations, for the one-shot commands;
      - `SetupLogger(w, format, level)` as the single place a slog handler is built;
      - the standard log attribute keys as constants
- [ ] `internal/util/exit`:
      - `OK = 0` and `Error = 1`;
      - a silent `*exit.Err` carrier, for errors that should exit non-zero without printing anything;
      - `main.go` prints errors once through the writer, including `error_kind`, and is the only place that calls `os.Exit`
- [ ] `main.go`: `signal.NotifyContext(ctx, SIGINT, SIGTERM)` feeds `ExecuteContext`
- [ ] Keep `version` and `--version` byte-identical to `v0.1.0`, which the role's molecule test depends on
- [ ] Architecture pages `overview.md` (the CLI wiring) and `logging.md`

## Tests

- Stream assertions: results go to stdout, narration and logs to stderr for one-shot commands, and `serve` logs to stdout.
- A golden-file test for the pretty and NDJSON writers, with a `-update` flag. `task test:update` finds the flag with grep, as in labelsync.
- `main` test: an error prints once, with its `error_kind`; the silent carrier prints nothing.

## Depends on

Blocked by #A4

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
