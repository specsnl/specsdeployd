---
id: C1
title: "schema: embed the §12 JSON Schemas and compile them once"
type: Task
milestone: "M1 · Contracts"
epic: C
depends_on: [A4]
repo: specsnl/specsdeployd
---

## Goal

`internal/schema`. The four contracts — the deployment payload (§12.1), the event envelope (§12.2), the config (§12.3,
amended) and the secrets (§12.4, new) — are JSON Schemas shared across three repositories. Embed them as
files, compile them once with `santhosh-tekuri/jsonschema/v6` (draft 2020-12), and turn validation errors
into sentinels with a readable instance path.

This deliberately departs from labelsync, which rejected a schema library. Record why on the library-decisions page:
the contracts **are** schemas, and a second, hand-written copy of each one in Go would drift away from it.

## Design reference

[`docs/design.md` § Contracts](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#contracts),
[§ Schemas are embedded](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#schemas-are-embedded),
[§ Dependencies](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#dependencies)

## Scope

- [ ] `internal/schema/{deployment-payload,github-deployment-event,specsdeployd-config,specsdeployd-secrets}.schema.json`,
      each byte-identical to the corresponding block in design.md
- [ ] The event schema resolves the payload `$ref` against the embedded payload schema, **with no network loader**.
      Any attempt to fetch a remote `$ref` is an error
- [ ] Turn `format` assertions on (`uri`)
- [ ] `schema.Validate(kind, value) error` returns an error wrapping `ErrInvalidConfig`, `ErrInvalidEvent`,
      `ErrInvalidPayload` or `ErrInvalidSecrets`. The message names the first failing instance path
      (`environments[1].probes.readyz.url`), not the library's whole error tree
- [ ] `EnvelopeWithoutPayload`: validate §12.2 with `payload` left unchecked. Routing comes first and the payload afterwards,
      as design.md § Request pipeline describes
- [ ] Architecture pages `contracts.md` and `library-decisions.md`: cobra, go-github, age, jsonschema and reference, plus
      a "what was rejected" section (sops, a hand-written validator, a router)

## Tests

- Every schema compiles.
- **Doc-sync test:** extract the JSON code blocks under design.md § Contracts, compare each schema block with the embedded
  file, and validate every example against its schema. This catches the doc and the code drifting apart
- A table of broken documents per schema, each asserting the sentinel and the reported path.

## Depends on

Blocked by #A4

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
