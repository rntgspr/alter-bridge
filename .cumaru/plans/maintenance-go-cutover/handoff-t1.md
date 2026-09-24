---
human_revised: false
plan: maintenance-go-cutover
task: T1
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-cutover T1 — identity gate failed on auto-named Claude sessions, then passed for every live Claude and Codex session after maintenance-go-agent-names.
---

# Hand-off — maintenance-go-cutover / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`index.md`](index.md) | modified — Blocker recorded, then marked resolved (option A) |
<!-- /cumaru:touched -->

## Decisions made during implementation

- First run failed AC6 (auto-named Claude sessions resolved to their id in
  Go). Renato chose option A; fixed by `maintenance-go-agent-names`.

## Commands run / verification

- Fresh re-run after `go/build.sh`: for every session `bin/alter-bridge who`
  lists, the same SessionStart payload into bash `watchpaths` and Go
  `watchpaths <provider>` against scratch roots:
  - `6e4c01e3-...` bash=`claude/vault-79` go=`claude/vault-79` SAME
  - `4ff25905-...` bash=`claude/alter-bridge-11` go=`claude/alter-bridge-11` SAME
  - Codex `01a0d13b`, `01a0b1dd`, `01a0b72c`, `01a0c282`: bash and Go both
    `codex/bira-probe`, `codex/assumir-papel-de-lead`, `codex/cumaru`,
    `codex/explore-o-reposit-rio`.
- `go test -count=1 ./...` — all 7 packages ok.

## Pending / follow-ups

- Auto-generated Claude names change on resume (`alter-bridge-ef` became
  `alter-bridge-11` for the same id); `/rename` is the stable fix, as in bash.

## Suggestions for the Lead

- None.
