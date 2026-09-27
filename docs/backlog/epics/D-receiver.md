---
id: D
title: "Epic: Webhook receiver"
type: Feature
parent: specsnl/specsops#13
repo: specsnl/specsdeployd
---

`serve --listen`:

- a hardened `http.Server` that accepts only loopback addresses;
- HMAC verification before any parsing;
- the twelve steps of the request pipeline, each with its own response code, and a rule for whether a status is posted;
- the per-environment queue: supersede, monotonic IDs, and shutdown in two branches.

Tested with `httptest` against signed requests, and with a fake clock and a fake deployer for the queue. No real
GitHub and no sudo.

Design reference: [`docs/design.md`](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#webhook-handling) § Webhook handling, § Queue, § HTTP surface.
