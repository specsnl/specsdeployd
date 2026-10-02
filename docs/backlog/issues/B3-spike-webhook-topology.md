---
id: B3
title: "Spike: confirm repository webhooks per (repo, server) pair, created by OpenTofu"
type: Task
milestone: "M0 · Foundation"
epic: B
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

A GitHub environment has no webhook of its own. The `deployment` event belongs to the repository. The design
therefore delivers it through **one repository webhook for each (repository, server) pair**. Each webhook points at the server's own
hostname and signs with that server's secret, and **OpenTofu creates it together with the server**. The design rejects:

- org webhooks, which would cap the fleet at 20 servers and send every server every deployment;
- the App's single webhook;
- a central relay, which would be a single point of failure.

This spike confirms that the design works in practice, and decides how each server's webhook secret gets from its 1Password item into
`secrets.age`.

## Design reference

[`docs/design.md` § Topology (spike B3)](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#topology-spike-b3),
[§ Replay and duplicates](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#replay-and-duplicates)

## Scope

- [ ] Using OpenTofu's `integrations/github` provider, create a `github_repository_webhook` on a scratch repository, with the `deployment`
      event only, `content_type = "json"` and a secret. Record the provider version and the token permissions it needs
      (repository webhooks: read and write, on the app repositories only)
- [ ] Create two deployments for two environments of the scratch repo, and confirm each is delivered once to the
      webhook, with the headers the daemon relies on: `X-GitHub-Event`, `X-GitHub-Delivery` and `X-Hub-Signature-256`
- [ ] Redelivery from the GitHub UI: is `X-GitHub-Delivery` the same GUID or a new one? This decides whether delivery-ID dedupe
      alone blocks a redelivery. It shouldn't; deployment-ID dedupe is the real guard (#D3)
- [ ] GitHub's timeout and retry behaviour for a `202`, a `4xx`, a `5xx`, and a delivery that times out
- [ ] **Who assembles `secrets.age`.** Decided already: the secret lives in the **server's own 1Password item**, and the
      server never holds 1Password credentials. Choose one of two builders:
      - OpenTofu, at creation, which already reads the item for the webhook;
      - an Ansible playbook run off-host, using the `community.general.onepassword` lookup.
      Either commits the ciphertext to specsops-ansible. Compare how each handles rotation, and whether the plaintext stays off disk.
      Also check whether the 1Password provider keeps the secret in OpenTofu's state
- [ ] The rotation flow: add the new secret to the item, rebuild `secrets.age` with both entries, update the webhooks, then drop
      the old entry
- [ ] Moving an environment to another server in one apply: the old webhook is removed and the new one created, and
      neither server's config disagrees with them for longer than one ansible-pull run

## Output

A comment with the observations and the decision on who builds `secrets.age`. It feeds:

- design.md § Topology, and open question 2;
- the webhook module in X3;
- the runbook in #G5;
- the §8 amendment in #C3.

## Depends on

Nothing. **Parallel-safe.**
