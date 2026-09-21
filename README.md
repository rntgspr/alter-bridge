# agent-bridge

A filesystem mailbox shared by every agent on a machine. One file per message,
delivered by an atomic `mv`, archived on read. No daemon, no network.

Agents are addressed as `provider:slug` — `claude:pikachu`, `codex:bridge`.

## Why it exists

Two agents from different vendors, on the same machine, had no way to talk.
A directory and an atomic rename turned out to be enough: delivery never depends
on the recipient being awake, and the message waits as long as it has to.

## Install

agent-bridge is a plugin on both sides (`.claude-plugin/plugin.json` for
Claude, `plugin.json` for Codex — see [Layout](#layout)), installed via
`install.sh` at the repo root through each runtime's own local marketplace
(`plugins/.claude-plugin/marketplace.json`): `claude plugin install
agent-bridge@agents-marketplace` for Claude, `codex plugin add
agent-bridge@agents-marketplace` for Codex. Neither is symlinked into a
skills directory — that was the first approach and it worked, but it left the
plugin looking like a plain skill in `~/.claude/skills/`, which was confusing
next to the real skills there.

Being a plugin does not wire up its hooks — Codex does not execute
plugin-bundled hooks yet ([openai/codex#16430](https://github.com/openai/codex/issues/16430),
open), so hooks stay installed the same way on both sides: manually, into each
runtime's global config. From a checkout of this repository, run once:

```
skills/agent-bridge/scripts/install-hooks.sh
```

This installs both runtimes' hooks:

- **Claude** — writes `SessionStart`, `UserPromptSubmit`, and `FileChanged`
  entries into `~/.claude/settings.json`. These register the mailbox watch
  paths, drain pending messages into each turn, and wake the session when a
  message lands.
- **Codex** — writes a `UserPromptSubmit` hook to `~/.codex/hooks.json`, backing
  the file up before any rewrite. Codex tracks per-hook trust in `config.toml`,
  so a newly written hook needs to be approved there before it runs — the
  script says so when it writes. Skipped if `~/.codex` does not exist.

Pass `claude` or `codex` to install one side only:

```
skills/agent-bridge/scripts/install-hooks.sh claude
skills/agent-bridge/scripts/install-hooks.sh codex
```

Re-running is safe. An entry already pointing at one of our commands is
reconciled rather than duplicated, and a file with nothing to change is not
rewritten at all.

## How a message travels

1. `send` writes the message to a temp file and `mv`s it into the recipient's
   mailbox. The rename is atomic, so a reader never sees half a message.
2. On the recipient's side, `FileChanged` fires and `relay` checks the arrival is
   really theirs.
3. `relay` spawns a throwaway headless session — Haiku, low effort, one tool call
   — that wakes the target session and deletes its own transcript on the way out.
   Claude has no CLI equivalent of `codex queue`, and this is the gap it fills.
4. `UserPromptSubmit` drains the mailbox into the turn.
5. Draining archives: every file moves to `.archive/`, renamed to record who read
   it. Colliding names are disambiguated, so archiving never loses a message.

Codex sessions are nudged directly with `codex queue`, which its own CLI provides.

## Known rough edge

The `FileChanged` watcher has been observed dying silently after a couple of
hours — no error, no log, the session simply stops being woken. `watchPaths` is
registered at `SessionStart` alone, so a session cannot re-register or even
notice. Restarting is the only known recovery.

`relay` logs every invocation to `$AGENT_BRIDGE_ROOT/.tmp/relay.log` precisely so
this is diagnosable: if the last entry predates a message that never arrived, the
watcher is gone.

## Layout

```
plugin.json                            Codex manifest
.claude-plugin/plugin.json             Claude manifest
skills/agent-bridge/SKILL.md           how an agent is meant to use it
skills/agent-bridge/scripts/agent-bridge       the whole bridge
skills/agent-bridge/scripts/install-hooks.sh   installs/reconciles hooks in both runtimes
```

Both runtimes install from this local marketplace into a cache
(`~/.claude/plugins/cache/agents-marketplace/agent-bridge/` and
`~/.codex/plugins/cache/agents-marketplace/agent-bridge/`), not a live symlink
back to this directory. After editing a file here, run `./install.sh` at the
repo root again (it reinstalls both sides) to propagate the change — Claude
picks it up at the next session start or `/reload-plugins`, Codex needs the
reinstall itself to recopy the snapshot.

## Configuration

`AGENT_BRIDGE_ROOT` overrides the mailbox root (default `~/.agent-bridge`).
`AGENT_SLUG` overrides this session's address, which otherwise comes from the
session name.

## License

MIT
