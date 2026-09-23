---
human_revised: false
plan: maintenance-go-peek
status: draft
date: 2026-09-23
summary: Delta for maintenance-go-peek — record the Go peek as a decision and reference its source, leaving bash requirements authoritative until cutover.
---

# Delta draft — maintenance-go-peek

The Go `peek` is built and tested but not wired into hooks or skills. Following
the `maintenance-go-send` and `maintenance-go-inbox` precedent, the bash
Requirements stay authoritative and the Go behavior is recorded under Decisions
until cutover.

## specs/bridge/index.md

### Added Requirements

- None. Bash requirements remain authoritative until cutover.

### Modified Requirements

- None.

### Removed Requirements

- None.

### Added Decisions

- 2026-09 (`maintenance-go-peek`): the Go `peek` is built but not yet wired
  into hooks or the skill. It shares `inbox`'s CLI path (`runDrain`: required
  `provider:value` address with no `self_addr` fallback, same session
  resolver, same exit codes) and calls `mailbox.Drain` with archiving off, so
  it prints oldest first in `inbox`'s format, byte-identical to bash `peek`,
  and moves nothing.

### Added Files / Reference rows

- `go/cmd/alter-bridge/peek.go` — Go `peek`: non-archiving drain through the runner shared with `inbox`.
- Update the `go/cmd/alter-bridge/` Files entry to name `peek`.
