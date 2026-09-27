---
id: C6
title: "event: decode a deployment event and check it against its environment"
type: Task
milestone: "M1 · Contracts"
epic: C
depends_on: [C1, C2]
repo: specsnl/specsdeployd
---

## Goal

`internal/event`. This is the pure half of the request pipeline. It decodes the envelope, and once routing is done it validates the
payload and every rule that ties a payload to its environment. It does no HTTP. The handler (#D3) calls it with bytes and
a `config.Environment`, and gets back a typed `Deployment` or a sentinel.

## Design reference

[`docs/design.md` § Deployment payload (§12.1, unchanged)](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#deployment-payload-121-unchanged),
[§ Deployment event subset (§12.2, unchanged)](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#deployment-event-subset-122-unchanged),
[§ Request pipeline](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#request-pipeline) (steps 6, 10, 11)

## Scope

- [ ] `DecodeEnvelope([]byte) (Envelope, error)`:
      - validates against §12.2 **without** the payload;
      - requires `repository.full_name`;
      - returns `action`, `deployment.{id, environment, ref}`, the repository, and the payload as `json.RawMessage`
- [ ] `Envelope.Created() bool`: only `action == "created"` is acted on
- [ ] `Check(env config.Environment, e Envelope) (Deployment, error)`:
      - the repository matches (`ErrRepositoryMismatch`);
      - the payload is valid against §12.1 (`ErrInvalidPayload`);
      - the image parses (`ErrInvalidImageRef`), with a tag or digest, and belongs to `env.image` (`ErrImageNotAllowed`);
      - `ref == deployment.ref`, and for the `tag` strategy the image tag equals `ref` (`ErrRefMismatch`);
      - `app` matches, when set (`ErrAppMismatch`)
- [ ] `Deployment` carries everything the queue and the pipeline need: id, environment, repository, strategy, ref,
      a normalised image reference, requested_by
- [ ] Pure imports; the #C2 boundary test covers this package too

## Tests

- A table per rule, using the design.md example event as the base case.
- Hostile images, each checked to fail with the right sentinel:
  - `--help`
  - `ghcr.io/specsnl/specs:v1 --rm`
  - `ghcr.io/specsnl/specs@sha256:…` (valid)
  - `docker.io/library/specs:v1` (wrong repository)
  - `ghcr.io/specsnl/specs-evil:v1` (a prefix, which must not match)
- Events that are not `created` are recognised as such.

## Depends on

Blocked by #C1, #C2

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
