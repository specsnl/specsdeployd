# specsdeployd

The deploy agent of the Specs golden images.

`specsdeployd` is a small Go daemon that runs as a host systemd service on an application server.
It receives GitHub deployment webhooks through Caddy on `127.0.0.1`, verifies the signature, checks
the repository against an allowlist, and rolls the target application forward — `podman pull`, then
a restart of the Quadlet unit — before reporting the deployment status back to GitHub.

Its place in the wider picture is
[§9 of the golden-images architecture](https://github.com/specsnl/specsops-golden-images/blob/main/docs/architecture.md).
The [`specsdeployd` Ansible role](https://github.com/specsnl/specsops-ansible-collection) creates
the user, the sudoers rule, the config directory and the unit, and installs the binary from a
GitHub release. This repository ships **only the binary**.

## Status

Early. The binary answers its version and nothing else:

```sh
specsdeployd version     # specsdeployd version 1.2.3
specsdeployd --version   # 1.2.3
```

That is enough for the Ansible role to install it, verify its checksum and assert that it is on the
host. The webhook receiver lands on top of it — no config file is read, no privileged command is
run, and nothing listens on a port yet.

## Contributing

Every command runs through [Task](https://taskfile.dev), which wraps the Docker Compose services
that pin the Go, golangci-lint and Node versions — so a check runs the same way locally as it does
in CI. Building from a checkout needs nothing but Docker and Task. Run `task --list` for the full
set.

```sh
task build      # the binary, into the working directory
task checkall   # tidy:check, lint, test, md:check — run this before opening a pull request
```

## License

MIT — see [LICENSE](./LICENSE).
