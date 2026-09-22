---
human_revised: false
scope: [bridge]
status: pending
summary: Port the bash `archive` command to Go — archive pending messages without printing them, for one mailbox or every mailbox.
targets: [cli]
aux: []
---

# Go rewrite — `archive`

## Overview

Port `archive` to Go. Not currently documented as a formal EARS requirement
in `specs/bridge/index.md` (a spec gap predating this plan) — criteria
below are derived from the bash script's current behavior
(`skills/alter-bridge/scripts/alter-bridge` lines ~401-421), and a "no
formal spec yet" fact should be closed alongside implementation, not
deferred again. Depends on `inbox`'s archive-move logic landing first —
see `plans/index.md`'s `go-rewrite-order` tag.

## Acceptance Criteria (derived from current bash behavior)

- WHEN an address is given THE SYSTEM SHALL archive every pending message for that mailbox only, and report the count.
- WHEN no address is given THE SYSTEM SHALL sweep every mailbox under the root (excluding dotfiles/dirs) and report a per-mailbox count plus a total.
- Archiving here MUST reuse the same archive-and-rename logic `inbox` uses — no duplicated implementation.
- Archived messages are NOT printed (that's the difference from `inbox`).

## Plan / DAG

To be broken into tasks when work starts on this plan.

## Out of scope

- `purge` (separate plan; deletes archived messages, this only archives pending ones).
- Any other subcommand.

## Risks

- None identified.
