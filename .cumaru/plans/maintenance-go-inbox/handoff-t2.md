---
human_revised: false
plan: maintenance-go-inbox
task: T2
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-inbox T2 — inbox subcommand wired with a required address and byte-identical output to bash.
---

# Hand-off — maintenance-go-inbox / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/cmd/alter-bridge/main.go`](go/cmd/alter-bridge/main.go) | modified — `inbox` dispatch; root and resolver setup hoisted into `setup()` shared with `send` |
| [`go/cmd/alter-bridge/inbox.go`](go/cmd/alter-bridge/inbox.go) | created — `inboxEnv` and `runInbox` (required address, slug resolution, archiving drain) |
| [`go/cmd/alter-bridge/inbox_test.go`](go/cmd/alter-bridge/inbox_test.go) | created — in-process tests for print+archive, id resolution, empty mailbox, and every refusal |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Exit codes mirror `send`: `0` success (including an empty or missing mailbox); `2` usage (missing/extra argument, `-h`, malformed address, unknown provider); `1` ambiguous or empty slug, or a filesystem error.
- `setup()` exits before any subcommand runs when `EnsureRoot` fails, as before; it replaces the inline setup in the `send` case rather than copying it.

## Commands run / verification

- `go test ./cmd/...` before implementation — failed on undefined `runInbox`/`inboxEnv` (expected red); the `-h` case failed on `address: must be provider:value (got "-h")` before the usage guard was added.
- `go test -count=1 ./...` — passed; `go vet ./...` clean; `gofmt -l .` empty.
- Smoke with the built binary against a scratch `ALTER_BRIDGE_ROOT`: two Go `send`s (one `type: question`) plus a hand-dropped `note.md`, then the root copied; Go `inbox claude:smoke-box` and bash `inbox claude:smoke-box` on the copy produced identical stdout (`diff` clean) and identical `.archive/` listings; mailbox left empty. Re-dropping `note.md` and draining again archived it as `note.md__to_claude_smoke-box__note.<8 hex>.md` without touching the first copy. No argument → usage, exit 2; `bogus` → exit 2.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- `peek` can reuse `inboxEnv` as is and call `mailbox.Drain(..., false, ...)`.
