#!/bin/sh
# Tests the go/alter-bridge launcher against stub binaries and a fake uname.
#
# Usage: go/launcher_test.sh
set -eu

here="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
fails=0

mkdir -p "$work/plugin/go" "$work/fakebin"
cp "$here/alter-bridge" "$work/plugin/go/alter-bridge"

for target in darwin-arm64 darwin-amd64 linux-amd64 linux-arm64; do
  printf '#!/bin/sh\necho "%s $*"\n' "$target" >"$work/plugin/go/alter-bridge-$target"
  chmod +x "$work/plugin/go/alter-bridge-$target"
done

cat >"$work/fakebin/uname" <<'EOF'
#!/bin/sh
case "$1" in
  -s) echo "$FAKE_S" ;;
  -m) echo "$FAKE_M" ;;
esac
EOF
chmod +x "$work/fakebin/uname"

# Runs the launcher as the given host and checks exit code, stdout, and the
# stderr line count.
check() {
  name="$1" s="$2" m="$3" want_out="$4" want_err_lines="$5"

  out="$(FAKE_S="$s" FAKE_M="$m" PATH="$work/fakebin:$PATH" \
    "$work/plugin/go/alter-bridge" hook 'claude x' 2>"$work/err")" && code=0 || code=$?
  err_lines="$(wc -l <"$work/err" | tr -d ' ')"

  if [ "$code" -eq 0 ] && [ "$out" = "$want_out" ] && [ "$err_lines" = "$want_err_lines" ]; then
    echo "ok   $name"
  else
    echo "FAIL $name: code=$code out='$out' stderr=$(cat "$work/err")"
    fails=$((fails + 1))
  fi
}

check darwin-arm64 Darwin arm64 'darwin-arm64 hook claude x' 0
check darwin-amd64 Darwin x86_64 'darwin-amd64 hook claude x' 0
check linux-amd64 Linux x86_64 'linux-amd64 hook claude x' 0
check linux-arm64 Linux aarch64 'linux-arm64 hook claude x' 0
check unknown-arch Linux riscv64 '' 1
check unknown-os FreeBSD amd64 '' 1

[ "$fails" -eq 0 ]
