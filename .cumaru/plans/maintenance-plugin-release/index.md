---
human_revised: false
scope: [bridge]
status: in-progress
summary: Ship the Claude Code plugin from a GitHub Release archive built by a root workflow that runs the builder in go/, so the standard marketplace install works without a local checkout, Go toolchain, or install.sh.
targets: [plugin, hooks]
aux: []
---

# Claude plugin shipped from a GitHub Release archive

## Overview

`maintenance-plugin-standard` made the repository a standard Claude Code
marketplace, but only a local checkout works. `bin/alter-bridge` is gitignored,
so a marketplace added from GitHub (`/plugin marketplace add
rntgspr/alter-bridge`) clones a tree with no binary, and the bundled hooks fail.
`install.sh` only exists to build that binary first.

This plan closes the gap using the release channel Claude Code already
supports. A GitHub Actions workflow at the repository root runs a builder that
lives in `go/`. The builder cross-compiles the CLI into `go/`, packs the whole
plugin tree into a zip and attaches the zip to the GitHub Release. The
marketplace entry then points at that zip through the `archive` source, which
takes an HTTPS zip `url` plus an optional `sha256`. Claude Code has no
dedicated GitHub Release source type, but a release download URL is a valid
`archive` url
([plugin-marketplaces](https://code.claude.com/docs/en/plugin-marketplaces)).

Decisions from Renato (2026-09-24):

- **Workflow location:** the workflow lives where GitHub requires it,
  `.github/workflows/` at the root
  ([about workflows](https://docs.github.com/en/actions/writing-workflows/about-workflows)).
- **Builder and binaries:** the builder runs inside `go/`, and the release
  binaries land in `go/`.
- **Codex is out of scope:** Renato designs the Codex side separately. Nothing
  here may break the current Codex wiring (`hooks/codex.json`,
  `.codex-plugin/plugin.json`, `bin/alter-bridge`).

Relationship to `maintenance-plugin-standard`: its T3 (live cutover) is
blocked. This plan's T4 does the Claude half of that cutover through the
release-based install. Reconcile T3 before either plan is absorbed.

## Acceptance Criteria (EARS / RFC 2119)

- AC1: WHEN `go/release.sh` runs THE SYSTEM SHALL write one binary per target
  platform into `go/` and a zip of the plugin tree whose root holds
  `.claude-plugin/plugin.json`, `hooks/hooks.json`, `skills/`, and those `go/`
  binaries plus the launcher.
- AC2: WHEN a Claude hook runs from an unpacked release zip THE SYSTEM SHALL
  exec the binary for the host OS/arch through a launcher in `go/`. WHEN no
  matching binary exists the launcher SHALL exit 0 with a one-line stderr
  notice. A hook must never block the prompt.
- AC3: WHEN a `v*` tag is pushed THE SYSTEM SHALL run
  `.github/workflows/release.yml`, which calls `go/release.sh`, attaches the
  zip to the GitHub Release, and updates the `archive` entry (`url`,
  `sha256`) in `.claude-plugin/marketplace.json` on the default branch.
- AC4: `.claude-plugin/marketplace.json` MUST declare `alter-bridge` with an
  `archive` source whose `sha256` matches the released zip.
- AC5: WHEN a user runs `/plugin marketplace add rntgspr/alter-bridge` and
  `/plugin install alter-bridge@alter-bridge` THE SYSTEM SHALL install from the
  release zip, and the three hooks SHALL work without Go or a checkout.
- AC6: `README.md` and `SKILL.md` MUST describe the standard install as the
  Claude install path, with no `install.sh` step for Claude. Local development
  MUST be documented via `claude --plugin-dir`.
- AC7: The Codex wiring MUST keep working unchanged: `hooks/codex.json`,
  `.codex-plugin/plugin.json`, and the `bin/alter-bridge` build path.

## Plan / DAG

| Task | Title | Status | Depends on |
|------|-------|--------|-----------|
| [T1](t1.md) | Builder `go/release.sh` and launcher | done | — |
| [T2](t2.md) | Root release workflow | done | T1 |
| [T3](t3.md) | Hooks, marketplace `archive` source, docs | pending | T1 |
| [T4](t4.md) | First release and live Claude cutover (Renato-gated) | pending | T2, T3 |

## Out of scope

- Everything Codex: its marketplace file, manifest, hooks, and install path.
- Code signing and notarization of the macOS binaries.
- Automatic plugin updates beyond what `/plugin marketplace update` does.

## Risks

- **Target platforms (assumed):** darwin-arm64, darwin-amd64, linux-amd64,
  linux-arm64. Renato has not confirmed the list. Each binary adds about
  4.5 MB to the zip.
- **Unsigned binaries:** macOS quarantine may block an unsigned binary that
  came from a download. Verify on the first real install, in T4.
- **Workflow writes to main:** committing the `sha256` back to the default
  branch needs `contents: write`, and a branch protection rule could reject the
  push.
- **Outward-facing steps:** pushing a tag and creating a release are
  outward-facing, so T4 requires Renato's explicit go-ahead.
- **Double hooks during cutover:** the release plugin's hooks and the manual
  settings hooks must not overlap. This is the same ordering risk recorded in
  `maintenance-plugin-standard`.
