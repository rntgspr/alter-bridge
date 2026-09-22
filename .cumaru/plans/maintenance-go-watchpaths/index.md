---
human_revised: false
scope: [bridge]
status: pending
summary: Port the bash `watchpaths` command to Go — SessionStart hook that registers the mailbox with the FileChanged watcher.
targets: [cli]
aux: []
---

# Go rewrite — `watchpaths`

## Overview

Port `watchpaths` to Go: the `SessionStart` hook entry point that resolves
this session's address, ensures its mailbox directory exists, and emits
the `hookSpecificOutput.watchPaths` JSON the harness expects. Depends on
`send`'s addressing and the already-implemented `internal/broker.EnsureRoot`
— see `plans/index.md`'s `go-rewrite-order` tag.

## Acceptance Criteria (EARS / RFC 2119)

Derived from `specs/bridge/index.md` ("Doorbell"):

- WHEN a Claude session starts THE SYSTEM SHALL register a `watchPaths` entry for the whole provider directory (not just this agent's own mailbox), so a later `/rename` or a mailbox that does not exist yet still rings.
- THE SYSTEM SHALL create both the provider directory and this agent's own mailbox subdirectory before emitting the watch registration (the directory must exist at registration time).
- Output MUST be the exact JSON shape the harness expects: `{hookSpecificOutput:{hookEventName:"SessionStart",watchPaths:[...]}}`.

## Plan / DAG

To be broken into tasks when work starts on this plan.

## Out of scope

- `relay` (separate plan; the `FileChanged` event this registration triggers).
- Any other subcommand.

## Risks

- None identified.
