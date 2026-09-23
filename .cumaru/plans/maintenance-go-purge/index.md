---
human_revised: false
scope: [bridge]
status: in-progress
summary: Port the bash `purge` command to Go — permanently delete the archived trail (`*.md` directly under `.archive/`), without creating the root or prompting.
targets: [cli]
aux: []
---

# Go rewrite — `purge`

## Overview

Port `purge` (`skills/alter-bridge/scripts/alter-bridge`, lines ~423-434) to
Go. Bash deletes every `*.md` directly under `<root>/.archive` (glob, so no
dot names, no recursion, no other types), prints
`purged <n> archived message(s) from <root>/.archive`, and prints
`nothing to purge` when `.archive/` is not a directory. It takes no
arguments, never prompts, and never creates the mailbox root.

`purge` has no formal requirement in `specs/bridge/index.md` (same spec gap
`archive` had); absorption closes it with a Message delivery requirement,
not only a Decision.

The delete lives next to `Archive` in `internal/mailbox` (`mailbox.Purge`),
which owns the `.archive/` layout. To match bash and `who`, `purge` must not
create the root, so root resolution and its HOME guard are split out of
`broker.EnsureRoot` into `broker.Resolve` (no mkdir) instead of duplicated.

## Acceptance Criteria (EARS / RFC 2119)

Derived from the bash `purge`:

- AC1: WHEN `<root>/.archive` does not exist or is not a directory THE SYSTEM SHALL print `nothing to purge`, delete nothing, and exit 0.
- AC2: WHEN `<root>/.archive` is a directory THE SYSTEM SHALL delete every non-dot entry named `*.md` directly under it that resolves to an existing non-directory (regular file or symlink to one; a symlink is removed, not its target), leave subdirectories, their contents, dot-named files, broken symlinks, and other file types untouched, print `purged <n> archived message(s) from <root>/.archive`, and exit 0 — `n` may be 0.
- AC3: `purge` MUST NOT prompt for confirmation (bash parity) and MUST NOT create the mailbox root; the root is resolved with the same `ALTER_BRIDGE_ROOT` / `HOME` rules and HOME guard as every other subcommand (guard failure exits 1).
- AC4: WHEN any argument is given (including `-h`/`--help`) THE SYSTEM SHALL print usage on stderr, delete nothing, and exit 2; WHEN a deletion fails THE SYSTEM SHALL report it on stderr and exit 1.
- AC5: Smoke parity: on two copies of the same scratch `ALTER_BRIDGE_ROOT` (archived `*.md`, a dot `.md`, a non-`.md` file, a nested directory with `.md` inside), Go and bash `purge` produce identical stdout and identical `find . | sort` listings; with `.archive/` absent, and with the root itself absent, both print `nothing to purge` and neither creates anything.

## Plan / DAG

| Task | Title | Depends on | Status |
|---|---|---|---|
| T1 | `mailbox.Purge` and `mailbox.ErrNoArchive` | — | pending |
| T2 | `broker.Resolve` and `purge` CLI wiring | T1 | pending |

## Out of scope

- Any other subcommand; replacing the bash script or rewiring hooks/skill.
- A confirmation prompt or dry-run flag (not in bash; only on user request).

## Risks

- Irreversible delete — same risk profile as the bash script, not introduced by this port.
- **Directory named `*.md` under `.archive/`.** Bash `rm -f` fails on it and `set -e` aborts with exit 1 after a partial purge; Go skips it (same deviation `inbox` recorded for mailboxes). Excluded from the smoke fixture.
- **Printed path.** Bash echoes `$ROOT/.archive` verbatim; Go uses `filepath.Join`, which cleans it (a trailing `/` in `ALTER_BRIDGE_ROOT` prints once). Identical for clean paths.
