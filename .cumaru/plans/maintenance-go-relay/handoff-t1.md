---
human_revised: false
plan: maintenance-go-relay
task: T1
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-relay T1 — session.ClaudeBin extracted from ClaudeAgents, and nudge.RingClaude starts a detached throwaway claude -p relay that removes its own transcript by exact path.
---

# Hand-off — maintenance-go-relay / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/session/agents.go`](go/internal/session/agents.go) | modified — `ClaudeBin(home)` (PATH, else `~/.local/bin/claude`) extracted; `ClaudeAgents` calls it |
| [`go/internal/session/agents_test.go`](go/internal/session/agents_test.go) | modified — `TestClaudeBin_PathThenLocalBin` |
| [`go/internal/nudge/claude.go`](go/internal/nudge/claude.go) | created — `RingClaude(bin, home, name, prompt)`, `ringScript`, `newUUID` |
| [`go/internal/nudge/claude_test.go`](go/internal/nudge/claude_test.go) | created — builtins-only fake `claude`: exact argv/cwd/stdin, lowercase v4 session id, transcript removed and sibling kept, fresh id per ring, non-executable/missing/directory bin rings nothing |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Detachment through `/bin/sh -c` with all values positional (`$1`..`$5`) and `Setsid`: the shell outlives the Go process and runs `claude` then `rm -f <exact transcript>`, as bash's `( ... ) &` subshell does. The prompt is an argument piped by `printf`, so no Go-side stdin pipe must stay alive.
- stdio left nil (`/dev/null`), so the ring never holds the hook's stdout open.
- Lowercase v4 UUID from `crypto/rand` instead of macOS `uuidgen` uppercase.
- Unlike bash, a missing `~/.claude` makes `Start` fail on chdir (error returned, nothing runs); bash would skip `claude` via `cd &&` and still run a no-op `rm`.

## Commands run / verification

- Red: `go test -count=1 ./internal/...` — build failed on undefined `ClaudeBin` / `RingClaude`.
- Green: `go test -count=1 ./...` all ok; `go vet ./...` clean; `gofmt -l .` empty.
- Mutation: replacing the `rm -f "$5"` line with `:` made `TestRingClaude_SpawnsDetachedRelayAndRemovesItsTranscript` fail (timed out waiting for transcript removal); restored, green.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- None.
