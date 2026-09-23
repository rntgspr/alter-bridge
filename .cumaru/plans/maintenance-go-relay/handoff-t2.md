---
human_revised: false
plan: maintenance-go-relay
task: T2
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-relay T2 — Go relay logs every call, rings only for add events of message files in its own mailbox (claude via RingClaude, codex via codex queue), and prints the systemMessage line byte-identical to bash.
---

# Hand-off — maintenance-go-relay / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/cmd/alter-bridge/relay.go`](go/cmd/alter-bridge/relay.go) | created — `relayUsage`, `relayEnv`, `runRelay`, `logRelay` (append, never creates `.tmp`), `messageOrigin` (bash glob and parameter-expansion parity) |
| [`go/cmd/alter-bridge/relay_test.go`](go/cmd/alter-bridge/relay_test.go) | created — own `add` rings and prints; explicit address; 13 silent-but-logged cases (`change`/`unlink`, foreign/nested/other-provider mailbox, outside root, root-prefix sibling, bad basenames, no id, bad/empty payload); log appends without creating `.tmp`; codex nudges; sender/msgid parsing with `&<>"\` raw in prompt and unescaped JSON; usage exits 2 without logging; ambiguous name exits 1 |
| [`go/cmd/alter-bridge/main.go`](go/cmd/alter-bridge/main.go) | modified — `relay` case (`broker.Resolve`, `session.NewResolver`, `Ring` bound to `nudge.RingClaude(session.ClaudeBin(home), ...)`, `Nudger` as in `send`) and a usage line |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Argument validated before stdin is read, so `-h` never blocks and a bad argument is not logged (AC6).
- Prompt built by concatenation, not `%q`: the first draft escaped `"` and `\` inside the note (test failed), bash inserts them raw.
- `json.Encoder` with HTML escaping off; re-enabling escaping failed `TestRelay_SenderAndMsgidParsing`.
- Ring and nudge failures never change the exit code or the notice (bash prints the notice even when no `claude` exists).
- Deviation: when `.tmp` is missing bash prints a redirection error on stderr (`2>/dev/null` guards `printf`, not the `>>`); Go stays silent.

## Commands run / verification

- Red: `go test -count=1 ./cmd/alter-bridge/` — build failed on undefined `runRelay` / `relayEnv`; then the raw-prompt test failed with `%q`.
- Green: `go test -count=1 ./...` all ok; `go vet ./...` clean; `gofmt -l .` empty.
- Smoke (AC7), scratch `HOME` and root copied to the same paths per run, `env -i`, `PATH` = fake `claude`/`codex` (sh builtins) + `/usr/bin:/bin`, bash `relay` vs Go `relay claude`, same payload: own-add, change, foreign, outside-root, non-message — stdout, stderr, exit, `find . | sort`, `relay.log` (time masked), and fake `claude -p` record (UUID masked: cwd `~/.claude`, identical argv and stdin prompt) all identical; relay transcript removed, sibling `other.jsonl` kept. own-add-no-tmp: identical except bash's stderr redirection error (deviation above).

## Pending / follow-ups

- AC8 live verification (T3), user-run.
- Cutover must pass the provider (`relay claude`) in the FileChanged wiring.

## Suggestions for the Lead

- None.
