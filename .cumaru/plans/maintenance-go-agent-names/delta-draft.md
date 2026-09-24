---
human_revised: false
plan: maintenance-go-agent-names
status: draft
date: 2026-09-23
summary: Delta draft for maintenance-go-agent-names — the Go Claude session store falls back to the live `claude agents --json` name.
---

# Delta draft — maintenance-go-agent-names

## specs/bridge/index.md

### Modified Requirements

- Decision `maintenance-go-send`, "Session state is read live": Claude sessions
  are named by their transcript `custom-title`, else by the live
  `claude agents --json` name for that `sessionId` (running sessions without a
  transcript are listed too); a failing `claude` binary leaves transcript
  titles only. (was: Claude `custom-title` entries only)
- Decision `maintenance-go-hook`: an unrenamed running session's mailbox is
  keyed by its live automatic name (bash parity), not its id; the id remains
  the key only when neither a title nor a live name is known.
  (was: an unnamed session's mailbox is keyed by its id)

### Added Requirements

- WHEN bash `watchpaths` and Go `watchpaths claude` receive the same
  SessionStart payload for a live unrenamed Claude session THE SYSTEM SHALL
  register the same `claude/<automatic-name>` mailbox (`maintenance-go-agent-names`).
