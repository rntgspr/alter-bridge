#!/usr/bin/env bash
# Cross-compiles the alter-bridge CLI into go/alter-bridge-<os>-<arch> for every
# target below, then packs separate Claude and Codex plugin archives.
# The Claude archive keeps its existing layout and versioned filename.
#
# Usage: go/release.sh [version]   (default 0.0.0-dev; a leading v is dropped
#                                   for the manifest)
set -euo pipefail

TARGETS=(darwin-arm64 darwin-amd64 linux-amd64 linux-arm64)

version="${1:-0.0.0-dev}"
godir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(dirname "$godir")"
zip="$repo/dist/alter-bridge-$version.zip"
codex_zip="$repo/dist/alter-bridge-codex.zip"

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

codex_stage="$(mktemp -d)"
trap 'rm -rf "$stage" "$codex_stage"' EXIT
mkdir -p "$codex_stage/.agents/plugins" "$codex_stage/.codex-plugin" "$codex_stage/hooks" "$codex_stage/bin"
jq --arg v "${version#v}" '.version = $v' "$repo/.codex-plugin/plugin.json" >"$codex_stage/.codex-plugin/plugin.json"
cp "$repo/hooks/codex.json" "$codex_stage/hooks/"
cp -R "$repo/skills" "$codex_stage/"
cp "$godir/alter-bridge" "$codex_stage/bin/"
for target in "${TARGETS[@]}"; do
  cp "$godir/alter-bridge-$target" "$codex_stage/bin/"
done
cat >"$codex_stage/.agents/plugins/marketplace.json" <<'EOF'
{
  "name": "alter-bridge-codex",
  "interface": {"displayName": "Alter Bridge"},
  "plugins": [{
    "name": "alter-bridge",
    "source": {"source": "local", "path": "./"},
    "policy": {"installation": "AVAILABLE", "authentication": "ON_INSTALL"},
    "category": "Productivity"
  }]
}
EOF

rm -f "$codex_zip"
(cd "$codex_stage" && zip -qrX "$codex_zip" . -x '*.DS_Store')
echo "zip: $codex_zip"
echo "sha256: $(shasum -a 256 "$codex_zip" | cut -d' ' -f1)"
