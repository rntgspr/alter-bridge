---
human_revised: false
plan: maintenance-plugin-release
task: T5
status: done
date: 2026-09-24
summary: Hand-off for maintenance-plugin-release T5 - bash script removed from the skill folder, Codex nudge notice repointed at bin/alter-bridge.
---

# Handoff T5

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`skills/alter-bridge/scripts/alter-bridge`](skills/alter-bridge/scripts/alter-bridge) | deleted - unwired bash fallback |
| [`go/internal/nudge/nudge.go`](go/internal/nudge/nudge.go) | modified - BridgeScript points at bin/alter-bridge |
| [`go/internal/nudge/nudge_test.go`](go/internal/nudge/nudge_test.go) | modified - notice text pinned to the literal path |
| [`go/internal/address/address_test.go`](go/internal/address/address_test.go) | modified - comment points at git history |
| [`README.md`](README.md) | modified - fallback mentions removed |
| [`skills/alter-bridge/SKILL.md`](skills/alter-bridge/SKILL.md) | modified - fallback mention removed |
<!-- /cumaru:touched -->

## Evidence

- Red: `go test ./internal/nudge/` failed, with a got/want mismatch on the
  old script path.
- Green: `go vet ./...` clean, `gofmt -l .` empty, `go test -count=1 ./...`
  all packages ok.
- `skills/alter-bridge/` now contains only `SKILL.md`. `go/release.sh`
  copies `skills/` as a whole, so the zip ships just `SKILL.md`.

## For the delta

`specs/bridge/index.md` still lists the bash script and `install-hooks.sh`
in Files and Reference (lines ~128, 289, 291, 309, 311). Remove those rows at
absorb.
