---
human_revised: false
plan: maintenance-go-inbox
status: draft
date: 2026-09-23
summary: Delta for maintenance-go-inbox — record the Go inbox and its shared drain helper as a decision and reference their sources, leaving bash requirements authoritative until cutover.
---

# Delta draft — maintenance-go-inbox

The Go `inbox` is built and tested but not wired into hooks or skills. Following
the `maintenance-go-send` precedent, the bash Requirements stay authoritative and
the Go behavior is recorded under Decisions until cutover.

## specs/bridge/index.md

### Added Requirements

- None. Bash requirements remain authoritative until cutover.

### Modified Requirements

- None.

### Removed Requirements

- None.

### Added Decisions

- 2026-09 (`maintenance-go-inbox`): the Go `inbox` requires its
  `provider:value` address (no `self_addr` fallback, per the go-send identity
  model) and resolves it through the same session resolver as `send`, refusing
  ambiguous or empty slugs. Its output and `.archive/` naming (split at the
  last `__`, `__to_<provider>_<slug>__`, `.<8 hex>` suffix on collision) are
  byte-identical to bash. The read path is `internal/mailbox.Drain`,
  parameterized by whether it archives, for `peek`, `archive`, and `hook` to
  reuse. Unlike bash, a directory named `*.md` in a mailbox is skipped rather
  than archived.

### Added Files / Reference rows

- `go/cmd/alter-bridge/inbox.go` — Go `inbox`: required address, resolution, archiving drain.
- `go/internal/mailbox/mailbox.go` — Go mailbox drain: oldest-first read, optional bash-parity archive move.
