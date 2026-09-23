---
human_revised: false
plan: maintenance-go-archive
status: draft
date: 2026-09-23
summary: Delta for maintenance-go-archive — close the archive spec gap with a Message delivery requirement, record the Go archive as a decision, and reference its source.
---

# Delta draft — maintenance-go-archive

The Go `archive` is built and tested but not wired into hooks or skills. Unlike
earlier ports, `archive` had no requirement at all, so this delta adds one
(true of both bash and Go); the Go-only deviations go under Decisions.

## specs/bridge/index.md

### Added Requirements

Under "Message delivery", after the `peek` requirement:

- WHEN `archive` is given an address THE SYSTEM SHALL move every pending
  message of that mailbox to `.archive/` exactly as `inbox` does, print none
  of them, and report `archived <n> from <address>`.
- WHEN `archive` is given no address THE SYSTEM SHALL do the same for every
  `<provider>/<slug>` mailbox under the root, skipping dot-named directories,
  report one line per mailbox that had messages, then `total: <n>`.

### Modified Requirements

- None.

### Removed Requirements

- None.

### Decisions (added)

- 2026-09 (`maintenance-go-archive`): the Go `archive` is built but not yet
  wired into hooks or the skill. It archives through the same
  `internal/mailbox` loop and `archiveOne` as `inbox` (`mailbox.Archive`, a
  count-only mode that never reads the message), and sweeps
  `mailbox.Boxes` (non-dot `<provider>/<slug>` directories, byte order).
  Stdout and resulting `.archive/` names are byte-identical to bash for a
  sweep and a single address. A given address is resolved through the
  session resolver (go-send model) and echoed as given. Deviations: a
  missing mailbox reports `archived 0` (bash prints an empty count); `-h`,
  `--help`, extra arguments, and unknown providers are usage errors (exit 2);
  ambiguous or empty slugs exit 1.

### Files / Reference (added)

- Files list: `go/cmd/alter-bridge/` line gains `archive`; `mailbox.go` line
  mentions count-only archiving and mailbox listing.
- Reference row: `go/cmd/alter-bridge/archive.go` — `archive`, one resolved
  address or a sweep of every mailbox, counts only.
- Reference row for `go/internal/mailbox/mailbox.go` mentions `Archive` and
  `Boxes`.
