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

Early — nothing is implemented yet. The repository is being scaffolded around a binary that answers
`specsdeployd version` and nothing else, which is enough for the Ansible role to install and assert
against. The webhook receiver lands on top of that.

## License

MIT — see [LICENSE](./LICENSE).
