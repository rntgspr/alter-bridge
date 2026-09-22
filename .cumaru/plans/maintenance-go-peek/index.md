---
human_revised: false
scope: [bridge]
status: pending
summary: Port the bash `peek` command to Go — read pending messages without archiving them.
targets: [cli]
aux: []
---

# Go rewrite — `peek`

## Overview

Port `peek` (the `drain no` case) to Go. Shares its read path with `inbox`
in the bash script (both call `drain`); the Go port should reuse the same
internal read helper `inbox`'s plan introduces, parameterized by whether
archiving happens. Land after `inbox` — see `plans/index.md`'s
`go-rewrite-order` tag.

## Acceptance Criteria (EARS / RFC 2119)

Derived from `specs/bridge/index.md` ("Message delivery"):

- `peek` SHALL read pending messages without archiving them.
- Messages MUST print oldest first, in the same format as `inbox`.

## Plan / DAG

To be broken into tasks when work starts on this plan.

## Out of scope

- Any archiving behavior — that is `inbox`'s job.
- Any other subcommand.

## Risks

- None identified.
