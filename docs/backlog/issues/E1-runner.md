---
id: E1
title: "runner: sudo runner with argument guards, timeouts and a fake"
type: Task
milestone: "M3 · Deploy"
epic: E
depends_on: [A4]
repo: specsnl/specsdeployd
---

## Goal

`internal/runner` is the **only** code in the repository that starts a process. It runs exactly the commands the
role's sudoers rule allows, the way it allows them: `sudo -n`, absolute paths, an argument vector, no shell. It
also refuses anything that looks like an option in a place where a value belongs.

## Design reference

[`docs/design.md` § Exec safety](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#exec-safety),
[§ Where it fits](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#where-it-fits) (sudoers contract)

## Scope

- [ ] `Runner` interface: `Run(ctx, Command) (Result, error)`. `Command` is one of a closed set of constructors:
      `PodmanPull(image)`, `PodmanTag(src, dst)`, `SystemctlRestart(unit)`, `SystemctlIsActive(unit)`. There is no
      generic `Command{Name, Args}`, so no caller can build an arbitrary command
- [ ] Paths are constants: `/usr/bin/sudo`, `/usr/bin/podman`, `/usr/bin/systemctl`. `SystemctlIsActive` runs without sudo
- [ ] `ErrUnsafeArgument` for any argument that starts with `-` or contains a NUL or a newline
- [ ] `exec.CommandContext`:
      - `Env` is `PATH=/usr/sbin:/usr/bin:/sbin:/bin` and `LANG=C.UTF-8`, and nothing else;
      - `SysProcAttr{Setpgid: true}`, and cancelling the context kills the process group;
      - stdout and stderr are combined and capped at 64 KiB;
      - hitting the timeout is `ErrCommandTimeout`
- [ ] The command's `Result` carries the exit code, the captured output and the duration. A non-zero exit is an error that carries its
      `Result`
- [ ] `runner.Fake` records `[]Command` in order and returns scripted results, matched by command
- [ ] `testdata/sudoers`, a copy of the role's `sudoers.j2` rendered for the `specsdeployd` user **plus** the
      `podman tag *` line from #B2. Put a comment at the top linking to the collection's template
- [ ] The architecture page `deploy-pipeline.md` § Exec safety

## Tests

- Every constructor yields a vector that one of the lines in `testdata/sudoers` matches. The test implements sudoers'
  `*` glob, which matches spaces too.
- Unsafe arguments are refused before any process starts.
- `//go:build integration`, using real processes and no sudo: a timeout kills a `sleep` and its child, and the output cap
  holds.

## Depends on

Blocked by #A4

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
