---
human_revised: false
plan: maintenance-go-watchpaths
task: T1
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-watchpaths T1 — hook's argument parsing, payload decoding, and workspace gate lifted into hookinput.go; hook tests unchanged and green.
---

# Hand-off — maintenance-go-watchpaths / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/cmd/alter-bridge/hookinput.go`](go/cmd/alter-bridge/hookinput.go) | created — `hookAddress` (bare provider or `provider:value`), `hookSession` (payload session id + `~/agentic-workspace` gate), `firstString` moved from `hook.go` |
| [`go/cmd/alter-bridge/hook.go`](go/cmd/alter-bridge/hook.go) | modified — `runHook` calls the helpers; usage, resolve, and drain unchanged |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Argument-count/usage handling stays in each command (usage text differs); only the argument shape, payload, and gate are shared.
- `hookSession` returns the id and the gate verdict together, since both come from the one payload decode.

## Commands run / verification

- `go test -count=1 ./cmd/alter-bridge/` green before the move (baseline).
- After the move: `hook_test.go` unmodified (`git diff --stat` shows only `hook.go`); `go test -count=1 ./...` all ok; `go vet ./...` clean; `gofmt -l .` empty.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- `relay` can reuse `hookSession` for its gate if its payload carries `cwd`.
