---
human_revised: false
plan: maintenance-go-hook
task: T1
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-hook T1 — Go hook takes a provider or provider:value argument, reads session id and cwd from the payload, gates on ~/agentic-workspace, and drains like inbox; bash parity for explicit addresses.
---

# Hand-off — maintenance-go-hook / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/cmd/alter-bridge/hook.go`](go/cmd/alter-bridge/hook.go) | created — `hookUsage`, `hookEnv`, `runHook` (argument check, payload decode, workspace gate, `Resolver.Slug`, `mailbox.Drain` archiving), `firstString` |
| [`go/cmd/alter-bridge/hook_test.go`](go/cmd/alter-bridge/hook_test.go) | created — in-process tests: id from `session_id`/`thread_id`/`threadId`, first non-empty wins, explicit address ignores payload id, workspace gate cases, unusable payloads silent, missing mailbox creates nothing, usage refusals exit 2, ambiguous name exits 1 |
| [`go/cmd/alter-bridge/main.go`](go/cmd/alter-bridge/main.go) | modified — `hook` case with `broker.Resolve` (root never created), `session.NewResolver`, `os.Getwd`, and a usage line |
| [`go/internal/address/address.go`](go/internal/address/address.go) | modified — `IsProvider`, so a bare provider argument is validated against the one provider list |
| [`go/internal/address/address_test.go`](go/internal/address/address_test.go) | modified — `IsProvider` accepts the three providers only |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Identity follows the go-send model (plan Overview): no `self_addr`, no env. The provider is the argument; a bare provider takes its value from the payload session id, resolved like `inbox`; no id means a silent exit 0 rather than a cwd-derived slug.
- Root from `broker.Resolve`, not `setup()`: bash `hook` never creates the root or the mailbox.
- Payload keys are read as strings only; a non-string `session_id` counts as absent (jq would print it).
- Deviations: unknown providers and a bad explicit address exit 2 even outside the workspace (bash checks the shape only, and only after the gate); argument count is enforced.

## Commands run / verification

- `go test -count=1 ./...` before implementation — build failed on undefined `runHook`/`hookEnv` and `IsProvider` (expected red).
- `go test -count=1 ./...` — all ok; `go vet ./...` clean; `gofmt -l .` empty.
- Smoke (scratch `HOME` with `agentic-workspace/papa`, fixture copied to the same `ALTER_BRIDGE_ROOT` path per run, bash vs Go `hook claude:papa-work`): payload `cwd` inside — stdout (3 messages incl. a hand-dropped `note.md`) and `find . | sort` identical, exit 0; `cwd` outside — both silent, trees unchanged; missing mailbox — both silent, nothing created; absent root — both exit 0 and neither creates it. `gemini:x` — bash exit 0, Go exit 2 (recorded deviation).
- Go-only smoke: `hook claude` with a scratch `~/.claude/projects/p/sid-1.jsonl` titled `Papa Work` drained `claude/papa-work`; garbage payload and a payload with no id exit 0 silently; no argument prints usage, exit 2.

## Pending / follow-ups

- Cutover must pass the provider (`hook claude` / `hook codex`) in the hook wiring.

## Suggestions for the Lead

- `watchpaths` and `relay` read the same payload fields and the same workspace gate; lift `firstString` and the gate out of `hook.go` when the second caller lands, not before.
