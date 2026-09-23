---
human_revised: false
plan: maintenance-go-archive
task: T1
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-archive T1 — mailbox.Archive counts through Drain's own loop without reading, and mailbox.Boxes lists every non-dot mailbox.
---

# Hand-off — maintenance-go-archive / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/mailbox/mailbox.go`](go/internal/mailbox/mailbox.go) | modified — `Drain`'s loop moved into private `drain` (returns a count, skips read/print when `w` is nil); new `Archive` (count-only archiving) and `Boxes` (non-dot `<provider>/<slug>` dirs, byte order) |
| [`go/internal/mailbox/mailbox_test.go`](go/internal/mailbox/mailbox_test.go) | modified — `Archive` moves with the same names as `Drain`, counts 2 with an unreadable message (never read), missing mailbox 0 without `.archive/`; `Boxes` order and dot/file skipping, missing root empty |
<!-- /cumaru:touched -->

## Decisions made during implementation

- One loop for `Drain` and `Archive` (`drain`), so move, naming, and collision suffix have a single implementation (AC3). `Drain`'s signature is unchanged.
- `Boxes` returns directory names as-is (no provider validation), matching bash's `"$ROOT"/*/*` sweep.

## Commands run / verification

- `go test -count=1 ./internal/mailbox/` before implementation — build failed on undefined `Archive` / `Boxes` (expected red).
- `go test -count=1 ./...` — all packages ok; `go vet ./...` clean; `gofmt -l .` empty.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- `purge` can locate `.archive` the same way (`root/.archive`); no new helper needed yet.
