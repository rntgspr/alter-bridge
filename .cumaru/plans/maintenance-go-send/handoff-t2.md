---
human_revised: false
plan: maintenance-go-send
task: T2
status: complete
date: 2026-09-22
summary: Hand-off for maintenance-go-send T2 — session resolver with archive-aware ambiguity checks over live Claude, Codex and OpenCode stores.
---

# Hand-off — maintenance-go-send / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/internal/session/session.go`](go/internal/session/session.go) | created — `Resolver.Slug`, `CodexThreadForSlug`, `AmbiguousError`, `ErrEmptySlug` |
| [`go/internal/session/session_test.go`](go/internal/session/session_test.go) | created — resolution rules against injected stores |
| [`go/internal/session/store.go`](go/internal/session/store.go) | created — Claude transcript, Codex sqlite/index and OpenCode sqlite readers |
| [`go/internal/session/store_test.go`](go/internal/session/store_test.go) | created — store readers against temp jsonl and sqlite fixtures |
<!-- /cumaru:touched -->

## Decisions made during implementation

- SQLite is read by shelling out to `sqlite3 -json` with `mode=ro`, as bash does; no Go driver dependency. Requires `sqlite3` on `PATH` (3.33+ for `-json`; 3.41.2 here).
- OpenCode's human-facing name is `title`; `slug` is an auto-generated word pair (`happy-rocket`, `calm-island`). Confirmed from live rows; unnamed sessions carry a `New session - <ts>` title, which slugifies uniquely.
- Store readers live in `store.go`, separate from resolution rules in `session.go` (different reasons to change: CLI storage formats vs. addressing policy). Not predicted in `files:`.
- An id match resolves even for an archived session, so `provider:<id>` always works; only name matching filters archived sessions.
- Empty slugs are rejected here with `ErrEmptySlug` (follow-up from T1).

## Commands run / verification

- `go test ./internal/session/...` before implementation — failed on undefined `Session`/`Resolver`/`Store` (expected red).
- `go test -count=1 ./...` — passed (14 session tests, none skipped); `go vet ./...` clean; `gofmt -l .` empty.
- Mutation: dropping the `!s.Archived` filter failed `TestSlug_ArchivedDuplicateIsNotAmbiguous`; restored → pass.
- Live check (temporary build-tagged test, removed): this Claude session id → its id (never renamed); a renamed Claude session → `testing`; Codex thread id → `assumir-papel-de-lead`; OpenCode id → `carregar-contexto-local`; `opencode:Activating Cumaru lead role` → `AmbiguousError` listing 9 ids; `CodexThreadForSlug("assumir-papel-de-lead")` → the live thread id.

## Pending / follow-ups

- The Claude store reads every transcript on each call (~28 MB here). Acceptable now; revisit if `send` latency becomes noticeable.

## Suggestions for the Lead

- Auto-titled OpenCode sessions collide heavily; addressing OpenCode by name only works after a deliberate rename.
