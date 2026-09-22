---
human_revised: false
scope: [bridge]
status: pending
summary: Port the bash `hook` command to Go — the UserPromptSubmit entry point that registers this session and drains its mailbox.
targets: [cli]
aux: []
---

# Go rewrite — `hook`

## Overview

Port `hook` to Go: reads the harness's `UserPromptSubmit` JSON payload from
stdin, resolves this session's address, scopes itself to
`~/agentic-workspace`, and drains (archives) its own mailbox. Depends on
`send`'s addressing and `inbox`'s drain logic — see `plans/index.md`'s
`go-rewrite-order` tag.

## Acceptance Criteria (EARS / RFC 2119)

Derived from `specs/bridge/index.md` ("Addressing") and current bash
behavior (`hook()`, lines ~278-295):

- WHEN invoked as a `UserPromptSubmit` hook THE SYSTEM SHALL read `session_id`/`thread_id` and `cwd` from the JSON payload on stdin.
- WHEN the resolved `cwd` is outside `~/agentic-workspace` THE SYSTEM SHALL return without draining anything.
- WHEN no explicit address is given THE SYSTEM SHALL resolve this session's own address using the same rule as `send`'s `self_addr`.
- THE SYSTEM SHALL then archive and print every pending message for that address (same behavior as `inbox`).

## Plan / DAG

To be broken into tasks when work starts on this plan.

## Out of scope

- `watchpaths` and `relay` (separate plans; different hook events, more complex payloads).
- Any other subcommand.

## Risks

- None identified.
