---
human_revised: false
plan: maintenance-bridge-root-guard
status: draft
date: 2026-09-21
summary: Delta draft for absorbing the Go mailbox-root guard into specs/bridge.
---

# Delta draft

## Target

`specs/bridge/index.md` — this is the only area; the plan's `scope: [bridge]`
matches exactly, no other area owns this claim.

## Acceptance criteria coverage

All five criteria in the plan are implementation-level guarantees of a new
package (`go/internal/broker.EnsureRoot`), verified by Go tests +
`go run`/`build` evidence in `handoff-t1.md`. None of them changes the
bash script's documented behavior (out of scope, untouched) — so none maps
to the existing "Message delivery" / "Doorbell" requirement subsections,
which describe the bash CLI's contract. Recording this as a **Decision**
(a durable fact about the codebase, not a behavioral requirement of "the
system" as a whole) is the accurate fit — the bash script does not enforce
the empty/`/` `$HOME` guard, so a requirement phrased as "THE SYSTEM SHALL"
would overclaim.

## Added / Modified

- **Decisions** (append): the Go rewrite has started — `go/` module,
  `internal/broker.EnsureRoot` guards and creates the mailbox root,
  `cmd/alter-bridge/main.go` calls it at startup. Notes the guard is
  stricter than the bash script (bash never validates `$HOME`).
- **Files** (append): `go/` module entries.
- **Reference** (append): `go/internal/broker/root.go` reference row.

## Removed Requirements

None.

## No spec-behavior change

The bash CLI's documented Requirements are unchanged — this plan added a
parallel, still-unwired Go implementation of one narrow piece (root
resolution), not a change to the system's externally observable behavior.
