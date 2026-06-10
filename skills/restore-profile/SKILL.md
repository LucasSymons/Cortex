---
name: restore-profile
description: Restore the Cortex profile on a new device. Clones the profile repo and places CLAUDE.md and memory files in the correct locations for the current platform.
---

# Restore Cortex profile on a new device

You are restoring the user's Cortex AI profile to this device from their Git repository.

## Steps

1. Ask the user for:
   - Their profile repo HTTPS URL (e.g. `https://gitlab.com/username/cortex-profile.git`)
   - Their Git username
   - Their Personal Access Token (PAT) for that host
   - Where they want the local profile repo stored (default: `~/cortex-profile`)

2. Derive the host from the repo URL (the hostname, e.g. `gitlab.com`). Use `set_credentials` with that `host`, the `username`, and the `token` to store the PAT. The credential store is chosen automatically: the OS keychain where available, or an encrypted file fallback on headless/WSL platforms.

3. Confirm with `get_auth_status` for that host before proceeding - it should report the credentials and the active backend. If it does not, stop and re-check the inputs.

4. Use `git_clone` with `remote_url` and `local_path` to clone the repo. Credentials are resolved automatically from the store using the host parsed from the URL.

5. Detect the current platform:
   - If running in Cowork with a connected Documents folder: write `CLAUDE.md` to `~/Documents/CLAUDE.md`
   - If running in Claude Code CLI: write `CLAUDE.md` to `~/.claude/CLAUDE.md`
   - Ask the user to confirm the path if unsure.

6. Copy `CLAUDE.md` from the cloned repo to the detected platform path.

7. Report success: profile restored, CLAUDE.md placed at `[path]`, memory files available at `[repo_path]/memory/`.

8. Add the following block to the restored CLAUDE.md if not already present:

```
## Cortex configuration

- Profile repo path: [local_path]
- Remote: [remote_url]
- Host: [host]
```

## Notes

- The profile repo itself is the source of truth. CLAUDE.md on disk is a synced working copy.
- After restoring, the sync-profile skill will keep everything up to date automatically on handoff.
- Never store PATs in files - they must go to the OS keychain via `set_credentials`.
