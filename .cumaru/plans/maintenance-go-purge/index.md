---
human_revised: false
scope: [bridge]
status: pending
summary: Port the bash `purge` command to Go — permanently delete the archived trail.
targets: [cli]
aux: []
---

# Go rewrite — `purge`

## Overview

Port `purge` to Go. Not currently documented as a formal EARS requirement
in `specs/bridge/index.md` (same spec gap as `archive`) — criteria below
are derived from the bash script's current behavior
(`skills/alter-bridge/scripts/alter-bridge` lines ~425-434). Depends on
`archive` landing first, since both touch `.archive/` — see
`plans/index.md`'s `go-rewrite-order` tag.

## Acceptance Criteria (derived from current bash behavior)

- WHEN `.archive/` does not exist THE SYSTEM SHALL report nothing to purge and exit 0, without error.
- WHEN `.archive/` exists THE SYSTEM SHALL delete only `*.md` files directly under it (not recursively, not other file types), and report the count deleted.
- This operation is NOT reversible — no confirmation prompt exists in the bash version either; keep behavior parity unless the user asks for a safety change.

## Plan / DAG

To be broken into tasks when work starts on this plan.

## Out of scope

- Any other subcommand.

## Risks

- Irreversible delete — same risk profile as the bash script, not introduced by this port.
