---
id: G
title: "Epic: Ship"
type: Feature
parent: specsnl/specsops#13
repo: specsnl/specsdeployd
---

From tested packages to a deployment on a real VM:

- `serve` wired up end to end;
- an e2e suite on a GitHub-hosted runner, with real Podman, Quadlet and systemd;
- the user and runbook docs;
- the GitHub App, plus the servers and their webhooks from OpenTofu;
- `v0.2.0-rc.N` running on a staging golden-app VM, then `v0.2.0`.

Done means the "Done when" of specsnl/specsops#13 is met. The cross-repo changes to the collection,
specsops-ansible and specsops-opentofu are tracked in their own repositories and linked from G6.

Design reference: [`docs/design.md`](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#distribution) § Distribution, § Testing strategy, § Cross-repo changes.
