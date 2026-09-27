---
id: B4
title: "Spike: how root pulls private GHCR images on an app server"
type: Task
milestone: "M0 · Foundation"
epic: B
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

`sudo -n /usr/bin/podman pull ghcr.io/specsnl/<app>:<tag>` runs as root with a clean environment, so
`REGISTRY_AUTH_FILE` cannot be passed through, and the sudoers rule has no room for `podman login`. Application
images are most likely private. Decide where root's registry credentials come from, **before** the first real
deploy fails with `unauthorized`.

## Design reference

[`docs/design.md` § Open questions](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#open-questions) (question 3)

## Scope

- [ ] Option A: ansible-pull lays down `/etc/containers/auth.json` (or root's `${XDG_RUNTIME_DIR}/containers/auth.json`), holding a
      read-only package token. Which token (a fine-grained PAT, or a machine user)? Where does it live, and how is it rotated?
- [ ] Option B: the daemon mints a short-lived token for the pull using the GitHub App (`packages: read`). Check whether GHCR
      accepts App installation tokens for `docker login`. If it does, the daemon would have to pass it to podman, for example with
      `--creds`, which cannot be done under the current sudoers rule without leaking the token into `ps`. Record why that
      rules the option out, if it does
- [ ] Which Podman search path does root use for `auth.json` under `sudo`, on Ubuntu 26.04?
- [ ] Confirm a pull of a **public** image needs nothing, for the e2e suite and for public apps

## Output

A comment with the decision, and a note in design.md (open question 3). If ansible-pull provides the file, add it to X2.
If the daemon needs to change, open a follow-up issue in this repo.

## Depends on

Nothing. **Parallel-safe.**
