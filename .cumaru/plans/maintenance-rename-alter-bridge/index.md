---
human_revised: false
scope:
  - bridge
status: in-progress
summary: Full technical rename of the agent-bridge plugin, CLI, env var, and default mailbox root to alter-bridge.
targets: [cli, hooks, plugin, skill]
aux: []
---

# Rename agent-bridge to alter-bridge

## Overview

Full technical rename of the plugin from `agent-bridge` to `alter-bridge`
(after the band). Covers plugin manifests, the skill directory and its
script/binary, `AGENT_BRIDGE_ROOT` -> `ALTER_BRIDGE_ROOT`, the default
mailbox root `~/.agent-bridge` -> `~/.alter-bridge` with migration of the
live mailbox, `install-hooks.sh`, and reinstalling this machine's hooks. The
GitHub remote was already renamed to `rntgspr/alter-bridge`; `origin` on
this checkout was repointed ahead of this plan, and `repository` fields in
the manifests are updated to match. The local checkout directory
(`agentic-workspace/papa/agent-bridge/` -> `.../alter-bridge/`) is renamed
by Renato after this plan's tasks land, not by this plan.

## Acceptance Criteria (EARS / RFC 2119)

- The plugin MUST be named `alter-bridge` in both `plugin.json` and
  `.claude-plugin/plugin.json`, with their `description` and `repository`
  fields updated to `alter-bridge`.
- The skill directory MUST be `skills/alter-bridge/` and its CLI script
  MUST be named `alter-bridge`.
- Every internal reference to the `agent-bridge` command, path, or skill
  name (README.md, SKILL.md, install-hooks.sh, the CLI script itself)
  MUST be updated to `alter-bridge`.
- The environment override MUST be renamed `ALTER_BRIDGE_ROOT`, and the
  default mailbox root MUST be `~/.alter-bridge`.
- WHEN this plan is executed on this machine THE SYSTEM SHALL migrate the
  live mailbox contents from `~/.agent-bridge` to `~/.alter-bridge` without
  losing any pending or archived message.
- WHEN this plan is executed on this machine THE SYSTEM SHALL reinstall the
  Claude and Codex hooks so `~/.claude/settings.json` and
  `~/.codex/hooks.json` point at the new `alter-bridge` script path.
- The `specs/bridge` area MUST be updated (via the delta/absorb flow) to
  document the new naming, paths, and env var.

## Plan / DAG

| Task | Title | Status | Depends on |
|------|-------|--------|-----------|
| [T1](t1.md) | Rename plugin manifests | done | — |
| [T2](t2.md) | Rename skill directory, CLI script, and SKILL.md | done | — |
| [T3](t3.md) | Rename env var and default mailbox root inside the CLI script and install-hooks.sh | done | T2 |
| [T4](t4.md) | Update README.md references | done | T2, T3 |
| [T5](t5.md) | Migrate live mailbox and reinstall this machine's hooks | done | T3, T4 |

## Out of scope

- Renaming the local checkout directory (`agent-bridge/` -> `alter-bridge/`) — Renato does this by hand once every task lands, since it's this session's own working directory.
- Reinstalling hooks on any other machine or agent that already wires this bridge — that is each owner's manual follow-up, not something this plan can reach.
- Updating the marketplace-enablement key (`"agent-bridge@agents-marketplace": true`) in `~/.claude/settings.json`'s plugin list — that key lives outside this repo's plan surface and is covered by T5's manual note, not automated.
- Any behavior change to the bridge's messaging semantics (addressing, delivery, doorbell) — pure rename, no functional delta.

## Risks

- **Live hooks currently point at the old path** (`~/.claude/settings.json`, `~/.codex/hooks.json` on this machine) — until T5 reinstalls them, delivery breaks for this session's own mailbox. Mitigate by running T5 last and verifying `agent-bridge watchpaths`/`hook` still resolve before declaring done.
- **Mailbox migration is destructive if mishandled** — `~/.agent-bridge` currently holds live `claude/`, `codex/`, and `.archive/` content. T5 must move, not copy-then-delete-blind, and must be verified before the old directory is removed.
- **Other machines/agents silently break** — any other agentic-workspace checkout with hooks already installed against `agent-bridge` will keep working against the old binary name until someone reinstalls there; this plan cannot reach those machines.
