---
id: C
title: "Epic: Contracts — schemas, config, secrets, events"
type: Feature
parent: specsnl/specsops#13
repo: specsnl/specsdeployd
---

Everything the daemon trusts, and how it decides whether to trust it:

- the four JSON Schemas (§12.1–§12.4), embedded and compiled once;
- loading and validating `config.json`, including the new `repository` and `image` bindings for each environment;
- decrypting the age-encrypted `secrets.age` into redacting types;
- parsing a deployment event;
- `config validate`, so ansible-pull can refuse a broken config.

The whole epic is pure code with no network, tested with golden files and tables of broken inputs. It also
carries the amendment of architecture §8–§12 in specsops-golden-images, so that the two documents agree again.

Design reference: [`docs/design.md`](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#contracts) § Contracts, § Secrets.
