---
human_revised: false
plan: maintenance-go-purge
task: T1
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-purge T1 — mailbox.Purge deletes the non-dot *.md entries directly under .archive and counts them; ErrNoArchive marks a missing archive.
---

# Hand-off — maintenance-go-purge / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/mailbox/mailbox.go`](go/internal/mailbox/mailbox.go) | modified — new `ErrNoArchive` sentinel and `Purge(root) (int, error)`: non-dot `*.md` entries directly under `root/.archive` that resolve to a non-directory are removed and counted |
| [`go/internal/mailbox/mailbox_test.go`](go/internal/mailbox/mailbox_test.go) | modified — `Purge` removes 2 archived files and a file symlink (target kept), keeps dot `.md`, `notes.txt`, `sub/nested.md`, a `dir.md` directory, and a broken symlink; empty `.archive` is 0; missing `.archive` and a plain-file `.archive` return `ErrNoArchive` |
<!-- /cumaru:touched -->

## Decisions made during implementation

- `.archive` not being a directory is a sentinel error rather than a zero count, because bash distinguishes `nothing to purge` from `purged 0 ...`.
- Entries are checked with `os.Stat` (follows symlinks, like bash `-e`): broken symlinks are skipped, file symlinks are removed as links.
- Deviation: a directory named `*.md` is skipped; bash `rm -f` fails on it and aborts with exit 1 mid-purge. Carries a one-line DX comment in the code.

## Commands run / verification

- `go test -count=1 ./internal/mailbox/` before implementation — build failed on undefined `Purge` / `ErrNoArchive` (expected red).
- `go test -count=1 ./...` — all packages ok; `go vet ./...` clean; `gofmt -l .` empty.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- None.
