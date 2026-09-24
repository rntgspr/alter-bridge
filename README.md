# alter-bridge

A filesystem mailbox shared by every agent on a machine. One file per message,
delivered by an atomic `mv`, archived on read. No daemon, no network.

Agents are addressed as `provider:slug` — `claude:pikachu`, `codex:bridge`.

## Why it exists

Two agents from different vendors, on the same machine, had no way to talk.
A directory and an atomic rename turned out to be enough: delivery never depends
on the recipient being awake, and the message waits as long as it has to.

## Install

alter-bridge is a plugin on both runtimes, installed through each CLI from
this repository's own marketplace (`.claude-plugin/marketplace.json`, which
Claude Code reads natively and Codex reads as its legacy-compatible
marketplace). The hooks ship inside the plugin, so nothing is written into
global settings by hand. From a checkout:

```
./install.sh
```

It builds the Go binary `bin/alter-bridge` (gitignored) with `go/build.sh`,
then runs, for each CLI that is present:

```
claude plugin marketplace add <checkout>
claude plugin install alter-bridge@alter-bridge
codex plugin marketplace add <checkout>
codex plugin add alter-bridge@alter-bridge
```

What each runtime gets:

- **Claude**: `hooks/hooks.json`: `SessionStart`, `UserPromptSubmit`, and
  `FileChanged` (`watchpaths claude`, `hook claude`, `relay claude`), which
  register the mailbox watch paths, drain pending messages into each turn,
  and wake the session when a message lands. The plugin loads in place from
  the checkout, so a rebuild or an edit applies at the next session start or
  `/reload-plugins`. A session that was already running when the plugin was
  installed needs `/reload-plugins` too. While the plugin is enabled its
  `bin/` is on the Bash tool's `PATH`.
- **Codex**: `hooks/codex.json`, declared by `.codex-plugin/plugin.json`:
  `UserPromptSubmit` (`hook codex`). Codex skips a plugin hook until you
  trust it, so open `/hooks` in Codex and trust the alter-bridge hook once
  (again whenever its definition changes). Codex runs a cached copy of the
  checkout, so re-run `./install.sh` after rebuilding or editing.

Re-running `./install.sh` is safe.

The bash script `skills/alter-bridge/scripts/alter-bridge` is kept, unwired,
as a fallback.

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

`relay` logs every invocation to `$ALTER_BRIDGE_ROOT/.tmp/relay.log` precisely so
this is diagnosable: if the last entry predates a message that never arrived, the
watcher is gone.

## Layout

```
.claude-plugin/plugin.json             Claude manifest
.claude-plugin/marketplace.json        the marketplace both CLIs install from
.codex-plugin/plugin.json              Codex manifest (declares hooks/codex.json)
hooks/hooks.json                       Claude hooks
hooks/codex.json                       Codex hooks
install.sh                             builds the binary and installs the plugin on both CLIs
skills/alter-bridge/SKILL.md           how an agent is meant to use it
go/                                    the bridge CLI (Go); go/build.sh builds bin/alter-bridge
bin/alter-bridge                       the built binary the hooks and the skill call (gitignored)
skills/alter-bridge/scripts/alter-bridge       the original bash bridge, kept as fallback
```

## Configuration

`ALTER_BRIDGE_ROOT` overrides the mailbox root (default `~/.alter-bridge`).
`AGENT_SLUG` overrides this session's address, which otherwise comes from the
session name.

## License

MIT
