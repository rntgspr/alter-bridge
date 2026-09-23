---
human_revised: false
plan: maintenance-go-purge
task: T2
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-purge T2 — broker.Resolve split out of EnsureRoot and purge wired without creating the root; stdout and resulting tree byte-identical to bash.
---

# Hand-off — maintenance-go-purge / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/broker/root.go`](go/internal/broker/root.go) | modified — new `Resolve` (override or `home/.alter-bridge`, HOME guard, no mkdir); `EnsureRoot` now calls it, then `MkdirAll` |
| [`go/internal/broker/root_test.go`](go/internal/broker/root_test.go) | modified — `Resolve` picks override / default without creating it and refuses empty and `/` HOME |
| [`go/cmd/alter-bridge/purge.go`](go/cmd/alter-bridge/purge.go) | created — `purgeUsage`, `runPurge`: no arguments, `mailbox.Purge`, `nothing to purge` on `ErrNoArchive`, else `purged <n> archived message(s) from <root>/.archive` |
| [`go/cmd/alter-bridge/purge_test.go`](go/cmd/alter-bridge/purge_test.go) | created — in-process tests: count line and deletion, empty archive `purged 0`, missing archive and missing root `nothing to purge` (root not created), arguments refused with nothing deleted, deletion failure exits 1 |
| [`go/cmd/alter-bridge/main.go`](go/cmd/alter-bridge/main.go) | modified — `purge` case resolving the root with `broker.Resolve` (not `setup()`), and a usage line |
<!-- /cumaru:touched -->

## Decisions made during implementation

- `purge` needs the root but, like bash, must not create it; the guard is shared via `broker.Resolve` instead of copied into `main.go`.
- `runPurge` takes the root and writers directly; it needs no resolver, so it does not reuse `inboxEnv`.
- Deviation: any argument, including `-h`/`--help`, is a usage error (exit 2); bash ignores arguments.

## Commands run / verification

- `go test -count=1 ./...` before implementation — build failed on undefined `Resolve` / `runPurge` (expected red).
- `go test -count=1 ./...` — all ok; `go vet ./...` clean; `gofmt -l .` empty.
- Smoke (scratch roots, `HOME` = empty scratch dir): fixture with 3 archived `*.md`, a file symlink `link.md`, a broken symlink, a dot `.md`, `notes.txt`, `sub/nested.md`, and a pending message, copied twice. Go vs bash stdout `diff` clean (`purged 4 ...`) and `find . | sort` `diff` clean. No `.archive`: both `nothing to purge`, trees identical. Absent root: both `nothing to purge`, exit 0, neither created it; default root under the empty `HOME` not created. `-h`/extra argument exit 2; `HOME=` without override exit 1.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- `hook`, `watchpaths`, and `relay` can use `broker.Resolve` when they must not create the root.
