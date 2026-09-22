---
human_revised: false
plan: maintenance-go-send
task: T5
status: complete
date: 2026-09-22
summary: Hand-off for maintenance-go-send T5 — best-effort codex queue nudge after delivery, verified against a fake codex binary.
---

# Hand-off — maintenance-go-send / T5

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/nudge/nudge.go`](go/internal/nudge/nudge.go) | created — `Nudger.Notify`, `Exec`, `BridgeScript` |
| [`go/internal/nudge/nudge_test.go`](go/internal/nudge/nudge_test.go) | created — codex args, slug fallback, non-codex no-op, swallowed failure, zero value |
| [`go/cmd/alter-bridge/send.go`](go/cmd/alter-bridge/send.go) | modified — `sendEnv.Nudger`; notify after printing the path, msgid taken from the delivered name |
| [`go/cmd/alter-bridge/send_test.go`](go/cmd/alter-bridge/send_test.go) | modified — nudge receives the msgid actually written |
| [`go/cmd/alter-bridge/main.go`](go/cmd/alter-bridge/main.go) | modified — wires `nudge.Exec` and `resolver.CodexThreadForSlug` (not predicted in `files:`) |
<!-- /cumaru:touched -->

## Decisions made during implementation

- The msgid is read back from the delivered file name instead of widening `message.Deliver`'s return; the name layout is already a contract `inbox` depends on.
- A zero `Nudger` is a no-op, so tests and future callers that do not want a nudge pass nothing.
- No `command -v codex` pre-check: `exec` failing on a missing binary is swallowed the same way.

## Commands run / verification

- `go test ./...` before implementation — failed on undefined `Nudger`/`BridgeScript` and the unknown `sendEnv.Nudger` field (expected red).
- `go test -count=1 ./...` — 6 packages ok; `go vet ./...` clean; `gofmt -l .` empty.
- Mutation: removing the codex-only guard failed `TestNotify_NonCodexRecipientRunsNothing`; restored → pass.
- `codex queue --help` confirms `--thread <THREAD>` (UUID or exact session name) and `--message <TEXT>`.
- End to end with the built binary and a fake `codex` on `PATH`: args were `queue --thread <live thread uuid> --message "alter-bridge: new message from claude:papa (msgid …). Run: … inbox codex:assumir-papel-de-lead"`; a fake `codex` exiting 7 and no `codex` on `PATH` both left `send` at exit 0.

## Pending / follow-ups

- Not run: a nudge into a real live Codex session (it would inject a notice into the user's session). Needs explicit approval to run or to defer.

## Suggestions for the Lead

- `BridgeScript` must flip to the Go binary at cutover, together with the hooks.
