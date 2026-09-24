#!/bin/sh
# Tests go/archive-source.sh on a copy of .claude-plugin/marketplace.json plus
# an unrelated plugin entry that must stay untouched.
#
# Usage: go/archive-source_test.sh
set -eu

here="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
fails=0

url='https://github.com/rntgspr/alter-bridge/releases/download/v1.2.3/alter-bridge-v1.2.3.zip'
sha='eb028797e94cc4df2e2138e90261ac7bf328448cb63eaf2a551ca5ea709e43b2'

jq '.plugins += [{"name": "other", "source": "./other"}]' \
  "$here/../.claude-plugin/marketplace.json" >"$work/marketplace.json"
cp "$work/marketplace.json" "$work/before.json"

# Prints ok or FAIL for one named condition.
assert() {
  name="$1"
  shift
  if "$@"; then
    echo "ok   $name"
  else
    echo "FAIL $name"
    fails=$((fails + 1))
  fi
}

"$here/archive-source.sh" "$url" "$sha" "$work/marketplace.json"

assert "archive source set" [ "$(jq -c '.plugins[] | select(.name == "alter-bridge") | .source' "$work/marketplace.json")" \
  = "{\"source\":\"archive\",\"url\":\"$url\",\"sha256\":\"$sha\"}" ]
assert "rest of the entry kept" [ "$(jq -c '.plugins[] | select(.name == "alter-bridge") | del(.source)' "$work/marketplace.json")" \
  = "$(jq -c '.plugins[] | select(.name == "alter-bridge") | del(.source)' "$work/before.json")" ]
assert "other plugin untouched" [ "$(jq -c '.plugins[] | select(.name == "other")' "$work/marketplace.json")" \
  = '{"name":"other","source":"./other"}' ]
assert "top level kept" [ "$(jq -c 'del(.plugins)' "$work/marketplace.json")" = "$(jq -c 'del(.plugins)' "$work/before.json")" ]

cp "$work/marketplace.json" "$work/good.json"
assert "bad sha256 refused" sh -c '! "$1" "$2" not-a-digest "$3" 2>/dev/null' _ "$here/archive-source.sh" "$url" "$work/marketplace.json"
assert "file unchanged after refusal" cmp -s "$work/marketplace.json" "$work/good.json"

[ "$fails" -eq 0 ]
