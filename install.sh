#!/usr/bin/env bash
# Builds bin/alter-bridge and installs this checkout as the alter-bridge plugin
# on Claude Code and Codex, through the repository's own marketplace
# (.claude-plugin/marketplace.json, read by both CLIs). Re-running is safe and
# is how a rebuilt binary reaches Codex, which runs a cached copy.
#
# Usage: install.sh
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"$repo/go/build.sh"

if command -v claude >/dev/null; then
  claude plugin marketplace add "$repo"
  claude plugin install alter-bridge@alter-bridge
else
  echo "install: claude not found, skipped"
fi

if command -v codex >/dev/null; then
  codex plugin marketplace add "$repo"
  codex plugin add alter-bridge@alter-bridge
  echo "install: approve the alter-bridge hook in Codex with /hooks before it runs"
else
  echo "install: codex not found, skipped"
fi
