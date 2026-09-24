---
human_revised: false
plan: maintenance-go-cutover
task: T2
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-cutover T2 — installer, SKILL.md and README moved to the Go binary with provider arguments; bash kept as fallback.
---

# Hand-off — maintenance-go-cutover / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`skills/alter-bridge/scripts/install-hooks.sh`](skills/alter-bridge/scripts/install-hooks.sh) | modified — wires `"<repo>/bin/alter-bridge" <sub> <provider>`, refuses a missing binary, matches bash and Go forms by subcommand |
| [`skills/alter-bridge/SKILL.md`](skills/alter-bridge/SKILL.md) | modified — Go binary as the command, `--from` / address usage, live name resolution, bash as fallback |
| [`README.md`](README.md) | modified — build step, wired commands, bash fallback, layout |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Commit `424b685`. The installer matches `^(?:bash )?"(.*)" (\S+)(?: \S+)?$`,
  so a legacy bash entry is FIXED in place, never duplicated.
- The binary path is the repo checkout's `bin/`, not the plugin cache; `bin/`
  stays gitignored, so `go/build.sh` is a documented install step.

## Commands run / verification

- Scratch `HOME` seeded with copies of the live (bash-form) files: run 1
  FIXED all four entries (one group per event, no duplicate), run 2 `left
  untouched`; stored strings keep `$HOME` unexpanded.
- Scratch checkout with no `bin/`: exit 1, `... is missing; run go/build.sh first`.
- `go test -count=1 ./...` — all 7 packages ok; `go/build.sh` built the binary.

## Pending / follow-ups

- None.

## Suggestions for the Lead

- None.
