#!/bin/sh
# Verifies the one-command installer from a release archive without network access.
set -eu

repo="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/bin" "$work/home"

cat >"$work/bin/curl" <<'EOF'
#!/bin/sh
while [ "$#" -gt 0 ]; do
  if [ "$1" = -o ]; then
    cp "$TEST_ARCHIVE" "$2"
    exit
  fi
  shift
done
exit 1
EOF

cat >"$work/bin/codex" <<'EOF'
#!/bin/sh
printf '%s\n' "$*" >>"$TEST_LOG"
EOF
chmod +x "$work/bin/curl" "$work/bin/codex"

plugin="$work/home/.local/share/alter-bridge-codex"
HOME="$work/home" PATH="$work/bin:$PATH" TEST_ARCHIVE="$repo/dist/alter-bridge-codex.zip" \
  TEST_LOG="$work/commands" sh "$repo/install-codex.sh"

printf 'stale manifest\n' >"$plugin/.codex-plugin/plugin.json"
chmod 444 "$plugin/.codex-plugin/plugin.json"

HOME="$work/home" PATH="$work/bin:$PATH" TEST_ARCHIVE="$repo/dist/alter-bridge-codex.zip" \
  TEST_LOG="$work/commands" sh "$repo/install-codex.sh"

test -x "$plugin/bin/alter-bridge"
unzip -p "$repo/dist/alter-bridge-codex.zip" .codex-plugin/plugin.json | cmp - "$plugin/.codex-plugin/plugin.json"
test "$(rg -c '^plugin marketplace add ' "$work/commands")" -eq 2
test "$(rg -c '^plugin add alter-bridge@alter-bridge-codex$' "$work/commands")" -eq 2
