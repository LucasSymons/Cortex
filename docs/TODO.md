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

Cortex targets **Claude Code CLI** today (plugin marketplace + local MCP binary +
`SessionStart` hook + `~/.claude/CLAUDE.md`). Cowork runs inside the **Claude Desktop
app**, which installs differently (no `/plugin`) but **does run local MCP servers** -
so the `cortex-git` binary IS viable there. Design the Cowork install path around the
Desktop app's mechanisms.

**Surface support matrix.** Cortex runs wherever Claude can run a **local MCP server
and read a local folder**: **Claude Code CLI** (done) and the **Claude Desktop app**
(where Cowork runs - viable, see below). The **browser (`claude.ai`) and mobile have
no local runtime**, so the native-binary + local-git-profile model does not apply
there; supporting them would need a different remote approach (remote MCP server /
connector / web-project instructions) and is out of scope.

What carries over:
- **Skills** (`/setup`, `/sync-profile`, ...) - same format, work on both surfaces.
- **The profile** (`CLAUDE.md` + `memory/`) - placed where the app auto-loads it (the
  global `.claude` rules space; see the loading model below).

**Goal:** make **Git the single source of truth** for the profile across ALL
surfaces, **independent of OneDrive** - including Cowork (which today only gets
`CLAUDE.md` because the connected Documents folder happens to be OneDrive-synced).

**What the Desktop/Cowork app offers (confirmed from the app UI, 2026-06-10):**
- **No `/plugin` command** - the CLI marketplace install path is absent in Cowork.
- **Settings > Extensions:** install local extensions by dragging a **`.mcpb`/`.dxt`**
  bundle (or "Browse extensions"). The shipped ones - Filesystem, Chrome control,
  Snyk - are local MCP servers running on the machine.
- **Settings > Developer > Local MCP servers ("Edit Config"):** add any local MCP
  server by command + args (`claude_desktop_config.json`).
- **Directory > Plugins:** a *curated* catalog (Your organization / Anthropic &
  Partners); some bundle MCP servers. Curated, not "add any repo".
- **Context loading (confirmed with Lucas, 2026-06-10):** the Desktop/Cowork app
  **always auto-loads global rules from `~/.claude/CLAUDE.md`** (Windows
  `C:\Users\<user>\.claude\CLAUDE.md`; the dir also carries `commands/`, `templates/`).
  Cowork additionally has a **Settings > Cowork** instructions box (global free text -
  can say "read file X at session start") and a **Cowork files** folder
  (`Documents\Claude`, for artifacts + scheduled tasks). Connected **working folders**
  are readable during a task but are NOT auto-loaded as steering unless an instruction
  points at them. **So the default Cortex placement on Desktop/Cowork is the global
  `.claude\CLAUDE.md`** (Git-synced, always loads - the exact analog of the CLI's
  `~/.claude/CLAUDE.md`), no custom instruction needed.

So the native binary is fine on Desktop. The two real gaps are: (a) no `/plugin`
install, and (b) the launcher `bin/cortex-git-launch.sh` is POSIX-only - it will not
run on native Windows, so Desktop needs a `.cmd`/`.ps1` launcher (or the bundled
`.exe`).

**Install paths for Cortex on Desktop/Cowork:**
1. **Manual, works today:** add `cortex-git` via Developer > Local MCP servers >
   Edit Config (point at the Windows binary), and connect a **git clone of the profile
   repo** as a working folder so `CLAUDE.md` + `memory/` come from Git, not OneDrive.
2. **Distributable (the product path):** ship Cortex as an **MCP Bundle (`.mcpb`,
   formerly `.dxt`)** - a zip of `manifest.json` + the server, drag-drop installed from
   Settings > Extensions. The format **natively supports a compiled-binary server**
   (`server.type: "binary"`), so we bundle the Go binary directly - **no fetch/verify
   launcher needed** (the `.mcpb` is the verified distribution). Build with the `mcpb`
   CLI (`mcpb init` / `mcpb pack`). Concrete design (manifest spec v0.3):
   - `server.type: "binary"`, `mcp_config.command` = the bundled binary via
     `${__dirname}`; Windows gets `.exe` auto-appended.
   - **`platform_overrides`** (`win32` / `darwin` / `linux`) carry per-OS binaries in
     one bundle. Arch (amd64 vs arm64) is NOT a manifest dimension - use a macOS
     universal build + a tiny arch-selecting wrapper, or per-arch bundles.
   - **`user_config`** declares `host` / `username` / `token` (token `sensitive: true`),
     injected via `${user_config.token}` as env (e.g. `CORTEX_GIT_TOKEN`) - no separate
     `set_credentials` step.
3. *(Org-wide, later)* publish to the org's Directory > Plugins if relevant.

Sub-tasks: **(i)** small server change - read creds from env (`CORTEX_GIT_*`) so
`user_config` can inject the PAT (today creds come only from the keychain/file store);
**(ii)** the `.mcpb` manifest + per-platform binary packaging (goreleaser can build a
macOS universal binary + the bundle); **(iii)** a Windows launcher (`.cmd`/`.ps1`) -
needed only for the *manual* `claude_desktop_config.json` path, not the `.mcpb`;
**(iv)** the git-clone-as-working-folder flow for profile + memory.

Still to verify in a real Cowork session (Lucas):
- Does `CLAUDE.md` auto-load from a connected **non-OneDrive** working folder (a git
  clone)? (High confidence yes - it's just a folder - but confirm.)
- Does Cowork read the `memory/` files, or only `CLAUDE.md` at the folder root?
- Can a Cowork skill shell `git` for in-app sync (else sync the clone out-of-session)?

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
