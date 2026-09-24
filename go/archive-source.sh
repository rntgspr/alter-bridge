#!/bin/sh
# Points the alter-bridge entry of a Claude marketplace file at a release zip by
# replacing its source with an archive source (url + sha256). Every other field
# and entry is kept. Refuses a sha256 that is not 64 hex characters and leaves
# the file untouched.
#
# Usage: go/archive-source.sh <url> <sha256> [marketplace.json]
#        (default: .claude-plugin/marketplace.json of this repository)
set -eu

url="$1"
sha="$2"
file="${3:-$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)/.claude-plugin/marketplace.json}"

if ! printf '%s' "$sha" | grep -Eq '^[0-9A-Fa-f]{64}$'; then
  echo "archive-source: sha256 must be 64 hex characters, got '$sha'" >&2
  exit 1
fi

tmp="$file.tmp.$$"
jq --arg url "$url" --arg sha "$sha" \
  '(.plugins[] | select(.name == "alter-bridge") | .source) = {source: "archive", url: $url, sha256: $sha}' \
  "$file" >"$tmp"
mv "$tmp" "$file"
