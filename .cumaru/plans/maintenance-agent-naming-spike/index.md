---
human_revised: false
scope: [bridge]
status: in-progress
summary: Investigate agent names and mailbox identity across providers, then decide a stable addressing contract before changing delivery behavior.
targets: [cli, plugin]
aux: []
---

# Agent naming and mailbox identity spike

## Overview

The Go `relay` command is implemented, and the user reports that relay delivery
works in Claude today. The open design question is the final naming contract:
what name a user or agent should address, how it maps to a live session and a
mailbox, and what happens after rename, resume, name reuse, or collision.

The current resolver accepts `provider:value` with a session name or id and
keys mailboxes by the slugified name. The bridge spec records that Claude's
automatic names may change on resume. This spike evaluates that behavior
across Claude, Codex, and OpenCode and records a decision before any change to
delivery semantics. The live verification criterion in
`maintenance-go-relay` remains open in that plan; the user's observation is
context, not recorded acceptance evidence for its AC8.

## Acceptance Criteria (EARS / RFC 2119)

- The spike MUST document the current name-to-session-to-mailbox path for
  `send`, `hook`, `watchpaths`, `relay`, and `who`, identifying provider-specific
  sources and observable behavior.
- The spike MUST examine rename, resume, duplicate names, archived or reused
  names, unavailable session stores, and sending by id, using a reproducible
  check or an explicitly labeled unverified case for each.
- The spike MUST compare at least two concrete naming contracts against
  address stability, delivery correctness, user ergonomics, and migration of
  existing mailboxes, then record one recommendation and its trade-offs.
- The spike MUST identify the exact spec claims, code paths, and follow-up
  acceptance criteria needed to implement the recommendation, without
  changing runtime behavior in this plan.

## Plan / DAG

| Task | Title | Status | Depends on |
|------|-------|--------|-----------|
| [T1](t1.md) | Map current identity behavior and recommend a naming contract | pending | — |

## Out of scope

- Implementing a new naming system or migrating existing mailboxes.
- Closing or weakening `maintenance-go-relay` AC8.
- Changing Claude or Codex plugin installation.

## Risks

- A name-based mailbox may change after resume or be inherited by another
  session; the spike must distinguish observed behavior from inferred risk.
- A more stable identity may require a mailbox migration. Do not assume that
  existing messages can be moved safely without examining the current layout.
