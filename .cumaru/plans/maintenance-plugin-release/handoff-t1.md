---
human_revised: false
plan: maintenance-plugin-release
task: T1
status: complete
date: 2026-09-24
summary: Hand-off for maintenance-plugin-release T1 - go/release.sh builder, go/alter-bridge host launcher and its shell test, gitignored release outputs.
---

# Hand-off — maintenance-plugin-release / T1

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`go/alter-bridge`](go/alter-bridge) | created - POSIX sh launcher; maps `uname -s`/`uname -m` to `<os>-<arch>` and execs `go/alter-bridge-<os>-<arch>`; no match prints one stderr line and exits 0 |
| [`go/launcher_test.sh`](go/launcher_test.sh) | created - shell test with a fake `uname` on `PATH` and stub binaries (4 targets, unknown arch, unknown OS) |
| [`go/release.sh`](go/release.sh) | created - cross-compiles every target in the `TARGETS` list at its top, stages and zips the plugin tree into `dist/alter-bridge-<version>.zip`, prints the sha256 |
| [`.gitignore`](.gitignore) | modified - ignores `go/alter-bridge-*` and `dist/` |
<!-- /cumaru:touched -->

## Decisions made during implementation

- The zip is built from a temp staging directory, not from the repo root, so
  the staged `.claude-plugin/plugin.json` gets `version` stamped from the
  release version (leading `v` dropped). The Claude docs say a declared
  `version` is the update signal for an `archive` source: without the stamp,
  every release would stay `1.0.0` and users would keep the cached copy
  ([plugin-marketplaces](https://code.claude.com/docs/en/plugin-marketplaces)).
  The checked-in `plugin.json` is untouched.
- `release.sh` takes an optional version (default `0.0.0-dev`), and it needs `jq`.
- The launcher test is a separate shell file, `go/launcher_test.sh`, which the
  task's `files:` did not list. `go test` does not cover it.
- The targets are a single bash array, `TARGETS=(...)`, at the top of
  `release.sh`. Changing that one line changes both the build and the packing.
- `zip` keeps the Unix modes (`-rwxr-xr-x` on the launcher and the binaries,
  per `unzip -Z`). The docs do not say whether Claude Code keeps them when it
  extracts, so T4 has to check this.

## Commands run / verification

- `go/launcher_test.sh` against an empty launcher: 6 FAIL, exit 1 (red). After
  the launcher was written: 6 ok, exit 0 (green). The cases are darwin-arm64,
  darwin-amd64, linux-amd64, linux-arm64 (`aarch64`), unknown arch
  (`riscv64`), and unknown OS (`FreeBSD`). Each unknown case exits 0 with
  exactly one stderr line.
- `go/release.sh v0.1.0` built 4 binaries (`file` shows 2 Mach-O and 2
  static ELF). It wrote `dist/alter-bridge-v0.1.0.zip`; `unzip -l` lists
  `.claude-plugin/plugin.json`, `hooks/hooks.json`, `skills/alter-bridge/**`,
  `go/alter-bridge`, and `go/alter-bridge-{darwin,linux}-{arm64,amd64}`. The
  manifest version inside the zip is `"0.1.0"`.
- `go test ./...` in `go/`: all 7 packages ok.
- The real launcher on this host (darwin-arm64), with a scratch
  `ALTER_BRIDGE_ROOT`, ran `peek claude:t1probe` and exited 0.

## Pending / follow-ups

- `hooks/hooks.json` still points at `bin/`. T3 fixes this.

## Suggestions for the Lead

- Running `go build ./cmd/alter-bridge` inside `go/` without `-o` writes
  `go/alter-bridge` and overwrites the launcher. `go/build.sh` and
  `release.sh` both pass `-o`.
