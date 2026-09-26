---
human_revised: false
scope: [bridge]
status: blocked
summary: Package alter-bridge as a standard Claude Code and Codex plugin installed from the repository's own marketplace, with each runtime's hooks bundled in the plugin instead of hand-written into global config, then cut the live machine over.
targets: [plugin, hooks]
aux: []
---

# Standard plugin packaging on Claude Code and Codex

## Overview

alter-bridge claims to be a plugin on both runtimes, but its install story
does not meet either platform's standard: the README points at an
`install.sh` and a `plugins/.claude-plugin/marketplace.json` that do not
exist, the machine still runs the old `agent-bridge@agents-marketplace`
plugin, and every hook is hand-written into `~/.claude/settings.json` and
`~/.codex/hooks.json` by `install-hooks.sh`, justified by
[openai/codex#16430](https://github.com/openai/codex/issues/16430).

This plan ships the repository as its own marketplace and bundles each
runtime's hooks in the plugin, so both runtimes install and wire the bridge
through their official CLIs.

The accepted distribution path now uses two release archives: Claude loads
the SHA-pinned archive from `.claude-plugin/marketplace.json`; Codex downloads
`alter-bridge-codex.zip` through `install-codex.sh` and registers the archive's
local `alter-bridge-codex` marketplace. This supersedes the original shared
checkout marketplace and `install.sh` design for future cutover evidence.

### Research (2026-09-24, Claude Code 2.1.282, codex-cli 0.156.1)

Claude Code:

- Marketplace: `.claude-plugin/marketplace.json` at the repository root; a
  relative `source` starting with `./` resolves against the marketplace root,
  and `"source": "./"` (the root itself) is allowed
  ([plugin-marketplaces](https://code.claude.com/docs/en/plugin-marketplaces)).
- Plugin hooks: `hooks/hooks.json` at the plugin root, same schema as
  settings hooks, `SessionStart`, `UserPromptSubmit`, and `FileChanged`
  supported; `${CLAUDE_PLUGIN_ROOT}` is substituted and exported
  ([plugins-reference](https://code.claude.com/docs/en/plugins-reference)).
- A relative-path plugin in a marketplace added from a local directory loads
  in place, so `CLAUDE_PLUGIN_ROOT` is the checkout and edits apply at the
  next session start or `/reload-plugins` (same page). Observed:
  `claude plugin install` reports "it loads in place from
  <checkout>". The gitignored `bin/alter-bridge` is therefore present.
- A plugin's `bin/` is added to the Bash tool's `PATH` while enabled (same
  page).
- `claude plugin install` from a shell does not activate in running
  sessions; they need `/reload-plugins` or a restart. Settings-file hook
  edits, by contrast, are picked up live by a file watcher
  ([discover-plugins](https://code.claude.com/docs/en/discover-plugins),
  [hooks](https://code.claude.com/docs/en/hooks)). A plugin's hook and an
  identical settings hook are not deduplicated (hooks page).

Codex:

- Plugin hooks run now. [PR #19705](https://github.com/openai/codex/pull/19705)
  (merged 2026-04-28) added plugin hook discovery behind `plugin_hooks`;
  `codex features list` on 0.156.1 reports `plugin_hooks removed`, i.e. the
  gate is gone, and the [hooks docs](https://learn.chatgpt.com/docs/hooks)
  list plugin-bundled hooks as a source. Issue #16430 is stale-open.
- Plugin hooks are non-managed: they are skipped until the user trusts the
  exact definition in `/hooks`; trust is per-hash, stored under
  `[hooks.state."<plugin>@<marketplace>:<file>:<event>:0:0"]` in
  `config.toml` (hooks docs, observed on the machine).
- Hook env: `PLUGIN_ROOT` / `PLUGIN_DATA`, plus `CLAUDE_PLUGIN_ROOT` /
  `CLAUDE_PLUGIN_DATA` for compatibility
  ([build](https://developers.openai.com/codex/plugins/build)).
- Marketplace: `.agents/plugins/marketplace.json`, or the documented
  "legacy-compatible" `.claude-plugin/marketplace.json` (build page).
  Observed: `codex plugin marketplace add <checkout>` reads the Claude file.
- Manifest, observed on 0.156.1: with only `.codex-plugin/plugin.json`
  declaring `"hooks": "./hooks/codex.json"`, `hooks/list` reports exactly the
  Codex hook. With a root portable `plugin.json`, Codex ignores
  `extensions.com.openai.hooks` (no hooks at all); with no manifest `hooks`,
  it loads the default `hooks/hooks.json`, i.e. the Claude hooks.
- Local plugins are copied to `~/.codex/plugins/cache/<mkt>/<plugin>/<version>/`
  by a filesystem copy that includes the gitignored `bin/` (observed), so a
  rebuilt binary reaches Codex only by re-running `codex plugin add`.

## Acceptance Criteria (EARS / RFC 2119)

- AC1: WHEN `claude plugin marketplace add rntgspr/alter-bridge` and
  `claude plugin install alter-bridge@alter-bridge` run THE SYSTEM SHALL list
  `alter-bridge@alter-bridge` enabled with the `alter-bridge` skill and the
  `SessionStart`, `UserPromptSubmit`, and `FileChanged` hooks.
- AC2: WHEN `install-codex.sh` installs the separate Codex release archive THE
  SYSTEM SHALL list `alter-bridge@alter-bridge-codex` installed and enabled,
  and Codex SHALL expose exactly one bundled `userPromptSubmit` hook running
  `hook codex` after its definition is trusted.
- AC3: Plugin hook commands MUST reach the binary through the plugin root
  (`${CLAUDE_PLUGIN_ROOT}`, `${PLUGIN_ROOT}`), never an absolute checkout path.
- AC4: WHEN each effective hook command is fed a payload against a scratch
  `ALTER_BRIDGE_ROOT` THE SYSTEM SHALL behave as the manual hooks did:
  `watchpaths` prints the `watchPaths` JSON, `hook` drains and archives a
  pending message, `relay` logs and ignores a `change` event.
- AC5: The release MUST provide separate Claude and Codex archives with
  working platform binaries; `install-codex.sh` MUST install or update the
  Codex archive, while Claude MUST install through its release-backed
  marketplace.
- AC6: `README.md` and `SKILL.md` MUST describe the plugin install flow and
  MUST NOT reference `install-hooks.sh` or files that do not exist.
- AC7: WHEN the live machine is cut over THE SYSTEM SHALL have exactly one
  alter-bridge entry per event on each runtime, the old
  `agent-bridge@agents-marketplace` uninstalled and its marketplace removed on
  both runtimes, and a `.bak.<timestamp>` copy of every global file touched.
- AC8: `bin/alter-bridge who` MUST run read-only after the cutover.

## Plan / DAG

| Task | Title | Status | Depends on |
|------|-------|--------|-----------|
| [T1](t1.md) | Marketplace, manifests, bundled hooks, `install.sh` | done | — |
| [T2](t2.md) | Docs and installer retirement | done | T1 |
| [T3](t3.md) | Live cutover of this machine | blocked | T2 |

## Out of scope

- Further release features beyond the two-archive distribution path.
- `nudge.BridgeScript`, which still names the bash script in the Codex
  `codex queue` notice text.
- The in-progress `maintenance-go-relay` plan.

## Risks

- Cutover ordering: Claude settings hooks are removed live, while a shell
  plugin install activates only after `/reload-plugins`; Codex plugin hooks
  run only after `/hooks` trust. Removing manual hooks before those steps
  leaves running sessions without delivery (messages wait, not lost);
  removing them after runs both (duplicate relay wake-ups; drains stay
  exactly-once because a message prints only after its rename wins).
- The plugin puts `alter-bridge` on the Claude Bash `PATH`, relaxing the
  "deliberately off PATH" decision.
- Codex runs a cached copy: a rebuilt binary is stale there until
  `install.sh` (or `codex plugin add`) runs again.
