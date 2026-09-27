---
id: B
title: "Epic: Spikes"
type: Feature
parent: specsnl/specsops#13
repo: specsnl/specsdeployd
---

Four questions that code cannot settle, and that each gate a contract:

- **B1**, the host age identity, gates the default `identity_file` path and possibly a unit change;
- **B2**, retagging under Quadlet, gates the Quadlet template and the `podman tag` sudoers line;
- **B3**, the webhook topology, gates the hostnames and how the webhooks are set up;
- **B4**, pulling private GHCR images, gates the registry credentials on hosts.

Each spike ends with a comment on its issue recording what was observed, and an update to `docs/design.md`,
which folds the answer into [§ Open questions](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#open-questions). They need no code and are fully parallel, so
start them on day one.

Design reference: [`docs/design.md`](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#open-questions) § Open questions.
