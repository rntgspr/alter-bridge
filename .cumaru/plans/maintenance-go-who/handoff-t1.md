---
human_revised: false
plan: maintenance-go-who
task: T1
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-who T1 — session.ClaudeAgents reads live interactive sessions from `claude agents --json`, with the ~/.local/bin fallback.
---

# Hand-off — maintenance-go-who / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/session/agents.go`](go/internal/session/agents.go) | created — `Agent{ID, Name, Cwd}` and `ClaudeAgents(home)`: PATH lookup then `<home>/.local/bin/claude`, `agents --json`, `kind == "interactive"` only, CLI order |
| [`go/internal/session/agents_test.go`](go/internal/session/agents_test.go) | created — fake `claude` shell script on a temp PATH: interactive filter and order, `~/.local/bin` fallback, missing binary / non-zero exit / invalid JSON errors |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Unnamed interactive agents are returned; skipping them is `who`'s formatting rule, not the reader's.
- Any failure (missing binary, non-zero exit, invalid JSON) is an error; the caller decides to degrade.
- The fake binary uses only `sh` builtins, because the tests set `PATH` to the fake's directory alone.

## Commands run / verification

- `go test ./internal/session/` before implementation — failed on undefined `Agent` / `ClaudeAgents` (expected red).
- `go test -count=1 ./...` — passed; `go vet ./...` clean; `gofmt -l .` empty.
- Mutation: replacing the fallback with a bare `claude` made `TestClaudeAgents_FallsBackToLocalBin` fail; restored, green.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- `relay`/`hook` can reuse `ClaudeAgents` for any live Claude session lookup (e.g. bash `claude_session_name`) instead of shelling out again.
