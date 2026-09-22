---
human_revised: false
plan: maintenance-rename-alter-bridge
task: T3
status: complete
date: 2026-09-21
summary: Hand-off for T3 — env var and default mailbox root renamed.
---

# Hand-off — maintenance-rename-alter-bridge / T3

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`skills/alter-bridge/scripts/alter-bridge`](../../../skills/alter-bridge/scripts/alter-bridge) | `AGENT_BRIDGE_ROOT`/`~/.agent-bridge` -> `ALTER_BRIDGE_ROOT`/`~/.alter-bridge`; all remaining `agent-bridge` command-name text in usage/error strings -> `alter-bridge`; `AGENT-BRIDGE MESSAGE` marker -> `ALTER-BRIDGE MESSAGE` |
| [`skills/alter-bridge/scripts/install-hooks.sh`](../../../skills/alter-bridge/scripts/install-hooks.sh) | `bridge="$here/agent-bridge"` -> `alter-bridge`; error message and Codex hook description seed text updated |
| [`skills/alter-bridge/SKILL.md`](../../../skills/alter-bridge/SKILL.md) | remaining `AGENT_BRIDGE_ROOT`/`~/.agent-bridge` occurrences (root line, doorbell section, archive rule, purge note) -> `ALTER_BRIDGE_ROOT`/`~/.alter-bridge`; title `# Agent Bridge` -> `# Alter Bridge` |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Task scope named only the env var and mailbox root, but the CLI script's usage text (`agent-bridge send ...`, `agent-bridge inbox ...`, etc., lines ~129-144) and the `AGENT-BRIDGE MESSAGE` printf marker also carried the old command name and nothing else in the plan covers CLI-script body content — renamed both for consistency with the plan's Acceptance Criteria ("every internal reference ... MUST be updated"), via a single `sed` pass rather than one-by-one edits.
- Renamed `# Agent Bridge` -> `# Alter Bridge` heading in SKILL.md, left untouched by T2 since it wasn't a command/path reference but is clearly the skill's display title.

## Commands run / verification

- `grep -rn 'AGENT_BRIDGE\|agent-bridge' skills/alter-bridge/` — empty.
- `ALTER_BRIDGE_ROOT=$HOME/.agent-bridge bash skills/alter-bridge/scripts/alter-bridge who` — ran clean against the (at-the-time) old mailbox root name, confirming the rename didn't break parsing ahead of T5.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- None.
