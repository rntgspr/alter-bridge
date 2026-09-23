---
human_revised: false
scope: [bridge]
status: in-progress
summary: Port the bash `relay` command to Go — the FileChanged entry point that logs every call, rings only for its own mailbox's `add` events, wakes a Claude session through a detached throwaway `claude -p` (transcript removed by exact path) or a Codex session through `codex queue`, and prints the `systemMessage` notice.
targets: [cli]
aux: []
---

# Go rewrite — `relay`

## Overview

Port `relay` (`skills/alter-bridge/scripts/alter-bridge`, lines ~329-397) to
Go. Bash reads the `FileChanged` payload on stdin, appends
`<HH:MM:SS UTC> <payload>` to `$ROOT/.tmp/relay.log` (never creating `.tmp`,
failures ignored), then returns 0 silently unless `.event` is `add`,
`.file_path` lies under `$ROOT/`, `dirname(file_path)` is this session's own
mailbox (`$ROOT/<provider>/<slug>`), and the basename matches
`*__from_*__*.md`. It derives `sender` (between the first `__from_` and the
next `__`, first `_` shown as `:`) and `msgid` (after the last `__`, minus
`.md`), spawns a detached throwaway `claude -p` (Haiku, low effort, only
`SendMessage` allowed, prompt on stdin, cwd `~/.claude`, named
`relay-<slugified sender>`, pre-chosen `--session-id`) that rings the session
named `<slug>`, deletes that relay's transcript
`~/.claude/projects/<proj>/<rid>.jsonl` by exact path when it exits, and
prints `{"systemMessage":"alter-bridge: new message from <sender> (msgid <msgid>)"}`.

### Identity (same model as the Go `hook` and `watchpaths`)

Per the `maintenance-go-send` / `-hook` / `-watchpaths` Decisions in
`specs/bridge/index.md`, no `self_addr` and no environment variable decides
identity. The one argument is a bare provider (value = the payload's first
non-empty `session_id` / `thread_id` / `threadId`) or an explicit
`provider:value`, resolved through `session.Resolver.Slug`. There is no
workspace gate (bash `relay` has none).

### Ring per provider

- `claude` — the detached throwaway `claude -p` above. The binary is found on
  `PATH`, else `~/.local/bin/claude` (the lookup `session.ClaudeAgents`
  already does, extracted as `session.ClaudeBin`); it is skipped when that
  path is not an executable file.
- `codex` — `nudge.Nudger.Notify` (the `send` nudge: `codex queue --thread
  <thread or slug>`), errors swallowed. Codex has no `FileChanged` hook today,
  so this only runs on an explicit `relay codex` call.
- `opencode` — no ring; the notice is still printed.

## Acceptance Criteria (EARS / RFC 2119)

- AC1 (log): WHEN `relay` is invoked with a valid argument THE SYSTEM SHALL append one line `<HH:MM:SS> <payload>` (UTC time, payload with trailing newlines stripped) to `<root>/.tmp/relay.log` before any other decision, creating the file but never `.tmp` or the root, and ignoring write failures.
- AC2 (own mailbox only): WHEN the event is not `add`, the `file_path` is not under `<root>/`, its directory is not `<root>/<provider>/<resolved slug>`, or its basename does not match `*__from_*__*.md`, THE SYSTEM SHALL print nothing, ring nothing, and exit 0; so SHALL an empty/invalid payload or a bare provider with no session id.
- AC3 (notice): WHEN an `add` event lands in the session's own mailbox THE SYSTEM SHALL print `{"systemMessage":"alter-bridge: new message from <sender> (msgid <msgid>)"}` plus a newline, compact and with HTML escaping off, byte-identical to bash `jq -nc`, and exit 0 whether or not the ring could start.
- AC4 (claude ring): WHEN the provider is `claude` and the Claude binary is an executable file THE SYSTEM SHALL start, detached from the hook (own session, stdio on `/dev/null`, not waited for), `<claude> -p --session-id <rid> --name relay-<slugified sender> --model claude-haiku-4-5-20251001 --effort low --allowedTools=SendMessage` in `~/.claude` with the prompt `Use the SendMessage tool to send the session named "<slug>" exactly this message: "<notice>". Do nothing else.` on stdin, and after it exits remove exactly `~/.claude/projects/<~/.claude with / and . as ->/<rid>.jsonl` and nothing else; `<rid>` is a fresh random UUID chosen before the spawn.
- AC5 (codex ring): WHEN the provider is `codex` THE SYSTEM SHALL nudge through `nudge.Nudger.Notify` (`codex queue --thread <live thread id, else slug> --message ...`), swallowing its failure, and SHALL NOT spawn Claude.
- AC6 (errors): WHEN the argument is missing, extra, `-h`/`--help`, or not a known provider / valid `provider:value` THE SYSTEM SHALL print usage or the address error on stderr, exit 2, and read or log nothing; WHEN slug resolution fails (ambiguous or empty) THE SYSTEM SHALL report it on stderr and exit 1. The root uses `broker.Resolve` (guard failure exits 1) and is never created.
- AC7 (smoke parity): with `HOME` an empty scratch dir, `claude` a fake builtins-only script on a temp `PATH`, and two copies of one `ALTER_BRIDGE_ROOT` fixture at the same path, bash `relay` and Go `relay claude` fed the same payloads produce identical stdout, stderr, exit code, `find . | sort`, and `relay.log` lines (time masked), and the fake `claude -p` argv/cwd/stdin (UUID masked), with the transcript removed and a sibling transcript kept, for: own-mailbox `add`, `change`, foreign mailbox, outside root, non-message basename.
- AC8 (live): the Go `relay`, wired by hand into a disposable live Claude session, rings that idle session when a message lands, and the throwaway relay leaves no transcript behind. **Manual only** (T3): automated tests cannot reach a real `FileChanged` watcher or the watcher-death class of bug.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| T1 | `session.ClaudeBin` + `nudge.RingClaude`: detached throwaway Claude ring with exact-path transcript cleanup | — | done |
| T2 | `relay` subcommand: log, gates, notice, per-provider ring, wiring | T1 | done |
| T3 | Live verification against a disposable Claude session (user-run) | T2 | pending |

## Out of scope

- Any other subcommand; rewiring the live hooks (`install-hooks.sh`, `settings.json`, `hooks.json`), the plugin manifests, or `bin/`. Cutover must pass the provider argument (`relay claude`).
- Detecting or recovering a dead `FileChanged` watcher; the log only makes it diagnosable.

## Risks

- Highest-risk port in the set: spawns a subprocess (`claude -p --session-id ...`), deletes another session's transcript by exact path, and bash has a documented flaky failure mode (the `FileChanged` watcher dying silently after hours, `alter-bridge:336-340`). Automated tests use fake binaries only; AC8 needs a user-run live check (T3) before this plan may close.
- **Detachment.** Go cannot leave a subshell behind like bash `( ... ) &`; the ring runs as `/bin/sh -c` in its own session (`Setsid`) with every argument passed positionally (no string interpolation), so it outlives the hook and never holds the hook's stdout open.
- **Identity deviation.** As in `hook`: no `AGENT_SLUG`, no cwd fallback; bash reads only `.session_id`.
- **UUID case.** macOS `uuidgen` prints uppercase; Go generates lowercase (as Claude's own ids). Same file on the default case-insensitive APFS; lowercase is the safe choice on case-sensitive volumes.
- **Codex double nudge.** If a Codex `relay` were ever wired, `send` already nudges Codex recipients, so the recipient would get two queue notices.
- **Logging order.** Go validates the argument before reading stdin, so a mis-wired call (bad argument) is reported on stderr instead of logged.
