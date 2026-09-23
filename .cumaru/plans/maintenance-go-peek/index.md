---
human_revised: false
scope: [bridge]
status: in-progress
summary: Port the bash `peek` command to Go — print a mailbox oldest first in inbox's exact format, leaving every message in place.
targets: [cli]
aux: []
---

# Go rewrite — `peek`

## Overview

Port `peek` (the `drain no` case in `skills/alter-bridge/scripts/alter-bridge`,
line ~273) to Go. It reuses everything `maintenance-go-inbox` landed:
`internal/mailbox.Drain` with archiving off, and the `inbox` CLI path
(`inboxEnv`, argument checks, address parsing, slug resolution) — extracted
into one shared runner parameterized by archiving instead of copied.

Identity follows the `maintenance-go-send` model: the mailbox address is an
explicit argument, never derived from the environment. Bash `peek` with no
argument falls back to `self_addr`; the Go `peek` requires the address.

## Acceptance Criteria (EARS / RFC 2119)

Derived from `specs/bridge/index.md` ("Message delivery") and the bash `drain`:

- `peek` MUST require one `provider:value` address argument; WHEN it is missing, extra, `-h`, or malformed THE SYSTEM SHALL exit 2 with a message on stderr, and it MUST NOT read any environment variable to pick the mailbox.
- The address value MUST resolve to a mailbox slug through the same `session.Resolver` `inbox` uses; WHEN resolution is ambiguous or empty THE SYSTEM SHALL exit 1 without printing any message.
- WHEN a mailbox is read via `peek` THE SYSTEM SHALL leave every message in place: no file is moved, renamed, or created, and `<root>/.archive/` is not created.
- Messages MUST print oldest first, in the same format as `inbox` (`===== ALTER-BRIDGE MESSAGE: <file name> =====`, content verbatim, trailing newline) — byte-identical to bash `peek` output for the same mailbox.
- WHEN the mailbox directory does not exist THE SYSTEM SHALL print nothing and exit 0.
- `peek` and `inbox` MUST share one argument/resolution path in the CLI; they differ only in the archive flag and usage text.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| T1 | `peek` CLI wiring on a shared drain runner | — | pending |

## Out of scope

- Any archiving behavior — that is `inbox`'s job.
- The bash `quiet` count mode (`archive`'s plan).
- Any other subcommand; replacing the bash script or rewiring hooks.

## Risks

- **Sort order.** Inherited from `Drain`: byte order vs bash locale collation; identical for bridge-generated names.
