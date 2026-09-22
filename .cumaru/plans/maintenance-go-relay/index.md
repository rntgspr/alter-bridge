---
human_revised: false
scope: [bridge]
status: pending
summary: Port the bash `relay` command to Go — the FileChanged entry point that wakes a live session and rings the doorbell.
targets: [cli]
aux: []
---

# Go rewrite — `relay`

## Overview

Port `relay` to Go: the most complex command in the bridge. Reacts to a
harness `FileChanged` event, decides whether the arrival is this session's
own, and either nudges Codex directly (`codex queue`) or spawns a
throwaway headless Claude session (via the `SendMessage` tool) to ring the
target. Land last — it depends on `watchpaths` (the registration this
reacts to) and on `send`'s addressing — see `plans/index.md`'s
`go-rewrite-order` tag.

## Acceptance Criteria (EARS / RFC 2119)

Derived from `specs/bridge/index.md` ("Doorbell"):

- WHEN a filesystem `add` event lands inside a session's own mailbox THE SYSTEM SHALL spawn a throwaway headless relay session that wakes the target session and deletes its own transcript on exit; `change` events and files outside the bridge SHALL be ignored.
- The relay notice MUST reach the human via a `systemMessage` (the JSON output shape the `FileChanged` hook expects).
- WHEN sending to a live Codex session THE SYSTEM SHALL nudge it directly via `codex queue`, and SHALL skip the nudge silently when no session by that name is running.
- Every `relay` invocation MUST be logged to `$ALTER_BRIDGE_ROOT/.tmp/relay.log` so a silently dead `FileChanged` watcher is diagnosable from the log's last entry.

## Plan / DAG

To be broken into tasks when work starts on this plan.

## Out of scope

- Any other subcommand.

## Risks

- Highest-risk port in the set: spawns a subprocess (`claude -p --session-id ...`), does filesystem cleanup of another session's transcript by exact path, and has a documented flaky failure mode in the bash version (the `FileChanged` watcher dying silently after hours, per the comment at `alter-bridge:336-340`). Needs careful manual verification against a live session before considering this port done — automated tests alone won't catch the watcher-death class of bug.
