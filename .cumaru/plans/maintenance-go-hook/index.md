---
human_revised: false
scope: [bridge]
status: in-progress
summary: Port the bash `hook` command to Go — the UserPromptSubmit entry point that reads its session from the stdin payload, stays inside `~/agentic-workspace`, and drains that session's mailbox.
targets: [cli]
aux: []
---

# Go rewrite — `hook`

## Overview

Port `hook` (`skills/alter-bridge/scripts/alter-bridge`, lines ~278-295) to
Go. Bash reads the harness payload on stdin, takes the session id from
`.session_id // .thread_id // .threadId` and `cwd` from `.cwd`, `cd`s into
`cwd` when it can, returns 0 unless the working directory is
`$HOME/agentic-workspace` or below, then drains (archives and prints) the
mailbox of the optional explicit address or of `self_addr "$sid"`. It creates
nothing when the mailbox is missing; it only makes `.archive/` when it moves
a message.

### Identity reconciliation (`self_addr` vs the go-send model)

The scaffolded criterion said "the same rule as `send`'s `self_addr`". The Go
`send` has no `self_addr`: per the `maintenance-go-send` Decision in
`specs/bridge/index.md`, identity is option-only, no environment variable
(`AGENT_SLUG`, `CODEX_THREAD_ID`, `CLAUDE_CODE_SESSION_ID`, ...) decides it,
and the working-directory fallback and `default` slug are gone. That decision
already says it replaces the bash Addressing rules at cutover, so the Go
`hook` follows it rather than reintroducing `self_addr`:

- The session value comes from the payload (the one identity source a hook
  has), resolved through the same `session.Resolver.Slug` as `send`/`inbox`:
  a session id maps to its slugified name, or its id when unnamed.
- Bash picks the provider from `CODEX_*` env. With env excluded, the provider
  is the argument the hook wiring passes: `hook claude` / `hook codex`.
  An explicit `provider:value` keeps bash's explicit-address form.
- With a bare provider and no session id in the payload there is no identity;
  the hook drains nothing and exits 0 instead of guessing a cwd-derived slug.

## Acceptance Criteria (EARS / RFC 2119)

- AC1: WHEN invoked as `hook <provider>` THE SYSTEM SHALL read the JSON payload on stdin, take the session id as the first non-empty string of `session_id`, `thread_id`, `threadId`, and drain the mailbox `session.Resolver.Slug(<provider>, <id>)` resolves to, oldest first, archiving and printing exactly as `inbox` does (byte-identical to bash for the same mailbox).
- AC2: WHEN invoked as `hook <provider>:<value>` THE SYSTEM SHALL drain that address, resolved through the same resolver, ignoring the payload's session id (bash explicit-address parity).
- AC3: WHEN the payload's `cwd` names an existing directory THE SYSTEM SHALL judge the workspace scope by it (relative to the process working directory when relative), otherwise by the process working directory; WHEN that directory is neither `$HOME/agentic-workspace` nor below it THE SYSTEM SHALL print nothing, touch nothing, and exit 0.
- AC4: WHEN the payload is empty, not JSON, or carries no session id and the argument is a bare provider THE SYSTEM SHALL print nothing, touch nothing, and exit 0, so a malformed payload never blocks the user's prompt.
- AC5: `hook` MUST NOT create the mailbox root or the mailbox; `.archive/` appears only when a message is moved (bash parity). The root uses the shared `ALTER_BRIDGE_ROOT` / `HOME` rules and guard (guard failure exits 1).
- AC6: WHEN the argument is missing, extra, `-h`/`--help`, or not a known provider / valid `provider:value` THE SYSTEM SHALL print usage or the address error on stderr and exit 2 without reading the mailbox; WHEN resolution or the drain fails THE SYSTEM SHALL report it on stderr and exit 1.
- AC7: Smoke parity: with `HOME` an empty scratch dir holding a fake `agentic-workspace/<agent>` cwd and two copies of the same `ALTER_BRIDGE_ROOT` at the same path, bash `hook claude:<slug>` and Go `hook claude:<slug>` fed the same payload produce identical stdout and `find . | sort`; with `cwd` outside the workspace, both print nothing and change nothing.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| T1 | `hook` subcommand: payload, workspace gate, drain | — | pending |

## Out of scope

- `watchpaths` and `relay` (separate plans; different hook events, more complex payloads).
- Any other subcommand; rewiring the live hooks (`install-hooks.sh`, `settings.json`, `hooks.json`), the plugin manifests, or `bin/`. Cutover must pass the provider argument.

## Risks

- **Identity deviation.** Bash's `AGENT_SLUG` override, `CODEX_*` provider detection, and cwd-segment fallback are dropped (go-send model). An unnamed session's Go mailbox is keyed by its id, where bash used the cwd segment; the cutover wiring and senders must address it accordingly.
- **Name source.** Go resolves Claude names from transcript `custom-title` entries (the `send` store), bash from `claude agents --json`; both reflect `/rename`, but a session with no transcript yet resolves to its id.
- **Exit 2 on UserPromptSubmit blocks the prompt.** Only a miswired argument reaches it (bash also exits 2 on a bad explicit address); payload problems always exit 0.
