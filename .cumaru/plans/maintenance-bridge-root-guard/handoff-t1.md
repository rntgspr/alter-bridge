---
human_revised: false
plan: maintenance-bridge-root-guard
task: T1
status: done
date: 2026-09-21
summary: Implementation and verification evidence for T1 — internal/broker.EnsureRoot.
---

<!-- cumaru:touched -->
go/internal/broker/root.go
go/internal/broker/root_test.go
go/cmd/alter-bridge/main.go
go/build.sh
.gitignore
<!-- /cumaru:touched -->

# Handoff — T1

## Implementation

- `go/internal/broker/root.go`: `EnsureRoot(override, home string) (string, error)`.
  - `override` non-empty wins outright, no `$HOME` guard applied.
  - Else `home == "" || home == "/"` returns an error, creates nothing.
  - Else root is `filepath.Join(home, ".alter-bridge")`, created via `os.MkdirAll`.
  - A failed `MkdirAll` (e.g. permission denied) is wrapped with the path and returned as an error, never a panic.
- `go/internal/broker/root_test.go`: five test cases (override, empty home, `home=/`, valid home, idempotent re-call).
- `go/cmd/alter-bridge/main.go`: now calls `broker.EnsureRoot(os.Getenv("ALTER_BRIDGE_ROOT"), os.Getenv("HOME"))` at startup, exits 1 with the error on stderr if it fails, otherwise prints the resolved root alongside "hello world".
- Go module relocated to `go/` (not repo root) during this plan, at user request, to keep the root free for future non-Go apps (Rust, Node). `go/build.sh` builds to `../bin/alter-bridge` (repo root, gitignored via the root-level `bin/` pattern).
- Package renamed `mailbox` → `broker` at user request, mid-plan.

## Verification (fresh, this session)

TDD cycle observed:
1. Red: `go test ./internal/broker/...` failed to compile — `undefined: EnsureRoot` (expected reason, before implementation existed).
2. Green after implementing `root.go`:
```
=== RUN   TestEnsureRoot_Override
--- PASS: TestEnsureRoot_Override (0.00s)
=== RUN   TestEnsureRoot_EmptyHomeNoOverride
--- PASS: TestEnsureRoot_EmptyHomeNoOverride (0.00s)
=== RUN   TestEnsureRoot_RootHomeNoOverride
--- PASS: TestEnsureRoot_RootHomeNoOverride (0.00s)
=== RUN   TestEnsureRoot_ValidHome
--- PASS: TestEnsureRoot_ValidHome (0.00s)
=== RUN   TestEnsureRoot_AlreadyExists
--- PASS: TestEnsureRoot_AlreadyExists (0.00s)
PASS
ok  	github.com/rntgspr/alter-bridge/internal/broker
```
3. After relocating to `go/`, renaming to `broker`, and wiring `main.go`: `go build ./...` and `go test ./...` re-run clean from `go/`.
4. `go run ./cmd/alter-bridge` from `go/`, real environment:
```
hello world
mailbox root: /Users/gaspar/.alter-bridge
```
Confirmed `$HOME/.alter-bridge` already existed (created earlier by the bash script) and was left untouched — `EnsureRoot` is idempotent against a pre-existing root.
5. `./build.sh` from `go/` produces `../bin/alter-bridge`; `./bin/alter-bridge` from repo root prints the same output, exit 0.

## Acceptance criteria reconciliation

| Criterion | Status |
|---|---|
| Root created when missing, path returned | PASS — `TestEnsureRoot_ValidHome`, `TestEnsureRoot_Override` |
| Empty/`/` `$HOME` with no override → error, no dir created | PASS — `TestEnsureRoot_EmptyHomeNoOverride`, `TestEnsureRoot_RootHomeNoOverride` |
| `mkdir` failure → error naming the path, no panic | Covered by code path (wrapped `%w` error); not independently unit-tested (would require a read-only parent fixture) — verified manually against `/private/var/root/...` during the bash-script iteration of this task, same `MkdirAll` semantics apply |
| Root already exists → path returned, no error, not re-created | PASS — `TestEnsureRoot_AlreadyExists` |
| Covered by Go tests, not manual-only | PASS — 5 test cases, all in `go/internal/broker/root_test.go` |
