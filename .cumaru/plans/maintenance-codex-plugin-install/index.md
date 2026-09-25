---
human_revised: false
scope: [bridge]
status: in-progress
summary: Install and activate alter-bridge as a Codex plugin on this machine, replace legacy Codex wiring, and verify one working hook without changing the Claude installation.
targets: [plugin, hooks]
aux: []
---

# Codex plugin installation

## Overview

Install the existing alter-bridge plugin through the Codex marketplace on this
machine. The checkout already contains `.codex-plugin/plugin.json`,
`hooks/codex.json`, and `install.sh`; the work is the local installation,
activation, cutover from manual hooks, and evidence that message delivery still
works. The Claude installation is owned by a separate thread.

`maintenance-plugin-standard` T3 currently includes both runtimes and is
blocked. This plan owns its Codex portion; reconcile T3 before closing either
plan so the Codex cutover is performed and recorded once. The
`maintenance-plugin-release` plan owns the Claude release path and does not
change the Codex install path.

## Acceptance Criteria (EARS / RFC 2119)

- AC1: WHEN the Codex marketplace and plugin install commands run from this
  checkout THE SYSTEM SHALL list `alter-bridge@alter-bridge` as installed and
  enabled with the `alter-bridge` skill.
- AC2: WHEN the installed Codex plugin is inspected THE SYSTEM SHALL expose
  exactly one alter-bridge `userPromptSubmit` hook invoking `hook codex` through
  the plugin root.
- AC3: Before changing Codex global configuration or plugin registrations, the
  procedure MUST save timestamped backups of every affected existing file.
- AC4: WHEN the Codex cutover finishes THE SYSTEM SHALL have one effective
  alter-bridge prompt hook and no active legacy `agent-bridge` plugin or manual
  alter-bridge hook.
- AC5: WHEN the trusted hook receives a pending message under a scratch
  `ALTER_BRIDGE_ROOT` THE SYSTEM SHALL drain and archive it once; `who` SHALL
  run read-only after installation.
- AC6: The Codex installation MUST NOT modify Claude settings, registrations,
  or release artifacts.

## Plan / DAG

| Task | Title | Status | Depends on |
|------|-------|--------|-----------|
| [T1](t1.md) | Inventory and prepare Codex installation | pending | — |
| [T2](t2.md) | Activate, cut over, and verify Codex | pending | T1 |

## Out of scope

- Claude Code installation, release publication, and its manual hook cleanup.
- Changes to the bridge's message format or Go command behavior.
- Absorption of `maintenance-plugin-standard` or `maintenance-plugin-release`.

## Risks

- Codex plugin hooks require trust in `/hooks`; until the hook is trusted,
  removing the manual hook would interrupt prompt delivery. Keep the manual
  entry active until trust and effective hook execution are verified.
- Codex caches a local plugin copy. Reinstall after rebuilding the binary or
  changing plugin files so validation uses the intended version.
- The old joint cutover task overlaps this plan. Record the Codex result there
  before either plan closes, without taking over the Claude work.
