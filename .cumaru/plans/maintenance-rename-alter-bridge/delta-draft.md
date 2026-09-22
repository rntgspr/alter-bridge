---
human_revised: false
plan: maintenance-rename-alter-bridge
status: draft
date: 2026-09-21
summary: Delta draft for maintenance-rename-alter-bridge — rename specs/bridge naming, paths, and env var to alter-bridge.
---

# Delta draft — maintenance-rename-alter-bridge

Proposed changes to `specs/` resulting from this plan. **The Lead validates and
absorbs these claims directly into their owning spec areas** through the
`cumaru-absorb` workflow. Do not edit `specs/` before that workflow.

---

## specs/bridge/index.md

### Modified Requirements

- Every `relay` invocation MUST be logged to `$ALTER_BRIDGE_ROOT/.tmp/relay.log`
  so a silently dead `FileChanged` watcher is diagnosable from the log's last
  entry.
  (was: `$AGENT_BRIDGE_ROOT/.tmp/relay.log`)

### Prose updates (not EARS/RFC 2119 requirements, but factual claims that changed)

- `## Overview` — first sentence: "`alter-bridge` is a filesystem mailbox..." (was `agent-bridge`).
- `## Decisions` — "The bridge script lives under `skills/alter-bridge/scripts/`..." (was `skills/agent-bridge/scripts/`).
- `## Files` — all six entries repoint from `skills/agent-bridge/...` to `skills/alter-bridge/...` (script, install-hooks.sh, SKILL.md); the plugin manifests, README, and their descriptions now name `alter-bridge`.
- `## Reference` table — all `skills/agent-bridge/...` links and labels repoint to `skills/alter-bridge/...`; the CLI's own display name changes from `agent-bridge` to `alter-bridge`.

> No new or removed requirements — pure rename across paths, the env var, and the default mailbox root (`~/.agent-bridge` -> `~/.alter-bridge`, `AGENT_BRIDGE_ROOT` -> `ALTER_BRIDGE_ROOT`). No behavior change to addressing, delivery, or the doorbell.
