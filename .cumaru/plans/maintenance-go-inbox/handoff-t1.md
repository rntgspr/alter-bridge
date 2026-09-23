---
human_revised: false
plan: maintenance-go-inbox
task: T1
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-inbox T1 — archive-parameterized mailbox drain landed with bash-parity naming and collision tests.
---

# Hand-off — maintenance-go-inbox / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/mailbox/mailbox.go`](go/internal/mailbox/mailbox.go) | created — `Drain(root, box, archive, w)` and the private `archiveOne` move with bash naming and collision suffixes |
| [`go/internal/mailbox/mailbox_test.go`](go/internal/mailbox/mailbox_test.go) | created — order and exact output bytes, archive on/off, collision chain, hand-dropped names, skipped entries, missing mailbox |
| [`go/internal/message/message.go`](go/internal/message/message.go) | modified — `randomID` exported as `RandomID` so the archive suffix reuses it |
<!-- /cumaru:touched -->

## Decisions made during implementation

- `Drain` returns only an error; the bash `quiet` count is left to `maintenance-go-archive` (pass `io.Discard` and add the count then).
- Directories named `*.md` are skipped; bash would `mv` them into `.archive/` and then fail on `cat`. Deliberate, harmless divergence.
- `.archive/` is created lazily, only when a message is archived, as in bash.

## Commands run / verification

- `go test ./internal/mailbox/...` before implementation — failed on undefined `Drain`/`newID` (expected red).
- `go test -count=1 ./...` — passed; `go vet ./...` clean; `gofmt -l .` empty.
- Mutations: splitting at the first `__` and dropping the dotfile filter failed three tests; replacing the collision suffix with `break` failed `TestDrain_ArchiveCollisionAppendsRandomSuffix`; restored → pass.

## Pending / follow-ups

- Byte parity with bash output is proven end to end in T2's smoke run.

## Suggestions for the Lead

- `peek` is `mailbox.Drain(root, box, false, w)`; `archive` is `Drain(..., true, io.Discard)` plus a count.
