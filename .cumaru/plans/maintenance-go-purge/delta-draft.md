---
human_revised: false
plan: maintenance-go-purge
status: draft
date: 2026-09-23
summary: Delta for maintenance-go-purge — close the purge spec gap with a Message delivery requirement, record the Go purge and broker.Resolve as a decision, and reference their sources.
---

# Delta draft — maintenance-go-purge

The Go `purge` is built and tested but not wired into hooks or skills. Like
`archive`, `purge` had no requirement at all, so this delta adds one (true of
both bash and Go); the Go-only deviations go under Decisions.

## specs/bridge/index.md

### Added Requirements

Under "Message delivery", after the `archive` requirements:

- WHEN `purge` is invoked THE SYSTEM SHALL permanently delete every `*.md`
  entry directly under `<root>/.archive` (not dot-named, not recursive, no
  other file types), without confirmation and without creating the root, and
  report `purged <n> archived message(s) from <root>/.archive`, or
  `nothing to purge` when `.archive/` is not a directory.

### Modified Requirements

- None.

### Removed Requirements

- None.

### Decisions (added)

- 2026-09 (`maintenance-go-purge`): the Go `purge` is built but not yet wired
  into hooks or the skill. The delete is `mailbox.Purge` next to `Archive`
  (`ErrNoArchive` when `.archive` is not a directory), and the root comes from
  `broker.Resolve`, the guard `EnsureRoot` now wraps, so `purge` never creates
  the root. Stdout and the resulting tree are byte-identical to bash for a
  populated, an absent, and a root-less archive; symlinked `*.md` entries are
  removed as links, broken ones kept. Deviations: a directory named `*.md` is
  skipped (bash `rm` fails and aborts mid-purge with exit 1); any argument,
  including `-h`/`--help`, is a usage error (exit 2); the printed path is
  `filepath.Join`-cleaned.

### Files / Reference (added)

- Files list: `go/cmd/alter-bridge/` line gains `purge`; `mailbox.go` line
  mentions purging the archive; `root.go` line mentions resolving without
  creating.
- Reference row: `go/cmd/alter-bridge/purge.go` — `purge`, deletes the
  archived trail without arguments or confirmation.
- Reference rows for `go/internal/mailbox/mailbox.go` (mentions `Purge`) and
  `go/internal/broker/root.go` (mentions `Resolve`).
