# Cortex — roadmap

Open items, grouped by theme. Each becomes a branch + PR.

## Publishing / install (current focus)

- [ ] **Binary distribution.** A freshly installed plugin has no compiled
      `cortex-git-server` binary. A bundled launcher script fetches the correct
      release binary for the host platform on first run into `${CLAUDE_PLUGIN_DATA}`,
      verifying its SHA-256 against the release `checksums.txt` before running it.
- [ ] **Release integrity.** Releases emit `checksums.txt`; optionally add cosign
      signatures and verify them in the launcher.
- [ ] **Marketplace.** Add `.claude-plugin/marketplace.json` so the plugin is
      installable via `/plugin marketplace add` → `/plugin install`.
- [ ] **Manifest polish.** Keep `plugin.json` `version` in sync with the released tag.
- [ ] **Validate + test install.** `claude plugin validate`, then a local end-to-end
      install from a marketplace through to a working sync round-trip.
- [ ] *(Optional)* Submit to the `anthropics/claude-plugins-community` marketplace.

## Enhancements (later)

- [ ] **Passphrase mode** for the encrypted-file credential fallback, for real at-rest
      protection on headless boxes (currently machine-bound obfuscation).
- [ ] **Real `git_diff`** — a content-level change preview (the stub was removed as it
      duplicated `git_status`).
- [ ] **Better pull conflict strategy** than last-write-wins (`Force: true`).
- [ ] **`golangci-lint`** config + CI job for stricter static analysis beyond `go vet`.
- [ ] **`CORTEX_CONFIG_DIR` / force-file-backend override** — pin the encrypted-file
      backend at a given dir regardless of whether an OS keychain is present. Enables
      deterministic, fully isolated E2E on every platform and a clean headless override.

## Code-review leftovers (optional)

- [ ] Note the weakened-key fallback in code: `machineID`
      (`internal/keychain/file_store.go`) falls back to the hostname and finally a
      constant when `/etc/machine-id` is absent, deriving the file-store key from
      non-secret inputs. Add a one-line comment so it isn't mistaken for strong
      at-rest crypto. (Ties into passphrase mode.)
- [ ] Unique temp file in the file store: `fileStore.save` writes to a fixed
      `path + ".tmp"` guarded only by an in-process mutex; use `os.CreateTemp` in the
      same dir.

## Docs

- [ ] Finalise the install/setup steps in `README.md` and `docs/usage.md` once
      marketplace publishing lands.
