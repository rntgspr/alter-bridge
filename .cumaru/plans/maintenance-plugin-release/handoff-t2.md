---
human_revised: false
plan: maintenance-plugin-release
task: T2
status: complete
date: 2026-09-24
summary: Hand-off for maintenance-plugin-release T2 - root release workflow on v* tags plus the tested go/archive-source.sh marketplace rewrite it calls.
---

# Hand-off — maintenance-plugin-release / T2

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`.github/workflows/release.yml`](.github/workflows/release.yml) | created - on `v*` tag push: tests, `go/release.sh $TAG`, `gh release create` with the zip, then `go/archive-source.sh` on the default branch, commit, and push |
| [`go/archive-source.sh`](go/archive-source.sh) | created - `jq` rewrite of the `alter-bridge` entry's `source` to `{source: archive, url, sha256}`; refuses a malformed sha256 and leaves the file untouched |
| [`go/archive-source_test.sh`](go/archive-source_test.sh) | created - shell test on a copy of `marketplace.json` plus an unrelated entry |
<!-- /cumaru:touched -->

## Decisions made during implementation

- The `jq` rewrite is in `go/archive-source.sh`, not inline in the YAML. That
  keeps it testable and inside `go/`, next to the rest of the build logic, and
  the workflow stays a thin caller. The task's `files:` listed only the
  workflow.
- The workflow runs `go test ./...` and both shell tests before it builds, so a
  red tree never publishes a release.
- The workflow computes the sha256 again with `sha256sum`, instead of parsing
  the output of `release.sh`.
- The marketplace commit is made on a fresh checkout of the default branch
  (`github.event.repository.default_branch`) and not on the tag's detached
  HEAD. When the file already matches, the step is a no-op, so a re-run is
  safe.
- The commit author is `github-actions[bot]`. The message is
  `chore(release): point marketplace at <tag>` and has no `[papa]` suffix,
  because the bot writes it, not the agent.
- Actions used: `actions/checkout@v4` and `actions/setup-go@v5`, with the Go
  version read from `go/go.mod`. The release itself uses the `gh` CLI, with no
  third-party actions.

## Commands run / verification

- `go/archive-source_test.sh` against an empty script: 2 FAIL, exit 1 (red).
  After the script was written: 6 ok, exit 0 (green).
- `go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/release.yml`
  (v1.7.12): no findings, exit 0. `shellcheck` is not installed, so the
  embedded `run:` scripts were not shellchecked.
- Dry run on a scratch copy of `marketplace.json` with `TAG=v0.1.0` and the
  sha256 of the T1 zip: the source became `{"source":"archive","url":"https://github.com/rntgspr/alter-bridge/releases/download/v0.1.0/alter-bridge-v0.1.0.zip","sha256":"eb0287...43b2"}`,
  and the sha256 matches `sha256sum`. The repo's `marketplace.json` was not
  changed.
- The workflow's Test step commands, run from `go/`, pass locally.

## Pending / follow-ups

- The workflow has not run on GitHub yet. Its first run is the T4 tag push.

## Suggestions for the Lead

- A branch protection rule on the default branch rejects the bot's push. The
  release is still published, but the marketplace stays stale. Check the
  repository settings before T4.
