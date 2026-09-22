---
human_revised: false
plan: maintenance-go-send
status: draft
date: 2026-09-22
summary: Delta for maintenance-go-send — record the Go send identity model as a decision and reference its sources, leaving bash requirements authoritative until cutover.
---

# Delta draft — maintenance-go-send

The Go `send` is built and tested but not wired into hooks or skills; the bash
CLI is still what runs. Following the precedent set when `EnsureRoot` landed,
the bash Requirements stay authoritative and the Go behavior is recorded under
Decisions. The Requirements are rewritten at cutover, when every command is
ported.

## specs/bridge/index.md

### Added Requirements

- None. Bash requirements remain authoritative until cutover.

### Modified Requirements

- None.

### Removed Requirements

- None.

### Added Decisions

- 2026-09 (`maintenance-go-send`): the Go `send` diverges from the bash
  identity model on purpose, and will replace it at cutover:
  - Identity is option-only: `--from provider:value` is required and no
    environment variable (`AGENT_SLUG`, `CLAUDE_CODE_SESSION_ID`,
    `CODEX_THREAD_ID`, ...) decides sender or recipient. The working-directory
    fallback and the `default` slug are gone. Only `HOME` (to locate session
    stores) and `ALTER_BRIDGE_ROOT` are read.
  - Providers are `claude`, `codex`, `opencode`.
  - A value matching a session id resolves to that session's slugified name,
    or its slugified id when unnamed; any other value is a name. The mailbox is
    keyed by the name slug, so a new session reusing an archived session's name
    inherits its mailbox.
  - Name matching considers only active sessions (Codex `archived = 0`,
    OpenCode `time_archived IS NULL`; Claude has no archive flag, so every
    titled session counts). Two or more active sessions sharing a name refuse
    delivery and list their ids so the caller addresses one by id.
  - Session state is read live: Claude `custom-title` entries in
    `~/.claude/projects/*/<id>.jsonl`, Codex `state_5.sqlite` (fallback
    `session_index.jsonl`), OpenCode `opencode.db` `title`, via the `sqlite3`
    CLI with `mode=ro`.
  - `--type` is validated against `message | question | result | ack`; an
    empty resolved slug is refused. Frontmatter, file naming, and atomic
    temp-then-rename delivery are byte-compatible with bash (bash `peek` reads
    Go-written messages).
- Observed 2026-09-22 on Claude Code v2.1.280: `claude agents --json` failed
  with `too many arguments for 'agents'` and only covers background agents,
  so the bash Claude name lookup (`claude_session_name`, and `who`'s Claude
  half) likely resolves nothing for interactive sessions.

### Added Files / Reference rows

- `go/cmd/alter-bridge/main.go` — Go CLI entry: subcommand dispatch.
- `go/cmd/alter-bridge/send.go` — Go `send`: flags, required `--from`, resolution, delivery, nudge.
- `go/internal/address/address.go` — Go address parsing and bash-parity slugify.
- `go/internal/session/session.go` — Go resolver: id/name to mailbox slug, archive-aware ambiguity refusal.
- `go/internal/session/store.go` — Go readers for Claude, Codex, and OpenCode session stores.
- `go/internal/message/message.go` — Go message frontmatter and atomic delivery.
- `go/internal/nudge/nudge.go` — Go best-effort `codex queue` nudge.
