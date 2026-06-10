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

Cortex targets **Claude Code CLI** today (binary + skills + `~/.claude/CLAUDE.md`).
**Cowork** (the agentic mode in the Claude Desktop app) is a *different runtime* from
both the CLI and the plain Desktop chat. The findings below are **ground truth observed
from inside a live Cowork session (2026-06-10)** and supersede earlier screenshot-era
guesses.

**How Cowork actually works (observed from inside):**
- **Runtime:** a sandboxed **Ubuntu 22 VM**. Connected folders mount at
  `/sessions/<id>/mnt/<folder>/` (session id changes - never hardcode). **Only connected
  folders persist** across sessions; the rest of the sandbox is ephemeral.
- **CLAUDE.md:** auto-loaded from the **root of each connected folder**, injected as a
  `<system-reminder>`. Multiple connected folders each contribute their `CLAUDE.md`.
- **Skills:** `SKILL.md` files (plugins land under `mnt/.remote-plugins/plugin_<id>/`);
  invoking one is **prompt injection**, no subprocess. They work unchanged.
- **MCP - two paths:** the *plugin* connectors are **remote `"type": "http"`** (e.g.
  `mcp.atlassian.com`) - which is all Cowork-Bree saw, because her machine had no local
  servers configured. But the canonical doc (`claude.com/docs/cowork/3p/extensions`)
  confirms Cowork **also supports local MCP servers**: user-added via **Settings >
  Developer** (gated by the `isLocalDevMcpEnabled` admin toggle) or as a **`.mcpb`**
  installed from the **Connectors** page. Chrome control is exactly this - a local
  server driving local Chrome. **So a local binary MCP server CAN run for a Cowork
  session** (bridged from `claude_desktop_config.json`).
- **Network:** the sandbox has no direct egress - all traffic goes via a host proxy with
  a **host-controlled allowlist**. `github.com`/`gitlab.com` return **HTTP 403**; only
  allowlisted MCP hosts work. So git-over-HTTPS from inside Cowork is blocked.
- **Plugin install:** Cowork unpacks a `.plugin` bundle into `.remote-plugins/` (skills +
  an optional remote-HTTP `.mcp.json`) - a **subset** of the CLI `.claude-plugin` format:
  skills + remote MCP, **no binary launcher**.

**What this means for Cortex in Cowork:**
- ✅ **Skills work today, unchanged** - `/setup`, `/sync-profile`, `/restore-profile`,
  `/promote-lessons` run as prompt injection.
- ✅ **Profile loads** - connect a git clone of the profile repo as a folder; its root
  `CLAUDE.md` auto-loads and `memory/` is readable via the file tools.
- ⚠️ **The `cortex-git` binary CAN run in Cowork as a local MCP server** (Settings >
  Developer, or a `.mcpb` via Connectors) - the earlier "remote-only" read was wrong.
  Open, needs hands-on verification: **(a)** is `isLocalDevMcpEnabled` on for the
  Origin-managed account, or has the admin disabled it? **(b)** does the bridged server
  run on the **host** (host network -> can reach git hosts) or in the sandbox (git hosts
  403'd)? **(c)** it's historically flaky (several open Cowork local-MCP bugs).

**Viable Cowork path (no binary):**
- Deliver the **skills** to Cowork (its plugin install / a `.plugin` bundle).
- Put the **profile** (CLAUDE.md + memory) in a **connected folder that is a git clone**
  of the profile repo - Cowork reads it; OneDrive drops out.
- **Sync happens host-side, not in Cowork:** keep that clone current with `git` from the
  CLI Cortex or a host-side `git pull` (the sandbox can't reach git hosts). Cowork is a
  read-mostly consumer; memory edits it writes to the folder are pushed by the next
  host-side sync.

**Autonomous git sync *inside* Cowork** now looks possible via the local-MCP path above
(add `cortex-git` through Settings > Developer or a `.mcpb`), **if** the admin toggle
allows it and the bridged server has host network. If that's blocked on the managed
account, the fallbacks are (a) host-side sync (CLI / a scheduled `git pull`), or (b) a
**hosted HTTP MCP** (Cortex-as-a-service) that fits the remote-MCP model.

**Surface matrix.**
- **Claude Code CLI:** full (binary + skills + `~/.claude/CLAUDE.md`). Done.
- **Cowork agent:** skills + connected-folder profile; the binary is runnable as a
  local MCP server (Settings > Developer / `.mcpb`) subject to the admin toggle +
  verification, else host-side sync.
- **Desktop chat (non-Cowork):** *could* run the binary via `.mcpb` / Local MCP servers -
  a separate surface, lower priority. (`.mcpb` v0.3 supports `server.type: "binary"` +
  `platform_overrides` + `user_config` for the PAT, if we ever pursue it.)
- **Browser / mobile:** out (no local runtime).

**Sub-tasks for Cowork:** (i) package/deliver the skills to Cowork (confirm whether a
local `.plugin` can be side-loaded or it must go via the curated Directory); (ii) the
git-clone-as-connected-folder profile flow + host-side sync (CLI or a scheduled
`git pull`); (iii) *(optional, later)* a hosted HTTP MCP for in-Cowork sync.

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
