---
human_revised: false
scope: [bridge]
status: pending
summary: Port the bash `inbox` command to Go — drain and archive pending messages.
targets: [cli]
aux: []
---

# Go rewrite — `inbox`

## Overview

Port `inbox` (the `drain yes` case in `skills/alter-bridge/scripts/alter-bridge`)
to Go. Depends on `send`'s address-parsing and message-format work landing
first — see `plans/index.md`'s `go-rewrite-order` tag.

## Acceptance Criteria (EARS / RFC 2119)

Derived from `specs/bridge/index.md` ("Message delivery"):

- WHEN a message is read via `inbox` THE SYSTEM SHALL archive it by moving it to `.archive/`, renamed to record who read it, disambiguating colliding names so archiving never loses a message.
- Messages MUST print oldest first.
- A message of `type: question` MUST be treated as open until a reply is sent back to it (informational only at this stage — no reply-tracking behavior is required yet, just that the type is preserved and printed).

## Plan / DAG

To be broken into tasks when work starts on this plan.

## Out of scope

- `peek` (separate plan, though it shares the same underlying drain/read logic — reuse a common internal helper rather than duplicating).
- Any other subcommand.

## Risks

- None identified.
