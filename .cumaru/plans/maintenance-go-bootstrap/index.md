---
human_revised: false
scope: []
status: done
summary: Bootstrap a minimal Go module producing a "hello world" alter-bridge binary, ahead of a future Go rewrite of the bash CLI.
targets: [cli]
aux: []
---

# Bootstrap Go module (hello world)

## Overview

Scaffold a minimal Go module in this repo that builds and runs a "hello
world" executable. This is the first step toward a future full Go rewrite
of the bash CLI documented in `specs/bridge`, but this plan's scope stops
at the scaffold — no bridge behavior is implemented here, so `scope` is
empty and no spec delta is expected on close. Binary name follows the
`alter-bridge` rename landed by `maintenance-rename-alter-bridge`.

## Acceptance Criteria (EARS / RFC 2119)

- A `go.mod` file MUST exist under `go/` declaring the module (the Go
  module root is `go/`, not the repository root, to keep future
  non-Go apps — Rust, Node — from mixing into the same directory).
- `go/cmd/alter-bridge/main.go` MUST print `hello world` to stdout when run via `go run`.
- Running `./go/build.sh` MUST produce an executable at `bin/alter-bridge` (repo root).
- Running `./bin/alter-bridge` from the repo root MUST print `hello world` and exit `0`.
- `bin/` MUST be excluded from version control via `.gitignore`.

## Plan / DAG

| Task | Title | Status | Depends on |
|------|-------|--------|-----------|
| [T1](t1.md) | Init go.mod and hello-world main | done | — |
| [T2](t2.md) | Build to bin/ and gitignore it | done | T1 |

## Out of scope

- Any actual bridge CLI logic (addressing, send/inbox, hooks) — that is the future rewrite plan, not this bootstrap.
- Wiring `install-hooks.sh` or any hook config to the new binary.
- CI/build automation beyond a manual `go build` command.

## Risks

- Local Go toolchain confirmed present (`go1.27.0 darwin/arm64`) — no install risk on this machine.
