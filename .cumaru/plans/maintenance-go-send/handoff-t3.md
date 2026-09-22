---
human_revised: false
plan: maintenance-go-send
task: T3
status: complete
date: 2026-09-22
summary: Hand-off for maintenance-go-send T3 — atomic message delivery with bash-identical frontmatter and file naming.
---

# Hand-off — maintenance-go-send / T3

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/message/message.go`](go/internal/message/message.go) | created — `Message`, `ValidType`, `Deliver` (temp file in `.tmp/`, rename into mailbox) |
| [`go/internal/message/message_test.go`](go/internal/message/message_test.go) | created — frontmatter bytes, invalid type, name layout, no `.tmp` leftovers, collision |
<!-- /cumaru:touched -->

## Decisions made during implementation

- `Message.From`/`To` reuse `address.Address`, with `Value` holding the already-resolved slug; resolution stays in T2's resolver.
- Clock and id generator are package-level vars swapped by tests, instead of adding parameters no production caller needs.
- Timestamp milliseconds truncate (Go `.000`), matching bash's integer `us / 1000`.
- The collision check keeps bash semantics (exists-check then rename); the residual race is documented in one comment line.

## Commands run / verification

- `go test ./internal/message/...` before implementation — failed on undefined `Deliver` (expected red).
- `go test -count=1 ./...` — passed; `go vet ./...` clean; `gofmt -l .` empty.
- Mutation: forcing the collision check to always accept failed `TestDeliver_CollisionRegeneratesID`; restored → pass.
- Interop: a message written by `Deliver` into a scratch root was printed correctly by bash `ALTER_BRIDGE_ROOT=<root> alter-bridge peek codex:bridge`.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- None.
