---
human_revised: false
scope: [bridge]
status: in-progress
summary: Port the bash `who` command to Go — list live Claude agents and recent Codex threads in bash's exact format, degrading per provider instead of failing.
targets: [cli]
aux: []
---

# Go rewrite — `who`

## Overview

Port `who` (`skills/alter-bridge/scripts/alter-bridge`, line ~439) to Go. The
Claude half shells out to `claude agents --json` (live interactive sessions,
with their `cwd`); the Codex half reuses the Codex store `maintenance-go-send`
built in `internal/session` (`state_5.sqlite` through the `sqlite3` CLI,
falling back to `session_index.jsonl`) instead of re-querying the database.

`claude agents --json` works on this machine (Claude Code 2.1.281) when the
real binary is invoked; the "too many arguments" failure recorded under the
spec's Decisions came from an interactive shell wrapper, not the CLI. The
transcript-based Claude store used for addressing lists every titled session
ever recorded, not the running ones, so it is not the right source for `who`.

`who` only reads: it never creates the mailbox root or any file.

## Acceptance Criteria (EARS / RFC 2119)

Derived from `specs/bridge/index.md` ("Decisions") and the bash `who`:

- `who` MUST read addressable agents live from each CLI's own state on every run — no local registry, no cache — and MUST NOT create or modify any file, including the mailbox root.
- Claude agents MUST be listed from `claude agents --json`, keeping only `kind == "interactive"` entries with a non-empty `name`, in the CLI's order, each printed as bash does: `printf '%-26s %s  %s\n' "claude:<slugified name>" <sessionId> <cwd>`.
- The `claude` binary MUST be looked up on `PATH`, falling back to `$HOME/.local/bin/claude` (bash `claude_bin()`), so a caller whose PATH lacks `~/.local/bin` still lists Claude agents.
- Codex threads MUST come from the existing `internal/session` Codex store (`state_5.sqlite` newest first, `session_index.jsonl` when the database or `sqlite3` is unavailable); the first 10 rows are considered (bash `limit 10`), unnamed ones skipped, each printed as `printf '%-26s %s\n' "codex:<slugified name>" <id>`.
- WHEN the `claude` binary is missing or fails, or its output is not valid JSON, THE SYSTEM SHALL omit the Claude rows; WHEN neither Codex source is readable THE SYSTEM SHALL omit the Codex rows; in both cases `who` exits 0 with nothing on stderr.
- WHEN given any argument (including `-h`) THE SYSTEM SHALL print usage on stderr and exit 2.
- With both sources available, Go `who` stdout MUST be byte-identical to bash `who` stdout on the same machine.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| T1 | `session.ClaudeAgents` — live `claude agents --json` reader with the `~/.local/bin` fallback | — | done |
| T2 | `who` CLI wiring over the Claude agents reader and the Codex store | T1 | pending |

## Out of scope

- OpenCode rows (bash `who` has none).
- Changing the transcript-based Claude store used by `send`/`inbox`/`peek` addressing.
- Any other subcommand; replacing the bash script or rewiring hooks.

## Risks

- **JSONL fallback order.** `session_index.jsonl` has no recency, so in fallback the "first 10" are the oldest-indexed threads, not the newest. Bash has no fallback at all; accepted and recorded as a Decision.
- **Archived Codex threads** are listed, matching bash (its query has no `archived` filter); they remain addressable by name.
