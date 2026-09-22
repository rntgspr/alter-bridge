---
human_revised: false
name: bridge
summary: Filesystem mailbox letting Claude and Codex agents on the same machine exchange asynchronous messages via atomic file moves, with hook-driven delivery and a doorbell nudge.
depends-on: []
targets: [cli, hooks, plugin, skill]
---

# Bridge

## Overview

`alter-bridge` is a filesystem mailbox shared by every agent on a machine. One
file per message, delivered by an atomic `mv` into the recipient's directory
and archived on read. There is no daemon and no network: delivery never
depends on the recipient being awake, and a message waits in the mailbox as
long as it has to.

Agents are addressed as `provider:slug` (`claude:pikachu`, `codex:bridge`).
The system has two halves that only meet at the mailbox directory: the `send`
side, which any agent can invoke synchronously, and the delivery side, which
is driven entirely by each runtime's own hooks (`SessionStart`,
`UserPromptSubmit`, `FileChanged`) so a session is nudged even while idle.

## Requirements (EARS / RFC 2119)

### Addressing

- An address MUST take the form `provider:slug`, with `provider` being
  `claude` or `codex`.
- A leading `#` on an address MUST be accepted and stripped before parsing.
- WHEN a session's own address is resolved THE SYSTEM SHALL use, in order:
  an explicit `AGENT_SLUG` override, the session name, then the path segment
  after `agentic-workspace/` as the last resort.
- Session names MUST be slugified for use as a mailbox directory name (e.g.
  `workspace/fe` becomes `workspace-fe`).

### Message delivery

- WHEN `send` is invoked THE SYSTEM SHALL write the message body to a temp
  file and `mv` it into the recipient's mailbox directory, so a reader never
  observes a partially written message.
- Every message file MUST carry frontmatter: `from`, `to`, `ts`, `msgid`, and
  `type` (`message | question | result | ack`); `thread` and `in_reply_to`
  are optional.
- WHEN a message is read via `inbox` THE SYSTEM SHALL archive it by moving it
  to `.archive/`, renamed to record who read it, disambiguating colliding
  names so archiving never loses a message.
- `peek` SHALL read pending messages without archiving them.
- A message of `type: question` MUST be treated as open until a reply is
  sent back to it; `message`, `result`, and `ack` types are informational.
- Never hand-write a message file directly into an agent's mailbox directory
  — creation MUST go through the CLI so delivery stays atomic and naming
  stays parseable.

### Doorbell (hook-driven wake-up)

