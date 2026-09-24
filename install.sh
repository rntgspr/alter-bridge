#!/usr/bin/env bash
# Builds bin/alter-bridge and installs this checkout as the alter-bridge plugin
# on Codex, through the repository's own marketplace
# (.claude-plugin/marketplace.json, which Codex reads as its legacy-compatible
# marketplace). Re-running is safe and is how a rebuilt binary reaches Codex,
# which runs a cached copy. Claude Code installs from the GitHub Release zip
# instead; see the README.
#
# Usage: install.sh
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"$repo/go/build.sh"

if command -v codex >/dev/null; then
  codex plugin marketplace add "$repo"
  codex plugin add alter-bridge@alter-bridge
  echo "install: approve the alter-bridge hook in Codex with /hooks before it runs"
else
  echo "install: codex not found, skipped"
fi
