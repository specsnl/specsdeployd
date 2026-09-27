---
id: A2
title: "docs: scaffold the Hugo/Hextra site and publish it to Pages"
type: Task
milestone: "M0 · Foundation"
epic: A
depends_on: []
repo: specsnl/specsdeployd
---

## Goal

Stand up the docs site **before** any feature lands, so that every feature PR can update its
architecture page in the same change, as AGENTS.md requires. Copy labelsync's setup: Hextra as a Hugo module
from its own `docs/go.mod`, and a Pages deploy.

`docs/design.md` and `docs/backlog/` stay **outside** `docs/content/`. They are the plan, not the published
documentation.

## Design reference

[`docs/design.md` § Documentation](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#documentation)

## Scope

- [ ] `docs/go.mod`, `docs/hugo.toml`, `docs/content/_index.md` (a hero page and feature cards: "No SSH",
      "Signed and allowlisted", "Rolls back on a failed readyz", "Secrets encrypted at rest")
- [ ] `docs/content/docs/_index.md`, with `architecture/_index.md` and `usage/_index.md`. Create a stub page, with front matter and a
      one-paragraph summary, for every page listed in design.md § Documentation. Mark each stub "not built yet"
- [ ] `architecture/versioning.md` and `architecture/distribution.md`, filled in from what `v0.1.0` already does
- [ ] `taskfiles/Taskfile.docs.yml` with `docs:build`, `docs:serve` (:1313), `docs:preview` (nginx on :8080) and `docs:mod:tidy`.
      Add compose services `hugo` and `docs-preview` with pinned images and `# Latest version:` comments
- [ ] `.github/workflows/docs.yml`: a paths filter, `hugo --minify`, a deploy to Pages, and concurrency `pages` without cancelling
- [ ] A dependabot `gomod` entry for `/docs`. This lands together with A3; whichever PR merges second adds it
- [ ] Domain `specsdeployd.specs.dev` (decided): set the `CNAME` in Pages and add the DNS record

## Tests

`task docs:build` succeeds with no warnings, and `task md:check` passes.

## Depends on

Nothing. **Parallel-safe.**

## Done when

`task checkall` and `task docs:build` pass, every box above is ticked, and the site is live on Pages.