- WHEN a Claude session starts THE SYSTEM SHALL register a `watchPaths` entry
  for the whole provider directory (not just this agent's own mailbox), so a
  later `/rename` or a mailbox that does not exist yet still rings.
- WHEN a filesystem `add` event lands inside a session's own mailbox THE
  SYSTEM SHALL spawn a throwaway headless relay session that wakes the target
  session and deletes its own transcript on exit; `change` events and files
  outside the bridge SHALL be ignored.
- The relay notice MUST reach the human via a `systemMessage`, since
  `FileChanged` runs outside the REPL and has no model context channel; the
  message content itself SHALL still arrive on the target's next turn via
  `UserPromptSubmit`.
- WHEN sending to a live Codex session THE SYSTEM SHALL nudge it directly via
  `codex queue`, and SHALL skip the nudge silently when no session by that
  name is running.
- Every `relay` invocation MUST be logged to `$ALTER_BRIDGE_ROOT/.tmp/relay.log`
  so a silently dead `FileChanged` watcher is diagnosable from the log's last
  entry.

### Plugin packaging and install

- The bridge MUST ship as a plugin on both runtimes: `.claude-plugin/plugin.json`
  for Claude, `plugin.json` for Codex.
- Hook wiring is NOT part of the plugin manifest (Codex does not yet execute
  plugin-bundled hooks — openai/codex#16430) — `install-hooks.sh` MUST write
  hook entries directly into each runtime's global config
  (`~/.claude/settings.json`, `~/.codex/hooks.json`), backing up the Codex
  file before any rewrite.
- Re-running `install-hooks.sh` MUST be safe: an entry already pointing at one
  of the bridge's own commands is reconciled in place rather than duplicated,
  and a file with nothing to change is left unwritten.
- `install-hooks.sh` MAY be scoped to one runtime via a `claude` or `codex`
  argument; with no argument it SHALL install both.

## Decisions

- 2026-09 (repo bootstrap `847387a`..`95015e5`): extracted from a `.agents`
  monorepo into its own repository; root `plugin.json` restored for Codex
  discovery and repository URLs repointed. Reflects a packaging move, not a
  behavior change.
- The bridge script lives under `skills/alter-bridge/scripts/`, not `bin/`,
  so it is deliberately off `PATH` — every caller invokes it by absolute
  path, keeping the mailbox tool scoped to the skill that documents its use.
- `who` reads addressable agents live from each CLI's own state (`claude
  agents --json`, the Codex `state_5.sqlite` threads table) rather than
  caching a roster here, since the CLIs already hold that truth and a second
  copy would only drift. Observed 2026-09-22 on Claude Code v2.1.280:
  `claude agents --json` failed with `too many arguments for 'agents'` and
  only covers background agents, so the bash Claude name lookup
  (`claude_session_name`, and `who`'s Claude half) likely resolves nothing for
  interactive sessions.
- A Go rewrite has started under `go/` (its own module, kept out of the repo
  root so future non-Go tooling can sit alongside it without mixing). The
  first piece is `internal/broker.EnsureRoot`, which resolves and creates the
  mailbox root (`$ALTER_BRIDGE_ROOT` override, else `$HOME/.alter-bridge`)
  and refuses an empty or `/` `$HOME` before touching the filesystem —
  `cmd/alter-bridge/main.go` calls it before running a subcommand. This guard
  is stricter than the bash script, which never validates `$HOME`; the bash
  CLI's documented Requirements above are unchanged and still authoritative
  until the rewrite covers addressing, delivery, and the doorbell.
- 2026-09 (`maintenance-go-send`): the Go `send` is built but not yet wired
  into hooks or the skill, and it diverges from the bash identity model on
  purpose; it replaces the Addressing requirements above at cutover:
  - Identity is option-only: `--from provider:value` is required and no
    environment variable (`AGENT_SLUG`, `CLAUDE_CODE_SESSION_ID`,
    `CODEX_THREAD_ID`, ...) decides sender or recipient. The working-directory
    fallback and the `default` slug are gone; only `HOME` (to locate session
    stores) and `ALTER_BRIDGE_ROOT` are read.
  - Providers are `claude`, `codex`, `opencode`.
  - A value matching a session id resolves to that session's slugified name,
    or its slugified id when unnamed; any other value is a name. The mailbox
    is keyed by the name slug, so a new session reusing an archived session's
    name inherits its mailbox.
  - Name matching considers only active sessions (Codex `archived = 0`,
    OpenCode `time_archived IS NULL`; Claude has no archive flag, so every
    titled session counts). Two or more active sessions sharing a name refuse
    delivery and list their ids so the caller addresses one by id.
  - Session state is read live: Claude `custom-title` entries in
    `~/.claude/projects/*/<id>.jsonl`, Codex `state_5.sqlite` (fallback
    `session_index.jsonl`), OpenCode `opencode.db` `title`, through the
    `sqlite3` CLI in `mode=ro` so the binary carries no SQLite driver.
  - `--type` is validated against `message | question | result | ack`, and an
    empty resolved slug is refused. Frontmatter, file naming, and atomic
    temp-then-rename delivery are byte-compatible with bash, whose `peek`
    reads Go-written messages. A live Codex recipient is nudged via
    `codex queue`, best effort.

## Files

- [skills/alter-bridge/scripts/alter-bridge](/skills/alter-bridge/scripts/alter-bridge) — the CLI: addressing, `send`, `inbox`/`peek`, `hook`, `watchpaths`, `relay`, `archive`, `purge`, `who`.
- [skills/alter-bridge/scripts/install-hooks.sh](/skills/alter-bridge/scripts/install-hooks.sh) — installs/reconciles hooks in `~/.claude/settings.json` and `~/.codex/hooks.json`.
- [skills/alter-bridge/SKILL.md](/skills/alter-bridge/SKILL.md) — how an agent is meant to use the bridge (addressing, sending, replying, rules).
- [.claude-plugin/plugin.json](/.claude-plugin/plugin.json) — Claude plugin manifest.
- [plugin.json](/plugin.json) — Codex plugin manifest.
- [README.md](/README.md) — install, message flow, and layout docs for the repository.
- [go/internal/broker/root.go](/go/internal/broker/root.go) — Go rewrite: resolves and guards the mailbox root.
- [go/cmd/alter-bridge/](/go/cmd/alter-bridge/) — Go rewrite: CLI entry and the `send` subcommand.
- [go/internal/address/address.go](/go/internal/address/address.go) — Go rewrite: address parsing and bash-parity slugify.
- [go/internal/session/](/go/internal/session/) — Go rewrite: session-store readers and id/name-to-slug resolution.
- [go/internal/message/message.go](/go/internal/message/message.go) — Go rewrite: message frontmatter and atomic delivery.
- [go/internal/nudge/nudge.go](/go/internal/nudge/nudge.go) — Go rewrite: best-effort `codex queue` nudge.

## Reference

<!-- cumaru:reference -->
| Link | Description |
|------|-------------|
| [alter-bridge](skills/alter-bridge/scripts/alter-bridge) | Mailbox CLI: address parsing, atomic send, drain/archive, hook entry points, doorbell relay, roster lookup. |
| [install-hooks.sh](skills/alter-bridge/scripts/install-hooks.sh) | Installs/reconciles the Claude and Codex hook wiring that drives delivery and the doorbell. |
| [SKILL.md](skills/alter-bridge/SKILL.md) | Agent-facing usage contract for the bridge. |
| [.claude-plugin/plugin.json](.claude-plugin/plugin.json) | Claude-side plugin manifest. |
| [plugin.json](plugin.json) | Codex-side plugin manifest. |
| [README.md](README.md) | Install steps, message-flow walkthrough, layout, and the known `FileChanged` rough edge. |
| [go/internal/broker/root.go](go/internal/broker/root.go) | Go rewrite: `EnsureRoot` resolves and creates the mailbox root, guarding against an empty or `/` `$HOME`. |
| [go/cmd/alter-bridge/main.go](go/cmd/alter-bridge/main.go) | Go rewrite: CLI entry; dispatches subcommands and reads only `HOME` and `ALTER_BRIDGE_ROOT`. |
| [go/cmd/alter-bridge/send.go](go/cmd/alter-bridge/send.go) | Go rewrite: `send` — flags, required `--from`, slug resolution, delivery, nudge. |
| [go/internal/address/address.go](go/internal/address/address.go) | Go rewrite: `provider:value` parsing (claude, codex, opencode) and bash-parity slugify. |
| [go/internal/session/session.go](go/internal/session/session.go) | Go rewrite: resolves ids/names to mailbox slugs, skipping archived sessions and refusing ambiguous names. |
| [go/internal/session/store.go](go/internal/session/store.go) | Go rewrite: live readers for Claude transcripts, Codex and OpenCode session databases. |
| [go/internal/message/message.go](go/internal/message/message.go) | Go rewrite: bash-compatible frontmatter and file naming, atomic temp-then-rename delivery. |
| [go/internal/nudge/nudge.go](go/internal/nudge/nudge.go) | Go rewrite: best-effort `codex queue` nudge for live Codex recipients. |
<!-- /cumaru:reference -->
