---
id: A
title: "Epic: Foundation & repo hygiene"
type: Feature
parent: specsnl/specsops#13
repo: specsnl/specsdeployd
---

The spine every later change builds on:

- `AGENTS.md`;
- the Hugo docs site, so that each feature PR can update its architecture page from the start;
- the rulesets and dependabot;
- the sentinel errors and `KindOf`;
- the output writer, the slog setup and the `App` injection struct;
- tests that pin the release artifacts the Ansible role depends on.

None of this needs the network, and none of it adds behaviour. `v0.1.0` still answers only `version`
once the epic is done, but everything after it has somewhere to land.

It is modelled directly on labelsync's M0. Where a daemon differs from a CLI, the issue says so: logging is on by default, and there are only two exit codes.

Design reference: [`docs/design.md`](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md) § Package structure, § Error handling, § Logging, § Documentation, § Distribution.
