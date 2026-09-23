---
human_revised: false
scope: [bridge]
status: in-progress
summary: Port the bash `archive` command to Go — archive pending messages without printing them, for one mailbox or a sweep of every mailbox, with per-mailbox counts.
targets: [cli]
aux: []
---

# Go rewrite — `archive`

## Overview

Port `archive` (`skills/alter-bridge/scripts/alter-bridge`, lines ~401-421)
to Go. It is `drain yes <addr> quiet` per mailbox: every pending message is
moved into `.archive/` exactly as `inbox` does, nothing is printed, and only
counts are reported. Without an address it sweeps every `<provider>/<slug>`
directory under the root.

`archive` has no formal requirement in `specs/bridge/index.md` (a spec gap
predating this plan); absorption closes it with a Message delivery
requirement, not only a Decision.

It reuses `maintenance-go-inbox`'s `internal/mailbox` archive-and-rename
logic (`archiveOne`, the reader-stamped name, the `.<8hex>` collision
suffix) through a count-only mode of the same drain loop, and the CLI's
address parsing and slug resolution. Identity follows the go-send model:
an explicit address is resolved through `session.Resolver`; no environment
variable picks a mailbox.

## Acceptance Criteria (EARS / RFC 2119)

Derived from the bash `archive` and `drain ... quiet`:

- AC1: WHEN one `provider:value` address is given THE SYSTEM SHALL resolve it through the same `session.Resolver` as `inbox`, archive every pending message of that mailbox only, and print `archived <n> from <address as given>` — `n` may be 0, including for a missing mailbox — then exit 0.
- AC2: WHEN no address is given THE SYSTEM SHALL sweep every `<root>/<provider>/<slug>` directory, skipping any provider or slug entry whose name starts with `.` (so `.archive/` and `.tmp/` are never swept), print `archived <n> from <provider>:<slug>` for each mailbox with `n > 0` in byte order, then `total: <sum>`, and exit 0.
- AC3: Archiving MUST go through the same `internal/mailbox` loop and `archiveOne` that `inbox` uses — no second implementation of the move, naming, or collision suffix; resulting `.archive/` names are identical to what bash produces for the same input (modulo the random collision suffix).
- AC4: Archived message content MUST NOT be printed; stdout carries only the count lines.
- AC5: WHEN more than one argument, `-h`/`--help`, or a malformed/unknown-provider address is given THE SYSTEM SHALL exit 2 with usage or the parse error on stderr and move nothing; WHEN resolution is ambiguous or yields an empty slug THE SYSTEM SHALL exit 1 and move nothing.
- AC6: Smoke parity: on two copies of the same scratch `ALTER_BRIDGE_ROOT`, Go and bash `archive` (sweep, and one address) produce identical stdout and identical `.archive/` listings and leave the same mailboxes empty.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| T1 | `mailbox.Archive` count-only mode and `mailbox.Boxes` listing | — | pending |
| T2 | `archive` CLI wiring (one address or sweep) | T1 | pending |

## Out of scope

- `purge` (separate plan; deletes archived messages, this only archives pending ones).
- Any other subcommand; replacing the bash script or rewiring hooks/skill.

## Risks

- **Bash missing-mailbox quirk.** Bash `archive <addr>` on a missing mailbox prints `archived  from <addr>` (empty count, because `drain` returns before printing its count). Go prints `0`; recorded as a Decision, excluded from the smoke parity for that case.
- **Sort order.** Byte order vs bash glob collation; identical for bridge-generated slugs.
