---
human_revised: false
plan: maintenance-go-cutover
task: T3
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-cutover T3 — live Claude and Codex hooks rewired to the Go binary with backups, verified against scratch roots only.
---

# Hand-off — maintenance-go-cutover / T3

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
<!-- /cumaru:touched -->

No repository file. Outside the repo: `~/.claude/settings.json` (three
alter-bridge command strings; backup `settings.json.bak.20260923211158`) and
`~/.codex/hooks.json` (one command string; backup `hooks.json.bak.20260923211158`).

## Decisions made during implementation

- Edited by exact string replacement (not the installer) so formatting and
  every other key stay byte-identical; `diff` against the backups shows only
  the command lines.

## Commands run / verification

- `jq empty` passes on both files; the updated installer against copies of
  them reports `already present` / `left untouched`.
- Wired strings (read back from the live files) run through `sh -c` with
  `ALTER_BRIDGE_ROOT` on a scratch root:
  - SessionStart payload -> `watchpaths claude` printed the `watchPaths`
    JSON for `claude` and `claude/alter-bridge-11`, exit 0.
  - UserPromptSubmit payload -> `hook claude` printed and archived the seeded
    message into `.archive/..._to_claude_alter-bridge-11__...`, exit 0.
  - FileChanged `change` -> `relay claude` printed 0 bytes, exit 0, one
    `relay.log` line.
  - `hook codex`: exit 0, silent, outside and inside the workspace (empty mailbox).
- Live root read-only: `bin/alter-bridge who` and `peek` (by name and by id)
  exit 0; `find ~/.alter-bridge` checksum identical before and after.

## Pending / follow-ups

- Codex re-approval of the changed hook; a live Codex turn to confirm its
  payload id fields.

## Suggestions for the Lead

- None.
