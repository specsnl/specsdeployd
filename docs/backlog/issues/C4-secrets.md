---
id: C4
title: "secrets: decrypt secrets.age in memory and redact it everywhere"
type: Task
milestone: "M1 · Contracts"
epic: C
depends_on: [A4, C1]
repo: specsnl/specsdeployd
---

## Goal

`internal/secrets`. At startup, decrypt the age-encrypted `secrets.age` with the host identity, validate the
plaintext against §12.4, and hold it **only in memory**, as types that cannot be printed. **There is no plaintext
code path**: no config field, no env var, no flag. Tests generate throwaway identities.

## Design reference

[`docs/design.md` § Secrets](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#secrets),
[§ Always encrypted at rest](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#always-encrypted-at-rest),
[§ Why age, not sops](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#why-age-not-sops),
[§ Secrets (§12.4, new)](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#secrets-124-new)

## Scope

- [ ] `Load(file, identityFile string) (*Secrets, error)` using `filippo.io/age`:
      - armored and binary input;
      - X25519 identities (`AGE-SECRET-KEY-1…`);
      - `ssh-ed25519` private keys, through `agessh`;
      - an identity file holding several identities
- [ ] Sentinels:
      - a missing file is `ErrSecretsNotFound` or `ErrIdentityNotFound`;
      - no matching recipient is `ErrDecrypt`;
      - a schema failure is `ErrInvalidSecrets`
- [ ] **Permission check:** an identity file that is group- or world-readable gives `ErrIdentityPermissions`. Leave room for the
      `LoadCredential` path if #B1 chooses it (the `$CREDENTIALS_DIRECTORY` file is owned by root)
- [ ] `Secret` type:
      - `String()`, `GoString()`, `Format`, `LogValue()` and `MarshalJSON()` all give `[redacted]`;
      - `Reveal() []byte` is the only way to read the value, and it is used only by the HMAC check and the App key parser
- [ ] `github_app_private_key` is parsed (`x509.ParsePKCS1PrivateKey` or PKCS#8) during `Load`. A bad key is
      `ErrInvalidAppKey`, **at startup**, not at the first status report
- [ ] `WebhookSecrets() []Secret` returns one or two entries, for rotation
- [ ] The architecture page `secrets.md`, and the usage page `secrets.md` (authoring with the `age` CLI, the break-glass
      recipient, rotation), which G3 finishes

## Tests

- Round trips with a generated X25519 identity and a generated `ssh-ed25519` key, in armored and binary form.
- A wrong identity gives `ErrDecrypt`, and a truncated file gives `ErrDecrypt`.
- Invalid plaintext: one webhook secret too short, three secrets, a missing key, a key that doesn't parse.
- Permissions: 0644 fails and 0400 passes. Marked `//go:build integration`, because it needs real file modes.
- **Leak test:** format a `Secrets` value with `%v`, `%+v`, `%#v`, slog JSON and `json.Marshal`. The raw value must never appear.

## Depends on

Blocked by #A4, #C1

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
