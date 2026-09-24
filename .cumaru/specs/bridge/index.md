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

- An address MUST take the form `provider:value`, with `provider` being
  `claude`, `codex`, or `opencode`, and `value` a session name or id.
- A leading `#` on an address MUST be accepted and stripped before parsing.
- A session's own address MUST come from an explicit `provider:value` — the
  hook payload's session id for the hook entry points, `--from` for `send`,
  the argument for `inbox` / `peek` — and never from an environment variable
  or the working directory (see the `maintenance-go-send` and
  `maintenance-go-agent-names` Decisions for name resolution).
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
- WHEN the `UserPromptSubmit` hook runs THE SYSTEM SHALL read the session id
  (`session_id`, `thread_id`, or `threadId`) and `cwd` from the JSON payload
  on stdin, do nothing when that `cwd` (or, when it is unusable, the process
  working directory) is outside `~/agentic-workspace`, and otherwise archive
  and print every pending message of that session's mailbox exactly as
  `inbox` does, creating neither the root nor the mailbox.
- WHEN `archive` is given an address THE SYSTEM SHALL move every pending
  message of that mailbox to `.archive/` exactly as `inbox` does, print none
  of them, and report `archived <n> from <address>`.
- WHEN `archive` is given no address THE SYSTEM SHALL do the same for every
  `<provider>/<slug>` mailbox under the root, skipping dot-named directories,
  and report one line per mailbox that had messages, then `total: <n>`.
- WHEN `purge` is invoked THE SYSTEM SHALL permanently delete every `*.md`
  entry directly under `<root>/.archive` (not dot-named, not recursive, no
  other file types), without confirmation and without creating the root, and
  report `purged <n> archived message(s) from <root>/.archive`, or
  `nothing to purge` when `.archive/` is not a directory.
- A message of `type: question` MUST be treated as open until a reply is
  sent back to it; `message`, `result`, and `ack` types are informational.
- Never hand-write a message file directly into an agent's mailbox directory
  — creation MUST go through the CLI so delivery stays atomic and naming
  stays parseable.

### Doorbell (hook-driven wake-up)

