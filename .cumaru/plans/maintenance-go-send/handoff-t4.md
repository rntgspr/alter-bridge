---
human_revised: false
plan: maintenance-go-send
task: T4
status: complete
date: 2026-09-22
summary: Hand-off for maintenance-go-send T4 — send subcommand wired end to end with required --from and bash-compatible flag placement.
---

# Hand-off — maintenance-go-send / T4

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/cmd/alter-bridge/main.go`](go/cmd/alter-bridge/main.go) | modified — hello-world replaced by subcommand dispatch; reads only `HOME` and `ALTER_BRIDGE_ROOT` |
| [`go/cmd/alter-bridge/send.go`](go/cmd/alter-bridge/send.go) | created — `runSend` (flags, required `--from`, type check, parse) and `deliver` (resolve, body, write) |
| [`go/cmd/alter-bridge/send_test.go`](go/cmd/alter-bridge/send_test.go) | created — in-process tests for bodies, flag placement, id resolution and every refusal |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Exit codes: `0` delivered; `2` usage errors (missing recipient or `--from`, bad flag, malformed address or unknown provider — bash `check_addr` also exits 2); `1` for invalid `--type`, ambiguous or empty slug, unreadable body, or write failure.
- Flags may appear before or after the positional body, as in bash; `flag.FlagSet.Parse` is re-run past each positional.
- `sendEnv` bundles root, resolver and stdio so tests run in-process; the concrete `session.Resolver` is injected with fake stores rather than hidden behind an interface.

## Commands run / verification

- `go test ./cmd/...` before implementation — failed on undefined `runSend`/`sendEnv` (expected red).
- `go test -count=1 ./...` — passed; `go vet ./...` clean; `gofmt -l .` empty.
- Mutation: parsing only up to the first positional failed `TestSend_DashReadsStdinAndFlagsMayFollowBody` and `TestSend_BodyFromFile`; restored → pass.
- `rg Getenv` outside tests: only `main.go` (`HOME`, `ALTER_BRIDGE_ROOT`).
- Smoke with the built binary against a scratch root: `send codex:<live thread id> --from claude:<this session id> --type question` delivered to `codex/assumir-papel-de-lead/` signed `claude:<session id>`; bash `peek codex:assumir-papel-de-lead` printed it. Missing `--from` → usage, exit 2. `opencode:Activating Cumaru lead role` → ambiguity listing 9 ids, exit 1, nothing written.

## Pending / follow-ups

- `build.sh` still writes `bin/alter-bridge`; wiring that binary into hooks belongs to cutover, out of scope.

## Suggestions for the Lead

- A slug may contain `_`, so `__from_<provider>_<slug>__` is only unambiguous because providers never contain `_`; readers must split the provider at the first `_` after `__from_`. Worth noting for `maintenance-go-inbox`.
