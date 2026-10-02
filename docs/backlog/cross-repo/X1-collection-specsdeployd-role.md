---
id: X1
title: "specsdeployd: allow podman tag, drain on stop, and tighten the config dir"
type: Task
repo: specsnl/specsops-ansible-collection
parent: specsnl/specsops#13
depends_on: [B1, B2]
---

**Repo:** `specsnl/specsops-ansible-collection` · **Driven by:** specsnl/specsdeployd [`docs/design.md` § Cross-repo changes](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#cross-repo-changes)

## Why

The daemon deploys tag releases by retagging the pulled image as `localhost/<unit>:deployed`, the stable ref the
Quadlet runs, and it keeps `:previous` for rollback. The role's sudoers rule allows `podman pull` and `systemctl
restart` only. Three other small changes also fall out of the daemon's design.

## Scope

- [ ] `templates/sudoers.j2`: add `{{ specsdeployd_user }} ALL=(root) NOPASSWD: /usr/bin/podman tag *`, the exact line
      from spike B2. Update the comment ("escalates for exactly three commands"), and change molecule's "exactly 2 non-comment lines" to 3
- [ ] The unit gets `TimeoutStopSec=90s`, so the daemon's 60s drain (`--shutdown-timeout`) finishes before SIGKILL
- [ ] **If spike B1 chose the SSH host key:** add `LoadCredential=age-identity:/etc/ssh/ssh_host_ed25519_key`, and assert it in
      molecule
- [ ] The config dir becomes `root:specsdeployd` 0750, so the daemon can read its config and secrets but not rewrite them. Files
      inside it are 0640 `root:specsdeployd`, and the identity, when it is a dedicated file, is 0440 or 0400
- [ ] Molecule: the installed host gets a fixture `config.json` and `secrets.age`, then asserts that the unit **starts and stays
      active**, `/livez` answers on 127.0.0.1:9000, and `specsdeployd config validate` passes
- [ ] Release collection `0.4.0`, and bump the golden-images `requirements.yml`

## Verify

- [ ] `molecule test` passes for both scenarios
- [ ] `sudo -l -U specsdeployd` on the converged host lists exactly the three commands
