---
human_revised: false
plan: maintenance-go-hook
status: draft
date: 2026-09-23
summary: Delta for maintenance-go-hook — add the UserPromptSubmit hook requirement, record the Go hook's go-send identity model and deviations as a decision, and reference its sources.
---

# Delta draft — maintenance-go-hook

The Go `hook` is built and tested but not wired into hooks or skills. The
bash `hook` had no requirement of its own (only the Doorbell's "arrives on
the next turn via `UserPromptSubmit`"), so this delta adds one true of both
implementations; the Go identity model and deviations go under Decisions.

## specs/bridge/index.md

### Added Requirements

Under "Message delivery", after the `peek` requirement:

- WHEN the `UserPromptSubmit` hook runs THE SYSTEM SHALL read the session id
  (`session_id`, `thread_id`, or `threadId`) and `cwd` from the JSON payload
  on stdin, do nothing when that `cwd` (or, when it is unusable, the process
  working directory) is outside `~/agentic-workspace`, and otherwise archive
  and print every pending message of that session's mailbox exactly as
  `inbox` does, creating neither the root nor the mailbox.

### Modified Requirements

- None.

### Removed Requirements

- None.

### Decisions (added)

- 2026-09 (`maintenance-go-hook`): the Go `hook` is built but not yet wired
  into hooks or the skill. It follows the go-send identity model instead of
  `self_addr`: the argument is required, either a bare provider
  (`hook claude`, value taken from the payload session id) or an explicit
  `provider:value`, both resolved through the `send`/`inbox` session resolver
  and drained with `mailbox.Drain`. No environment variable picks the
  provider or slug, and a payload without a session id drains nothing (exit
  0) instead of falling back to the cwd segment, so an unnamed session's
  mailbox is keyed by its id. The root comes from `broker.Resolve` (never
  created). Stdout and the resulting tree are byte-identical to bash for an
  explicit address inside and outside the workspace, a missing mailbox, and
  an absent root. Payload problems (empty, not JSON, non-string fields) are
  silent exit 0, so the prompt is never blocked; deviations: a missing,
  extra, `-h`/`--help`, unknown-provider, or malformed argument exits 2 even
  outside the workspace (bash checks the shape only, after the gate), and
  ambiguous or empty slugs exit 1. Cutover must pass the provider in the
  hook wiring.

### Files / Reference (added)

- Files list: `go/cmd/alter-bridge/` line gains `hook`; `address.go` line
  mentions the provider check.
- Reference row: `go/cmd/alter-bridge/hook.go` — `hook`, the
  `UserPromptSubmit` entry point: provider argument, payload session id and
  `cwd`, workspace gate, archiving drain.
- Reference row for `go/internal/address/address.go` mentions `IsProvider`.
