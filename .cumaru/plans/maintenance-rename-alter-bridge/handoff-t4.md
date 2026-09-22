---
human_revised: false
plan: maintenance-rename-alter-bridge
task: T4
status: complete
date: 2026-09-21
summary: Hand-off for T4 — README.md references updated.
---

# Hand-off — maintenance-rename-alter-bridge / T4

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`README.md`](../../../README.md) | title, marketplace install commands, plugin cache paths, `skills/agent-bridge/...` paths, `AGENT_BRIDGE_ROOT`/`~/.agent-bridge`, and the file-layout table all updated to `alter-bridge` |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Applied a single `sed` pass (`agent-bridge` -> `alter-bridge`, `AGENT_BRIDGE_ROOT` -> `ALTER_BRIDGE_ROOT`) rather than per-line edits, since every occurrence in the file needed the same substitution with no exceptions (unlike T2/T3 in SKILL.md).

## Commands run / verification

- `grep -n 'agent-bridge\|AGENT_BRIDGE' README.md` — empty.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- None.
