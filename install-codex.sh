#!/bin/sh
# Installs the precompiled Codex plugin from the latest GitHub Release.
set -eu

for dependency in curl unzip codex; do
  if ! command -v "$dependency" >/dev/null 2>&1; then
    echo "install-codex: $dependency is required" >&2
    exit 1
  fi
done

if [ -z "${HOME:-}" ]; then
  echo "install-codex: HOME is required" >&2
  exit 1
fi

target="$HOME/.local/share/alter-bridge-codex"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

curl -fsSL https://github.com/rntgspr/alter-bridge/releases/latest/download/alter-bridge-codex.zip -o "$work/plugin.zip"
mkdir "$work/plugin"
unzip -q "$work/plugin.zip" -d "$work/plugin"

if [ ! -f "$work/plugin/.agents/plugins/marketplace.json" ] || \
   [ ! -f "$work/plugin/.codex-plugin/plugin.json" ] || \
   [ ! -x "$work/plugin/bin/alter-bridge" ]; then
  echo "install-codex: release archive is missing required plugin files" >&2
  exit 1
fi

mkdir -p "$target"
cp -R "$work/plugin/." "$target/"
codex plugin marketplace add "$target"
codex plugin add alter-bridge@alter-bridge-codex

echo "install-codex: open /hooks in Codex to trust the alter-bridge hook, then start a new thread"
