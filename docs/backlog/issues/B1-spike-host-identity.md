---
id: B1
title: "Spike: choose the host age identity and how its recipient reaches specsops-ansible"
type: Task
milestone: "M0 · Foundation"
epic: B
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

`secrets.age` is encrypted to each host's age recipient plus a team break-glass recipient. The daemon reads
both native age X25519 identities and `ssh-ed25519` keys, so the code does not have to choose between them. Operations does:
the choice decides where the identity file lives, who owns it, whether the role's unit changes, and how a
**new VM** gets a recipient before any secret can be encrypted for it.

## Design reference

[`docs/design.md` § Host identity (spike B1)](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#host-identity-spike-b1),
[§ Recipients and authoring](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#recipients-and-authoring)

## Scope

- [ ] **Option A, a dedicated identity:** `/etc/specsdeployd/identity` (0400 `specsdeployd`). Work out who generates it
      (the first ansible-pull run, cloud-init, or OpenTofu) and how its public recipient is committed to specsops-ansible
      before `secrets.age` can be built. Sketch the bootstrap sequence for a brand-new VM, step by step
- [ ] **Option B, the SSH host key:** handed to the unit with `LoadCredential=age-identity:/etc/ssh/ssh_host_ed25519_key`, so the
      daemon reads `$CREDENTIALS_DIRECTORY/age-identity`. Confirm that golden-app VMs get **unique host keys per VM**,
      regenerated on first boot and not baked into the image. Confirm that `ssh-keyscan` output converts to a working age recipient
      (`age -R` accepts `ssh-ed25519` lines)
- [ ] Compare the two options on bootstrap effort, blast radius, rotation, and the change each needs in the role
- [ ] Rotation runbook sketch: a new host identity, and a new break-glass identity
- [ ] Break-glass: where the identity lives (1Password), and who can decrypt with it

## Output

A comment on this issue recording the comparison and the decision. After that, updates to:

- design.md's § Host identity and its open question 1;
- the default `identity_file`;
- #C4, the permission check (when `LoadCredential` is used, the file is owned by root);
- X1 and X2.
Also a note on specsnl/specsops#14 describing the per-host step it gains.

## Depends on

Nothing. **Parallel-safe**, and it needs no code.
