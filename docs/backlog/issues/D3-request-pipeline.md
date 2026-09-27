---
id: D3
title: "webhook: route events, deduplicate, allowlist and hand off (steps 5–12)"
type: Task
milestone: "M2 · Receive"
epic: D
depends_on: [D1, D2, C6]
repo: specsnl/specsdeployd
---

## Goal

The rest of the request pipeline, after the signature has been verified. Each row of the design.md table becomes one step. Each step has its
own response code, a JSON body with its `error_kind`, and a fixed answer to whether a status is posted to GitHub. The most important
rule is **"not hosted here" → `202 ignored`, and no status**: another server owns that environment, and this one must stay silent.

## Design reference

[`docs/design.md` § Request pipeline](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#request-pipeline),
[§ Replay and duplicates](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#replay-and-duplicates),
[§ Status mapping](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#status-mapping)

## Scope

- [ ] Step 5, `X-GitHub-Event`: `ping` gets `200 {"status":"pong"}`; anything other than `deployment` gets `204`
- [ ] Step 6, `event.DecodeEnvelope`: failure gets `422 invalid_event`; an action other than `created` gets `204`
- [ ] Step 7, dedupe: a bounded LRU of 1024 entries, keyed by `X-GitHub-Delivery` and separately by `deployment.id`. A hit gets
      `200 duplicate`. Record in the code what #B3 found about redelivery GUIDs
- [ ] Step 8, the allowlist: a repo not on it gets `403 repo_not_allowed`, with no status
- [ ] Step 9, routing: an environment not hosted here gets `202 ignored`, logged at `info`, with no status
- [ ] Steps 10–11, `event.Check`: failure gets `422` with its sentinel, **and** an `error` status posted through the
      `Reporter` interface (implemented in #F2; faked here)
- [ ] Step 12, `Enqueue` (an interface; implemented in #D4): accepted gets `202 accepted`; shutting down gets `503 shutting_down`
- [ ] One log line per request, with `delivery_id`, `event`, `repository`, `environment`, `deployment_id`, the outcome and
      `error_kind`
- [ ] The architecture page `webhook.md`, holding the table as built

## Tests

- **One test per row** of the pipeline table, using signed requests from a helper `signedRequest(t, secret, event, body)`.
  Each asserts the status code, the body, whether `Reporter` was called and with which state, and whether `Enqueue` was
  called.
- Ordering: an unsigned request for an allowlisted repo and a hosted environment still gets `401`, and nothing is parsed or reported.

## Depends on

Blocked by #D1, #D2, #C6

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
