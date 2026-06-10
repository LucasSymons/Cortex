# Cortex - roadmap

**Status (2026-06-10):** v0.1.0 is released and publicly installable
(`/plugin marketplace add LucasSymons/Cortex` -> `/plugin install cortex@cortex`).
GitHub Actions CI, Dependabot, and the goreleaser release pipeline are all green.
The plugin has been verified end-to-end, including a real first-run setup against a
private GitLab profile repo.

**Next up:** a small **v0.1.1** carrying the two skill-only Setup / onboarding UX
wins below (import an existing `CLAUDE.md`; a clean "no persona" path) - no Go
change, so low risk. The community-marketplace submission is queued (manual, via
`clau.de/plugin-directory-submission`).

Open items, grouped by theme. Each becomes a branch + PR.

## Setup / onboarding UX

- [ ] **Import an existing setup.** Detect an existing `CLAUDE.md` (`~/.claude`, the
      Cowork Documents folder) during `/setup` and offer to import and adapt it into
      the profile, instead of only the cold questionnaire. Biggest win for users who
      already have a profile and persona.
- [ ] **Deeper guided persona builder.** When a user wants a full character, branch
      into a richer guided interview (name, background, personality, values, voice)
      rather than the current handful of questions.
- [ ] **Clean "no persona" path.** Make "just preferences, no character" an obvious
      first-class choice in the persona section for the many users who will not want
      a named persona.
- [ ] **Memory path resolution on Claude Code CLI.** When `CLAUDE.md` is placed at
      `~/.claude/CLAUDE.md`, the relative `memory/` reference should resolve to the
      profile repo regardless of working directory. `/setup` and `/restore-profile`
      should make the memory path explicit (or place a pointer).

## Cowork support

Cortex today targets **Claude Code CLI only** - plugin marketplace + a local native
MCP binary + a `SessionStart` hook + `~/.claude/CLAUDE.md`. **Cowork is a different,
more sandboxed runtime and this architecture does not port directly.** Design a
Cowork install path.

What carries over:
- **Skills** (`/setup`, `/sync-profile`, ...) - same format, work on both surfaces.
- **The profile** (`CLAUDE.md` + `memory/`) - Cowork auto-loads `CLAUDE.md` from the
  connected Documents folder root, so the profile can live there.

The problem child is the **`cortex-git` MCP binary**: Cowork's plugin model is
skill/connector-oriented (not the CLI `.mcp.json` + hooks model), and the session
sandbox makes "download + cache + exec a native binary" unviable. Note that in
Cowork the Documents folder is usually already synced (OneDrive / a GitHub
connector), so git-based profile sync may be unnecessary there.

**Goal:** make **Git the single source of truth** for the profile across ALL
surfaces, **independent of OneDrive** - including Cowork (which today only gets
`CLAUDE.md` because the connected Documents folder happens to be OneDrive-synced).

Likely shape: the profile repo is the source of truth and each surface reads a local
**git clone** of it. For Cowork, connect its workspace to a local clone of the
profile repo (instead of the OneDrive Documents folder); keep that clone current with
`git pull`/`push` - via the cortex-git plugin on the CLI, a Cowork skill that shells
plain `git` if the sandbox allows it, or synced out-of-session. The native
`cortex-git` MCP binary stays CLI-only; Cowork either reads a clone kept current
elsewhere or runs `git` directly from a skill (needs git + a PAT in the sandbox).

**Confirmed (2026-06-10, tested in Cowork):** Cowork has **no `/plugin` command** -
the Claude Code CLI marketplace install path does not exist there. Any Cowork install
is via the Settings UI (skills / connectors), so the CLI-plugin format (MCP binary +
hooks) is not installable in Cowork as-is. However, Cowork already auto-loads
`CLAUDE.md` from the connected Documents folder, so the *profile* (persona + prefs)
works there with no install - the open question is the *skills* and *sync*.

Open questions to verify in a real Cowork session (do this while setting Cowork up):
- Can a Cowork plugin/connector declare or reach an MCP server (local or remote), or
  is it skills-only?
- Is `${CLAUDE_PLUGIN_DATA}` (or any plugin storage) persistent across Cowork
  conversations?
- Does Cowork read a `memory/` dir, or only `CLAUDE.md`? How is memory meant to work?
- Does the Cowork sandbox have `git` + outbound network + a writable Documents folder
  for a skill-driven sync?

## Publishing / install

- [x] Shipped in **v0.1.0**: binary launcher (fetch + SHA-256 verify into
      `${CLAUDE_PLUGIN_DATA}`), `SessionStart` warm hook, `marketplace.json`,
      `plugin.json` polish, the goreleaser release pipeline (`checksums.txt`), and a
      verified end-to-end install + real-host sync test.
- [ ] *(Optional)* cosign-sign release artifacts and verify the signature in the
      launcher, on top of the existing SHA-256 check.
- [ ] *(Optional)* Submit to the `anthropics/claude-plugins-community` marketplace
      via `clau.de/plugin-directory-submission`.

## Enhancements (later)

- [ ] **Passphrase mode** for the encrypted-file credential fallback, for real at-rest
      protection on headless boxes (currently machine-bound obfuscation).
- [ ] **Real `git_diff`** - a content-level change preview (the stub was removed as it
      duplicated `git_status`).
- [ ] **Better pull conflict strategy** than last-write-wins (`Force: true`).
- [ ] **`golangci-lint`** config + CI job for stricter static analysis beyond `go vet`.
- [ ] **`CORTEX_CONFIG_DIR` / force-file-backend override** - pin the encrypted-file
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

- [x] Install/setup steps finalised in `README.md` and `docs/usage.md` (v0.1.0).
