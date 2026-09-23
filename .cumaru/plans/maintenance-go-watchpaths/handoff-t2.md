---
human_revised: false
plan: maintenance-go-watchpaths
task: T2
status: complete
date: 2026-09-23
summary: Hand-off for maintenance-go-watchpaths T2 — Go watchpaths takes a provider or provider:value argument, gates on ~/agentic-workspace, creates the mailbox and .tmp after the gate, and prints the SessionStart watchPaths line byte-identical to bash.
---

# Hand-off — maintenance-go-watchpaths / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/cmd/alter-bridge/watchpaths.go`](go/cmd/alter-bridge/watchpaths.go) | created — `watchpathsUsage`, `watchpathsOutput`, `runWatchpaths` (argument check, `hookAddress`, `hookSession` gate, `Resolver.Slug`, `MkdirAll` of the mailbox and `.tmp`, compact JSON with HTML escaping off) |
| [`go/cmd/alter-bridge/watchpaths_test.go`](go/cmd/alter-bridge/watchpaths_test.go) | created — in-process tests: id from `session_id`/`thread_id`/`threadId`, existing mailbox kept, explicit address ignores payload id, HTML characters unescaped, silent no-ops create no root, usage refusals exit 2, ambiguous name and uncreatable root exit 1 |
| [`go/cmd/alter-bridge/main.go`](go/cmd/alter-bridge/main.go) | modified — `watchpaths` case with `broker.Resolve`, `session.NewResolver`, `os.Getwd`, reusing `hookEnv`, and a usage line |
<!-- /cumaru:touched -->

## Decisions made during implementation

- Root from `broker.Resolve`, not `setup()`/`EnsureRoot`: those create the root before the gate, while bash creates nothing outside the workspace. `MkdirAll` of the mailbox creates the root after the gate, as bash's `mkdir -p` does; the `HOME` guard still applies (exit 1).
- `json.Encoder` with `SetEscapeHTML(false)` so `&`, `<`, `>` in paths match `jq -nc` (the test was seen failing with escaping on); `Encode` supplies jq's trailing newline.
- Same identity model and exit codes as `hook`; `hookEnv` is reused rather than a second identical struct.
- Deviation: bash reads only `.session_id` here; Go reads `session_id` / `thread_id` / `threadId` through the shared helper.

## Commands run / verification

- `go test -count=1 ./cmd/alter-bridge/` before implementation — build failed on undefined `runWatchpaths` (expected red); HTML test failed with escaping on, then passed.
- `go test -count=1 ./...` — all ok; `go vet ./...` clean; `gofmt -l .` empty.
- Smoke (scratch `HOME` with `agentic-workspace/papa`, fixture copied to the same `ALTER_BRIDGE_ROOT` path per run, bash vs Go `watchpaths <addr>`, stdout + stderr + exit + `find . | sort` diffed): existing mailbox, absent root, new mailbox — identical JSON line and trees; `cwd` outside with and without a root — both silent, exit 0, nothing created.
- Go-only smoke: `watchpaths claude` with a scratch transcript `sid-1.jsonl` titled `Papa Work` registered `claude/papa-work`; garbage payload and a payload with no id exit 0 and create no root; no argument prints usage, exit 2; `HOME=""` exits 1 on the guard.

## Pending / follow-ups

- Cutover must pass the provider (`watchpaths claude`) in the SessionStart wiring.

## Suggestions for the Lead

- None.
