# Backlog — drafts for the GitHub tracker

These files are the tracker for the first functional release, in draft form. Nothing in here exists on GitHub yet.
Each file becomes one issue, with a body identical to the text below its front matter. Once the issues exist,
[`docs/design.md` § Milestones](../design.md#milestones) gets the real issue numbers and this directory is
deleted, because GitHub is then the source of truth.

## Layout

| Path                        | Becomes                                                                                                 |
|-----------------------------|---------------------------------------------------------------------------------------------------------|
| `epics/<A–G>-<slug>.md`     | An issue of type `Feature`, titled `Epic: …`, with no milestone, attached as a sub-issue of specsops#13 |
| `issues/<id>-<slug>.md`     | An issue of type `Task` with a milestone, attached as a sub-issue of its epic                           |
| `cross-repo/<id>-<slug>.md` | An issue in **another** repository. Created only after a separate confirmation                          |

## Front matter

```yaml
id: C2                      # local id; references look like #C2 in bodies
title: "config: …"          # the issue title
type: Task                  # Feature (epics) | Task
milestone: "M1 · Contracts" # leaf issues only
epic: C                     # the parent epic's id
depends_on: [C1, A4]        # becomes native "blocked by" links; also listed under ## Depends on
repo: specsnl/specsdeployd  # target repository
```

In the bodies, `#A4`-style references are placeholders. Creation rewrites them to the real `#N`, or to
`owner/repo#N` for issues in other repositories.

## Milestones

| Milestone         | Description (as it will appear on GitHub)                                                                                    |
|-------------------|------------------------------------------------------------------------------------------------------------------------------|
| `M0 · Foundation` | Repo hygiene, docs site, sentinel errors, the output and App spine, release contract tests, and the four spikes. No network. |
| `M1 · Contracts`  | Embedded JSON Schemas, config, age-encrypted secrets, deployment-event parsing, `config validate`. No network.               |
| `M2 · Receive`    | `serve`, HMAC verification, the request pipeline, the per-environment queue. Loopback only.                                  |
| `M3 · Deploy`     | Sudo runner, readyz probe, pull → retag → restart → ready, rollback, and the `deploy` command.                               |
| `M4 · Report`     | GitHub App authentication and deployment statuses.                                                                           |
| `M5 · Ship`       | `serve` end to end, the e2e suite on a runner, docs, `v0.2.0`, GitHub setup, and acceptance on a golden-app VM.              |

## Index

| Epic                          | Issues                                                         |
|-------------------------------|----------------------------------------------------------------|
| A — Foundation & repo hygiene | A1–A7                                                          |
| B — Spikes                    | B1–B4                                                          |
| C — Contracts                 | C1–C6                                                          |
| D — Webhook receiver          | D1–D4                                                          |
| E — Deploy pipeline           | E1–E5                                                          |
| F — GitHub reporting          | F1–F2                                                          |
| G — Ship                      | G1–G6                                                          |
| Cross-repo                    | X1 (collection), X2 (specsops-ansible), X3 (specsops-opentofu) |

## Creating the tracker

This happens only after the drafts have been reviewed.

- [ ] Create the six milestones, using the descriptions above
- [ ] Create the epics (`type: Feature`), then attach each one to specsnl/specsops#13 as a sub-issue
- [ ] Create the issues (`type: Task`, with milestone), and attach each one to its epic as a sub-issue
- [ ] Add the native "blocked by" links from `depends_on`
- [ ] Rewrite the `#<id>` placeholders in every body into real numbers
- [ ] Update the "Where it stands" section of specsops#13
- [ ] Record the numbers in `docs/design.md` § Milestones, then delete `docs/backlog/`
- [ ] Separately, once confirmed: create X1, X2 and X3 in their repositories and link them from G6
