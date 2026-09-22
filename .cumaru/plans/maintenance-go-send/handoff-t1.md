---
human_revised: false
plan: maintenance-go-send
task: T1
status: complete
date: 2026-09-22
summary: Hand-off for maintenance-go-send T1 — address parsing and byte-level slugify landed with bash-parity tests.
---

# Hand-off — maintenance-go-send / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/address/address.go`](../../../go/internal/address/address.go) | created — `Address`, `Parse` (providers claude, codex, opencode; leading `#` stripped), `Slugify` |
| [`go/internal/address/address_test.go`](../../../go/internal/address/address_test.go) | created — table tests for valid/invalid addresses and bash-captured slugify outputs |
<!-- /cumaru:touched -->

## Decisions made during implementation

- `Slugify` works on bytes, not runes, to match bash `tr`: a multibyte UTF-8 character becomes dashes (`Ção São` → `o-s-o`, `ÁB` → `b`). Expected values were captured by running the bash `slugify` on this machine.
- `Parse` splits on the first `:` only, so `claude:a:b` keeps `a:b` as the value, matching bash `${addr#*:}`.

## Commands run / verification

- `go test ./internal/address/...` before implementation — failed on undefined `Parse`/`Address`/`Slugify` (expected red).
- `go test ./...` — passed; `go vet ./...` clean; `gofmt -l .` empty.
- Mutation check: dropping `opencode` from the provider set and removing the edge trim made `TestParse_Valid` and `TestSlugify_MatchesBash` fail; restoring made them pass.

## Pending / follow-ups

- `Slugify` can return an empty string (`---` → ``, or a name made only of non-ASCII characters). T2/T4 must reject an empty resolved slug instead of delivering to `<root>/<provider>/`.

## Suggestions for the Lead

- Non-ASCII session names collapse badly under bash-parity slugify (`Ção São` → `o-s-o`); transliteration would be friendlier but breaks parity with existing mailboxes. Revisit only at cutover.
