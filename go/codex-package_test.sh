#!/bin/sh
# Verifies that the Codex release archive installs without a checkout or Go.
set -eu

repo="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
archive="$repo/dist/alter-bridge-codex.zip"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/codex-home"

test -f "$archive"
unzip -q "$archive" -d "$work/plugin"
test -x "$work/plugin/bin/alter-bridge"
test -x "$work/plugin/bin/alter-bridge-$(go env GOOS)-$(go env GOARCH)"
test -f "$work/plugin/.codex-plugin/plugin.json"
test -f "$work/plugin/hooks/codex.json"
test -f "$work/plugin/skills/alter-bridge/SKILL.md"
test -x "$work/plugin/skills/alter-bridge/scripts/alter-bridge"
"$work/plugin/bin/alter-bridge" who >/dev/null
"$work/plugin/skills/alter-bridge/scripts/alter-bridge" who >/dev/null

if command -v codex >/dev/null; then
  CODEX_HOME="$work/codex-home" codex plugin marketplace add "$work/plugin" --json >/dev/null
  CODEX_HOME="$work/codex-home" codex plugin add alter-bridge@alter-bridge-codex --json >"$work/installed.json"
  jq -e '.pluginId == "alter-bridge@alter-bridge-codex"' "$work/installed.json" >/dev/null
  installed="$(jq -r '.installedPath' "$work/installed.json")"
  jq -e '.hooks.UserPromptSubmit[0].hooks[0].command == "\"${PLUGIN_ROOT}/bin/alter-bridge\" hook codex"' "$installed/hooks/codex.json" >/dev/null
  test -x "$installed/bin/alter-bridge"
fi
