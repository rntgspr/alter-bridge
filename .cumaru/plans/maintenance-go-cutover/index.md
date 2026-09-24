---
human_revised: false
scope: [bridge]
status: in-progress
summary: Cut the live bridge over from the bash script to the Go binary `bin/alter-bridge` — docs, `install-hooks.sh`, and the live Claude/Codex hook wiring — keeping bash as fallback.
targets: [cli, hooks, skill]
aux: []
---

# Go cutover — `bin/alter-bridge` becomes canonical

## Overview

Every subcommand of `skills/alter-bridge/scripts/alter-bridge` now has a Go
port under `go/` (built by `go/build.sh` into `bin/alter-bridge`, gitignored).
Renato authorized the cutover ("pode deixar o go assumir"): the Go binary
takes over from bash, including the live hook wiring outside the repository.
The bash script stays in place as a fallback.

The Go hook entry points take a provider argument (`watchpaths claude`,
`hook claude`, `relay claude`, `hook codex`) and no environment variable
decides identity, so every wiring and every documented invocation changes
shape, not only its path (`send` requires `--from`; `inbox` / `peek` require
an address).

## Acceptance Criteria (EARS / RFC 2119)

- AC1 (build): `go/build.sh` MUST produce `bin/alter-bridge` from the current tree, and `go test -count=1 ./...` in `go/` MUST pass.
- AC2 (docs): `skills/alter-bridge/SKILL.md` and `README.md` MUST name `~/agentic-workspace/papa/alter-bridge/bin/alter-bridge` as the canonical command, with every documented invocation matching the Go CLI's `-h` usage (provider argument on hook entry points, `--from` on `send`, an address on `inbox` / `peek`), and MUST describe the bash script as a fallback that is kept, not deleted.
- AC3 (installer): WHEN `install-hooks.sh` runs THE SYSTEM SHALL write `"$HOME/.../bin/alter-bridge" watchpaths claude`, `... hook claude`, `... relay claude` (matcher `.*__from_.*`) into `~/.claude/settings.json` and `... hook codex` into `~/.codex/hooks.json`, reconciling an existing bash-form entry for the same subcommand in place instead of duplicating it, and leaving a file already in the Go form untouched.
- AC4 (live wiring): the live `~/.claude/settings.json` and `~/.codex/hooks.json` MUST call the Go binary with the provider argument, each backed up to `<file>.bak.<timestamp>` before the edit, with only the alter-bridge command strings changed and both files still valid JSON.
- AC5 (entry points as wired): fed realistic payloads against a scratch `ALTER_BRIDGE_ROOT`, the wired commands MUST behave per spec: `watchpaths claude` prints the `watchPaths` JSON; `hook claude` drains the session's mailbox; a `change` event to `relay claude` prints nothing.
- AC6 (identity continuity): WHEN a currently live Claude or Codex session inside `~/agentic-workspace` is resolved by the Go identity rule THE SYSTEM SHALL key it to the same mailbox the address `who` advertises for it, so no live session's mailbox is orphaned by the cutover.
- AC7 (read-only live checks): `bin/alter-bridge who` and `bin/alter-bridge peek` against a real mailbox MUST run without archiving, moving, or creating anything under `~/.alter-bridge`.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| [T1](t1.md) | Identity continuity gate: Go resolver vs `who` names for live sessions | — | done |
| [T2](t2.md) | Repo cutover: rebuild, `SKILL.md`, `README.md`, `install-hooks.sh` | T1 | done |
| [T3](t3.md) | Live rewiring of `~/.claude/settings.json` and `~/.codex/hooks.json`, with backups and scratch verification | T2 | done |

## Blocker (T1, 2026-09-23) — resolved

**Resolution:** Renato chose option A. `maintenance-go-agent-names`
(`69f04b8`..`3ce2055`) made the Go Claude resolver fall back to the live
`claude agents --json` name; the re-run gate (see `handoff-t1.md`) shows bash
and Go registering the same mailbox for every live Claude and Codex session.
The original finding follows for the record.


AC6 fails on evidence, before any live file was touched:

- `bin/alter-bridge who` advertises this session as
  `claude:alter-bridge-ef  4ff25905-8f72-4c7a-8371-b2a90f60cf7c` (the
  auto-generated name from `claude agents --json`).
- Fed a SessionStart payload for that id against a scratch root, bash
  `watchpaths` registers `claude/alter-bridge-ef`, but Go `watchpaths claude`
  registers `claude/4ff25905-8f72-4c7a-8371-b2a90f60cf7c`. Same for
  `6e4c01e3-...` (`vault-9e` vs the id). Only a `/rename`d session
  (`claudio-probe`, a `custom-title` in its transcript) resolves the same on
  both.
- Cause: the Go resolver (`go/internal/session/store.go`) names a Claude
  session only by its transcript `custom-title`; the auto-name is not in the
  transcript (only `ai-title` entries), while bash asked `claude agents
  --json`. Every existing mailbox under `~/.alter-bridge/claude/`
  (`alter-bridge-*`, `teste-bridge-*`) is an auto-name.
- Consequence after cutover: Go `send claude:alter-bridge-ef` still delivers
  to `claude/alter-bridge-ef` (verified in scratch), which no Go hook drains
  and no Go relay rings for; the address `who` shows would silently stop
  reaching its session. No mailbox currently holds pending messages, so
  nothing is lost today; the break is in addressing.

Options (decision needed from Renato):

- A. New plan first: teach the Go Claude resolver to fall back to the live
  `claude agents --json` name (id -> name when no `custom-title`; name -> id
  for running sessions), restoring bash parity; then resume this plan
  unchanged. Downside: a `claude agents` subprocess per hook call (bash
  already paid it).
- B. Accept id-keyed mailboxes for un-renamed sessions (the go-send Decision
  as written) and cut over now; change `who` to advertise the id for
  sessions without a `custom-title`, and document `/rename` as the way to get
  a readable address. Downside: breaks the addresses humans read from `who`
  today, and the `alter-bridge-*` directories become dead.
- C. Cut over with the mismatch as a known risk. Not recommended: silent
  misdelivery.

## Out of scope

- Deleting the bash script, any live mailbox, or `~/.alter-bridge` content.
- The probe files (`teste-bridge/.claude/settings.local.json`,
  `~/.alter-bridge-probe`, `~/.local/share/ab-probe`), handled by the root
  agent.
- Closing `maintenance-go-relay` (pending its own visual confirmation).
- Plugin manifests and the marketplace `install.sh`.

## Risks

- Codex trust: Codex hashes each hook; editing the command in
  `~/.codex/hooks.json` disables it until approved again in Codex
  (`config.toml` `[hooks.state."<path>:user_prompt_submit:0:0"].trusted_hash`).
- Codex hook payload must carry `session_id` / `thread_id` / `threadId` for
  `hook codex` to drain anything; unverified against a live Codex turn.
- `bin/` is gitignored, so a fresh checkout has no binary until
  `go/build.sh` runs; the installer must refuse to wire a missing binary.
