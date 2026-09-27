---
id: G5
title: "ops: create the GitHub App and the per-server webhooks"
type: Task
milestone: "M5 · Ship"
epic: G
depends_on: [B3, F1]
repo: specsnl/specsdeployd
---

## Goal

This is the one-time setup that no code can do. Create the GitHub App the daemon reports statuses as, and the webhook
through which each server receives deployments. Follow `usage/github-setup.md` (#G3) **as written**, so the runbook
is tested by being used.

## Design reference

[`docs/design.md` § Authentication](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#authentication),
[§ Topology (spike B3)](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#topology-spike-b3)

## Scope

- [ ] Create the App `specsdeployd` in `specsnl`, with Deployments read and write, Metadata read, and **no webhook**. Install it on the
      repositories that will be deployed first
- [ ] Generate the App private key and store it in 1Password, next to the break-glass age identity
- [ ] Create a webhook for the staging server, following #B3: its URL, `application/json`, the `deployment` event only, and a
      generated secret of at least 32 bytes, also stored in 1Password
- [ ] Record the `app_id` and the webhook IDs in the ansible repository's host vars (X2), not in this repository
- [ ] Wherever the runbook turned out to be wrong, fix it in the same PR as the notes

## Tests

A `ping` delivery from GitHub's UI reaches the staging server and gets `200 pong`. This needs X2's Caddy site.

## Depends on

Blocked by #B3, #F1

## Done when

The App and the staging webhook exist, both secrets are in 1Password, and the runbook matches reality.
