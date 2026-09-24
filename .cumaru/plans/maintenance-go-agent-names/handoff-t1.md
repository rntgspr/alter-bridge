---
human_revised: false
plan: maintenance-go-agent-names
task: T1
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-agent-names T1 — the Go Claude store fills untitled transcripts with the live `claude agents` name and appends running sessions without a transcript.
---

# Hand-off — maintenance-go-agent-names / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/session/store.go`](go/internal/session/store.go) | modified — `claudeStore` merges `ClaudeAgents(home)` (error ignored): empty `custom-title` takes the live name, live agents with no transcript are appended |
| [`go/internal/session/store_test.go`](go/internal/session/store_test.go) | modified — live-name merge, broken-binary degradation (missing / non-zero / invalid JSON), resolver id -> name, name -> slug, and ambiguity across a titled transcript and a live name; existing title test pinned to an empty PATH |
<!-- /cumaru:touched -->

## Decisions made during implementation

- The merge lives in the store, not in `Resolver.Slug`, so id -> name and the ambiguity rule are reused unchanged and Codex/OpenCode are untouched.
- Transcript `custom-title` wins over the live name (after `/rename` they agree).

## Commands run / verification

- Red: `go test -count=1 ./internal/session/` failed `TestClaudeStore_LiveAgentNames` (s2 unnamed, s3 missing) and `TestResolver_ClaudeLiveNames` (`Slug(id) = "s2"`).
- Green: `go test -count=1 ./...` all `ok`; `go vet ./...` clean; `gofmt -l .` empty.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- None.
