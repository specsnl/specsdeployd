---
id: X3
title: "App servers: VM, deploy hostname, webhook secret and repository webhooks per server"
type: Task
repo: specsnl/specsops-opentofu
parent: specsnl/specsops#13
depends_on: [B1, B3]
---

**Repo:** `specsnl/specsops-opentofu` · **Driven by:** specsnl/specsdeployd [`docs/design.md` § Topology](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#topology-spike-b3)

## Why

Every app server receives deployments through **one repository webhook for each repository it hosts an environment of**.
Each webhook points at that server's own `deploy.<server>.specs.dev` and signs with the server's secret. Org
webhooks would cap the fleet at 20 servers, and a central relay would be a single point of failure. OpenTofu already creates the servers, so it
also creates their webhooks. That puts the mapping from environments to servers in one declared place.

## Scope

- [ ] A module for app servers. Its input is the server name plus the environments it hosts, each as a `{ repository, environment }` pair.
      It creates:
      - the golden-app VM;
      - the DNS record `deploy.<server>.specs.dev`;
      - the webhook secret, following spike B3's decision on its origin;
      - one `github_repository_webhook` for each **distinct** repository among the server's environments, with the `deployment`
        event only, `content_type = "json"`, `insecure_ssl = false`, the URL `https://deploy.<server>.specs.dev/hook/github`
        and the server's secret
- [ ] The GitHub provider token: repository webhooks read and write on the app repositories, and nothing else. Where it
      lives, and how it is rotated
- [ ] If spike B1 chose a dedicated age identity generated at creation time: the module creates it too, and outputs its
      recipient for specsops-ansible. Record that the private key then sits in state
- [ ] Outputs that specsops-ansible can consume, so each server's `config.json` lists exactly the environments the module
      declared for it

## Verify

- [ ] One `tofu apply` creates the server, and the webhooks appear in each hosted repository
- [ ] Moving an environment between servers in the input removes one webhook and creates the other
- [ ] GitHub's `ping` delivery reaches the server and gets `200 pong`. This needs X2's Caddy site
