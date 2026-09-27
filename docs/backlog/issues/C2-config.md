---
id: C2
title: "config: load, validate and normalise config.json"
type: Task
milestone: "M1 · Contracts"
epic: C
depends_on: [A4, C1]
repo: specsnl/specsdeployd
---

## Goal

`internal/config`. Turn `/etc/specsdeployd/config.json` into typed, defaulted, **fully validated** structs,
or fail with the first broken rule. The rest of the daemon never sees an invalid config and never applies a default
itself.

The rules that matter most are the new **bindings for each environment**. Environment names are free-form in GitHub, so every
environment must name the one `repository` allowed to deploy it and the one `image` repository it may run.

## Design reference

[`docs/design.md` § Configuration (§12.3, amended)](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#configuration-123-amended),
[§ Environment](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#environment),
[§ Local image refs](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#local-image-refs)

## Scope

- [ ] `Load(path) (*Config, error)`:
      - read the file;
      - validate it against the schema (#C1);
      - decode it strictly (`DisallowUnknownFields`, as defence in depth);
      - apply the defaults;
      - run the semantic rules
- [ ] A missing file wraps `ErrConfigNotFound`, and the message names the path
- [ ] The semantic rules from the design.md table, first broken rule wins:
      - unique names and units;
      - `repository ∈ allowed_repos`;
      - `image` is a repository reference without a tag or digest (`distribution/reference`);
      - the unit pattern;
      - `timeout_seconds ≤ deadline_seconds`;
      - `http` or `https` probe URLs
- [ ] Defaults: `secrets.file`, `secrets.identity_file`, `github.api_url`, and the probe timeout, period and deadline
- [ ] Types: `Environment.LocalRef(tag)` returns `localhost/<unit-basename>:<tag>`, for `deployed` and `previous`.
      `Config.Environment(name) (Environment, bool)`
- [ ] **Imports stay pure.** No `net/http`, `runner`, `probe` or `github` (the boundary rule)
- [ ] Architecture page `configuration.md`, with a table of rules and the concern each one covers, and the usage page `configuration.md`
      as the reference for every field

## Tests

- Golden files: the normalised config, after defaults, for each design.md example
- A table of broken configs, one per rule, asserting the sentinel and the message
- `LocalRef` for a handful of units
- An import-boundary test: parse the package's imports and fail on any forbidden one. This test is reused by #C6 and #E3

## Depends on

Blocked by #A4, #C1

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
