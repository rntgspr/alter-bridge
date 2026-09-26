# alter-bridge

```text
         /\     /\
        /  \   /  \
       / A  \ / B  \
      /      X      \
     /      / \      \
    /______/   \______\
    |                 |
    |  ALTER  BRIDGE  |
    |_________________|
```

A filesystem mailbox shared by every agent on a machine. One file per message,
delivered by an atomic `mv`, archived on read. No daemon, no network.

Agents are addressed as `provider:slug` — `claude:pikachu`, `codex:bridge`.

## Why it exists

Two agents from different vendors, on the same machine, had no way to talk.
A directory and an atomic rename turned out to be enough: delivery never depends
on the recipient being awake, and the message waits as long as it has to.

## Install

alter-bridge is a plugin on both runtimes, and its hooks ship inside the
plugin, so nothing is written into global settings by hand.

### Claude Code

Install it through the standard marketplace flow, straight from GitHub:

```
/plugin marketplace add rntgspr/alter-bridge
/plugin install alter-bridge@alter-bridge
```

You need no checkout and no Go toolchain. The release workflow makes the
`alter-bridge` entry in `.claude-plugin/marketplace.json` an `archive` source.
That source points at the zip attached to the latest GitHub Release and pins
that zip's `sha256`. Pushing a
`v*` tag runs `.github/workflows/release.yml`. The workflow calls
`go/release.sh`, which cross-compiles the CLI into `go/alter-bridge-<os>-<arch>`
and packs the plugin tree into a zip. The workflow then publishes the release
and commits the new `url` and `sha256` to the marketplace entry.

`hooks/hooks.json` wires `SessionStart`, `UserPromptSubmit`, and `FileChanged`
(`watchpaths claude`, `hook claude`, `relay claude`). Those hooks register the
mailbox watch paths, drain pending messages into each turn, and wake the
session when a message lands. Each hook calls the launcher `go/alter-bridge`,
which runs the binary built for the host. On a host with no matching binary,
the launcher prints one line to stderr and exits 0, so a prompt is never
blocked. A session that was already running when the plugin was installed needs
`/reload-plugins`. To pick up a new release, run `/plugin marketplace update`.

Local development runs the plugin from a checkout, with no install:

```
go/release.sh            # builds go/alter-bridge-<os>-<arch> (gitignored) and dist/*.zip
claude --plugin-dir .
```

### Codex

Install or update with one command. The comments show the equivalent manual
commands for v0.2.0. The script uses the latest release and a temporary
directory, validates the archive, then copies it into the persistent path.
It also checks for `curl`, `unzip`, and `codex`. No Git checkout or Go toolchain
is needed. Every run downloads the archive again and force-overwrites matching
files from an earlier installation:

```sh
# mkdir -p "$HOME/.local/share/alter-bridge-codex"
# curl -fL https://github.com/rntgspr/alter-bridge/releases/download/v0.2.0/alter-bridge-codex.zip -o /tmp/alter-bridge-codex.zip
# unzip -oq /tmp/alter-bridge-codex.zip -d "$HOME/.local/share/alter-bridge-codex"
# codex plugin marketplace add "$HOME/.local/share/alter-bridge-codex"
# codex plugin add alter-bridge@alter-bridge-codex
curl -fsSL https://raw.githubusercontent.com/rntgspr/alter-bridge/refs/heads/main/install-codex.sh | sh
```

Codex gets `hooks/codex.json`, which `.codex-plugin/plugin.json` declares:
`UserPromptSubmit` (`hook codex`). Codex skips a plugin hook until you trust
it, so open `/hooks` in Codex and trust the alter-bridge hook once, and again
whenever its definition changes. Start a new thread after installation or
update. Re-run the same command to update the plugin.

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
.claude-plugin/marketplace.json        Claude marketplace (archive source)
.codex-plugin/plugin.json              Codex manifest (declares hooks/codex.json)
.github/workflows/release.yml          on v* tags: build both zips, update the Claude marketplace
hooks/hooks.json                       Claude hooks (call go/alter-bridge)
hooks/codex.json                       Codex hooks (call bin/alter-bridge)
dist/alter-bridge-codex.zip            precompiled Codex plugin with a local marketplace
install-codex.sh                       downloads and installs the Codex release archive
skills/alter-bridge/SKILL.md           how an agent is meant to use it
go/                                    the bridge CLI (Go)
go/release.sh                          builds binaries and both release zips in dist/
go/alter-bridge                        launcher: runs the go/ binary for the host
skills/alter-bridge/scripts/alter-bridge  launcher used by the installed skill
```

## Configuration

`ALTER_BRIDGE_ROOT` overrides the mailbox root (default `~/.alter-bridge`).
`AGENT_SLUG` overrides this session's address, which otherwise comes from the
session name.

## License

MIT
