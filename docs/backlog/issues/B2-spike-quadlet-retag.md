---
id: B2
title: "Spike: confirm local-ref retagging works under Quadlet on Ubuntu 26.04 Podman"
type: Task
milestone: "M0 · Foundation"
epic: B
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

The whole deploy pipeline rests on one assumption: a Quadlet unit with
`Image=localhost/app-X:deployed` and `Pull=never` picks up whatever image `podman tag` last pointed that ref at, on a
plain `systemctl restart`, **with no `daemon-reload`**. This spike verifies that assumption on the actual `golden-app` stack
(Ubuntu 26.04, Podman from `universe`, rootful Quadlet), together with the edge cases the pipeline branches on.

## Design reference

[`docs/design.md` § Local image refs](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#local-image-refs),
[§ Deploy pipeline](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#deploy-pipeline)

## Scope

- [ ] On a golden-app VM, or on the systemd container the collection's molecule tests use: write a `.container` with
      `Image=localhost/app-spike-test:deployed` and `Pull=never`, then pull two tags of a small image and alternate
      `podman tag X localhost/app-spike-test:deployed` with `systemctl restart app-spike-test.service`. Confirm each restart runs the newly tagged image
- [ ] Before the first tag: record how the unit fails, whether it loops or ends in `failed`, and what a later deploy does
      afterwards
- [ ] `podman tag localhost/app-spike-test:deployed …:previous` when `:deployed` does not exist yet: record the exact exit code and
      stderr, so the executor can tell "first deploy" apart from a real failure (design.md open question 4)
- [ ] Run everything as `specsdeployd` through `sudo -n`, with the role's sudoers file **plus** `/usr/bin/podman tag *`. Confirm the
      image lands in root's store, the one rootful Quadlet reads
- [ ] Check whether `systemctl restart` returns only after the new container has started. That depends on Quadlet's `Notify=` and
      `sdnotify` defaults, and it decides whether `/readyz` can be polled straight away
- [ ] Record the Podman version and the Quadlet generator version

## Output

A comment containing:

1. the exact Quadlet template for specsops-ansible, which goes into X2;
2. the exact sudoers line for the collection, which goes into X1;
3. the exit code and message pattern that means "image not known", which goes into #E3;
4. anything that changes design.md § Local image refs.

## Depends on

Nothing. **Parallel-safe.**
