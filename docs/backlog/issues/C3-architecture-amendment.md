---
id: C3
title: "docs: amend architecture §8–§12 in specsops-golden-images to match this design"
type: Task
milestone: "M1 · Contracts"
epic: C
depends_on: [B1, B2, B3, C1]
repo: specsnl/specsdeployd
---

## Goal

`specsops-golden-images/docs/architecture.md` is "the single source of truth" and is "locked for Phase 1", but it
is incomplete in ways that only showed up while designing the daemon:

- there are no secrets anywhere;
- a tag deploy cannot switch the image;
- environments are not bound to a repository or an image;
- one webhook hostname cannot serve several servers, and an environment has no webhook of its own.

Open a PR there that brings it in line with `docs/design.md`, so both documents say the same thing again. The PR lives in
specsops-golden-images; this issue tracks it.

## Design reference

[`docs/design.md` § Cross-repo changes](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#cross-repo-changes),
[§ Contracts](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#contracts)

## Scope

- [ ] §8: one repository webhook per (repository, server) pair, created by OpenTofu, and one hostname per server, as decided in #B3
- [ ] §9:
      - the added sudoers line, as decided in #B2;
      - `in_progress`, `inactive` and `error` added to the reported states;
      - rollback;
      - the identity and secrets, as decided in #B1
- [ ] §11: the stable local ref (`localhost/<unit>:deployed`, `Pull=never`), and why a tag deploy needs it
- [ ] §12.3: the amended config schema and example, copied from the embedded file (#C1) and not retyped
- [ ] §12.4, new: the secrets schema
- [ ] Mention the change in §16, since the document declares itself locked: "amended for specsdeployd v0.2.0 (link)"

## Tests

None in this repository. The schemas in the PR must be byte-identical to `internal/schema/*.schema.json`. Paste a
`diff` in the PR to show it.

## Depends on

Blocked by #B1, #B2, #B3, #C1

## Done when

The specsops-golden-images PR has merged, and design.md links to it from § Cross-repo changes.
