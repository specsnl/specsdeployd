---
id: F1
title: "github: go-github wrapper with GitHub App auth, installation lookup and retries"
type: Task
milestone: "M4 · Report"
epic: F
depends_on: [A4, C2, C4]
repo: specsnl/specsdeployd
---

## Goal

`internal/github` is the only package that talks to api.github.com. It authenticates as the org's GitHub App: an app JWT
from `app_id` and the private key, then **one installation per repository**, looked up once and cached. It retries
`5xx` responses the way labelsync does, and classifies errors in one place.

## Design reference

[`docs/design.md` § Authentication](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#authentication),
[§ Dependencies](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#dependencies)

## Scope

- [ ] `New(appID, key, opts...) *Client` with functional options `WithBaseURL` (for `httptest` and `github.api_url`),
      `WithHTTPClient`, `WithClock`, `WithRetries` and `WithBackoff`, labelsync's set
- [ ] App transport via `bradleyfalzon/ghinstallation/v2`: `NewAppsTransport` for the JWT, and an installation transport for each
      repository. The installation for a repository comes from `GET /repos/{owner}/{repo}/installation` and is cached for the
      life of the process. A repository with no installation is `ErrGitHubAuth`, "the App is not installed on owner/repo"
- [ ] `retryTransport` ported from labelsync: retry `5xx` up to 3 times, doubling from 500ms. Clone and rewind the request,
      drain the response body, and use the injected clock
- [ ] `Classify(err)`: `401` and `403` on auth are `ErrGitHubAuth`; everything else wraps `ErrStatusReportFailed`, with the
      status code
- [ ] The key comes from `secrets.Secret`, and the token is never logged. Add a `debug` line for each request, without headers
- [ ] The dependency is on the latest go-github major. Record the choice on the library-decisions page
- [ ] The architecture page `github-client.md`

## Tests

- `httptest` fakes: the installation lookup happens once per repository; the token is refreshed after it expires, using a fake clock;
  a `5xx` is retried and succeeds on the third try; a `404` installation gives `ErrGitHubAuth`.
- The JWT's `iss` equals `app_id`, checked by decoding it in the fake.

## Depends on

Blocked by #A4, #C2, #C4

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
