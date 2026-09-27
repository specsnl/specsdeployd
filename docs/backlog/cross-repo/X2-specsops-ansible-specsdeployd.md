---
id: X2
title: "specsdeployd per-host config: Caddy route, Quadlet, config.json, secrets.age and identity"
type: Task
repo: specsnl/specsops-ansible
parent: specsnl/specsops#14
depends_on: [B1, B2, B3, B4, C5, X1]
---

**Repo:** `specsnl/specsops-ansible` · **Driven by:** specsnl/specsdeployd [`docs/design.md` § Cross-repo changes](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#cross-repo-changes)

## Why

The golden image bakes the daemon's scaffolding. Everything specific to a host is ansible-pull's job (architecture §13):
the Caddy site, the Quadlet units, `config.json`, the encrypted secrets, the identity, and which binary version to install.
Nothing in this list exists yet, and specsdeployd v0.2.0 cannot run without it.

## Scope

- [ ] **Caddy:** a site for the server's hostname (per spike B3, for example `deploy.web-1.specs.dev`) that proxies **only**
      `POST /hook/github` to `127.0.0.1:9000`, with `request_body max_size 1MB`. Every other path gets `404`
- [ ] **Quadlet:** the `app-<app>-<stage>.container` template uses `Image=localhost/app-<app>-<stage>:deployed` and
      `Pull=never`, exactly as spike B2 wrote it
- [ ] **`/etc/specsdeployd/config.json`:** from host vars, laid down with
      `validate: "specsdeployd config validate %s"`, and notifying a restart of `specsdeployd.service`
- [ ] **`/etc/specsdeployd/secrets.age`:** copied as-is from the repository, still encrypted, and notifying the restart
- [ ] **Host identity,** per spike B1: generated on first run with its recipient published, or no step at all if the SSH host key is used.
      Document the new-host bootstrap order
- [ ] **Registry credentials** for root, per spike B4
- [ ] `specsdeployd_version` pinned in group vars, starting at `0.2.0-rc.N`

## Verify

- [ ] A `ping` from the server's GitHub webhook gets `200 pong`
- [ ] Breaking `config.json` on purpose makes the ansible-pull run fail at `validate`, and the old file stays in place
- [ ] `curl https://deploy.<server>.specs.dev/livez` gets `404` from Caddy, not `200`: loopback-only health routes stay loopback-only
