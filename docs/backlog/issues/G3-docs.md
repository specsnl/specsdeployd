---
id: G3
title: "docs: usage pages, runbook and the security page"
type: Task
milestone: "M5 · Ship"
epic: G
depends_on: [G1, B1, B3, B4]
repo: specsnl/specsdeployd
---

## Goal

The architecture pages grow with each feature PR. This issue fills in the pages an **operator** needs, written for someone who has
never seen the daemon before: how to put it on a server, configure it, author its secrets, and diagnose it at 3am.

## Design reference

[`docs/design.md` § Documentation](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#documentation),
[§ Security](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#security)

## Scope

- [ ] `usage/getting-started.md`: what the role provides, and what ansible-pull must add: the config, `secrets.age`, the identity, the
      Caddy site, and a Quadlet that uses the local ref
- [ ] `usage/configuration.md`: every field, its default, and what is rejected. Where possible, generate the tables from the schema and
      check them in a test
- [ ] `usage/secrets.md`:
      - authoring with the `age` CLI;
      - the host and break-glass recipients;
      - the 1Password location of the break-glass identity and the App key;
      - rotating a webhook secret with no downtime, using the two-secret window;
      - rotating a host identity
- [ ] `usage/github-setup.md`: the permissions of the GitHub App, installing it, and one webhook per server, as decided in #B3
- [ ] `usage/commands.md`: `serve`, `config validate`, `deploy`, `version`
- [ ] `usage/runbook.md`, covering:
      - using `journalctl -u specsdeployd` with the standard log keys;
      - a failed deploy;
      - **rollback failed**;
      - a deployment stuck in `pending`, after the daemon was down;
      - the App is not installed;
      - a bad signature;
      - `unauthorized` on pull, per #B4
- [ ] `architecture/security.md`: the threat-model table from design.md, kept up to date with what was built
- [ ] `README.md`: rewrite the Status and Install sections, keeping the README an overview plus a pointer, as in labelsync

## Tests

`task docs:build` and `task md:check`. Also test whatever generated tables come out of the configuration item above.

## Depends on

Blocked by #G1, #B1, #B3, #B4

## Done when

Every page listed in design.md § Documentation has left its "not built yet" state, and the site is published.
