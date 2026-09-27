---
id: F2
title: "github: Reporter posts deployment statuses for every queue and pipeline outcome"
type: Task
milestone: "M4 · Report"
epic: F
depends_on: [F1, D4, E4]
repo: specsnl/specsdeployd
---

## Goal

This implements the `Reporter` interface that the handler (#D3), the queue (#D4) and the executor (#E3, #E4) already call.
Every state a deployment can reach becomes a deployment status whose description tells a person looking at the GitHub UI
which server did what. **Reporting is best effort**, and a failure to report never changes what happens on the host.

## Design reference

[`docs/design.md` § Status mapping](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#status-mapping),
[§ Failure handling](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#failure-handling),
[§ Rollback](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#rollback) (outcome table)

## Scope

- [ ] `Reporter.Report(ctx, repo, deploymentID, Status)`:
      - `POST /repos/{o}/{r}/deployments/{id}/statuses` with `state`, `description`, `environment` and
        `environment_url` when configured;
      - `auto_inactive: true` on `success`
- [ ] **The one mapping table, from outcome to state**, shared with #E4:
      - `pending` when accepted;
      - `in_progress` when started;
      - `inactive` when superseded;
      - `error` when rejected after routing, or on shutdown;
      - `success` or `failure` from the `Outcome`
- [ ] Descriptions always start with `server_id:` and are cut to 140 **runes**, never in the middle of a rune.
      Include the `error_kind`, when there is one
- [ ] Failures go through the client's retries, then are logged at `WARN` with `error_kind: status_report_failed`. The caller
      carries on regardless. Reporting has its own context and timeout (10s), so a slow GitHub never holds up the
      queue
- [ ] The architecture page `github-client.md` § Statuses

## Tests

- `httptest`: the request body for every row of the mapping table, including truncation at 140 runes with a multi-byte
  character on the boundary, and `auto_inactive` on success only.
- A GitHub that returns `500` forever: the call returns within its timeout and the error is only logged.

## Depends on

Blocked by #F1, #D4, #E4

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
