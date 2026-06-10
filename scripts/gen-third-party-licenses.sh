#!/usr/bin/env bash
#
# Regenerate THIRD_PARTY_LICENSES.md from the licences of every Go module that
# is actually compiled into the cortex-git-server binary (not the full module
# graph - test/tooling deps of dependencies are excluded).
#
# Run from the repo root: `make licenses`. Re-run whenever dependencies change.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SERVER_DIR="$ROOT/mcp/git-server"
OUT="$ROOT/THIRD_PARTY_LICENSES.md"
MODCACHE="$(go env GOMODCACHE)"

cd "$SERVER_DIR"

{
  echo "# Third-party licences"
  echo
  echo "The \`cortex-git-server\` binary statically links the Go modules listed"
  echo "below; their licences are reproduced in full to satisfy the attribution"
  echo "terms of the MIT, BSD, and Apache-2.0 licences."
  echo
  echo "This file is generated - do not edit by hand. Regenerate with"
  echo "\`make licenses\` after changing dependencies."
  echo
} > "$OUT"

# Unique modules linked into the binary.
mods="$(go list -deps -f '{{if and (not .Standard) .Module}}{{.Module.Path}}@{{.Module.Version}}{{end}}' ./cmd/server | sort -u)"

while IFS= read -r modver; do
  [ -z "$modver" ] && continue
  mod="${modver%@*}"
  ver="${modver#*@}"
  # Go escapes uppercase letters in cache paths as "!" + lowercase.
  esc="$(printf '%s' "$mod" | sed -E 's/([A-Z])/!\L\1/g')"
  dir="$MODCACHE/${esc}@${ver}"
  lic="$(ls "$dir" 2>/dev/null | grep -iE '^(LICENSE|LICENCE|COPYING)' | head -1 || true)"
  {
    echo "## ${mod}"
    echo
    echo "Version: \`${ver}\`"
    echo
    if [ -n "$lic" ] && [ -f "$dir/$lic" ]; then
      echo '```'
      # Strip trailing whitespace so the file passes the trailing-whitespace hook.
      sed 's/[[:space:]]*$//' "$dir/$lic"
      echo '```'
    else
      echo "_No licence file found in the module cache._"
    fi
    echo
  } >> "$OUT"
done <<< "$mods"

echo "Wrote $OUT ($(grep -c '^## ' "$OUT") modules)"
