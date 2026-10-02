---
id: C5
title: "cmd: config validate, for Ansible's validate: hook"
type: Task
milestone: "M1 · Contracts"
epic: C
depends_on: [A5, C2, C4]
repo: specsnl/specsdeployd
---

## Goal

`specsdeployd config validate [path]`. ansible-pull lays down `config.json` with
`template: … validate: "specsdeployd config validate %s"`, so a broken config never replaces a working one. A
config problem is then an Ansible task failure, instead of a daemon that won't start at 3am.

## Design reference

[`docs/design.md` § CLI](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#cli),
[§ Exit codes](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#exit-codes)

## Scope

- [ ] `config validate [path]`: the path defaults to `--config`. It runs `config.Load`, including every semantic rule
- [ ] `--secrets`: also decrypt and validate the configured `secrets.file` with `identity_file`. Ansible runs as root,
      so it can read the identity
- [ ] Exit `0` when valid and `1` otherwise. Pretty output is one line, `valid: /path` or the error.
      `--output json` gives `{"valid":bool,"path":…,"error_kind":…,"error":…}`
- [ ] It must work on a file whose name is not `config.json`, because Ansible's `%s` is a temporary file
- [ ] The usage page `commands.md`, with the exact Ansible snippet

## Tests

Command tests, driven through `executeCmd` with buffers:

- valid config, invalid config, missing file;
- `--secrets` with the right identity and with a wrong one;
- JSON output shape, exit codes, and which stream each line goes to.

## Depends on

Blocked by #A5, #C2, #C4

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
