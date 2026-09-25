---
human_revised: false
plan: maintenance-plugin-release
task: T3
status: done
date: 2026-09-24
summary: Hand-off for maintenance-plugin-release T3 - hooks on the go/ launcher, Claude-free install.sh and docs done; the archive switch of marketplace.json is blocked because it breaks the Codex install.
---

# Hand-off — maintenance-plugin-release / T3

## Files touched

<!-- cumaru:touched -->
| Link | Description |
|------|-------------|
| [`hooks/hooks.json`](hooks/hooks.json) | modified - the three Claude hooks call `"${CLAUDE_PLUGIN_ROOT}/go/alter-bridge"` instead of `bin/` |
| [`install.sh`](install.sh) | modified - Claude branch removed; Codex branch byte-identical |
| [`README.md`](README.md) | modified - Install split into Claude (standard marketplace from GitHub, `claude --plugin-dir .` for dev) and Codex (`./install.sh`); Layout lists the release files |
| [`skills/alter-bridge/SKILL.md`](skills/alter-bridge/SKILL.md) | modified - install per runtime; dropped the "on the Bash `PATH`" claim, false for the release zip (no `bin/`) |
<!-- /cumaru:touched -->

## Resolution (2026-09-24)

Renato ruled Codex out of scope for this plan. The checked-in marketplace
keeps `"source": "./"`, and the release workflow's `go/archive-source.sh` step
switches it to `archive` on the first `v*` release. That puts AC4's check in
T4. The Codex incompatibility recorded below is now Renato's Codex plan to
handle.

## Former blocker: `.claude-plugin/marketplace.json` NOT switched to `archive`

Codex reads `.claude-plugin/marketplace.json` as its legacy-compatible
marketplace, and it cannot resolve an `archive` source. Both runs below used a
scratch `CODEX_HOME` and a `git archive` copy of the repo, with codex-cli
0.156.1:

- Current file (`"source": "./"`): `codex plugin add alter-bridge@alter-bridge`
  reports "Added plugin `alter-bridge`".
- File rewritten by `go/archive-source.sh`: `Error: plugin \`alter-bridge\` was
  not found in marketplace \`alter-bridge\``.

Switching the file therefore breaks the Codex install and violates AC7. The
T2 workflow makes that same rewrite on every release, so the first tag push
would break Codex in the same way. Renato has to choose one of these:

- A: add a Codex-native `.agents/plugins/marketplace.json` that keeps
  `"source": "./"`. Codex prefers that file. In scratch, with the Claude file
  switched to `archive`, `codex plugin add` succeeded and cached the checkout.
  The catch is that it adds a Codex file under `.agents/`, which is Renato's
  Codex design space and currently holds untracked content from someone else.
- B: another Codex-side arrangement that Renato designs.

Once Renato decides, T3 finishes with one call: `go/archive-source.sh <url>
<sha256>`, or a placeholder entry.

## Decisions made during implementation

- The live machine is not affected. It uses manual settings hooks on
  `bin/alter-bridge`, and `claude plugin list` shows no alter-bridge plugin.
- `SKILL.md` still names the checkout's `bin/alter-bridge` as the command. A
  user who installed only the release has no CLI for `send`. See the
  follow-ups.

## Commands run / verification

- `go/release.sh v0.1.0` produced sha256
  `d63325a8833ff5fd4a898046825448ba3be45cd147699ae44033a0c14896dab9`. The zip
  holds the updated `hooks/hooks.json`.
- Every hook command from `hooks/hooks.json` was run with
  `CLAUDE_PLUGIN_ROOT` set to the checkout and a scratch `ALTER_BRIDGE_ROOT`:
  - `watchpaths claude` printed the `watchPaths` JSON.
  - `relay claude` with `event: change` logged the payload to `relay.log`,
    printed nothing, and exited 0.
  - `hook claude` printed the pending message and moved it to `.archive/`.
- With a scratch `CLAUDE_CONFIG_DIR`:
  - `claude plugin validate .` passed, and so did the validation of the
    unpacked zip and of the archive-variant marketplace.
  - `claude --plugin-dir . plugin details alter-bridge` listed Skills (1) and
    Hooks (3): SessionStart, UserPromptSubmit, FileChanged.
  - The same check on the unpacked zip reports version `0.1.0`.
  - The unpacked zip's `SessionStart` hook printed the `watchPaths` JSON.
- No interactive `claude --plugin-dir .` session was started. The hooks were
  checked by feeding each effective command a payload, as in
  `maintenance-plugin-standard`.
- `rg 'install.sh' README.md skills` matches only Codex lines.

## Pending / follow-ups

- The marketplace switch is blocked (see above). AC4 is open.
- The release zip ships no `bin/`, so `alter-bridge` is not on the Claude Bash
  `PATH`, and `SKILL.md`'s command needs a checkout. The Claude docs advise
  against a top-level `bin/` for org-distributed plugins
  ([plugin-marketplaces](https://code.claude.com/docs/en/plugin-marketplaces)).
  A possible fix is to point the skill at the plugin's `go/alter-bridge`.
  That needs a verified substitution variable in skill content first.
