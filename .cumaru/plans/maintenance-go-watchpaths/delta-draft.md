---
human_revised: false
plan: maintenance-go-watchpaths
status: draft
date: 2026-09-23
summary: Delta for maintenance-go-watchpaths — sharpen the SessionStart registration requirement, record the Go watchpaths identity model and deviations as a decision, and reference its sources.
---

# Delta draft — maintenance-go-watchpaths

The Go `watchpaths` is built and tested but not wired into hooks or skills.
The Doorbell's registration requirement omitted what both implementations do
(the leaf mailbox, the directory creation, the workspace gate, the exact
output), so it is sharpened; the Go identity model and deviations go under
Decisions.

## specs/bridge/index.md

### Added Requirements

- None.

### Modified Requirements

Under "Doorbell", the first requirement becomes:

- WHEN a Claude session starts inside `~/agentic-workspace` THE SYSTEM SHALL
  create the provider directory, this session's own mailbox, and
  `<root>/.tmp`, then print
  `{"hookSpecificOutput":{"hookEventName":"SessionStart","watchPaths":[<root>/<provider>, <root>/<provider>/<slug>]}}`
  as one compact JSON line, registering the whole provider directory (not
  just this agent's own mailbox) so a later `/rename` or a mailbox that does
  not exist yet still rings; outside `~/agentic-workspace` it SHALL print and
  create nothing.

### Removed Requirements

- None.

### Decisions (added)

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

### Files / Reference (added)

- Files list: `go/cmd/alter-bridge/` line gains `watchpaths`.
- Reference row: `go/cmd/alter-bridge/watchpaths.go` — `watchpaths`, the
  `SessionStart` entry point: provider argument, workspace gate, mailbox and
  `.tmp` creation, compact `watchPaths` JSON.
- Reference row: `go/cmd/alter-bridge/hookinput.go` — shared hook helpers:
  argument shape, payload session id, `~/agentic-workspace` gate.
