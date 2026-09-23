---
human_revised: false
scope: [bridge]
status: in-progress
summary: Port the bash `watchpaths` command to Go — the SessionStart entry point that reads its session from the stdin payload, stays inside `~/agentic-workspace`, creates the provider directory and its own mailbox, and registers both with the FileChanged watcher.
targets: [cli]
aux: []
---

# Go rewrite — `watchpaths`

## Overview

Port `watchpaths` (`skills/alter-bridge/scripts/alter-bridge`, lines ~297-326)
to Go. Bash reads the harness payload on stdin, takes `cwd` from `.cwd` and
the session id from `.session_id`, `cd`s into `cwd` when it can, returns 0
silently unless the working directory is `$HOME/agentic-workspace` or below,
then resolves the address (explicit argument, else `self_addr "$sid"`),
runs `mkdir -p "$ROOT/<provider>/<slug>" "$ROOT/.tmp"` and prints
`jq -nc '{hookSpecificOutput:{hookEventName:"SessionStart",watchPaths:[$p,$m]}}'`
with `$p = $ROOT/<provider>` and `$m = $p/<slug>`. Outside the workspace it
creates nothing.

### Identity (same model as the Go `hook`)

Per the `maintenance-go-send` and `maintenance-go-hook` Decisions in
`specs/bridge/index.md`, no `self_addr` and no environment variable decides
identity. The one argument is a bare provider (value = the payload's first
non-empty `session_id` / `thread_id` / `threadId`) or an explicit
`provider:value`, both resolved through `session.Resolver.Slug`. A bare
provider with no session id registers nothing and exits 0.

### Shared hook input

`watchpaths` is the second caller of `hook`'s argument parsing, payload
decoding, and `~/agentic-workspace` gate, so T1 lifts that code out of
`runHook` into shared helpers (behavior unchanged, hook tests stay green)
before T2 builds on it.

## Acceptance Criteria (EARS / RFC 2119)

- AC1: WHEN invoked as `watchpaths <provider>` inside the workspace THE SYSTEM SHALL take the session id as the first non-empty string of `session_id`, `thread_id`, `threadId`, resolve it with `session.Resolver.Slug(<provider>, <id>)`, create `<root>/<provider>/<slug>` and `<root>/.tmp` (and the root when missing), and print `{"hookSpecificOutput":{"hookEventName":"SessionStart","watchPaths":["<root>/<provider>","<root>/<provider>/<slug>"]}}` followed by a newline, byte-identical to bash `jq -nc` for the same paths.
- AC2: WHEN invoked as `watchpaths <provider>:<value>` THE SYSTEM SHALL register that address, resolved through the same resolver, ignoring the payload's session id.
- AC3: WHEN the payload's `cwd` (or, when it is not an existing directory, the process working directory) is neither `$HOME/agentic-workspace` nor below it THE SYSTEM SHALL print nothing, create nothing, and exit 0.
- AC4: WHEN the payload is empty, not JSON, or carries no session id and the argument is a bare provider THE SYSTEM SHALL print nothing, create nothing, and exit 0, so a malformed payload never breaks session start.
- AC5: WHEN the argument is missing, extra, `-h`/`--help`, or not a known provider / valid `provider:value` THE SYSTEM SHALL print usage or the address error on stderr and exit 2, creating nothing; WHEN resolution or directory creation fails THE SYSTEM SHALL report it on stderr and exit 1. The root uses the shared `ALTER_BRIDGE_ROOT` / `HOME` rules and guard (guard failure exits 1).
- AC6: `hook`'s behavior is unchanged by the shared-helper extraction: every existing `hook` test passes unmodified.
- AC7: Smoke parity: with `HOME` an empty scratch dir holding a fake `agentic-workspace/<agent>` cwd and two copies of the same `ALTER_BRIDGE_ROOT` fixture at the same path, bash `watchpaths claude:<slug>` and Go `watchpaths claude:<slug>` fed the same payload produce identical stdout and `find . | sort` (mailbox already present, and absent root); with `cwd` outside the workspace both print nothing and change nothing.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| T1 | Extract shared hook argument, payload, and workspace-gate helpers | — | pending |
| T2 | `watchpaths` subcommand: payload, gate, mkdir, watchPaths JSON | T1 | pending |

## Out of scope

- `relay` (separate plan; the `FileChanged` event this registration triggers).
- Any other subcommand; rewiring the live hooks (`install-hooks.sh`, `settings.json`, `hooks.json`), the plugin manifests, or `bin/`. Cutover must pass the provider argument.

## Risks

- **Identity deviation.** As in `hook`: no `AGENT_SLUG`, no cwd-segment fallback; an unnamed session registers its id-keyed mailbox. Bash reads only `.session_id` here; Go reads the same three keys as `hook`.
- **Path form.** Go builds paths with `filepath.Join`, so an `ALTER_BRIDGE_ROOT` with a trailing slash or `..` prints cleaned paths where bash concatenates the raw string.
