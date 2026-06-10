#!/usr/bin/env bash
# End-to-end test harness for cortex-git. Stands up a disposable Gitea over
# HTTPS, provisions a user + access token + empty repo, then runs the
# e2e-tagged Go test which drives the real cortex-git binary through the full
# sync lifecycle (set_credentials -> git_init -> clone -> commit_push -> pull)
# against that remote over real HTTPS with PAT auth.
#
# Requires: docker (with the compose plugin), openssl, curl, go.
# Usage: ./e2e/run.sh   (or: make e2e)
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
repo_root="$(cd "$here/.." && pwd)"
cd "$here"

# How this job reaches Gitea.
#  - Local dev: defaults to localhost:3443 (the published port on this host).
#  - CI on the host-socket runner: the job is a container that drives the host's
#    Docker daemon, so Gitea is a *sibling* container, not on this container's
#    network and not on its localhost. CI therefore sets E2E_GITEA_HOST=gitea +
#    E2E_GITEA_PORT=3000 and E2E_JOIN_NETWORK to the compose network, and we
#    attach this job container to that network below so `gitea:3000` resolves.
gitea_host="${E2E_GITEA_HOST:-localhost}"
gitea_port="${E2E_GITEA_PORT:-3443}"
gitea_url="https://${gitea_host}:${gitea_port}"
join_network="${E2E_JOIN_NETWORK:-}"
self_cid="" # this job's own container id, resolved after compose up (CI only)
admin_user="e2e"
# Random per-run password: nothing static or secret-shaped lands in the repo.
admin_pass="$(openssl rand -hex 16)"
repo_name="cortex-profile"

cleanup() {
	echo "==> tearing down Gitea"
	# If we attached this job container to the compose network, detach first:
	# otherwise `compose down` cannot remove a network with an active endpoint
	# and would leak it.
	if [[ -n "$join_network" && -n "$self_cid" ]]; then
		docker network disconnect -f "$join_network" "$self_cid" >/dev/null 2>&1 || true
	fi
	# Use an explicit -f: by the time this fires on EXIT the script has cd'd into
	# the Go module to run the test, so a bare `docker compose` would find no
	# compose file here and silently no-op, leaving Gitea running.
	docker compose -f "$here/docker-compose.yml" down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "==> generating self-signed cert"
E2E_GITEA_HOST="$gitea_host" ./gen-certs.sh

# Clean slate up front: a prior run's teardown can lag, leaving a container
# Compose would otherwise reuse with stale data (a leftover repo/user). Don't
# trust the previous run's trap - nuke anything here, then force a fresh container.
docker compose down -v --remove-orphans >/dev/null 2>&1 || true

echo "==> starting Gitea (HTTPS) and waiting for it to become healthy"
docker compose up -d --wait --force-recreate --build

# Host-socket CI: this job is a sibling container of Gitea on the host daemon, so
# it isn't on Gitea's network by default. Attach it so `gitea:3000` resolves.
# We must connect THIS container by its real id: a CI job container's
# internal hostname is only a prefix of its docker name, so `connect $(hostname)`
# matches nothing. The id is in /proc/self/mountinfo (the /etc/hosts etc. bind
# mounts reference <docker-root>/containers/<id>/); fall back to a name-prefix
# lookup. Fail loudly if unresolved - a silent miss looks like a Gitea timeout.
if [[ -n "$join_network" ]]; then
	self_cid="$(grep -oE '/containers/[0-9a-f]{64}/' /proc/self/mountinfo 2>/dev/null | head -1 | grep -oE '[0-9a-f]{64}' || true)"
	if [[ -z "$self_cid" ]]; then
		self_cid="$(docker ps --no-trunc --filter "name=$(hostname)" --format '{{.ID}}' | head -1)"
	fi
	if [[ -z "$self_cid" ]]; then
		echo "ERROR: could not resolve this job's own container id to join $join_network" >&2
		exit 1
	fi
	echo "==> joining compose network $join_network as $self_cid"
	docker network connect "$join_network" "$self_cid"
fi

# The container healthcheck can flip green before a first-boot Gitea has finished
# initialising its DB, which makes the next admin/API call race-y. Poll the
# unauthenticated API until it actually answers before provisioning.
echo "==> waiting for the Gitea API to be ready"
for attempt in $(seq 1 30); do
	if [[ "$(curl -sk -o /dev/null -w '%{http_code}' "$gitea_url/api/v1/version")" == "200" ]]; then
		break
	fi
	if [[ "$attempt" -eq 30 ]]; then
		echo "ERROR: Gitea API did not become ready within 30s" >&2
		exit 1
	fi
	sleep 1
done

# Provisioning runs the gitea CLI against the same SQLite DB the running server
# holds open. Writes from a CLI process do not always propagate to the next
# command instantly on first boot, so each step is retried until it takes.
echo "==> provisioning admin user"
create_out=""
for attempt in $(seq 1 10); do
	if create_out="$(docker compose exec -T -u git gitea \
		gitea admin user create \
		--username "$admin_user" --password "$admin_pass" \
		--email "e2e@example.com" --admin --must-change-password=false 2>&1)" \
		|| printf '%s' "$create_out" | grep -qi "already exists"; then
		break
	fi
	sleep 1
	if [[ "$attempt" -eq 10 ]]; then
		echo "ERROR: creating the admin user failed:" >&2
		printf '%s\n' "$create_out" >&2
		exit 1
	fi
done

# Token via the CLI (not the API): the API token endpoint authenticates with the
# user's password, which on first boot can lag the create. The CLI only needs
# the username. A unique token name per attempt avoids a "name already used"
# wedge if a previous attempt half-applied before erroring.
echo "==> generating access token (via CLI)"
token=""
token_raw=""
for attempt in $(seq 1 10); do
	if token_raw="$(docker compose exec -T -u git gitea \
		gitea admin user generate-access-token \
		--username "$admin_user" --token-name "cortex-e2e-$attempt" \
		--scopes "write:repository,write:user" --raw 2>&1)"; then
		token="$(printf '%s' "$token_raw" | tr -d '[:space:]')"
		[[ -n "$token" ]] && break
	fi
	sleep 1
done
if [[ -z "$token" ]]; then
	echo "ERROR: could not generate an access token. Last output:" >&2
	printf '%s\n' "$token_raw" >&2
	exit 1
fi

# Create the empty repo using token auth (not the password) for the same reason.
echo "==> creating empty remote repo ($admin_user/$repo_name)"
repo_resp="$(curl -sk -w '\nHTTP_CODE=%{http_code}' \
	-H "Authorization: token $token" \
	-X POST "$gitea_url/api/v1/user/repos" \
	-H 'Content-Type: application/json' \
	-d "{\"name\":\"$repo_name\",\"private\":true,\"auto_init\":false}")"
if ! printf '%s' "$repo_resp" | grep -q 'HTTP_CODE=201'; then
	echo "ERROR: creating the remote repo failed:" >&2
	printf '%s\n' "$repo_resp" >&2
	exit 1
fi

echo "==> running e2e test"
cd "$repo_root/mcp/git-server"
SSL_CERT_FILE="$here/certs/server.crt" \
	E2E_REMOTE_URL="$gitea_url/$admin_user/$repo_name.git" \
	E2E_HOST="$gitea_host" \
	E2E_USERNAME="$admin_user" \
	E2E_TOKEN="$token" \
	go test -tags e2e -count=1 -run TestE2ESyncRoundTrip ./cmd/server/ -v

echo "==> e2e passed"
