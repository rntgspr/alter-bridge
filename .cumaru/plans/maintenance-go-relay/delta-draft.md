---
human_revised: false
plan: maintenance-go-relay
status: draft
date: 2026-09-23
summary: Delta for maintenance-go-relay — sharpen the FileChanged relay requirement, record the Go relay identity model, detachment, and deviations as a decision, and reference its sources. Absorb only after AC8 (T3) is PASS.
---

# Delta draft — maintenance-go-relay

The Go `relay` is built and tested (fake binaries, bash smoke parity) but not
wired into hooks or skills. **Not absorbable yet:** AC8 (T3, user-run live
verification) is Blocked. The relay requirement omitted the own-mailbox
check, the message-name gate, and the exact notice, so it is sharpened; the
Go identity model, detachment, and deviations go under Decisions.

## specs/bridge/index.md

### Added Requirements

- None.

### Modified Requirements

Under "Doorbell", the second and third requirements become:

- WHEN a filesystem `add` event lands directly inside a session's own
  mailbox (`<root>/<provider>/<slug>`) for a file named
  `*__from_<provider>_<slug>__<msgid>.md` THE SYSTEM SHALL wake that session
  — a Claude session through a detached throwaway headless `claude -p`
  (Haiku, low effort, `SendMessage` only, run from `~/.claude`, pre-chosen
  session id) that deletes its own transcript by exact path on exit — and
  print the notice; `change` events, other mailboxes, files outside the
  bridge, and other file names SHALL be ignored silently.
- The relay notice MUST reach the human as the one compact JSON line
  `{"systemMessage":"alter-bridge: new message from <provider>:<slug> (msgid <msgid>)"}`,
  since `FileChanged` runs outside the REPL and has no model context channel;
  the message content itself SHALL still arrive on the target's next turn via
  `UserPromptSubmit`.

### Removed Requirements

- None.

### Decisions (added)

- 2026-09 (`maintenance-go-relay`): the Go `relay` is built but not yet wired
  into hooks or the skill. It shares `hook`'s argument helper and identity
  model (a bare provider takes the payload's first non-empty `session_id` /
  `thread_id` / `threadId`; bash reads `session_id` only; or an explicit
  `provider:value`, resolved through the session resolver) and, like bash,
  has no workspace gate. After the argument check it appends
  `<HH:MM:SS UTC> <payload>` to `<root>/.tmp/relay.log` on every call, never
  creating `.tmp` or the root (`broker.Resolve`). The Claude ring is
  `nudge.RingClaude`: `/bin/sh -c` with every value positional, in its own
  session (`Setsid`), stdio on `/dev/null`, never waited for, so it outlives
  the hook as bash's `( ... ) &` does; the binary comes from
  `session.ClaudeBin` (PATH, else `~/.local/bin/claude`, shared with
  `ClaudeAgents`). A `codex` mailbox is nudged with `nudge.Nudger` (`codex
  queue`) instead; `opencode` only gets the notice. Stdout, stderr, exit
  code, tree, `relay.log`, and the spawned `claude` argv/cwd/stdin are
  byte-identical to bash for an own `add`, `change`, a foreign mailbox, a
  file outside the root, and a non-message name. Deviations: argument errors
  exit 2 before anything is read or logged; ambiguous or empty slugs exit 1;
  the relay session id is a lowercase v4 UUID (macOS `uuidgen` prints
  uppercase); a missing `.tmp` is silent (bash leaks a redirection error on
  stderr). Live ringing and the watcher-death failure mode were verified by
  hand against a disposable session (T3). Cutover must pass the provider in
  the `FileChanged` wiring.

### Files / Reference (added)

- Files list: `go/cmd/alter-bridge/` line gains `relay`; `go/internal/nudge/nudge.go` line becomes "best-effort `codex queue` nudge and the throwaway Claude relay ring".
- Reference row: `go/cmd/alter-bridge/relay.go` — `relay`, the `FileChanged` entry point: provider argument, `relay.log` append, own-mailbox `add` gate, message-name parsing, per-provider ring, compact `systemMessage` JSON.
- Reference row: `go/internal/nudge/claude.go` — `RingClaude`: detached throwaway `claude -p` relay with exact-path transcript cleanup.
- Reference row `go/internal/session/agents.go` gains `ClaudeBin`, the shared Claude CLI lookup.