- WHEN a Claude session starts inside `~/agentic-workspace` THE SYSTEM SHALL
  create the provider directory, this session's own mailbox, and
  `<root>/.tmp`, then print
  `{"hookSpecificOutput":{"hookEventName":"SessionStart","watchPaths":[<root>/<provider>, <root>/<provider>/<slug>]}}`
  as one compact JSON line, registering the whole provider directory (not
  just this agent's own mailbox) so a later `/rename` or a mailbox that does
  not exist yet still rings; outside `~/agentic-workspace` it SHALL print and
  create nothing.
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
- `install-hooks.sh` MUST wire the Go binary with the provider argument —
  `"$HOME/.../bin/alter-bridge" watchpaths claude`, `... hook claude`,
  `... relay claude` (matcher `.*__from_.*`), and `... hook codex` — and MUST
  refuse (exit 1) when `bin/alter-bridge` is missing.
- Re-running `install-hooks.sh` MUST be safe: an entry already pointing at one
  of the bridge's own commands, in either the legacy `bash "<script>" <sub>`
  form or the Go `"<bin>" <sub> <provider>` form, is reconciled in place
  rather than duplicated, and a file with nothing to change is left unwritten.
- `install-hooks.sh` MAY be scoped to one runtime via a `claude` or `codex`
  argument; with no argument it SHALL install both.

## Decisions

- 2026-09 (repo bootstrap `847387a`..`95015e5`): extracted from a `.agents`
  monorepo into its own repository; root `plugin.json` restored for Codex
  discovery and repository URLs repointed. Reflects a packaging move, not a
  behavior change.
- The bridge CLI is deliberately off `PATH` — every caller invokes it by
  absolute path, keeping the mailbox tool scoped to the skill that documents
  its use. Originally the bash script under `skills/alter-bridge/scripts/`;
  since `maintenance-go-cutover`, `bin/alter-bridge`.
- `who` reads addressable agents live from each CLI's own state (`claude
  agents --json`, the Codex `state_5.sqlite` threads table) rather than
  caching a roster here, since the CLIs already hold that truth and a second
  copy would only drift. `claude agents --json` works when the real binary
  runs (verified on Claude Code 2.1.281: it lists live `kind: interactive`
  sessions with `sessionId`, `name`, `cwd`); an earlier `too many arguments
  for 'agents'` failure came from an interactive-shell wrapper function
  named `claude`, not the CLI.
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
    `~/.claude/projects/*/<id>.jsonl`, else the live `claude agents --json`
    name for that id (see `maintenance-go-agent-names`), Codex `state_5.sqlite` (fallback
    `session_index.jsonl`), OpenCode `opencode.db` `title`, through the
    `sqlite3` CLI in `mode=ro` so the binary carries no SQLite driver.
  - `--type` is validated against `message | question | result | ack`, and an
    empty resolved slug is refused. Frontmatter, file naming, and atomic
    temp-then-rename delivery are byte-compatible with bash, whose `peek`
    reads Go-written messages. A live Codex recipient is nudged via
    `codex queue`, best effort.
- 2026-09 (`maintenance-go-inbox`): the Go `inbox` is built but not yet wired
  into hooks or the skill. It requires its `provider:value` address (no
  `self_addr` fallback, per the go-send identity model) and resolves it
  through the same session resolver as `send`, refusing ambiguous or empty
  slugs. Its output and `.archive/` naming (split at the last `__`,
  `__to_<provider>_<slug>__`, a `.<8 hex>` suffix on collision) are
  byte-identical to bash. The read path is `internal/mailbox.Drain`,
  parameterized by whether it archives, for `peek`, `archive`, and `hook` to
  reuse. Unlike bash, a directory named `*.md` inside a mailbox is skipped
  rather than archived.
- 2026-09 (`maintenance-go-peek`): the Go `peek` is built but not yet wired
  into hooks or the skill. It shares `inbox`'s CLI path (`runDrain`: required
  `provider:value` address with no `self_addr` fallback, same session
  resolver, same exit codes) and calls `mailbox.Drain` with archiving off, so
  it prints oldest first in `inbox`'s format, byte-identical to bash `peek`,
  and moves nothing.
- 2026-09 (`maintenance-go-who`): the Go `who` is built but not yet wired into
  hooks or the skill. Its stdout is byte-identical to bash `who`: Claude rows
  from `claude agents --json` (`kind == "interactive"`, named only,
  `%-26s %s  %s` with the `cwd`), then the first 10 Codex threads, named only
  (`%-26s %s`), archived ones included as in bash. Deviations: the `claude`
  binary falls back to `~/.local/bin/claude` when absent from `PATH`; the
  Codex half reuses the `send` resolver's Codex store, so it also falls back
  to `session_index.jsonl` when `state_5.sqlite` or `sqlite3` is unavailable
  (that index has no recency, so the 10-row window then shows the
  oldest-indexed threads); any argument is a usage error (exit 2). An
  unavailable provider only omits its rows (exit 0, empty stderr), and `who`
  never creates the mailbox root. The live Claude reader
  (`session.ClaudeAgents`) is also merged into the transcript store used for
  addressing (see `maintenance-go-agent-names`).
- 2026-09 (`maintenance-go-archive`): the Go `archive` is built but not yet
  wired into hooks or the skill. It archives through the same
  `internal/mailbox` loop and `archiveOne` as `inbox` (`mailbox.Archive`, a
  count-only mode that never reads the message) and sweeps `mailbox.Boxes`
  (non-dot `<provider>/<slug>` directories, byte order). Stdout and the
  resulting `.archive/` names are byte-identical to bash for a sweep and for
  a single address. A given address is resolved through the session resolver
  (go-send model) and echoed as given. Deviations: a missing mailbox reports
  `archived 0` (bash prints an empty count); `-h`, `--help`, extra
  arguments, and unknown providers are usage errors (exit 2); ambiguous or
  empty slugs exit 1.
- 2026-09 (`maintenance-go-purge`): the Go `purge` is built but not yet wired
  into hooks or the skill. The delete is `mailbox.Purge` next to `Archive`
  (`ErrNoArchive` when `.archive` is not a directory), and the root comes from
  `broker.Resolve`, the guard `EnsureRoot` now wraps, so `purge` never creates
  the root. Stdout and the resulting tree are byte-identical to bash for a
  populated, an absent, and a root-less archive; symlinked `*.md` entries are
  removed as links, broken ones kept. Deviations: a directory named `*.md` is
  skipped (bash `rm` fails and aborts mid-purge with exit 1); any argument,
  including `-h`/`--help`, is a usage error (exit 2); the printed path is
  `filepath.Join`-cleaned.
- 2026-09 (`maintenance-go-hook`): the Go `hook` is built but not yet wired
  into hooks or the skill. It follows the go-send identity model instead of
  `self_addr`: the argument is required, either a bare provider
  (`hook claude`, value taken from the payload session id) or an explicit
  `provider:value`, both resolved through the `send`/`inbox` session resolver
  and drained with `mailbox.Drain`. No environment variable picks the
  provider or slug, and a payload without a session id drains nothing (exit
  0) instead of falling back to the cwd segment, so a session with neither a
  title nor a live agent name has its mailbox keyed by its id. The root comes from `broker.Resolve` (never
  created). Stdout and the resulting tree are byte-identical to bash for an
  explicit address inside and outside the workspace, a missing mailbox, and
  an absent root. Payload problems (empty, not JSON, non-string fields) are
  silent exit 0, so the prompt is never blocked; deviations: a missing,
  extra, `-h`/`--help`, unknown-provider, or malformed argument exits 2 even
  outside the workspace (bash checks the shape only, after the gate), and
  ambiguous or empty slugs exit 1. Cutover must pass the provider in the
  hook wiring.
- 2026-09 (`maintenance-go-watchpaths`): the Go `watchpaths` is built but not
  yet wired into hooks or the skill. It shares `hook`'s argument, payload,
  and workspace-gate helpers (`go/cmd/alter-bridge/hookinput.go`) and its
  identity model: a bare provider (value = the payload's first non-empty
  `session_id` / `thread_id` / `threadId`; bash reads `session_id` only) or an
  explicit `provider:value`, both resolved through the session resolver; no
  session id registers nothing (exit 0). The root comes from `broker.Resolve`
  and is created only after the gate, by the mailbox `MkdirAll`, as bash's
  `mkdir -p` does. The JSON line is encoded with HTML escaping off, and it
  and the resulting tree are byte-identical to bash for an existing mailbox,
  a new mailbox, an absent root, and a `cwd` outside the workspace.
  Deviations: argument errors exit 2 even outside the workspace, ambiguous or
  empty slugs exit 1, and paths are `filepath.Join`-cleaned. Cutover must
  pass the provider in the SessionStart wiring.
- 2026-09 (`maintenance-go-agent-names`): the Go Claude session store names a
  session by its transcript `custom-title`, else by the name
  `claude agents --json` reports for that `sessionId` (the automatic name of
  a session never `/rename`d, e.g. `alter-bridge-11`), and also lists running
  sessions that have no transcript yet. Id -> name and name -> id both go
  through the unchanged resolver, so a titled transcript and a live agent
  sharing a name are ambiguous like any two active sessions. A missing or
  failing `claude` binary (absent, non-zero exit, invalid JSON) leaves
  transcript titles only, with no error. This restores bash parity: fed the
  same SessionStart payload for a live unrenamed session, bash `watchpaths`
  and Go `watchpaths claude` register the same `claude/<automatic-name>`
  mailbox, which Go `hook claude` drains. Costs one `claude agents`
  subprocess per store read, as bash did. Automatic names can change when a
  session resumes (`alter-bridge-ef` became `alter-bridge-11`); bash and Go
  both follow the current name. Codex and OpenCode resolution is unchanged.
- 2026-09 (`maintenance-go-cutover`): Go is canonical. The live hooks, the
  installer, `SKILL.md`, and `README.md` call `bin/alter-bridge` (built by
  `go/build.sh`; `bin/` is gitignored, so a checkout must build before
  installing) with the provider argument; the "built but not yet wired"
  notes in the per-command Go decisions above are superseded, and those
  decisions' deviations from bash are now the live behavior. The bash script
  is kept, unwired, as a fallback; returning to it means restoring the
  `.bak.<timestamp>` copies the installer (or the cutover) left next to each
  settings file. Before cutover, bash and Go were fed the same SessionStart
  payload for every live Claude and Codex session and registered the same
  mailbox. Residual risks: automatic Claude names change when a session
  resumes (`/rename` is the stable fix, as with bash); a changed Codex hook
  stays disabled until its trust is re-approved inside Codex; the Codex
  payload's id fields (`session_id` / `thread_id` / `threadId`) are not yet
  verified against a live Codex turn.

## Files

- [skills/alter-bridge/scripts/alter-bridge](/skills/alter-bridge/scripts/alter-bridge) — the original bash CLI, kept unwired as a fallback: addressing, `send`, `inbox`/`peek`, `hook`, `watchpaths`, `relay`, `archive`, `purge`, `who`.
- [go/build.sh](/go/build.sh) — builds the canonical binary `bin/alter-bridge` (gitignored) that the hooks and the skill call.
- [skills/alter-bridge/scripts/install-hooks.sh](/skills/alter-bridge/scripts/install-hooks.sh) — installs/reconciles the Go-binary hooks in `~/.claude/settings.json` and `~/.codex/hooks.json`, rewriting legacy bash entries in place.
- [skills/alter-bridge/SKILL.md](/skills/alter-bridge/SKILL.md) — how an agent is meant to use the bridge (addressing, sending, replying, rules).
- [.claude-plugin/plugin.json](/.claude-plugin/plugin.json) — Claude plugin manifest.
- [plugin.json](/plugin.json) — Codex plugin manifest.
- [README.md](/README.md) — install, message flow, and layout docs for the repository.
- [go/internal/broker/root.go](/go/internal/broker/root.go) — Go rewrite: resolves and guards the mailbox root, with or without creating it.
- [go/cmd/alter-bridge/](/go/cmd/alter-bridge/) — Go rewrite: CLI entry and the `send`, `inbox`, `peek`, `archive`, `purge`, `who`, `hook`, and `watchpaths` subcommands.
- [go/internal/mailbox/mailbox.go](/go/internal/mailbox/mailbox.go) — Go rewrite: oldest-first mailbox drain with optional bash-parity archiving, count-only archiving, mailbox listing, and purging the archive.
- [go/internal/address/address.go](/go/internal/address/address.go) — Go rewrite: address parsing, the provider check, and bash-parity slugify.
- [go/internal/session/](/go/internal/session/) — Go rewrite: session-store readers, the live `claude agents --json` reader, and id/name-to-slug resolution.
- [go/internal/message/message.go](/go/internal/message/message.go) — Go rewrite: message frontmatter and atomic delivery.
- [go/internal/nudge/nudge.go](/go/internal/nudge/nudge.go) — Go rewrite: best-effort `codex queue` nudge.

## Reference

<!-- cumaru:reference -->
| Link | Description |
|------|-------------|
| [alter-bridge](skills/alter-bridge/scripts/alter-bridge) | Original bash mailbox CLI, kept unwired as a fallback: address parsing, atomic send, drain/archive, hook entry points, doorbell relay, roster lookup. |
| [go/build.sh](go/build.sh) | Builds the canonical Go binary `bin/alter-bridge` (gitignored) that the hooks and the skill call. |
| [install-hooks.sh](skills/alter-bridge/scripts/install-hooks.sh) | Installs/reconciles the Claude and Codex hook wiring to `bin/alter-bridge <sub> <provider>`, refusing a missing binary and rewriting legacy bash entries in place. |
| [SKILL.md](skills/alter-bridge/SKILL.md) | Agent-facing usage contract for the bridge, written against the Go CLI (`--from`, explicit addresses). |
| [.claude-plugin/plugin.json](.claude-plugin/plugin.json) | Claude-side plugin manifest. |
| [plugin.json](plugin.json) | Codex-side plugin manifest. |
| [README.md](README.md) | Install steps, message-flow walkthrough, layout, and the known `FileChanged` rough edge. |
| [go/internal/broker/root.go](go/internal/broker/root.go) | Go rewrite: `Resolve` picks the mailbox root, guarding against an empty or `/` `$HOME`, without creating it; `EnsureRoot` resolves and creates it. |
| [go/cmd/alter-bridge/main.go](go/cmd/alter-bridge/main.go) | Go rewrite: CLI entry; dispatches subcommands and reads only `HOME` and `ALTER_BRIDGE_ROOT`. |
| [go/cmd/alter-bridge/send.go](go/cmd/alter-bridge/send.go) | Go rewrite: `send` — flags, required `--from`, slug resolution, delivery, nudge. |
| [go/internal/address/address.go](go/internal/address/address.go) | Go rewrite: `provider:value` parsing (claude, codex, opencode), `IsProvider` for a bare provider, and bash-parity slugify. |
| [go/internal/session/session.go](go/internal/session/session.go) | Go rewrite: resolves ids/names to mailbox slugs, skipping archived sessions and refusing ambiguous names. |
| [go/internal/session/store.go](go/internal/session/store.go) | Go rewrite: live readers for Claude transcripts, Codex and OpenCode session databases. |
| [go/internal/message/message.go](go/internal/message/message.go) | Go rewrite: bash-compatible frontmatter and file naming, atomic temp-then-rename delivery. |
| [go/internal/nudge/nudge.go](go/internal/nudge/nudge.go) | Go rewrite: best-effort `codex queue` nudge for live Codex recipients. |
| [go/cmd/alter-bridge/inbox.go](go/cmd/alter-bridge/inbox.go) | Go rewrite: `inbox` — required address, slug resolution, archiving drain; hosts `runDrain`, the argument and resolution path shared with `peek`. |
| [go/internal/mailbox/mailbox.go](go/internal/mailbox/mailbox.go) | Go rewrite: `Drain` reads a mailbox oldest first, optionally archiving each message under a bash-parity, collision-safe name; `Archive` does the same archiving without reading, returning a count; `Boxes` lists every non-dot `<provider>/<slug>` mailbox; `Purge` deletes the non-dot `*.md` entries directly under `.archive` (`ErrNoArchive` when it is not a directory). |
| [go/cmd/alter-bridge/peek.go](go/cmd/alter-bridge/peek.go) | Go rewrite: `peek` — `runDrain` with archiving off; prints pending messages and keeps them. |
| [go/cmd/alter-bridge/archive.go](go/cmd/alter-bridge/archive.go) | Go rewrite: `archive` — one resolved address or a sweep of every mailbox, archiving without printing and reporting counts. |
| [go/cmd/alter-bridge/purge.go](go/cmd/alter-bridge/purge.go) | Go rewrite: `purge` — no arguments, deletes the archived trail without confirmation or creating the root, bash-parity output. |
| [go/cmd/alter-bridge/who.go](go/cmd/alter-bridge/who.go) | Go rewrite: `who` — live Claude agents and the 10 most recent Codex threads in bash's format, degrading per provider. |
| [go/cmd/alter-bridge/hook.go](go/cmd/alter-bridge/hook.go) | Go rewrite: `hook`, the `UserPromptSubmit` entry point — provider or `provider:value` argument, payload session id and `cwd`, `~/agentic-workspace` gate, archiving drain without creating the root. |
| [go/cmd/alter-bridge/watchpaths.go](go/cmd/alter-bridge/watchpaths.go) | Go rewrite: `watchpaths`, the `SessionStart` entry point — provider or `provider:value` argument, `~/agentic-workspace` gate, mailbox and `.tmp` creation after the gate, compact `watchPaths` JSON. |
| [go/cmd/alter-bridge/hookinput.go](go/cmd/alter-bridge/hookinput.go) | Go rewrite: helpers shared by the hook entry points — argument shape (bare provider or `provider:value`), payload session id, and the `~/agentic-workspace` gate. |
| [go/internal/session/agents.go](go/internal/session/agents.go) | Go rewrite: `ClaudeAgents` reads running interactive sessions from `claude agents --json`, falling back to `~/.local/bin/claude`. |
<!-- /cumaru:reference -->
