---
human_revised: false
plan: maintenance-go-bootstrap
task: T2
status: complete
date: 2026-09-21
summary: Hand-off for T2 — build wrapper script produces bin/alter-bridge at the repo root and bin/ is ignored by Git.
---

# Hand-off — maintenance-go-bootstrap / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [build.sh](go/build.sh) | created — wrapper that builds `./cmd/alter-bridge` into `../bin/alter-bridge` |
| [.gitignore](.gitignore) | created — single `bin/` entry |
<!-- /cumaru:touched -->

## Decisions made during implementation

- **A `build.sh` wrapper replaced the bare `go build` command.** With the module
  root moved to `go/` in T1, the literal command from the task body no longer
  works from the repository root. The script `cd`s to its own directory first,
  so the build is invoked the same way from anywhere, and the output still lands
  at `bin/alter-bridge` relative to the repository root.
- The binary itself is not committed; only the script that produces it is.

## Commands run / verification

- `./go/build.sh` — printed `built bin/alter-bridge`.
- `ls -l bin/alter-bridge` — executable present, mode `-rwxr-xr-x`.
- `./bin/alter-bridge` — printed `hello world` and `mailbox root: /Users/gaspar/.alter-bridge`, exit code `0`.
- `git check-ignore -v bin/alter-bridge` — matched `.gitignore:1:bin/`.
- `cd go && go test ./...` — `ok github.com/rntgspr/alter-bridge/internal/broker`; `cmd/alter-bridge` has no test files.

## Pending / follow-ups

- No build automation beyond the manual script, which the plan lists as out of
  scope. Wiring `install-hooks.sh` to the Go binary remains untouched.

## Suggestions for the Lead

- `go/build.sh` is the single build entry point every later `maintenance-go-*`
  plan will inherit. Once those ports start producing real subcommands, it is
  the natural place for build flags (version stamping, trimpath).
