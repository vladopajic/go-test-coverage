#!/usr/bin/env bash
set -euo pipefail

# Comprehensive entrypoint args validation tests.
# For each scenario, we simulate runner-provided env vars (including hyphenated
# names via `env 'NAME=val'`) and assert the expected args are printed by a
# copy of `docker-entrypoint.sh` that prints args instead of execing the binary.

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

cp "$(pwd)/docker-entrypoint.sh" "$tmpdir/entrypoint.sh"
sed -i.bak 's|^exec .*|printf "%s\n" "${args[@]}"|' "$tmpdir/entrypoint.sh"
chmod +x "$tmpdir/entrypoint.sh"

failures=0

run_case() {
  local name="$1"; shift
  local envcmd=(env)
  local expect=()
  local not_expect=()

  # Parse arguments: env=..., include=..., exclude=...
  while (($#)); do
    case "$1" in
      env=*) envcmd+=("${1#env=}"); shift;;
      include=*) expect+=("${1#include=}"); shift;;
      exclude=*) not_expect+=("${1#exclude=}"); shift;;
      *) echo "Unknown token $1"; exit 2;;
    esac
  done

  echo "-- CASE: $name"
  out=$("${envcmd[@]}" "$tmpdir/entrypoint.sh") || true
  echo "$out" | sed 's/^/   /'

  for e in "${expect[@]}"; do
    if ! echo "$out" | grep -Fxq -- "$e"; then
      echo "[FAIL] expected to find: $e"
      failures=$((failures+1))
    else
      echo "[OK] found: $e"
    fi
  done

  for ne in "${not_expect[@]}"; do
    if echo "$out" | grep -Fxq -- "$ne"; then
      echo "[FAIL] did not expect: $ne"
      failures=$((failures+1))
    else
      echo "[OK] not present: $ne"
    fi
  done
}

# Cases
run_case "underscore threshold" env='INPUT_THRESHOLD_FILE=5' include='--threshold-file=5' exclude='--threshold-file=-1'

run_case "hyphenated threshold" env='INPUT_THRESHOLD-FILE=7' include='--threshold-file=7' exclude='--threshold-file=-1'

run_case "threshold sentinel -1" env='INPUT_THRESHOLD_FILE=-1' exclude='--threshold-file=-1' exclude='--threshold-file=0'

run_case "diff threshold sentinel -101" env='INPUT_DIFF_THRESHOLD=-101' exclude='--diff-threshold=-101'

run_case "diff threshold value" env='INPUT_DIFF_THRESHOLD=100' include='--diff-threshold=100'

run_case "debug true" env='INPUT_DEBUG=true' include='--debug=true'

run_case "debug false (absent)" exclude='--debug=true'

run_case "cdn force path style true" env='INPUT_CDN_FORCE_PATH_STYLE=true' include='--cdn-force-path-style=true'

run_case "cdn hyphenated file name" env='INPUT_CDN-FILE-NAME=test.svg' include='--cdn-file-name=test.svg'

run_case "git repository hyphenated" env='INPUT_GIT-REPOSITORY=owner/repo' include='--git-repository=owner/repo'

run_case "badge file name" env='INPUT_BADGE_FILE_NAME=coverage.svg' include='--badge-file-name=coverage.svg'

run_case "profile set" env='INPUT_PROFILE=cover.out' include='--profile=cover.out'

# Summary
if ((failures==0)); then
  echo "All entrypoint args tests passed"
  exit 0
else
  echo "Entry point args tests had $failures failures"
  exit 2
fi
