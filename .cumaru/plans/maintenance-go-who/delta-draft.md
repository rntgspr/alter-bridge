---
human_revised: false
plan: maintenance-go-who
status: draft
date: 2026-09-23
summary: Delta for maintenance-go-who — record the Go who as a decision, correct the stale `claude agents --json` observation, and reference the new sources.
---

# Delta draft — maintenance-go-who

The Go `who` is built and tested but not wired into hooks or skills. Following
the earlier go-rewrite precedent, bash Requirements stay authoritative and the
Go behavior is recorded under Decisions until cutover.

## specs/bridge/index.md

### Added Requirements

- None. Bash requirements remain authoritative until cutover.

### Modified Requirements

- None.

### Removed Requirements

- None.

### Modified Decisions

- The `who` decision's 2026-09-22 observation is stale: `claude agents --json`
  works when the real binary runs (Claude Code 2.1.281 lists live
  `kind: interactive` sessions with `sessionId`, `name`, `cwd`); the
  `too many arguments for 'agents'` failure came from an interactive-shell
  wrapper function named `claude`, not the CLI.

### Added Decisions

- 2026-09 (`maintenance-go-who`): the Go `who` is built but not yet wired into
  hooks or the skill. Its stdout is byte-identical to bash `who`: Claude rows
  from `claude agents --json` (`kind == "interactive"`, named only,
  `%-26s %s  %s` with the `cwd`), then the first 10 Codex threads, named only
  (`%-26s %s`), archived ones included as in bash. Deviations: the `claude`
  binary falls back to `~/.local/bin/claude` when absent from `PATH`; the
  Codex half reuses the `send` resolver's Codex store, so it also falls back
  to `session_index.jsonl` when `state_5.sqlite` or `sqlite3` is unavailable
  (that index has no recency, so the 10-row window then shows the
  oldest-indexed threads); any argument is a usage error (exit 2). An
  unavailable provider only omits its rows (exit 0, empty stderr), and `who`
  never creates the mailbox root. The live Claude reader
  (`session.ClaudeAgents`) is separate from the transcript store used for
  addressing, which lists every titled session rather than the running ones.

### Added Files / Reference rows

- `go/cmd/alter-bridge/who.go` — Go `who`: live Claude agents and recent Codex threads in bash's format, degrading per provider.
- `go/internal/session/agents.go` — Go rewrite: `ClaudeAgents`, the live `claude agents --json` reader with the `~/.local/bin` fallback.
- Update the `go/cmd/alter-bridge/` Files entry to name `who`.
