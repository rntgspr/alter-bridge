---
human_revised: false
scope: [bridge]
status: pending
summary: Port the bash `send` command to Go — addressing, atomic message write, and the live-session nudge.
targets: [cli]
aux: []
---

# Go rewrite — `send`

## Overview

Port `send` from `skills/alter-bridge/scripts/alter-bridge` (lines ~160-225)
to the Go binary. This is the foundation command: it establishes address
parsing (`provider:slug`), the message file format, and `self_addr`
resolution that every other command reuses for reading its own mailbox.
Land this first — see `plans/index.md`'s `go-rewrite-order` tag.

## Acceptance Criteria (EARS / RFC 2119)

Derived from `specs/bridge/index.md` ("Addressing", "Message delivery"):

- An address MUST take the form `provider:slug`; a leading `#` MUST be accepted and stripped.
- WHEN this session's own address is resolved THE SYSTEM SHALL use, in order: an explicit `AGENT_SLUG` override, the session name, then the path segment after `agentic-workspace/`.
- Session names MUST be slugified for use as a mailbox directory name.
- WHEN `send` is invoked THE SYSTEM SHALL write the message body to a temp file and rename it into the recipient's mailbox directory, so a reader never observes a partial write.
- Every message file MUST carry frontmatter: `from`, `to`, `ts`, `msgid`, `type` (`message | question | result | ack`); `thread` and `in_reply_to` optional.
- The mailbox root MUST be resolved via the existing `internal/broker.EnsureRoot` (already implemented) — do not duplicate that logic.

## Plan / DAG

To be broken into tasks when work starts on this plan.

## Out of scope

- Reading the Claude/Codex session name live (`claude agents --json`, Codex sqlite) — may land as a follow-up inside this same plan or a shared helper; decide at task-breakdown time.
- The live-session nudge (`codex queue`) — can be a separate task within this plan; not required for `send` to be correct, only for immediacy.
- Any other subcommand.

## Risks

- None identified.
