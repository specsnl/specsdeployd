---
id: G6
title: "acceptance: a GitHub deployment rolls a golden-app VM forward, then cut v0.2.0"
type: Task
milestone: "M5 · Ship"
epic: G
depends_on: [G4, G5, X1, X2, X3]
repo: specsnl/specsdeployd
---

## Goal

This issue meets specsnl/specsops#13's "Done when" on a real server:

> A GitHub deployment for an environment on a `golden-app` VM makes that VM pull the new image, restart its Quadlet
> unit, pass `/readyz`, and report `success`, with no human or CI touching the server.

Once that holds, cut `v0.2.0` and pin it from ansible-pull.

## Design reference

[`docs/design.md` § Goals](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#goals)

## Scope

- [ ] A staging golden-app VM runs collection `0.4.0` (X1) with `specsdeployd_version: 0.2.0-rc.N` and the
      ansible-pull changes (X2). `systemctl status specsdeployd` shows it active
- [ ] Create a deployment from a workflow in an allowlisted app repository, with a `tag` payload. It ends in `success`, and the new
      tag is running
- [ ] A `branch` deployment of `main` ends in `success`
- [ ] A deliberately broken image ends in `failure` with "rolled back", and the previous version is still serving
- [ ] Server creation, webhooks included, came from one `tofu apply` (X3), with no webhook clicked by hand
- [ ] A deployment for an environment on another server is ignored, and no status appears from this server
- [ ] Restart the VM: the unit comes back with the last deployed image, without any deploy
- [ ] Tag `v0.2.0`, point `specsdeployd_version` at it, and update specsops#13's "Where it stands"

## Tests

The checks above, each recorded with a link to its deployment in a comment on this issue.

## Depends on

Blocked by #G4, #G5, X1 (specsnl/specsops-ansible-collection), X2 (specsnl/specsops-ansible), X3 (specsnl/specsops-opentofu)

## Done when

Every box above is ticked, `v0.2.0` is released, and specsops#13 can be closed.
