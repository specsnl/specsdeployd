---
id: F
title: "Epic: GitHub reporting"
type: Feature
parent: specsnl/specsops#13
repo: specsnl/specsdeployd
---

GitHub App authentication and deployment statuses:

- an App JWT, then an installation lookup per repository, cached;
- the go-github wrapper and `retryTransport` from labelsync;
- a `Reporter` that maps every queue and pipeline outcome onto `pending`, `in_progress`, `inactive`, `error`,
  `success` and `failure`.

Reporting is best effort by design. A status that cannot be posted is logged and never changes what happens on the host.

Tested entirely against `httptest` fakes. This epic depends only on the foundation and the config types, so it is
**fully parallel with the contracts and deploy epics**.

Design reference: [`docs/design.md`](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#github-reporting) § GitHub reporting.
