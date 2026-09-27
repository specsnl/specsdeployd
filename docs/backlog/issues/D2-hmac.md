---
id: D2
title: "webhook: verify X-Hub-Signature-256 before anything is parsed"
type: Task
milestone: "M2 · Receive"
epic: D
depends_on: [A4, C4]
repo: specsnl/specsdeployd
---

## Goal

This is the first line of defence. Steps 1–4 of the request pipeline check the route, the method, the content type, the size cap and the signature, **before a
single byte is parsed**. It is small code, and it has to be exactly right.

## Design reference

[`docs/design.md` § Request pipeline](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#request-pipeline) (steps 1–4),
[§ Security](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#security)

## Scope

- [ ] `POST /hook/github` only: other paths get `404`, other methods `405`
- [ ] Any `Content-Type` other than `application/json` gets `415`. GitHub's form-encoded format is refused, never
      decoded
- [ ] The body is read through `http.MaxBytesReader` with a limit of 1 MiB; anything larger gets `413`
- [ ] `X-Hub-Signature-256`: it must be present and prefixed with `sha256=`, with a hex digest of the right length. The body's
      HMAC-SHA256 is compared with `hmac.Equal` against **each** configured webhook secret (one or two, for rotation). If none
      matches, the request gets `401` with `ErrBadSignature`. Log which secret index matched at `debug`, so a rotation
      can be seen to finish
- [ ] Never log the signature header, and never echo the body
- [ ] Hand the verified raw body and the headers to the next stage as a value. The handler's type makes an
      unverified body unreachable (for example, `type verified struct{ body []byte; … }` with no exported constructor)

## Tests

- Check against GitHub's published test vector from "Validating webhook deliveries" (secret `It's a Secret to Everybody`, payload
  `Hello, World!`).
- Failure cases:
  - the header is missing, or has the wrong prefix;
  - the digest is truncated or its hex is invalid;
  - the right digest was made with the wrong secret;
  - the second secret matches;
  - the body is exactly 1 MiB, or 1 MiB + 1 byte;
  - the wrong content type.
- A test that a handler stub placed after verification is never called on any failure path.

## Depends on

Blocked by #A4, #C4

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
