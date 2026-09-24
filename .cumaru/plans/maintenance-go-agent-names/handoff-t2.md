---
human_revised: false
plan: maintenance-go-agent-names
task: T2
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-agent-names T2 — scratch evidence that bash and Go `watchpaths` register the same auto-named mailbox for live unrenamed sessions, and Go `hook claude` drains it.
---

# Hand-off — maintenance-go-agent-names / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`bin/alter-bridge`](bin/alter-bridge) | rebuilt by `go/build.sh` (gitignored) |
<!-- /cumaru:touched -->

## Decisions made during implementation

- The vault session's payload used the repo `cwd` so it passes the workspace gate; the comparison is about name resolution, not the gate.

## Commands run / verification

- `bin/alter-bridge who`: `claude:alter-bridge-11 4ff25905-...` and `claude:vault-79 6e4c01e3-...`; neither transcript has a real `custom-title` entry (jq count 0).
- Same SessionStart payload, separate scratch roots: bash `watchpaths` -> `claude/alter-bridge-11`, Go `watchpaths claude` -> `claude/alter-bridge-11`; for `6e4c01e3-...` both -> `claude/vault-79`.
- Regression direction: the pre-fix tree (`69f04b8`) built into scratch registers `claude/4ff25905-8f72-4c7a-8371-b2a90f60cf7c` for the same payload.
- Go `send claude:alter-bridge-11 --from claude:parity-probe` into the scratch root, then `hook claude` with the session payload printed the message (exit 0) and archived it to `.archive/..__to_claude_alter-bridge-11__865d7a43.md`.
- Real `~/.alter-bridge` untouched by these runs (all `ALTER_BRIDGE_ROOT` under the scratchpad).

## Pending / follow-ups

- Observed: this session's automatic name changed from `alter-bridge-ef` (Blocker, earlier today) to `alter-bridge-11` after a resume; both bash and Go follow the current name, so mail sent to the old address is not drained. Pre-existing bash behavior, relevant to `maintenance-go-cutover` AC6.

## Suggestions for the Lead

- None.
