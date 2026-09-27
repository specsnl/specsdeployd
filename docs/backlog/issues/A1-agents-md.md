---
id: A1
title: "AGENTS.md: conventions, workflow and the package-boundary rule"
type: Task
milestone: "M0 · Foundation"
epic: A
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

Give every contributor, human or agent, the same rules labelsync's contributors work by, adapted for a daemon.
`CLAUDE.md` is the single line `@AGENTS.md`.

The file should open with the binary name, the module path, one paragraph on what the daemon does, and the line
"Structure, conventions and library choices deliberately mirror specsnl/labelsync."

## Design reference

[`docs/design.md` § Package structure](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#package-structure),
[§ Boundary rule](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#boundary-rule),
[§ Error handling](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#error-handling),
[§ Documentation](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#documentation)

## Scope

- [ ] **Commands.** Everything runs through Task, never `docker compose`, `go` or `npx` directly. Include a table of tasks, and explain
      the `volume-init` teardown. `task checkall` is the gate before every PR
- [ ] **Workflow.** Branch off `main` as `<type>/GH-<n>-<slug>`: one issue, one branch, one PR, merge commits only.
      "Implement milestone X" means every open issue in it, ordered so that blocked issues follow their blockers
      and spikes are *merged* before the work that depends on them. Deliver a milestone as a stack of PRs with `gh-stack`. Close the loop
      by ticking the Scope and Done-when boxes and stating in the PR what was left unticked
- [ ] **Commits.** Conventional Commits with a scope where it helps, the subject ending in `(GH-N)`,
      and a body that explains *why*, wrapped at about 80 columns. **PRs** follow the structure of #3 and #4: `## Summary`,
      `## Notes for review`, `## Verified locally`, then `Closes #N`
- [ ] **Changes ship complete.** Code changes ship with, in the same PR: tests, the architecture page (mandatory,
      even for internal fixes), the usage page when the change is user-facing, and the README only when *what it is*
      or *how it is installed* changes
- [ ] **Tests.** stdlib `testing` only; table-driven; `httptest`; an injected clock; no test sleeps for real;
      `runner.Fake` for anything that would run a command
- [ ] **Errors.** The `%w` rule, and "adding a sentinel = `KindOf` + the error table on the docs page + `allSentinels`"
- [ ] **Boundaries.** The rule that `config`, `event` and the plan half of `deploy` never import `runner`, `probe`,
      `github` or `net/http`
- [ ] **Security invariants** that no PR may weaken without a design.md change:
      - nothing is parsed before the HMAC check;
      - no shell;
      - absolute sudo paths;
      - secrets only through `secrets.Secret`;
      - loopback-only listening
- [ ] **Comment rules.** Package docs with `# Heading` sections explain why, and inline comments give the non-obvious reason or the
      failure mode behind the code, never a restatement of the code
- [ ] **Docs.** Links between Hugo pages use `{{< ref >}}`, and pages link to design.md by absolute URL

## Tests

None. This is documentation only, and `task md:check` passes.

## Depends on

Nothing. **Parallel-safe.**

## Done when

`task md:check` passes and every box above is ticked. The file is short enough to read in full: aim for about the length of
labelsync's.
