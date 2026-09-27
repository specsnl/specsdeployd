---
id: E
title: "Epic: Deploy pipeline"
type: Feature
parent: specsnl/specsops#13
repo: specsnl/specsdeployd
---

The part that changes the host:

- a sudo runner, the only code that starts a process;
- the readyz poller;
- a pure `Plan()` plus an executor for pull → preserve previous → promote → restart → ready;
- rollback to `:previous`;
- `specsdeployd deploy`, which runs the same pipeline without GitHub.

Every test uses `runner.Fake` and a fake prober. The real Podman, systemd and Quadlet run is the e2e suite in
the Ship epic. This epic depends only on the foundation and on the config and event types, so it can run
alongside the receiver.

Design reference: [`docs/design.md`](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#deploy-pipeline) § Deploy pipeline, § Local image refs.
