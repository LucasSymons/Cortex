#!/usr/bin/env bash
# Generate a self-signed TLS certificate for the local Gitea instance used by
# the E2E suite. The certificate doubles as its own trust anchor: Gitea serves
# it, and the cortex-git subprocess trusts it via SSL_CERT_FILE (see run.sh).
#
# Valid for localhost / 127.0.0.1 only - this is a throwaway cert for local and
# CI testing, never for anything reachable off the box.
set -euo pipefail

cd "$(dirname "$0")"
certs_dir="certs"
mkdir -p "$certs_dir"

crt="$certs_dir/server.crt"
key="$certs_dir/server.key"

# Always valid for localhost; add the override host (e.g. "docker" under CI's
# docker-in-docker) so the cert matches whatever URL the client connects to.
host="${E2E_GITEA_HOST:-localhost}"
san="DNS:localhost,IP:127.0.0.1"
if [[ "$host" != "localhost" && "$host" != "127.0.0.1" ]]; then
	san="$san,DNS:$host"
fi

echo "gen-certs: generating self-signed cert (SAN: $san)"
openssl req -x509 -newkey rsa:2048 -nodes \
	-keyout "$key" \
	-out "$crt" \
	-days 365 \
	-subj "/CN=localhost" \
	-addext "subjectAltName=$san"

# World-readable on purpose: this is a disposable localhost-only cert, and the
# Gitea container's unprivileged user must read the key through the bind mount
# regardless of host UID. Never reuse this cert for anything real.
chmod 644 "$key"
echo "gen-certs: wrote $crt and $key"
