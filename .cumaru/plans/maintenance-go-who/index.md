---
human_revised: false
scope: [bridge]
status: pending
summary: Port the bash `who` command to Go — list addressable agents read live from each CLI's own session state.
targets: [cli]
aux: []
---

# Go rewrite — `who`

## Overview

Port `who` to Go: shells out to `claude agents --json` and reads the Codex
`state_5.sqlite` threads table (or `session_index.jsonl` fallback) to list
addressable agents. Mostly independent of the other commands' shared
addressing/drain code — see `plans/index.md`'s `go-rewrite-order` tag for
why it's sequenced where it is.

## Acceptance Criteria (EARS / RFC 2119)

Derived from `specs/bridge/index.md` ("Decisions"):

- `who` MUST read addressable agents live from each CLI's own state — no local registry, no caching, so nothing here can drift from the CLIs' own truth.
- Claude agents are listed via `claude agents --json`, filtered to `kind == "interactive"`, formatted as `claude:<slugified-name>`.
- Codex threads are listed via the `state_5.sqlite` `threads` table (or the JSONL fallback when the database is unreadable), formatted as `codex:<slugified-name>`.
- A missing `claude` or `sqlite3` binary, or an unreadable Codex state file, MUST degrade to omitting that provider's rows, not erroring out.

## Plan / DAG

To be broken into tasks when work starts on this plan.

## Out of scope

- Any other subcommand.

## Risks

- Shelling out to `claude`/`sqlite3` from Go needs the same PATH robustness the bash script already handles (`claude_bin()` falls back to `$HOME/.local/bin/claude`) — carry that fallback over.
