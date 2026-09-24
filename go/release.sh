#!/usr/bin/env bash
# Cross-compiles the alter-bridge CLI into go/alter-bridge-<os>-<arch> for every
# target below, then packs the Claude plugin tree (manifest, hooks, skills, the
# go/ launcher, and those binaries) into dist/alter-bridge-<version>.zip and
# prints its sha256. The manifest version inside the zip is stamped from
# <version>, because Claude Code treats it as the plugin's update signal.
#
# Usage: go/release.sh [version]   (default 0.0.0-dev; a leading v is dropped
#                                   for the manifest)
set -euo pipefail

TARGETS=(darwin-arm64 darwin-amd64 linux-amd64 linux-arm64)

version="${1:-0.0.0-dev}"
godir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(dirname "$godir")"
zip="$repo/dist/alter-bridge-$version.zip"

cd "$godir"
for target in "${TARGETS[@]}"; do
  CGO_ENABLED=0 GOOS="${target%-*}" GOARCH="${target#*-}" \
    go build -trimpath -ldflags '-s -w' -o "alter-bridge-$target" ./cmd/alter-bridge
  echo "built go/alter-bridge-$target"
done

stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT
mkdir -p "$stage/.claude-plugin" "$stage/hooks" "$stage/go"
jq --arg v "${version#v}" '.version = $v' "$repo/.claude-plugin/plugin.json" >"$stage/.claude-plugin/plugin.json"
cp "$repo/hooks/hooks.json" "$stage/hooks/"
cp -R "$repo/skills" "$stage/"
cp "$godir/alter-bridge" "$stage/go/"
for target in "${TARGETS[@]}"; do
  cp "$godir/alter-bridge-$target" "$stage/go/"
done

mkdir -p "$repo/dist"
rm -f "$zip"
(cd "$stage" && zip -qrX "$zip" . -x '*.DS_Store')
echo "zip: $zip"
echo "sha256: $(shasum -a 256 "$zip" | cut -d' ' -f1)"
