---
human_revised: false
scope: [bridge]
status: in-progress
summary: Add a Go mailbox-root guard (internal/broker) that ensures $ALTER_BRIDGE_ROOT exists, failing loudly on permission errors or a dangerous $HOME. First real logic in the Go rewrite, beyond the hello-world scaffold.
targets: [cli]
aux: []
---

# Bridge mailbox root guard (Go)

## Overview

The bash `alter-bridge` script computes `ROOT="${ALTER_BRIDGE_ROOT:-$HOME/.alter-bridge}"`
and creates parts of it lazily, scattered across `send` and `watchpaths`
(`mkdir -p` calls). Read-only commands (`inbox`, `peek`, `who`) never
create it, so a fresh machine's first read-only call silently no-ops.

This plan does NOT touch the bash script — it targets the Go rewrite
(`maintenance-go-bootstrap` scaffolded `cmd/alter-bridge` as a hello-world
binary; this is the first real behavior added on top). Add an
`internal/broker` package with a function that resolves and ensures the
mailbox root exists, so every future Go subcommand can call it once instead
of repeating `mkdir -p` at each call site. It must fail loudly instead of
proceeding into undefined behavior when:

- creating the root directory fails (e.g. permission denied).
- `$HOME` is unset, empty, or `/` and no explicit override was given —
  creating a `.alter-bridge` directory under `/` is almost certainly a
  misconfigured environment, not intent.

Listing mailbox contents (a `list`/`who`-style enumeration of the root) is
explicitly deferred to a future plan — out of scope here. `cmd/alter-bridge/main.go`
now calls `broker.EnsureRoot` once at startup (using `$ALTER_BRIDGE_ROOT` and
`$HOME` from the environment) and prints the resolved root, so the current
hello-world binary already proves the guard end-to-end even before any real
subcommand dispatch exists.

## Acceptance Criteria (EARS / RFC 2119)

- WHEN the mailbox root does not exist THE SYSTEM SHALL create it,
  returning its resolved path.
- WHEN `$HOME` is empty or equal to `/` and no explicit root override is
  given THE SYSTEM SHALL return an error without creating any directory.
- WHEN creating the root directory fails (e.g. permission denied) THE
  SYSTEM SHALL return an error naming the path, not a panic.
- WHEN the mailbox root already exists THE SYSTEM SHALL return its path
  without error and without re-creating it.
- The behavior MUST be covered by Go tests (TDD: a failing test observed
  before the implementation), not manual verification only, since this is
  a testable pure-Go unit unlike the bash script.

## Plan / DAG

| Task | Title | Status | Depends on |
|------|-------|--------|-----------|
| [T1](t1.md) | internal/broker.EnsureRoot with TDD tests | done | — |

## Out of scope

- Any changes to the bash script `skills/alter-bridge/scripts/alter-bridge`.
- A `list`/mailbox-enumeration subcommand — deferred to a future plan.
- A real subcommand dispatch in the Go binary — `main.go` still just calls
  `EnsureRoot` once and prints "hello world", it does not yet parse args.

## Risks

- None identified — pure Go, no external deps beyond the standard library.
