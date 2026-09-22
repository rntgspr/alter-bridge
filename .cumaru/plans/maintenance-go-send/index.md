---
human_revised: false
scope: [bridge]
status: in-progress
summary: Port the bash `send` command to Go — addressing, option-only identity resolved by session name or id, atomic message write, and the Codex nudge.
targets: [cli]
aux: []
---

# Go rewrite — `send`

## Overview

Port `send` from `skills/alter-bridge/scripts/alter-bridge` (lines ~160-225)
to the Go binary. This is the foundation command: it establishes address
parsing (`provider:slug`), identity resolution, and the message file format
that every other command reuses. Land this first — see `plans/index.md`'s
`go-rewrite-order` tag.

The port deliberately diverges from the bash identity model:

- Identity is **option-only**. No environment variable (`AGENT_SLUG`,
  `CLAUDE_CODE_SESSION_ID`, `CODEX_THREAD_ID`, ...) is read to decide who is
  sending; the sender comes from `--from`, which is required.
- The working-directory fallback (`agentic-workspace/<segment>`) and the
  `default` slug are removed.
- Claude, Codex, and OpenCode all assign a session id to every chat. A
  session's mailbox slug is its slugified name; WHEN the session has no name,
  the slugified session id is the fallback. The id is always present, so
  resolution never needs a guess.
- `opencode` joins `claude` and `codex` as a valid provider.

`ALTER_BRIDGE_ROOT` stays: it selects the mailbox root, not an identity, and is
already handled by `internal/broker.EnsureRoot`.

## Acceptance Criteria (EARS / RFC 2119)

### Addressing

- An address MUST take the form `provider:value`, with `provider` one of `claude`, `codex`, `opencode`; a leading `#` MUST be accepted and stripped.
- An address with an unknown provider or an empty half MUST be rejected with a non-zero exit and a message on stderr.
- Mailbox slugs MUST be slugified exactly as the bash `slugify` does (lowercase; characters outside `a-z0-9._-` become `-`; runs of `-` collapse; leading and trailing `-` trimmed).

### Identity resolution

- `send` MUST NOT read any environment variable to determine the sender or the recipient.
- `send` MUST require `--from provider:value`; WHEN `--from` is missing THE SYSTEM SHALL exit non-zero with a usage message.
- WHEN an address value matches a session id in that provider's own session state THE SYSTEM SHALL use that session's name as the slug, or the session id itself WHEN the name is empty.
- WHEN an address value matches no session id THE SYSTEM SHALL treat it as a name and use it as the slug.
- The mailbox is keyed by the name slug, so a new session reusing an archived session's name inherits its mailbox, pending messages included.
- Name resolution MUST consider only active sessions: Codex rows with `archived = 0`, OpenCode rows with `time_archived IS NULL`; Claude has no archive flag, so every titled session counts as active.
- WHEN exactly one active session carries the name, or none does, THE SYSTEM SHALL deliver to the name's mailbox (a message never depends on the recipient being awake).
- WHEN two or more active sessions of the same provider carry the same slugified name THE SYSTEM SHALL refuse delivery with a non-zero exit, listing the candidate session ids, so the caller retries with `provider:<id>`.
- Session state MUST be read live from each CLI's own store (Claude `custom-title` entries in `~/.claude/projects/*/<session-id>.jsonl`; Codex `~/.codex/state_5.sqlite` with `session_index.jsonl` as fallback; OpenCode `~/.local/share/opencode/opencode.db`), never from a cache kept by the bridge.
- WHEN a provider's session store is unavailable THE SYSTEM SHALL fall back to treating the value as a name, without failing.

### Message write

- `--type` MUST be one of `message | question | result | ack` (default `message`); any other value MUST be rejected before anything is written.
- WHEN `send` is invoked THE SYSTEM SHALL write the message to a temp file under `<root>/.tmp/` and rename it into `<root>/<provider>/<slug>/`, so a reader never observes a partial write.
- The delivered file MUST be named `<ts>__from_<sender-provider>_<sender-slug>__<msgid>.md`, where `ts` is UTC `YYYYMMDDTHHMMSS.mmmZ` and `msgid` is 8 lowercase hex characters; WHEN the name already exists THE SYSTEM SHALL regenerate `msgid`.
- Every message file MUST carry frontmatter `from`, `to`, `ts`, `msgid`, `type`, plus `thread` and `in_reply_to` only when given, in the bash field order.
- The body MUST be read from the positional file argument, or from stdin WHEN it is omitted or `-`.
- WHEN delivery succeeds THE SYSTEM SHALL print the delivered file path on stdout and exit zero.
- The mailbox root MUST be resolved via the existing `internal/broker.EnsureRoot` — do not duplicate that logic.

### Nudge

- WHEN the recipient provider is `codex` THE SYSTEM SHALL run `codex queue --thread <thread-id>` pointing the session at `inbox`, and SHALL skip silently WHEN `codex` is absent or the call fails; delivery MUST NOT depend on the nudge.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| T1 | Address parsing and slugify | — | pending |
| T2 | Session lookup by id per provider | T1 | pending |
| T3 | Message file write | T1 | pending |
| T4 | `send` CLI wiring | T2, T3 | pending |
| T5 | Codex nudge | T2, T4 | pending |

T2 and T3 may run in parallel; their `files:` do not overlap.

## Out of scope

- Any other subcommand (`inbox`, `hook`, `who`, ...), including how hooks learn their own session id (that comes from the hook payload, in `maintenance-go-hook`).
- Making `--thread` required or auto-filling it; it stays optional as in bash.
- Replacing the bash script or rewiring hooks to the Go binary.

## Risks

- **Bash/Go identity divergence.** Until cutover, bash still resolves identity from env and the working directory while Go does not; the same session can sign with different addresses depending on which binary runs. Cutover must be coordinated across all commands.
- **OpenCode "name".** The `session` table has both `title` and `slug`; which one is the human-facing name is unverified. T2 must confirm before coding.
- **Claude name source.** `claude agents --json` failed here (`too many arguments for 'agents'`, v2.1.280) and only covers background agents, so the bash `claude_session_name` likely resolves nothing for interactive sessions. The Go port reads the `custom-title` entries (latest per session wins) instead.
- **Claude ambiguity is heuristic.** Without an archive flag, an old Claude session titled `pikachu` still counts as active and can trigger the ambiguity refusal; the caller then addresses by id.
- **SQLite access.** Reading Codex and OpenCode stores needs either the `sqlite3` binary (as bash does) or a Go driver dependency; T2 decides, preferring no new dependency.
- **Spec change.** `specs/bridge/index.md` "Addressing" (provider set, resolution order, `AGENT_SLUG`) no longer matches; the delta draft must carry it.
