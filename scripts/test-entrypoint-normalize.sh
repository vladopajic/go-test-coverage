#!/usr/bin/env bash
set -euo pipefail

# Simple unit test for docker-entrypoint.sh normalization of hyphenated INPUT_*
# environment variables. Copies the real entrypoint, replaces the final exec
# with a printf of the args, then runs it with a hyphenated env var and asserts
# the underscore-style flag is present.

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

cp "$(pwd)/docker-entrypoint.sh" "$tmpdir/entrypoint.sh"
cp "$(pwd)/docker-entrypoint.sh" "$tmpdir/entrypoint.sh"
# Replace the final exec line robustly (match a line starting with exec)
sed -i.bak 's|^exec .*|printf "%s\\n" "${args[@]}"|' "$tmpdir/entrypoint.sh"
chmod +x "$tmpdir/entrypoint.sh"

# Run the modified entrypoint with a hyphenated INPUT_* env var (simulate runner)
# Note: shells cannot export hyphenated names, but `env 'NAME=val' cmd` can simulate it.
out=$(env 'INPUT_THRESHOLD-FILE=42' "$tmpdir/entrypoint.sh")

echo "Entrypoint output:"
echo "$out"

if echo "$out" | grep -q -- "--threshold-file=42"; then
  echo "[OK] normalization produced expected --threshold-file flag"
  exit 0
else
  echo "[FAIL] expected --threshold-file=42 in output"
  exit 2
fi
