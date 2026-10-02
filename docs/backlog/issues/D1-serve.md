---
id: D1
title: "cmd: serve --listen with a hardened, loopback-only http.Server"
type: Task
milestone: "M2 · Receive"
epic: D
depends_on: [A5, C2, C4]
repo: specsnl/specsdeployd
---

## Goal

`specsdeployd serve --listen 127.0.0.1:9000`: the exact command line the role's unit already runs. This issue
covers the server shell only: loading at startup, listening, the two health routes, and a graceful shutdown. The webhook handler
(#D3) and the queue (#D4) plug into it, and #G1 wires everything together.

## Design reference

[`docs/design.md` § HTTP surface](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#http-surface),
[§ Queue](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#queue) (shutdown),
[§ Where it fits](https://github.com/specsnl/specsdeployd/blob/main/docs/design.md#where-it-fits) (the unit contract)

## Scope

- [ ] Startup order:
      1. config (#C2);
      2. secrets (#C4), including the App key parse;
      3. listen.
      A failure at any step exits `1` with its `error_kind` logged. The unit's `Restart=on-failure` then retries every 5s,
      so the error message must say what to fix
- [ ] `--listen` refuses anything that is not a loopback address (`127.0.0.0/8`, `::1`) with `ErrNonLoopbackListen`. There is
      no override flag: a non-loopback listener is a design change, not an option
- [ ] `http.Server`: `ReadHeaderTimeout` 5s, `ReadTimeout` 15s, `WriteTimeout` 15s, `IdleTimeout` 60s, `MaxHeaderBytes`
      64 KiB, and `ErrorLog` routed into slog
- [ ] `GET /livez` returns `200` always. `GET /readyz` returns `200` once startup has finished, and `503` while shutting down
- [ ] When `ctx` is cancelled by SIGTERM: stop accepting, mark the daemon not ready, call the queue's drain (an interface stub until
      #D4), and exit within `--shutdown-timeout` (default 60s)
- [ ] One `info` line at startup, with `server_id`, the listen address, the hosted environments and the version
- [ ] The architecture page `overview.md`, updated with the `serve` data flow

## Tests

- Through the command tree:
  - a non-loopback address is refused;
  - a missing config, a missing identity, or a bad App key fails fast with the right `error_kind`;
  - `/livez` and `/readyz` answer across the startup and shutdown transitions.
- Tests use `127.0.0.1:0`, a cancellable context, and no real sleeps.

## Depends on

Blocked by #A5, #C2, #C4

## Done when

`task checkall` passes, every box above is ticked, and the change ships with its tests and documentation in the same PR,
per `AGENTS.md`.
