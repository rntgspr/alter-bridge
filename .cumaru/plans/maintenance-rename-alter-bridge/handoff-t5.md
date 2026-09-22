---
human_revised: false
plan: maintenance-rename-alter-bridge
task: T5
status: complete
date: 2026-09-21
summary: Hand-off for T5 — mailbox migration (moot) and hook reinstall.
---

# Hand-off — maintenance-rename-alter-bridge / T5

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| `~/.claude/settings.json` | `SessionStart`/`UserPromptSubmit`/`FileChanged` hook commands repointed to `.../alter-bridge/skills/alter-bridge/scripts/alter-bridge`; backup written by `install-hooks.sh` |
| `~/.codex/hooks.json` | `UserPromptSubmit` hook command and description repointed/updated; backup written by `install-hooks.sh` |
<!-- /cumaru:touched -->

## Decisions made during implementation

- The plan's premise (`~/.agent-bridge` holds live `claude/`, `codex/`, `.archive/` content) no longer held at execution time: neither `~/.agent-bridge` nor `~/.alter-bridge` exists on this machine. Skipped the `mv` migration step as moot rather than fabricating a directory — the mailbox root is created lazily on first `send`, so there was nothing to lose.
- Ran `install-hooks.sh` with no argument to reconcile both runtimes in one pass, per its documented behavior.
- Confirmed the marketplace-enablement key `"agent-bridge@agents-marketplace": true` in `~/.claude/settings.json` is untouched, per the plan's Out of scope note — this plan does not update it.

## Commands run / verification

- `find ~/.agent-bridge -name '*.tmp' -o -name 'msg.*'` — path did not exist (nothing to check).
- `bash skills/alter-bridge/scripts/install-hooks.sh` — both Claude and Codex hooks FIXED and written; backups at `~/.claude/settings.json.bak.20260921182659` and `~/.codex/hooks.json.bak.20260921182659`.
- `grep -n alter-bridge ~/.claude/settings.json ~/.codex/hooks.json` — all three Claude hooks and the Codex hook show the new path.
- `grep -n 'skills/agent-bridge' ~/.claude/settings.json ~/.codex/hooks.json` — empty.
- `bash skills/alter-bridge/scripts/alter-bridge who` — resolved clean with no `ALTER_BRIDGE_ROOT` override, listing this and other live sessions.

## Pending / follow-ups

- Codex hook trust: `install-hooks.sh` reports the rewritten Codex hook needs re-approval in `config.toml` (`[hooks.state."<path>:user_prompt_submit:0:0"].trusted_hash`) before it fires — a one-time manual step for Renato, not something this plan automates.
- Other machines/agents with hooks already installed against `agent-bridge` still point at the old binary name until someone reinstalls there (plan's stated Out of scope).

## Suggestions for the Lead

- None.
