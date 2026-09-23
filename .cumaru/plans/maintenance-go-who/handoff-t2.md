---
human_revised: false
plan: maintenance-go-who
task: T2
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-who T2 — who subcommand wired, byte-identical to bash on this machine, degrading per provider and never touching the mailbox root.
---

# Hand-off — maintenance-go-who / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/cmd/alter-bridge/who.go`](go/cmd/alter-bridge/who.go) | created — `whoEnv`, `whoUsage`, `runWho`: Claude rows then the first 10 Codex store rows, unnamed skipped, bash `printf` formats, provider errors omit rows |
| [`go/cmd/alter-bridge/who_test.go`](go/cmd/alter-bridge/who_test.go) | created — in-process tests with fake readers: format and order, 10-row Codex window, per-provider degrade with empty stderr, argument refusal |
| [`go/cmd/alter-bridge/main.go`](go/cmd/alter-bridge/main.go) | modified — `who` dispatch case and usage line, wired from `HOME` only (`session.ClaudeAgents`, `NewResolver(home).Stores["codex"]`), bypassing `setup()` |
<!-- /cumaru:touched -->

## Decisions made during implementation

- `who` does not call `setup()`, so it never creates the mailbox root (bash `who` never touches it either).
- The Codex half reuses the resolver's Codex store (DRY); the 10-row window is applied before skipping unnamed rows, as bash's SQL `limit 10` is.
- Archived Codex threads are listed, matching bash's unfiltered query.
- Unlike bash, which ignores extra arguments, any argument is a usage error (exit 2), consistent with the other Go subcommands.

## Commands run / verification

- `go test ./cmd/...` before implementation — failed on undefined `runWho` / `whoEnv` / `whoUsage` (expected red).
- `go test -count=1 ./...` — passed; `go vet ./...` clean; `gofmt -l .` empty.
- Smoke with the built binary on this machine: Go `who` and bash `who` stdout `diff` clean (2 Claude + 4 Codex rows); `ALTER_BRIDGE_ROOT` pointing at a missing dir still missing afterwards.
- Degrade (`env -i`, `PATH` = empty dir): fake `HOME` with `~/.codex` → no Claude rows, Codex rows from `session_index.jsonl`, exit 0, empty stderr; fake `HOME` with only `state_5.sqlite` → no rows at all, exit 0, empty stderr; real `HOME` → Claude rows through the `~/.local/bin/claude` fallback.

## Pending / follow-ups

- In the JSONL fallback the 10-row window picks the oldest-indexed threads (the index has no recency) — recorded as a Decision.

## Suggestions for the Lead

- None.
