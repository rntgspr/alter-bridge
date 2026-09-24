---
human_revised: false
scope: [bridge]
status: in-progress
summary: Teach the Go Claude session resolver the live `claude agents --json` name, so unrenamed sessions resolve to the same mailbox bash and `who` use — unblocking `maintenance-go-cutover` (option A).
targets: [cli]
aux: []
---

# Go Claude resolver — live agent names

## Overview

`maintenance-go-cutover` is blocked (its `## Blocker`, option A chosen by
Renato): the Go Claude store in `go/internal/session/store.go` names a
session only by its transcript `custom-title`. A session never `/rename`d
has an automatic name (`alter-bridge-ef`, `vault-9e`) that only
`claude agents --json` reports, and bash (`claude_session_name`) and `who`
use that name. Go therefore keys such a session by its id, orphaning the
mailbox `who` advertises.

Fix: the Claude store merges the live agents from `session.ClaudeAgents`
into the transcript sessions. A transcript without a `custom-title` takes
the live name for its `sessionId`; a live agent with no transcript yet is
listed too. Both id -> name and name -> id then flow through the unchanged
`Resolver.Slug` rules (ambiguity included). A missing or failing `claude`
binary contributes nothing, leaving today's behavior.

Precedence: transcript `custom-title` first, live name as the fallback. After
a `/rename` both carry the same title, so this matches bash for every live
session; it differs only for a titled session that is no longer running,
which bash cannot resolve at all.

## Acceptance Criteria (EARS / RFC 2119)

- AC1 (id -> name): WHEN a Claude session id has no `custom-title` in its transcript and `claude agents --json` lists it with a name THE SYSTEM SHALL resolve the id to that name's slug.
- AC2 (no transcript yet): WHEN a live Claude agent has no transcript file THE SYSTEM SHALL still resolve its id to its live name's slug.
- AC3 (precedence): a transcript `custom-title` MUST win over the live name for the same session id.
- AC4 (name -> id): WHEN `claude:<name>` is resolved and two or more sessions carry that name (transcript title or live name) THE SYSTEM SHALL refuse it with an `AmbiguousError` listing their ids; a single match resolves to the name slug.
- AC5 (degradation): WHEN the `claude` binary is missing, exits non-zero, or prints invalid JSON THE SYSTEM SHALL resolve exactly as before (transcript titles only) with no error.
- AC6 (codex unchanged): Codex and OpenCode resolution MUST NOT change; existing tests stay green.
- AC7 (parity): WHEN the same SessionStart payload for the session id of a live unrenamed Claude session is fed to bash `watchpaths` and to Go `watchpaths claude` (each into its own scratch `ALTER_BRIDGE_ROOT`) THE SYSTEM SHALL register the same mailbox directory name, and Go `hook claude` with that payload SHALL drain that mailbox in scratch.
- AC8 (quality): `go test -count=1 ./...`, `go vet ./...` pass and `gofmt -l .` is empty in `go/`; `go/build.sh` rebuilds `bin/alter-bridge`.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| [T1](t1.md) | Merge live `claude agents` names into the Claude store (TDD) | — | pending |
| [T2](t2.md) | Rebuild and prove bash-vs-Go parity for a live unrenamed session in scratch | T1 | pending |

## Out of scope

- Resuming, editing, or closing `maintenance-go-cutover` and `maintenance-go-relay`.
- Hook wiring, docs, `who` output, Codex/OpenCode stores.
- Any write to the real `~/.alter-bridge`, messaging real sessions, or editing `~/.claude/settings.json` / `~/.codex/hooks.json`.

## Risks

- One `claude agents --json` subprocess per store read (a `send` reads it
  twice, for sender and recipient); bash paid the same per call.
- A stale titled transcript and a live agent sharing a name are now
  ambiguous by name, the same rule as two titled sessions today.
