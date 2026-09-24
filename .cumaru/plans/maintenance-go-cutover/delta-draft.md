---
human_revised: false
plan: maintenance-go-cutover
status: draft
date: 2026-09-23
summary: Delta draft for maintenance-go-cutover — Go binary becomes the canonical, hook-wired bridge CLI; bash stays as fallback; installer and Files/Reference rows follow.
---

# Delta draft — maintenance-go-cutover

## specs/bridge/index.md

### Modified Requirements

- Addressing: providers are `claude`, `codex`, `opencode`; identity comes
  from an explicit `provider:value` (hook payload session id, or `--from`),
  never from `AGENT_SLUG` or the working directory
  (was: `AGENT_SLUG` -> session name -> `agentic-workspace/` segment).
- Plugin packaging and install: `install-hooks.sh` MUST wire the Go binary
  `bin/alter-bridge` with the provider argument (`watchpaths claude`,
  `hook claude`, `relay claude` with matcher `.*__from_.*`, `hook codex`),
  refuse to run when the binary is missing, and reconcile an existing entry
  for the same subcommand in either the bash or the Go form in place.

### Decisions

- Add `maintenance-go-cutover`: Go is canonical; the live hooks and the skill
  call `bin/alter-bridge` (built by `go/build.sh`, gitignored); the bash
  script is a kept fallback. Residual risks: automatic Claude names change on
  resume (`/rename` is the stable fix, as in bash); the changed Codex hook
  needs trust re-approval inside Codex; the Codex payload id fields are
  unverified against a live Codex turn.
- Supersede the "script lives under `skills/`, not `bin/`" decision and the
  "built but not yet wired" wording of the per-command Go decisions.

### Files / Reference

- The bash script row becomes "fallback"; add `bin/alter-bridge`
  (via `go/build.sh`) and `go/cmd/alter-bridge/relay.go` rows where missing.
