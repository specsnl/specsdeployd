---
id: B3
title: "Spike: webhook topology — one webhook per server, hostnames and secrets"
type: Task
milestone: "M0 · Foundation"
epic: B
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

Architecture §8 names a single endpoint, `https://deploy.specs.dev/hook/github`, but every server has to receive
every `deployment` event and ignore those it does not host. A GitHub App has exactly one webhook URL, so it cannot
fan out to several servers. The design therefore proposes:

- **one org webhook per server**, at `deploy.<server>.specs.dev`, with its own secret;
- **the App only for statuses**.

Confirm that design, or replace it, before anyone builds the GitHub setup (#G5) or the Caddy site (X2).

## Design reference

[`docs/design.md` § Topology (spike B3)](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#topology-spike-b3)

## Scope

- [ ] Org webhook or repository webhook: which scales better as allowlisted repos are added? Check the limit on
      webhooks per event per org and per repo, and whether an org webhook can be limited to certain repos
- [ ] Subscribing to only the `deployment` event: confirm the delivery headers (`X-GitHub-Event`, `X-GitHub-Delivery`,
      `X-Hub-Signature-256`) and that the content type can be forced to `application/json`
- [ ] What a redelivery from the GitHub UI does to `X-GitHub-Delivery`: is it the same GUID or a new one? This decides whether
      delivery-ID dedupe alone blocks a redelivery. It shouldn't, but check (#D3)
- [ ] GitHub's timeout and retry behaviour for a `202`, a `4xx`, a `5xx`, and a delivery that times out
- [ ] DNS and hostnames for each server, and how they fit the Caddy site ansible-pull lays down
- [ ] Secrets: one per server, or shared? Also the rotation flow when `webhook_secrets` holds two entries

## Output

A comment with the decision. It feeds an update to design.md § Topology and open question 2, the runbook in #G5,
the Caddy part of X2, and the §8 amendment in #C3.

## Depends on

Nothing. **Parallel-safe.**
