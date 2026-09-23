---
human_revised: false
scope: [bridge]
status: in-progress
summary: Port the bash `inbox` command to Go — drain a mailbox oldest first, archiving each message under a reader-stamped, collision-safe name.
targets: [cli]
aux: []
---

# Go rewrite — `inbox`

## Overview

Port `inbox` (the `drain yes` case in `skills/alter-bridge/scripts/alter-bridge`,
lines ~229-273) to Go. It reuses `send`'s address parsing and slug resolution
(`internal/address`, `internal/session`) — see `plans/index.md`'s
`go-rewrite-order` tag.

The read path lives in one internal helper parameterized by whether it
archives, so `maintenance-go-peek` (`drain no`) and `maintenance-go-archive`
(`drain yes quiet`) reuse it instead of duplicating it.

Identity follows the `maintenance-go-send` model: the mailbox address is an
explicit argument, never derived from the environment. Bash `inbox` with no
argument falls back to `self_addr`; the Go `inbox` requires the address.

## Acceptance Criteria (EARS / RFC 2119)

Derived from `specs/bridge/index.md` ("Message delivery") and the bash `drain`:

### Reading

- `inbox` MUST require one `provider:value` address argument; WHEN it is missing or malformed THE SYSTEM SHALL exit 2 with a message on stderr, and it MUST NOT read any environment variable to pick the mailbox.
- The address value MUST resolve to a mailbox slug through the same `session.Resolver` `send` uses (a session id yields that session's name slug); WHEN resolution is ambiguous or empty THE SYSTEM SHALL exit 1 without touching any file.
- Messages MUST print oldest first (byte order of the file names, whose timestamp prefix sorts chronologically).
- Only `*.md` entries directly in `<root>/<provider>/<slug>/` are messages; dotfiles and directories MUST be skipped.
- Each message MUST print as `===== ALTER-BRIDGE MESSAGE: <original file name> =====`, a newline, the file content verbatim, and a trailing newline — byte-identical to bash `inbox` output for the same mailbox.
- WHEN the mailbox directory does not exist THE SYSTEM SHALL print nothing and exit 0.
- A message of `type: question` MUST keep its type in the printed frontmatter (informational only — no reply tracking is required yet).

### Archiving

- WHEN a message is read via `inbox` THE SYSTEM SHALL move it into `<root>/.archive/` before printing it, so a failed turn never reprocesses it.
- The archived name MUST be `<name up to its last "__">__to_<provider>_<slug>__<text after its last "__">`, exactly as bash builds it.
- WHEN the archived name already exists THE SYSTEM SHALL append `.<8 hex>` before `.md` until it is free, so archiving never overwrites a message.
- The read helper MUST take archiving as a parameter, so `peek` can reuse it with archiving off.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| T1 | Mailbox drain helper | — | done |
| T2 | `inbox` CLI wiring | T1 | pending |

## Out of scope

- `peek`, `archive`, `hook` (separate plans; they reuse T1's helper).
- The bash `quiet` count mode — it serves `archive`, which adds it when it lands.
- Replacing the bash script or rewiring hooks to the Go binary.

## Risks

- **Hand-dropped files.** A name without `__` archives as `<name>__to_<p>_<s>__<name>` (bash parity, odd but lossless). Kept as is.
- **Sort order.** Bash globs in the locale's collation; Go sorts bytes. Identical for bridge-generated names (digits-first timestamps); may differ only for hand-dropped names with mixed case.
