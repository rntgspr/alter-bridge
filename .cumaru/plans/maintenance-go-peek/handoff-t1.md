---
human_revised: false
plan: maintenance-go-peek
task: T1
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-peek T1 — peek subcommand wired through a runner shared with inbox, byte-identical to bash and leaving the mailbox untouched.
---

# Hand-off — maintenance-go-peek / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/cmd/alter-bridge/inbox.go`](go/cmd/alter-bridge/inbox.go) | modified — `runInbox` body extracted into `runDrain(args, env, archive, usage)`; `runInbox` delegates with archiving on |
| [`go/cmd/alter-bridge/peek.go`](go/cmd/alter-bridge/peek.go) | created — `peekUsage` and `runPeek` (`runDrain` with archiving off) |
| [`go/cmd/alter-bridge/peek_test.go`](go/cmd/alter-bridge/peek_test.go) | created — in-process tests for oldest-first print with no move and no `.archive/`, id resolution, empty mailbox, and every refusal |
| [`go/cmd/alter-bridge/main.go`](go/cmd/alter-bridge/main.go) | modified — `peek` dispatch case and usage line, reusing `setup()` and `inboxEnv` |
<!-- /cumaru:touched -->

## Decisions made during implementation

- `inboxEnv` is reused unchanged as the env for both commands (no rename, to keep the diff minimal).
- Exit codes identical to `inbox`: `0` success (including a missing mailbox); `2` usage; `1` ambiguous or empty slug, or a filesystem error.

## Commands run / verification

- `go test ./cmd/...` before implementation — failed on undefined `runPeek` (expected red). With a stub delegating to `runInbox`, the new tests failed on "message moved" and on the inbox usage text (behavioral red).
- `go test -count=1 ./...` — passed; `go vet ./...` clean; `gofmt -l .` empty.
- Smoke with the built binary against a scratch `ALTER_BRIDGE_ROOT`: two Go `send`s (one `type: question`) plus a hand-dropped `note.md`; Go `peek claude:smoke-box` and bash `peek claude:smoke-box` produced identical stdout (`diff` clean, 3 messages oldest first); the root's file listing was identical before and after, with no `.archive/`. `claude:nobody` → exit 0, silent; no argument → usage, exit 2; `bogus` → exit 2.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- `archive` can reuse `runDrain`'s argument path only if it adds a quiet count mode to `mailbox.Drain`; `hook` needs its own identity path, not `runDrain`.
