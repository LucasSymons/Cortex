# E2E tests

End-to-end tests for the `cortex-git` MCP server. They stand up a disposable
[Gitea](https://about.gitea.com/) instance over HTTPS in a container, then drive
the **real compiled `cortex-git` binary** over the MCP stdio protocol through the
full profile-sync lifecycle against that remote:

```
set_credentials -> git_init -> git_clone -> git_commit_push -> git_pull
```

Unlike the unit/integration tests in `mcp/git-server/internal/git`
(`TestSyncRoundTrip`), which use go-git's local filesystem transport, this suite
exercises the parts that only a real host can:

- the smart-HTTP transport with **HTTPS + PAT** BasicAuth,
- the `RequireHTTPS` guard (the remote URL really is `https://`),
- the MCP tool registration and stdio wiring of the shipped binary.

## Requirements

- Docker with the Compose plugin (`docker compose`)
- `openssl`, `curl`, and the Go toolchain on `PATH`

## Run

```sh
make e2e          # from the repo root
# or
./e2e/run.sh
```

`run.sh` generates a self-signed cert, brings Gitea up with `--wait`, provisions
an admin user + access token + an empty repo via the Gitea API, runs the
`e2e`-tagged Go test, and tears everything down on exit (`docker compose down -v`).

### Local vs CI networking

Locally the test reaches Gitea on the published port at `localhost:3443`. In CI,
Gitea runs as a *sibling* container reached over the host's Docker daemon, so the
harness is parameterised: set `E2E_GITEA_HOST`/`E2E_GITEA_PORT` to point at the
service and `E2E_JOIN_NETWORK` to the compose network the job container should join
so the service hostname resolves; `run.sh` handles the attach. (True
docker-in-docker is avoided.)

## How TLS trust works

`gen-certs.sh` writes a throwaway self-signed cert valid for `localhost` /
`127.0.0.1` into `e2e/certs/` (gitignored). Gitea serves it; the `cortex-git`
subprocess trusts it by having `SSL_CERT_FILE` point at the cert, which Go's
`crypto/x509` honours when building the system cert pool. No code changes and no
OS trust-store modification are needed.

## Credential isolation and the dev-machine caveat

The test points `XDG_CONFIG_HOME` at a temp dir, so on headless Linux, WSL2, and
CI - where there is no OS Secret Service - the encrypted-file backend writes only
into that temp dir and nothing persists.

**Caveat:** on a workstation with a live OS keychain (e.g. macOS, or Linux with
gnome-keyring), `set_credentials` stores under the OS keychain, which ignores
`XDG_CONFIG_HOME`. The throwaway token for host `localhost` would then be written
to your real keychain (harmless, but untidy). A future `CORTEX_CONFIG_DIR` /
force-file-backend override (see `docs/TODO.md`) would remove this wart by
letting the test pin the file backend on every platform.
