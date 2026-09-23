---
human_revised: false
plan: maintenance-go-archive
task: T2
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-archive T2 — archive subcommand wired; sweep and single-address runs byte-identical to bash in stdout and resulting tree.
---

# Hand-off — maintenance-go-archive / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/cmd/alter-bridge/archive.go`](go/cmd/alter-bridge/archive.go) | created — `archiveUsage`, `runArchive`: one address (parse, resolve, `mailbox.Archive`, `archived <n> from <arg>`) or a sweep over `mailbox.Boxes` (non-zero lines, then `total: <sum>`) |
| [`go/cmd/alter-bridge/archive_test.go`](go/cmd/alter-bridge/archive_test.go) | created — in-process tests: single mailbox only, id-to-name resolution, missing mailbox `0`, sweep order/total skipping dot dirs, empty root `total: 0`, refusals move nothing |
| [`go/cmd/alter-bridge/main.go`](go/cmd/alter-bridge/main.go) | modified — `archive` dispatch case through `setup()` and a usage line |
<!-- /cumaru:touched -->

## Decisions made during implementation

- `archive` reuses `inboxEnv`, `address.Parse`, and the resolver but not `runDrain`, since its address is optional (no address = sweep, as in bash).
- The count line echoes the address as given (bash prints `$addr`), even when it resolved to a different slug.
- Deviation: a missing mailbox reports `archived 0 from <addr>`; bash prints an empty count (`archived  from <addr>`) because `drain` returns before its quiet count.
- Deviation: `-h`/`--help` and more than one argument are usage errors (exit 2); bash treats `-h` as an address and ignores extras. Unknown providers are refused (go-send model); the sweep still takes every non-dot directory as-is.

## Commands run / verification

- `go test ./cmd/...` before implementation — build failed on undefined `runArchive` (expected red).
- `go test -count=1 ./...` — all ok; `go vet ./...` clean; `gofmt -l .` empty.
- Smoke (scratch roots, `HOME` = empty scratch dir): seeded claude/papa (3 incl. hand-dropped `note.md`), codex/bridge (1), opencode/work (1), empty codex/idle, dot dirs with `.md` files, a pre-existing archive entry; copied twice per mode. Sweep and `archive claude:papa`: Go vs bash stdout `diff` clean and `find . | sort` of both roots `diff` clean. Missing mailbox and refusals checked against the binary.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- None.
