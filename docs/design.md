# specsdeployd — Design Plan

> **Repo:** `specsnl/specsdeployd` · **module:** `github.com/specsnl/specsdeployd` · **binary:** `specsdeployd`
> Distributed as a `.deb` on GitHub Releases (plus a Homebrew cask for local use). It runs as a host
> systemd service and is **not** containerized.

The deploy agent of the Specs golden images. GitHub emits a `deployment` event, and the app server
rolls itself forward, with no SSH and no CI access to servers.

This document is the plan for the first functional release. It picks up where
[specsops#13](https://github.com/specsnl/specsops/issues/13) leaves off. The epic tracks the outcome at a high level; this
document holds the detail. The architecture it implements is
[`specsops-golden-images/docs/architecture.md`](https://github.com/specsnl/specsops-golden-images/blob/main/docs/architecture.md)
§7–§12. Where this plan goes beyond that document, the difference is called out and listed under
[Cross-repo changes](#cross-repo-changes), so both documents can be brought back in line.

Structure, conventions and library choices deliberately mirror
[`specsnl/labelsync`](https://github.com/specsnl/labelsync), which in turn mirrors
[`specsnl/specs-cli`](https://github.com/specsnl/specs-cli). Where a pattern from those CLIs does
not fit a daemon, this document says so.

---

## Goals

- Turn a GitHub `deployment` event into a deployment on the one server that hosts its
  environment:
  1. `podman pull`
  2. retag
  3. `systemctl restart`
  4. wait for `/readyz`
  5. report the outcome back to GitHub
- **Trust nothing that arrives over the wire.** Every request is signature-checked before anything is parsed. Every
  repository has to be on an allowlist, and every environment is bound to one repository and one image. Every
  payload is schema-validated.
- **Never leave a server half-deployed without saying so.** A failed readiness check rolls back to the
  previous image. Every terminal state is reported to GitHub, including "rollback failed — manual action
  needed".
- **Stay within the privileges the role grants.** Sudo is limited to `podman pull`, `podman tag` and
  `systemctl restart app-*.service`. There is no shell and no writing of unit files.
- **Keep secrets encrypted at rest, always.** They are decrypted only in memory, at startup.
- **Operators get the same behaviour without GitHub.** `specsdeployd deploy` runs the same pipeline by hand, and
  `specsdeployd config validate` lets ansible-pull refuse a broken config before laying it down.

## Non-goals

These carry over from architecture §15, plus:

- **Rolling or multi-instance deployments.** One environment is one server, so there are no targets and no
  ordering.
- **Writing or reloading Quadlet units.** ansible-pull owns `/etc/containers/systemd`. The daemon only
  retags an image and restarts a unit.
- **Running database migrations or other hooks.** If an application needs them, it runs them itself on start and
  reflects them in `/readyz`.
- **Polling GitHub.** Servers never poll (§8). A deployment lost while the daemon is down stays
  `pending` in GitHub until someone redeploys. Recovering such deployments is listed under [Later](#later).
- **GitHub Enterprise Server.** `github.api_url` exists only as a test seam.
- **Public reachability.** The daemon listens on loopback only, and Caddy is the only way in.

---

## Where it fits

```text
GitHub ──deployment event──▶ Caddy (TLS, deploy.<server>.specs.dev/hook/github)
                               │ reverse_proxy 127.0.0.1:9000
                               ▼
                         specsdeployd serve ──sudo -n──▶ podman pull / podman tag / systemctl restart
                               │                                              │
                               │◀───────────── GET /readyz (via Caddy) ───────┘
                               ▼
GitHub ◀──deployment statuses (GitHub App installation token)
```

The host side is already in place, delivered by the
[`specsdeployd` role](https://github.com/specsnl/specsops-ansible-collection/tree/main/roles/specsdeployd)
in collection `0.3.x`. The following items are a **fixed contract**, and this design builds on them without changing them:

| Contract         | Value                                                                                  |
|------------------|----------------------------------------------------------------------------------------|
| Binary           | `/usr/bin/specsdeployd`, owned by dpkg; the unit guards on `ConditionPathExists`       |
| Start command    | `specsdeployd serve --listen 127.0.0.1:9000`                                           |
| User             | `specsdeployd`, a system user with nologin and no home                                 |
| Config directory | `/etc/specsdeployd/`, 0750, laid down by ansible-pull                                  |
| Config file      | `/etc/specsdeployd/config.json`                                                        |
| Sudo             | `/usr/bin/podman pull *`, `/usr/bin/systemctl restart app-*.service`, `NOPASSWD`       |
| Unit             | `Type=simple`, `Restart=on-failure`, `ProtectHome`, `PrivateTmp`, no `NoNewPrivileges` |
| Release assets   | `specsdeployd_<version>_linux_<arch>.deb` plus `checksums.txt`, matched by basename    |

This design needs **one addition** to the sudoers rule, `/usr/bin/podman tag *`, explained in
[Local image refs](#local-image-refs). The other changes it needs in other repositories are listed under
[Cross-repo changes](#cross-repo-changes).

---

## Concepts

### Environment

An environment is a GitHub environment (`{app}-{stage}`, for example `specs-production`) that is **hosted on this
server**. The config declares each one it hosts and binds it to four things:

- `repository`: the one repository allowed to deploy it;
- `image`: the one registry repository whose images it may run;
- `unit`: the Quadlet unit that runs it (`app-<app>-<stage>.service`);
- `probes.readyz`: the endpoint that decides whether a deployment succeeded.

The repository and image bindings close a gap in the architecture. An environment name in GitHub is
**free-form**: anyone with write access to *any* allowlisted repository can create a deployment whose
`environment` is `specs-production`, pointing at any image they like. Checking the allowlist alone would let
`specsnl/atlas` deploy an arbitrary image onto the specs production server. Both bindings are required.

### Deployment

A deployment is one `deployment` event (action `created`) routed to a hosted environment. It is identified by
`deployment.id`, which GitHub assigns and which increases monotonically. Its outcome is reported as a sequence of
[deployment statuses](#status-mapping).

### Local image refs

**The problem.** A Quadlet `.container` file names a fixed `Image=`. For a *branch* deploy
(`ghcr.io/specsnl/specs:main`), `podman pull` followed by a restart is enough, because the moving tag now resolves to the new
digest. For a *tag* deploy (`ghcr.io/specsnl/specs:v2026.02.01-1`), the Quadlet still names whatever tag
it was written with, so pulling and restarting changes nothing.

**The decision.** The Quadlet never names a registry image. It names a **stable local ref**, and the
daemon moves that ref:

```ini
# /etc/containers/systemd/app-specs-production.container (laid down by ansible-pull)
[Container]
Image=localhost/app-specs-production:deployed
Pull=never
```

| Ref                                  | Meaning                                                                                |
|--------------------------------------|----------------------------------------------------------------------------------------|
| `localhost/<unit-basename>:deployed` | The image the unit runs. It is moved on every deploy.                                  |
| `localhost/<unit-basename>:previous` | The image that was `:deployed` before the current deploy. This is the rollback target. |

Here `<unit-basename>` is the unit name without `.service`. The refs are derived from the config and never taken from
the payload.

This approach has several advantages:

- It works the same way for both strategies.
- It needs no `daemon-reload`, and no write access to `/etc/containers/systemd`.
- It doesn't conflict with ansible-pull, which owns the Quadlet files.
- It gives a rollback anchor for free.

The cost is one sudoers line, `/usr/bin/podman tag *`. Alternatives that were rejected:

- **A drop-in (`<unit>.container.d/image.conf`) plus `daemon-reload`.** This needs write access to
  root-owned unit directories and a broader sudo rule, and it has two writers for the same unit.
- **Supporting only the `branch` strategy in v1.** This ships half of §11.

Until the first deploy there is no `:deployed` image, so the unit fails to start. That is expected, and
it is the same dormant-until-deployed behaviour the unit shows before the binary is installed. Spike B2
confirms the behaviour on Ubuntu 26.04's Podman.

---

## Contracts

The four JSON Schemas below are the single source of truth. They are embedded in the binary (see
`internal/schema`), and copies are proposed for architecture §12 (see [Cross-repo changes](#cross-repo-changes)).
Hand-written Go rules add the checks a schema cannot express. Those rules are listed with each schema.

### Deployment payload (§12.1, unchanged)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://specs.dev/schemas/deployment-payload.schema.json",
  "title": "Specs Deployment Payload",
  "type": "object",
  "additionalProperties": false,
  "required": ["schema", "strategy", "ref", "image"],
  "properties": {
    "schema": { "type": "integer", "const": 1 },
    "strategy": { "type": "string", "enum": ["tag", "branch"] },
    "ref": { "type": "string", "minLength": 1 },
    "image": { "type": "string", "minLength": 1 },
    "app": { "type": "string" },
    "requested_by": { "type": "string" }
  }
}
```

The semantic rules are checked after routing, and a violation is reported as an `error` status:

| Rule                                                                     | Sentinel             |
|--------------------------------------------------------------------------|----------------------|
| `image` parses as a fully qualified reference (`distribution/reference`) | `ErrInvalidImageRef` |
| `image`'s repository equals the environment's `image`                    | `ErrImageNotAllowed` |
| `image` carries a tag or a digest, never neither                         | `ErrInvalidImageRef` |
| `ref` equals `deployment.ref`                                            | `ErrRefMismatch`     |
| For `strategy: tag`, the image tag equals `ref`                          | `ErrRefMismatch`     |
| If `app` is set, it equals the environment's `app`                       | `ErrAppMismatch`     |

### Deployment event subset (§12.2, unchanged)

The envelope is validated **without** its `payload`, because the payload is checked only after routing. A
malformed payload for an environment this server hosts deserves an `error` status. A malformed envelope
cannot even be attributed to a deployment, so it gets a `422` and nothing else.

On top of the §12.2 schema, the daemon requires `action: "created"` (other actions are answered with `204`) and
requires `repository.full_name` to be present.

### Configuration (§12.3, amended)

What changes from §12.3:

- **`environments[].repository` and `environments[].image` are new and required.** They are the
  bindings described in [Environment](#environment).
- **`github.app_id` is new and required.** It is the GitHub App that reports statuses. The key lives in the
  [secrets](#secrets-124-new).
- **`secrets` is new.** It holds the paths to the encrypted secrets and to the host identity. Both have defaults.
- **`environments[].environment_url` is new and optional.** It is passed through to GitHub.
- **`probes.readyz` is now required.** Success is gated on it (§10).
- **New probe fields `period_seconds` and `deadline_seconds`.** `timeout_seconds` is defined as the timeout of a
  single attempt, the same as Kubernetes' `timeoutSeconds`.
- **`additionalProperties: false` everywhere.** A typo such as `deadline_secondz` would otherwise quietly fall
  back to its default.
- **`github.api_url` is new, optional and undocumented for users.** It exists only as the end-to-end test seam.
- `schema` stays at `1`. No config file with schema 1 has been deployed yet, so this is the moment to change it.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://specs.dev/schemas/specsdeployd-config.schema.json",
  "type": "object",
  "additionalProperties": false,
  "required": ["schema", "server_id", "github", "environments"],
  "properties": {
    "schema": { "type": "integer", "const": 1 },
    "server_id": { "type": "string", "pattern": "^[a-z0-9][a-z0-9-]*$" },
    "github": {
      "type": "object",
      "additionalProperties": false,
      "required": ["allowed_repos", "app_id"],
      "properties": {
        "allowed_repos": {
          "type": "array",
          "minItems": 1,
          "uniqueItems": true,
          "items": { "type": "string", "pattern": "^[A-Za-z0-9-]+/[A-Za-z0-9._-]+$" }
        },
        "app_id": { "type": "integer", "minimum": 1 },
        "api_url": { "type": "string", "format": "uri" }
      }
    },
    "secrets": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "file": { "type": "string", "minLength": 1 },
        "identity_file": { "type": "string", "minLength": 1 }
      }
    },
    "environments": {
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["name", "app", "repository", "image", "unit", "probes"],
        "properties": {
          "name": { "type": "string", "pattern": "^[a-z0-9][a-z0-9-]*$" },
          "app": { "type": "string", "pattern": "^[a-z0-9][a-z0-9-]*$" },
          "repository": { "type": "string", "pattern": "^[A-Za-z0-9-]+/[A-Za-z0-9._-]+$" },
          "image": { "type": "string", "minLength": 1 },
          "unit": { "type": "string", "pattern": "^app-[a-z0-9][a-z0-9-]*\\.service$" },
          "environment_url": { "type": "string", "format": "uri" },
          "probes": {
            "type": "object",
            "additionalProperties": false,
            "required": ["readyz"],
            "properties": {
              "livez": { "$ref": "#/$defs/probe" },
              "readyz": { "$ref": "#/$defs/probe" }
            }
          }
        }
      }
    }
  },
  "$defs": {
    "probe": {
      "type": "object",
      "additionalProperties": false,
      "required": ["url"],
      "properties": {
        "url": { "type": "string", "format": "uri" },
        "timeout_seconds": { "type": "integer", "minimum": 1, "default": 5 },
        "period_seconds": { "type": "integer", "minimum": 1, "default": 2 },
        "deadline_seconds": { "type": "integer", "minimum": 1, "default": 180 }
      }
    }
  }
}
```

The semantic rules are checked at load time, and the first broken rule wins:

| Rule                                                                                     | Sentinel                  |
|------------------------------------------------------------------------------------------|---------------------------|
| Environment `name`s are unique                                                           | `ErrDuplicateEnvironment` |
| Environment `unit`s are unique                                                           | `ErrDuplicateUnit`        |
| Every `repository` is listed in `github.allowed_repos`                                   | `ErrRepositoryNotAllowed` |
| `image` is a repository reference, with no tag and no digest (`distribution/reference`)  | `ErrInvalidImageRef`      |
| `unit` matches the sudoers glob `app-*.service` (enforced by the schema pattern as well) | `ErrInvalidUnit`          |
| `timeout_seconds` ≤ `deadline_seconds`                                                   | `ErrInvalidProbe`         |
| Probe URLs use `http` or `https`                                                         | `ErrInvalidProbe`         |

The defaults apply after validation: `secrets.file` becomes `/etc/specsdeployd/secrets.age`,
`secrets.identity_file` becomes `/etc/specsdeployd/identity` (unless spike B1 changes it), `github.api_url`
becomes `https://api.github.com`, and the probe defaults are as in the schema.

#### Example: a server hosting two environments

```json
{
  "schema": 1,
  "server_id": "web-1",
  "github": {
    "allowed_repos": ["specsnl/specs", "specsnl/atlas"],
    "app_id": 1234567
  },
  "environments": [
    {
      "name": "specs-production",
      "app": "specs",
      "repository": "specsnl/specs",
      "image": "ghcr.io/specsnl/specs",
      "unit": "app-specs-production.service",
      "environment_url": "https://specs.dev",
      "probes": {
        "livez": { "url": "https://specs.dev/livez", "timeout_seconds": 5 },
        "readyz": { "url": "https://specs.dev/readyz", "timeout_seconds": 10, "deadline_seconds": 300 }
      }
    },
    {
      "name": "atlas-staging",
      "app": "atlas",
      "repository": "specsnl/atlas",
      "image": "ghcr.io/specsnl/atlas",
      "unit": "app-atlas-staging.service",
      "probes": {
        "readyz": { "url": "https://atlas.staging.specs.dev/readyz" }
      }
    }
  ]
}
```

`livez` is accepted and documented, but the daemon does not probe it. Success is gated on `readyz` only (§10).
Keeping `livez` in the config keeps the file compatible with §12.3, and a later monitoring feature can use it.

### Secrets (§12.4, new)

The plaintext of `secrets.age`. It is never written to disk on the host:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://specs.dev/schemas/specsdeployd-secrets.schema.json",
  "type": "object",
  "additionalProperties": false,
  "required": ["schema", "webhook_secrets", "github_app_private_key"],
  "properties": {
    "schema": { "type": "integer", "const": 1 },
    "webhook_secrets": {
      "type": "array",
      "minItems": 1,
      "maxItems": 2,
      "items": { "type": "string", "minLength": 32 }
    },
    "github_app_private_key": { "type": "string", "pattern": "^-----BEGIN (RSA )?PRIVATE KEY-----" }
  }
}
```

- **`webhook_secrets` is an array** so that a secret can be rotated without downtime. During a rotation the
  daemon accepts a signature made with any of the listed secrets. The limit is two: the current secret and the next one.
- **`github_app_private_key` is parsed at startup.** A key that doesn't parse stops the daemon with
  `ErrInvalidAppKey`, rather than failing at the first status report.

### Schemas are embedded

The four schemas live in `internal/schema/*.schema.json`. They are compiled once with
`santhosh-tekuri/jsonschema/v6`, which supports draft 2020-12, and its validation errors are mapped onto sentinels.
This deliberately departs from labelsync, which rejected a schema library for its YAML config. The §12 contracts
*are* JSON Schemas, and they are shared across three repositories. Keeping a hand-written copy of each schema in
Go would mean the two drift apart. The semantic rules above remain in Go, because a schema cannot express them.

A test asserts that the embedded schemas and the examples in this document agree: every example validates, and
every example of a broken rule fails with the expected sentinel.

---

## Secrets

### What is secret

| Secret                  | Why it is secret                                                                                                       |
|-------------------------|------------------------------------------------------------------------------------------------------------------------|
| Webhook secret(s)       | Anyone holding a webhook secret can trigger a deployment to any hosted environment, within the bindings.               |
| GitHub App private key  | Anyone holding it can mint installation tokens with `deployments: write` for every repository the App is installed on. |
| The host's age identity | Anyone holding it can decrypt the two secrets above.                                                                   |

### Always encrypted at rest

The two secrets travel as **one age-encrypted file**, `secrets.age`. It goes from git (specsops-ansible),
through ansible-pull, to `/etc/specsdeployd/secrets.age` without ever being decrypted in between. ansible-pull
never sees the plaintext and never needs a key.

The daemon decrypts the file **once, at startup**, using the host's identity file. It holds the plaintext
in memory only. **There is no plaintext code path at all:** no `webhook_secret` field in the config, no
environment variable, no `--insecure` flag. The tests generate throwaway age identities. That is cheap,
so no test needs an escape hatch either.

In memory, each secret is a `secrets.Secret`. Its `String()`, `LogValue()` and `MarshalJSON()` methods all
return `[redacted]`, so a secret cannot reach a log line, an error message or JSON output by accident. This
is labelsync's `Token` pattern. A change to the secrets requires a restart. ansible-pull restarts the unit
when it lays down a new file, and the restart decrypts the file again.

### Why age, not sops

- **The daemon only needs whole-file encryption.** sops adds value through field-level encryption, which gives
  structured files readable diffs, and through its many key backends. The daemon needs neither for two
  secrets.
- **sops would be a heavy dependency.** Decrypting a sops file means linking `getsops/sops/v3/decrypt`, which
  pulls in every key service (AWS, GCP and Azure KMS, Vault, PGP). That is a lot of binary and attack surface
  to add to a network-facing daemon.
- **age is small.** `filippo.io/age` is a small, audited library with no transitive dependencies worth
  mentioning.
- **The daemon reads both identity formats.** It accepts native age X25519 identities (`AGE-SECRET-KEY-1…`)
  and OpenSSH `ssh-ed25519` private keys (through `filippo.io/age/agessh`). That leaves the choice of host
  identity to operations (spike B1).

### Why not 1Password at runtime

The daemon needs its secrets when it starts, and it starts after a reboot, after ansible-pull changes its
config, or after a crash. Fetching them from 1Password with a service account at that moment has three problems:

- It adds a **runtime dependency on the network and on 1Password's availability**, at exactly the moment a deploy
  may be most needed.
- It needs **the service-account token on disk**, which just moves the problem one level down.
- It adds the `op` CLI and a rate-limited API to every server.

1Password belongs on the **authoring side**. It is the source of truth for the GitHub App private key and for
a **break-glass age identity**, whose recipient is included in every `secrets.age`. That way a person can
always decrypt, inspect and re-encrypt the file.

### Recipients and authoring

Every `secrets.age` is encrypted to at least two recipients: the host's identity and the team's break-glass
recipient. Authoring uses the upstream `age` CLI, and the daemon ships no encrypt command:

```sh
age --encrypt --armor \
    --recipient "$(cat hosts/web-1.age.pub)" \
    --recipient "$(cat team/break-glass.age.pub)" \
    --output hosts/web-1/secrets.age secrets.json
```

Rotation, re-keying a host, and adding a host are runbook pages (G3).

### Host identity (spike B1)

The design stays neutral between the two options, because the daemon reads both formats:

| Option                                                                     | For                                                                                | Against                                                                                                                                                                |
|----------------------------------------------------------------------------|------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Dedicated age identity, `/etc/specsdeployd/identity` (0400 `specsdeployd`) | Its only purpose is this daemon, and it is easy to rotate.                         | Bootstrap: the recipient has to be generated on the host (or by OpenTofu or cloud-init) and registered in specsops-ansible before any secrets can be encrypted for it. |
| SSH host key, given to the unit with `LoadCredential=`                     | The recipient is just `ssh-keyscan`, so bootstrap is trivial (the agenix pattern). | It hands the host's SSH identity to a network-facing daemon. It needs a unit change in the role. The key must be unique per VM, so it cannot be baked into the image.  |

Spike B1 decides which is the default, how the recipient gets back into specsops-ansible, and how
rotation works. Its outcome is written back here and fed into
[specsops#14](https://github.com/specsnl/specsops/issues/14).

At startup, the daemon refuses an identity file that is group- or world-readable, with `ErrIdentityPermissions`.

---

## Webhook handling

### Topology (spike B3)

Every server must receive every deployment event and decide for itself whether it hosts the environment. That shapes how webhooks are delivered:

- **A GitHub App has exactly one webhook URL.** It cannot fan out to several servers, so the App is used
  **only for reporting statuses**.
- **Delivery uses one org webhook per server**, subscribed to the `deployment` event only, with content type
  `application/json` and a per-server secret. It points at that server's hostname, for example
  `https://deploy.web-1.specs.dev/hook/github`. The single `deploy.specs.dev` name in architecture §8 works
  only while there is one app server.
- GitHub allows up to 20 webhooks per event per organisation. That is the ceiling on app servers before this
  approach needs a relay, and it is comfortably above the Phase 1 scale.

Spike B3 confirms these points, and whether repository-level webhooks would be better than org-level ones.

### Request pipeline

A request goes through a fixed sequence of steps. **Each step can stop the request, and none runs before
the one above it.** In particular, **nothing is parsed before the signature has been verified.**

| #  | Step                                                                                                                 | On failure                    | Status posted to GitHub        |
|----|----------------------------------------------------------------------------------------------------------------------|-------------------------------|--------------------------------|
| 1  | Route: `POST /hook/github`                                                                                           | `404` / `405`                 | —                              |
| 2  | `Content-Type: application/json`                                                                                     | `415`                         | —                              |
| 3  | Body at most 1 MiB (`http.MaxBytesReader`)                                                                           | `413`                         | —                              |
| 4  | `X-Hub-Signature-256` HMAC-SHA256 over the raw body, compared in constant time against each webhook secret           | `401` `ErrBadSignature`       | —                              |
| 5  | `X-GitHub-Event`: `ping` → `200 pong`; anything other than `deployment` → `204`                                      | —                             | —                              |
| 6  | Decode the envelope and validate it against §12.2 (without the payload); require `action: created` (otherwise `204`) | `422` `ErrInvalidEvent`       | —                              |
| 7  | Deduplicate by `X-GitHub-Delivery` and by `deployment.id` (bounded LRU of 1024 entries)                              | `200 duplicate`               | —                              |
| 8  | `repository.full_name` is in `github.allowed_repos`                                                                  | `403` `ErrRepoNotAllowed`     | —                              |
| 9  | `deployment.environment` is hosted here                                                                              | `202 ignored`                 | — (**another server owns it**) |
| 10 | The environment's `repository` equals `repository.full_name`                                                         | `422` `ErrRepositoryMismatch` | `error`                        |
| 11 | Payload against §12.1, then the semantic rules                                                                       | `422` (sentinel)              | `error`                        |
| 12 | Hand the deployment to the environment's queue                                                                       | `503` `ErrShuttingDown`       | `pending` on success           |
| —  | Accepted                                                                                                             | `202 accepted`                | `pending`                      |

Every response body is small JSON, `{"status": "...", "error_kind": "..."}`, so a person looking at GitHub's delivery log sees
the reason. When the daemon rejects a request, it never includes the part of the payload it rejected.

Steps 1–12 never block on the network or on sudo. The handler returns well within GitHub's 10-second
delivery timeout, and the deployment itself runs in the background.

### Replay and duplicates

GitHub does not sign a timestamp, so a captured request stays valid forever. Three defences apply:

- **Deduplication.** Step 7 remembers delivery IDs and deployment IDs.
- **Monotonic deployment IDs.** The [queue](#queue) never runs a deployment whose ID is lower than one it has already started
  for that environment.
- **TLS.** Every delivery goes through Caddy over TLS.

A replay therefore can at most re-trigger a deployment of an image that was already deployed. Redeliveries after
a daemon restart are treated as new, which is the right behaviour: the earlier attempt may have been lost.

---

## Queue

Each hosted environment has one worker. The queue enforces these rules:

- **At most one running deployment per environment, and at most one pending.** A newer deployment that
  arrives while another is pending replaces it. The replaced deployment is reported `inactive` with the
  description "superseded by #N". Branch tracking can produce a burst of deployments, and only the newest one
  matters.
- **IDs only go forward.** A deployment whose ID is lower than the running one or the last started one is reported `inactive`
  ("superseded by #N") and never runs. This handles deliveries that arrive out of order.
- **Different environments run concurrently.** They share nothing except the image store, and Podman handles
  concurrent pulls.

**Shutdown** (`SIGTERM` or `SIGINT`) proceeds like this:

1. The server stops accepting requests, so new deliveries get `503` and GitHub retries them later.
2. Pending deployments are reported `error` ("specsdeployd stopped before running this deployment").
3. Running deployments continue up to their **point of no return**, which is the retag of `:deployed`:
   - **Before the retag,** nothing on the host has changed. The deployment is aborted and reported `error`.
   - **After the retag,** the restart is completed. A new image behind a stale unit would start
     unexpectedly on the next reboot. The readiness wait is then cut short, and the deployment is reported `error` ("interrupted
     before readiness was confirmed"). There is no rollback on shutdown.
4. The daemon exits within `--shutdown-timeout`, which defaults to 60s. The unit's `TimeoutStopSec` must be at least that
   plus a margin. This is a role change (see [Cross-repo changes](#cross-repo-changes)).

---

## Deploy pipeline

### Steps

`deploy.Plan(env, payload)` is a pure function. It returns the steps below, and the executor runs them.
`localhost/X` stands for the environment's local ref.

| # | Step              | Command (all `sudo -n`, no shell)                               | Default timeout    | Failure → outcome                                                                              |
|---|-------------------|-----------------------------------------------------------------|--------------------|------------------------------------------------------------------------------------------------|
| 1 | Pull              | `/usr/bin/podman pull <image>`                                  | 10m                | `failure`, nothing changed                                                                     |
| 2 | Preserve previous | `/usr/bin/podman tag localhost/X:deployed localhost/X:previous` | 30s                | skipped if `:deployed` does not exist yet (first deploy); otherwise `failure`, nothing changed |
| 3 | Promote           | `/usr/bin/podman tag <image> localhost/X:deployed`              | 30s                | `failure`, **point of no return**, roll back                                                   |
| 4 | Restart           | `/usr/bin/systemctl restart <unit>`                             | 3m                 | roll back                                                                                      |
| 5 | Ready             | poll `readyz` (see [Probes](#probes))                           | `deadline_seconds` | roll back                                                                                      |

The executor emits one event per step, with fields `step`, `duration_ms` and `error_kind`. Events become log lines, and the
last one becomes the GitHub status description, for example `web-1: ready after 23s (v2026.02.01-1)`.

### Rollback

Rollback runs when step 3, 4 or 5 fails and a `:previous` image exists. It consists of:

1. `podman tag localhost/X:previous localhost/X:deployed`
2. `systemctl restart <unit>`
3. poll `readyz` again, with the same deadline

| Outcome                   | Status    | Description (example)                                            |
|---------------------------|-----------|------------------------------------------------------------------|
| Succeeded                 | `success` | `web-1: ready after 23s`                                         |
| Failed, nothing changed   | `failure` | `web-1: pull failed (image_pull_failed)`                         |
| Failed, rolled back       | `failure` | `web-1: not ready after 180s; rolled back to the previous image` |
| Failed, no previous image | `failure` | `web-1: not ready after 180s; no previous image to roll back to` |
| Failed, rollback failed   | `failure` | `web-1: not ready; ROLLBACK FAILED, manual action needed`        |

A failed rollback is also logged at `ERROR` with every step's captured output, because it is the one outcome
that needs a person.

### Probes

The readiness poller works like this:

- It sends `GET <readyz.url>` every `period_seconds`, each attempt bounded by `timeout_seconds`, until the
  endpoint answers `200` or `deadline_seconds` has passed since the restart returned.
- It does not follow redirects, because a redirect to a login page must not count as ready. Only a `200` counts.
- It verifies TLS normally, because the probe goes through the host's Caddy with a real certificate.
- It is tested against `httptest` with an injected clock. No test sleeps.

`systemctl restart` returns after the old container has stopped and the new one has started. Any `/readyz` answer received
after that point therefore comes from the new container.

### Exec safety

`internal/runner` is the only package that starts a process.

- **Fixed commands.** It uses `exec.CommandContext` with the argument vector `sudo -n <absolute path> <args…>`. There is no
  shell. The absolute paths `/usr/bin/podman` and `/usr/bin/systemctl` are constants, and a test checks them against a copy of the role's
  sudoers lines kept in `testdata/sudoers`.
- **Guarded arguments.** Every argument has already been validated: the image by `distribution/reference`, the unit by the config pattern,
  and the local refs derived from the unit. As defence in depth, the runner refuses any argument that starts with `-`
  (`ErrUnsafeArgument`), so a crafted image string can never be read as an option.
- **Clean environment.** The environment passed to the command is `PATH=/usr/sbin:/usr/bin:/sbin:/bin`, `LANG=C.UTF-8`, and nothing else.
- **Bounded output and time.** stdout and stderr are captured together, capped at 64 KiB, and attached to the step's
  error. Each command has its own timeout, and when a context is cancelled the process group is killed.
- **Fake runner for tests.** `runner.Fake` records calls in order and returns scripted results. Every deploy-pipeline test uses it.

---

## GitHub reporting

### Authentication

The GitHub App is created by hand once for the org, following a runbook page. It has two permissions,
**Deployments: read and write** and **Metadata: read**, and is installed on the allowlisted
repositories. It uses no webhook of its own; see [Topology](#topology-spike-b3).

- **`app_id` from the config, plus the private key from the secrets.** Together they produce an app JWT, via
  `bradleyfalzon/ghinstallation/v2`.
- **One installation per repository.** The installation for a repository is looked up once with
  `GET /repos/{owner}/{repo}/installation` and cached for the life of the process. Installation tokens are
  refreshed by the transport.
- **The API client is `google/go-github`,** the same library as labelsync. It uses the same
  functional options (`WithBaseURL`, `WithHTTPClient`, `WithClock`, `WithRetries`) and the same `retryTransport`,
  which retries a `5xx` three times, doubling the delay from 500ms.

### Status mapping

| When                                 | State                 | Notes                                                                                                         |
|--------------------------------------|-----------------------|---------------------------------------------------------------------------------------------------------------|
| Accepted into the queue              | `pending`             | `queued on <server_id>`                                                                                       |
| The worker starts it                 | `in_progress`         | §9 names only `pending`/`success`/`failure`; `in_progress` is added so the GitHub UI shows that it is running |
| Superseded or out of order           | `inactive`            | `superseded by #N`                                                                                            |
| Rejected after routing (steps 10–11) | `error`               | `error_kind` in the description                                                                               |
| Pipeline outcome                     | `success` / `failure` | see [Rollback](#rollback); `success` sets `auto_inactive: true`                                               |
| Shutdown before completion           | `error`               | see [Queue](#queue)                                                                                           |

Every status sets `environment`, which is the deployment's own, and `environment_url` when it is configured.
Descriptions are cut to GitHub's 140-character limit, rune-safe, and always start with `server_id`.

### Failure handling

**Reporting is best effort and never changes what happens on the host.** A status that cannot be posted after the
client's retries is logged at `WARN` with `error_kind: status_report_failed`, and the pipeline carries on.
The trade-off is explicit: the host is the source of truth, GitHub's view may lag, and the journal
always has the full record.

---

## CLI

| Command                                                  | Purpose                                                                                                                                                                                      |
|----------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `specsdeployd serve --listen 127.0.0.1:9000`             | The daemon. Refuses a listen address that is not loopback (`ErrNonLoopbackListen`).                                                                                                          |
| `specsdeployd config validate [path]`                    | Validates a config (schema plus semantic rules). With a readable identity file, it also decrypts and validates the secrets. Built for Ansible's `validate: specsdeployd config validate %s`. |
| `specsdeployd deploy --environment <name> --image <ref>` | Runs the same pipeline (pull, retag, restart, readyz, rollback) without GitHub. Refuses while `specsdeployd.service` is active, unless given `--force`.                                      |
| `specsdeployd version` / `--version`                     | Unchanged from `v0.1.0`.                                                                                                                                                                     |

Global flags:

- `--config`, default `/etc/specsdeployd/config.json`;
- `--log-level` (`debug|info|warn|error`, default `info`);
- `--log-format` (`json|text`, default `json`);
- `-o/--output` (`pretty|json`), for the one-shot commands.

`serve` also takes `--shutdown-timeout`. The `deploy` command checks whether the service is active with `systemctl is-active specsdeployd.service`, which needs no
privilege. It exists to catch the obvious mistake, not to act as a lock.

### Exit codes

**The only exit codes are `0` and `1`.** labelsync's bitmask scheme is for reconcilers that report drift. None of these
commands has an outcome that a script needs to tell apart beyond success and failure. The reason
is always in `error_kind`, both in `--output json` and in the log. As in labelsync and specs-cli, `main.go` is the only
place that calls `os.Exit`, and root sets `SilenceUsage` and `SilenceErrors`.

## HTTP surface

| Route               | Purpose                                                                                                                      |
|---------------------|------------------------------------------------------------------------------------------------------------------------------|
| `POST /hook/github` | The only route Caddy proxies.                                                                                                |
| `GET /livez`        | `200` while the process is up. It depends on nothing.                                                                        |
| `GET /readyz`       | `200` once the config and secrets are loaded, the App key has parsed and the workers are running. `503` while shutting down. |

The server sets `ReadHeaderTimeout` to 5s, `ReadTimeout` to 15s, `WriteTimeout` to 15s, `IdleTimeout` to 60s, and `MaxHeaderBytes` to 64 KiB. Caddy
must proxy **only** `/hook/github`. `/livez` and `/readyz` are for local monitoring through the loopback address.

---

## Package structure

```text
main.go                        the only os.Exit; signal.NotifyContext(SIGINT, SIGTERM)
internal/
  specsdeployd/                AppName, default paths, sentinel errors + KindOf()
  cmd/                         one file per command; app.go holds the App (streams, Now, Runner, Prober, GitHub options)
  schema/                      embedded *.schema.json, compiled once; errors → sentinels
  config/                      load, decode, defaults, semantic rules, local refs
  secrets/                     age decrypt, identity parsing + permission check, Secret type
  event/                       decode a deployment event, semantic checks (pure)
  webhook/                     the http.Handler: the request pipeline, steps 1–12
  queue/                       per-environment worker, supersede, shutdown
  deploy/                      Plan() (pure) + Executor, rollback, outcomes
  runner/                      Runner interface, sudo runner, Fake
  probe/                       readyz poller
  github/                      go-github + App auth, retryTransport, Reporter
  util/exit/                   exit codes and the silent *exit.Err carrier
  util/output/                 pretty|json Writer for the one-shot commands; slog setup
```

### Boundary rule

`config`, `event`, and the plan half of `deploy` **never import `runner`, `probe`, `github` or
`net/http`**. They take plain structs and return plain structs, so the logic that decides *what* to do can be tested
without faking anything. This is the equivalent of labelsync's rule for `plan` and `palette`, and `AGENTS.md` enforces
it through review. A test that parses the imports of those packages enforces it in CI.

Seams, which labelsync uses too:

- **`App` fields**, which tests replace: `Now func() time.Time`, `Runner runner.Runner`, `Prober probe.Prober`,
  `GitHub []github.Option`, plus the streams.
- **Narrow consumer interfaces**, such as `deploy.Reporter` and `queue.Deployer`, defined in the package that consumes them.
- **An injected `Clock`** for the queue, the probe and the retries. No test sleeps for real.

---

## Error handling

Sentinel errors live in `internal/specsdeployd/errors.go`. They are always wrapped with `%w`, and mapped to a stable
snake_case `error_kind` by `KindOf()`. Kinds are a public contract: new kinds may be added, but existing ones are
never renamed. Adding a sentinel therefore touches three places, just as in labelsync: `KindOf`, the table on the
error-handling architecture page, and the `allSentinels` test table. A test that parses the source checks that every
exported `Err*` appears in that table.

| Group   | Sentinels (kind = snake_case of the name without `Err`)                                                                                                                              |
|---------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Config  | `ErrConfigNotFound`, `ErrInvalidConfig`, `ErrDuplicateEnvironment`, `ErrDuplicateUnit`, `ErrRepositoryNotAllowed`, `ErrInvalidImageRef`, `ErrInvalidUnit`, `ErrInvalidProbe`         |
| Secrets | `ErrSecretsNotFound`, `ErrIdentityNotFound`, `ErrIdentityPermissions`, `ErrDecrypt`, `ErrInvalidSecrets`, `ErrInvalidAppKey`                                                         |
| Webhook | `ErrBadSignature`, `ErrInvalidEvent`, `ErrRepoNotAllowed`, `ErrRepositoryMismatch`, `ErrInvalidPayload`, `ErrImageNotAllowed`, `ErrRefMismatch`, `ErrAppMismatch`, `ErrShuttingDown` |
| Deploy  | `ErrUnsafeArgument`, `ErrCommandTimeout`, `ErrImagePullFailed`, `ErrImageTagFailed`, `ErrRestartFailed`, `ErrNotReady`, `ErrRollbackFailed`, `ErrSuperseded`, `ErrServiceActive`     |
| GitHub  | `ErrGitHubAuth`, `ErrStatusReportFailed`                                                                                                                                             |
| CLI     | `ErrNonLoopbackListen`, `ErrUnknownEnvironment`                                                                                                                                      |

## Logging

Unlike labelsync's CLI, where slog is silent by default, the daemon **logs by default**:

- **Format.** It writes slog to **stdout** (JSON by default). journald collects it through the unit's default output. There is no log file,
  and no rotation of its own.
- **Levels.**
  - `info`: every accepted, ignored and finished deployment, plus every step event.
  - `warn`: status reports that failed and requests that were rejected.
  - `error`: a failed rollback, and a failure to start.
  - `debug`: request headers, except the signature.
- **Standard attribute keys.** `delivery_id`, `deployment_id`, `environment`, `repository`, `image`, `step`,
  `duration_ms`, `error_kind` and `server_id`. They are defined as constants in one place and documented in the
  logging architecture page.
- **What is never logged.** Secrets (enforced by type), request bodies, and the signature header.

The one-shot commands (`config validate`, `deploy`) write their result through `output.Writer` (pretty or
NDJSON, as in labelsync) and their narration through slog on stderr.

---

## Security

The threat model has its own architecture page. In summary:

| Threat                                                                                    | Mitigation                                                                                                                                                                         |
|-------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| A forged delivery                                                                         | HMAC checked before parsing; constant-time compare; loopback-only listener behind Caddy                                                                                            |
| An allowlisted repo deploying to another app's environment, or running an arbitrary image | The per-environment `repository` and `image` bindings                                                                                                                              |
| A replayed delivery                                                                       | Deduplication, monotonic IDs, TLS                                                                                                                                                  |
| Command or option injection through `image`                                               | `distribution/reference` parsing, argument vector with no shell, leading-`-` guard, absolute sudo paths                                                                            |
| Privilege escalation through sudo                                                         | Only `podman pull *`, `podman tag *`, `systemctl restart app-*.service`. The unit pattern mirrors the glob.                                                                        |
| Secrets leaking through logs or errors                                                    | The `Secret` type, age at rest, decryption in memory only                                                                                                                          |
| The host identity leaking                                                                 | The file must be mode 0400/0600, owned by the daemon user or by root through `LoadCredential` (spike B1)                                                                           |
| A compromised daemon                                                                      | Its blast radius equals its sudo rights: it can deploy any image to any `app-*` unit. That is inherent to the role, and it is why the daemon stays small and has few dependencies. |

`podman tag *` is as permissive as `podman pull *`, because sudoers wildcards match spaces. It does not widen what a
compromised daemon can do, since anything able to pull and restart can already run any image it chooses. It is
still recorded here because it adds to the sudo surface.

---

## Testing strategy

**Tools.** Only the stdlib `testing` package, with table-driven tests, `net/http/httptest` and an injected clock, just as in labelsync. No
assertion library. The build tags are `integration` (tests that need no network but do need real files or
processes) and none for unit tests. The e2e suite is bats, as labelsync's image tests are.

| Package         | What the tests assert                                                                                                                                                                                                                                                                                                                                       |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `specsdeployd`  | Every exported `Err*` is in `allSentinels`; kinds are unique; `KindOf` handles wrapped errors                                                                                                                                                                                                                                                               |
| `schema`        | Each schema compiles; the design examples validate; each broken-rule example fails with the expected sentinel                                                                                                                                                                                                                                               |
| `config`        | Golden files for normalised configs; every semantic rule, with a table of broken configs; defaults                                                                                                                                                                                                                                                          |
| `secrets`       | Round trip with generated X25519 and ssh-ed25519 identities; wrong identity → `ErrDecrypt`; loose permissions → `ErrIdentityPermissions`; `Secret` never shows up in `fmt`, slog or JSON                                                                                                                                                                    |
| `event`         | Every semantic rule; the envelope is validated without the payload; the `action` filter                                                                                                                                                                                                                                                                     |
| `webhook`       | Every row of the request pipeline table, including status codes, response bodies, and whether a status was posted; nothing is parsed before the HMAC check                                                                                                                                                                                                  |
| `queue`         | Supersede, monotonic IDs, concurrent environments, both shutdown branches, all with a fake clock and a fake deployer                                                                                                                                                                                                                                        |
| `deploy`        | `Plan()` golden files per strategy; every rollback outcome, using `runner.Fake` and a fake prober; the point of no return                                                                                                                                                                                                                                   |
| `runner`        | Command vectors match `testdata/sudoers`; the leading-`-` guard; timeouts kill the process group (integration tag, real `sleep`)                                                                                                                                                                                                                            |
| `probe`         | Success, non-200, a redirect that is not followed, per-attempt timeout, deadline, all with an injected clock                                                                                                                                                                                                                                                |
| `github`        | httptest fakes: installation lookup and caching, token refresh, `5xx` retry, status bodies, description truncation                                                                                                                                                                                                                                          |
| `cmd`           | End to end through the command tree: signed webhooks into `serve`, a fake runner, a fake GitHub (`harness_test.go`, as in labelsync); `config validate` and `deploy` output and exit codes                                                                                                                                                                  |
| top level       | `release_test.go`: the deb contains only `/usr/bin/specsdeployd`; asset names match the role's `specsdeployd_{v}_linux_{arch}.deb` and `checksums.txt`; ldflags path; cask anchors. `ci_test.go`: the required jobs exist and the e2e job runs the bats suite                                                                                               |
| `test/e2e.bats` | On an `ubuntu-24.04` runner with real Podman, Quadlet and systemd: a local registry, a tiny test app image in a good and a bad variant, a fake GitHub, and the built binary under a sudoers file equivalent to the role's. It asserts that a deploy switches the running image, that the bad variant rolls back, and that the expected statuses were posted |

## Dependencies

| Module                              | Why                                                                           |
|-------------------------------------|-------------------------------------------------------------------------------|
| `spf13/cobra`                       | Already in go.mod; the same as labelsync and specs-cli                        |
| `google/go-github` (latest major)   | Typed API, the same as labelsync                                              |
| `bradleyfalzon/ghinstallation/v2`   | GitHub App JWT and installation-token transport; widely used; a small surface |
| `filippo.io/age`                    | Secrets at rest (X25519 and `agessh`)                                         |
| `santhosh-tekuri/jsonschema/v6`     | Draft 2020-12, the §12 contracts as they are written                          |
| `github.com/distribution/reference` | The canonical grammar for image references; small                             |

Everything else uses the standard library: `net/http`, `log/slog`, `os/exec`, `crypto/hmac`. The library-decisions page records the alternatives
that were rejected: sops, a hand-written schema validator, `jferrl/go-githubauth`, and an HTTP router.

## Distribution

Distribution is unchanged from `v0.1.0`:

- GoReleaser builds the `.deb` for amd64 and arm64 (containing **only** `/usr/bin/specsdeployd`), the tarballs, `checksums.txt`, and the
  `specsdeployd` and `specsdeployd@rc` casks.
- `release_test.go` pins down everything in that list that the role depends on.
- The first functional release is **`v0.2.0`**. `v0.2.0-rc.N` tags are deployed to a staging golden-app VM first,
  by setting `specsdeployd_version`. `v1.0.0` is cut once production has run it for a while.

## Documentation

As in labelsync, there is a Hugo site on the Hextra theme at `docs/`, deployed to GitHub Pages at
`specsdeployd.specs.dev`, the same pattern as `labelsync.specs.dev` and `cli.specs.dev`. This file stays outside the published tree, as the plan.

| Section         | Pages                                                                                                                                                                                                                              |
|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `architecture/` | overview, contracts, configuration, secrets, webhook, queue, deploy-pipeline, probes, github-client, error-handling, logging, security, versioning, distribution, library-decisions                                                |
| `usage/`        | getting-started (host setup), configuration, secrets (authoring, rotation, break-glass), commands, github-setup (App and webhooks), runbook (troubleshooting with `journalctl`, failed deploys, failed rollback, rotating secrets) |

As `AGENTS.md` requires, every feature PR updates its architecture page in the same change.

---

## Cross-repo changes

This repository ships only the binary. The work below lives in other repositories and is drafted in
`docs/backlog/cross-repo/`.

| Repository                    | Change                                                                                                                                                                                                                                                                                                                      |
|-------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `specsops-golden-images`      | Amend architecture §8 (a webhook per server), §9 (sudoers and `in_progress`), §11 (local refs), §12.3 (as above), and add §12.4                                                                                                                                                                                             |
| `specsops-ansible-collection` | sudoers: add `/usr/bin/podman tag *`; unit: `TimeoutStopSec`, plus `LoadCredential` if B1 chooses the SSH host key; config directory ownership `root:specsdeployd` so the daemon cannot rewrite its own config; molecule: the unit starts and stays up with a fixture config; release `0.4.0`                               |
| `specsops-ansible`            | The Caddy site `deploy.<server>.specs.dev`, which proxies **only** `/hook/github`; the Quadlet template with `Image=localhost/<unit>:deployed` and `Pull=never`; the `config.json` template with `validate: specsdeployd config validate %s`; delivery of `secrets.age`; the host identity (per B1); `specsdeployd_version` |
| GitHub org (manual, runbook)  | Create the GitHub App and install it; one org webhook per server; registry credentials for root, per spike B4                                                                                                                                                                                                               |

---

## Milestones

| #  | Milestone         | Scope                                                                                             | Network? |
|----|-------------------|---------------------------------------------------------------------------------------------------|----------|
| M0 | `M0 · Foundation` | AGENTS.md, docs site, rulesets, sentinels, output, the App spine, release contract tests, spikes  | no       |
| M1 | `M1 · Contracts`  | Schemas, config, secrets, event parsing, `config validate`, amending the architecture             | no       |
| M2 | `M2 · Receive`    | `serve`, HMAC, the request pipeline, the queue                                                    | loopback |
| M3 | `M3 · Deploy`     | Runner, probe, pipeline, rollback, `deploy`                                                       | local    |
| M4 | `M4 · Report`     | GitHub App client, status reporter                                                                | yes      |
| M5 | `M5 · Ship`       | `serve` end to end, e2e on a runner, docs, the `v0.2.0` release, rollout across repos, acceptance | yes      |

M1 and most of M3 are the interesting core, and they need no GitHub access at all. They are worth building and
testing first.

### Build order

Work is tracked as seven epics, one per subsystem, and each epic has sub-issues that carry a milestone. The epics
themselves carry no milestone, because several span more than one. Each epic is also a sub-issue of
[specsops#13](https://github.com/specsnl/specsops/issues/13).

The critical path is **sentinels (A4) → schemas (C1) → config (C2) → event (C6) → deploy executor (E3) →
rollback (E4) → status reporter (F2) → `serve` end to end (G1) → e2e (G2) → release candidate (G4) → acceptance (G6)**.
The waves below come from the `depends_on` graph in `docs/backlog/`. An item can start as soon as everything it
depends on has merged.

| Wave | Available in parallel                                                                                                                   |
|------|-----------------------------------------------------------------------------------------------------------------------------------------|
| 1    | AGENTS.md (A1), docs site (A2), rulesets and CI (A3), sentinels (A4), chores (A6), release contract tests (A7), all four spikes (B1–B4) |
| 2    | App spine (A5), schemas (C1), runner (E1), probe (E2), collection role change (X1)                                                      |
| 3    | Config (C2), secrets (C4), architecture amendment (C3)                                                                                  |
| 4    | Event (C6), `config validate` (C5), `serve` (D1), HMAC (D2), GitHub client (F1)                                                         |
| 5    | Request pipeline (D3), queue (D4), deploy executor (E3), GitHub setup (G5), ansible-pull changes (X2)                                   |
| 6    | Rollback (E4)                                                                                                                           |
| 7    | `deploy` command (E5), status reporter (F2)                                                                                             |
| 8    | `serve` end to end (G1)                                                                                                                 |
| 9    | E2E (G2), docs (G3)                                                                                                                     |
| 10   | `v0.2.0-rc.1` (G4)                                                                                                                      |
| 11   | Acceptance on a golden-app VM, then `v0.2.0` (G6)                                                                                       |

The spikes are the ones to start straight away. They need no code, they can all run in parallel, and each of
them gates a contract: B1 the identity path, B2 the Quadlet template and the sudoers line, B3 the webhook setup,
and B4 pulling private images.

The GitHub client (F1) depends only on the foundation and on the config and secrets types, so it can run in parallel with
the receiver and deploy epics. Only the status reporter (F2) waits for the pipeline's outcome types.

## Open questions

1. **The host identity source.** Resolved by spike B1.
2. **Webhook topology and hostnames.** A webhook per server at `deploy.<server>.specs.dev`, or a
   relay? Resolved by spike B3.
3. **Pulling private GHCR images.** root's `/etc/containers/auth.json`, laid down by ansible-pull, or a
   token with `packages: read` minted by the App. `sudo` resets the environment, so `REGISTRY_AUTH_FILE`
   cannot be passed through. Resolved by spike B4.
4. **Is `podman tag` enough on Ubuntu 26.04's Podman?** Does `Pull=never` behave with a local ref, and is
   "image not known" from step 2 reliably distinguishable? Resolved by spike B2.
5. **Per-environment command timeouts.** Are the defaults (10m pull, 3m restart) enough for large images, or
   should `timeouts` go into the config? The defaults stand until a real image proves otherwise.

## Later

- On startup, find deployments for hosted environments that are still `pending` or `in_progress`, and report them `error`. This
  would be the first time the daemon reads from GitHub unprompted, so it waits until there is a real need.
- Reloading the config on `SIGHUP`, instead of restarting.
- Publishing the schemas at their `$id` URLs under `https://specs.dev/schemas/`.
- Metrics.
- Checking that the running container's digest equals the pulled one. This needs `podman inspect` under
  sudo.
- Pruning old images from the store.

## References

- [specsops#13](https://github.com/specsnl/specsops/issues/13): the epic
- [specsops#14](https://github.com/specsnl/specsops/issues/14): ansible-pull at runtime
- [Architecture §7–§12](https://github.com/specsnl/specsops-golden-images/blob/main/docs/architecture.md)
- [`specsdeployd` role](https://github.com/specsnl/specsops-ansible-collection/tree/main/roles/specsdeployd)
- [labelsync `docs/design.md`](https://github.com/specsnl/labelsync/blob/main/docs/design.md): the structure this document follows
- GitHub docs: [webhook events: deployment](https://docs.github.com/en/webhooks/webhook-events-and-payloads#deployment),
  [validating deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries),
  [deployment statuses](https://docs.github.com/en/rest/deployments/statuses),
  [authenticating as a GitHub App installation](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation)
- [age](https://github.com/FiloSottile/age), [Podman Quadlet](https://docs.podman.io/en/latest/markdown/podman-systemd.unit.5.html)
