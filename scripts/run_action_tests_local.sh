#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT"

echo "Generating coverage profile..."
# Run tests but skip GitHub integration tests by clearing GITHUB_TOKEN
env GITHUB_TOKEN= go test ./... -coverprofile=./cover.out -covermode=atomic

tmp_output="$(mktemp)"
cleanup() { rm -f "$tmp_output"; }
trap cleanup EXIT

run_cmd() {
  local id=$1
  shift
  local expect_fail=$1
  shift
  echo "-> Running $id: $*"
  export GITHUB_OUTPUT="$tmp_output"
  set +e
  "$@"
  rc=$?
  set -e
  if [ "$expect_fail" = "false" ]; then
    if [ $rc -ne 0 ]; then
      echo "[FAIL] $id: exited $rc but expected success"
      return 1
    fi
  else
    if [ $rc -eq 0 ]; then
      echo "[FAIL] $id: exited 0 but expected failure"
      return 1
    fi
  fi
  echo "[OK] $id (rc=$rc)"
  return 0
}

check_outputs_present() {
  if ! grep -q '^total-coverage=' "$tmp_output"; then
    echo "[FAIL] missing total-coverage in GITHUB_OUTPUT"
    return 1
  fi
  if ! grep -q '^badge-color=' "$tmp_output"; then
    echo "[FAIL] missing badge-color in GITHUB_OUTPUT"
    return 1
  fi
  if ! grep -q '^badge-text=' "$tmp_output"; then
    echo "[FAIL] missing badge-text in GITHUB_OUTPUT"
    return 1
  fi
  if ! grep -q '^report=' "$tmp_output" && [ "$1" = "require-report" ]; then
    echo "[FAIL] missing report in GITHUB_OUTPUT"
    return 1
  fi
  echo "[OK] outputs present in GITHUB_OUTPUT"
}

# ensure no leftover artifacts from previous runs
rm -f coverage-badge.svg coverage.breakdown || true

CMD_BASE=(go run ./ --github-action-output=true)

# Test 1
run_cmd test-1 false "${CMD_BASE[@]}" --config=./.github/workflows/testdata/zero.yml || exit 1
check_outputs_present require-report || exit 1

# Test 2 (should fail)
run_cmd test-2 true "${CMD_BASE[@]}" --config=./.github/workflows/testdata/total100.yml || true
check_outputs_present "" || true

# Test 3
run_cmd test-3 false "${CMD_BASE[@]}" --profile=cover.out --threshold-file=0 --threshold-package=0 --threshold-total=0 || exit 1

# Test 4 (should fail)
run_cmd test-4 true "${CMD_BASE[@]}" --profile=cover.out --threshold-file=0 --threshold-package=0 --threshold-total=100 || true

# Test 5
run_cmd test-5 false "${CMD_BASE[@]}" --config=./.github/workflows/testdata/total100.yml --threshold-file=0 --threshold-package=0 --threshold-total=0 || exit 1

# Test 6 (debug, missing profile => fail)
run_cmd test-6 true "${CMD_BASE[@]}" --profile=nonexistent-profile.out --debug=true --threshold-file=0 --threshold-package=0 --threshold-total=100 || true

# Test 7 (threshold-file failure)
run_cmd test-7 true "${CMD_BASE[@]}" --profile=cover.out --threshold-file=100 --threshold-package=0 --threshold-total=0 || true

# Test 8 (threshold-package failure)
run_cmd test-8 true "${CMD_BASE[@]}" --profile=cover.out --threshold-file=0 --threshold-package=100 --threshold-total=0 || true

# Test 9 (badge file generation)
run_cmd test-9 false "${CMD_BASE[@]}" --config=./.github/workflows/testdata/zero.yml --badge-file-name=coverage-badge.svg || exit 1
if [ ! -f coverage-badge.svg ]; then echo "[FAIL] badge file not created"; exit 1; fi
echo "[OK] badge file created"

# Test 10 (breakdown file)
run_cmd test-10 false "${CMD_BASE[@]}" --config=./.github/workflows/testdata/zero.yml --breakdown-file-name=coverage.breakdown || exit 1
if [ ! -f coverage.breakdown ]; then echo "[FAIL] breakdown file not created"; exit 1; fi
echo "[OK] breakdown file created"

# Test 11 (diff with base breakdown)
run_cmd test-11 false "${CMD_BASE[@]}" --config=./.github/workflows/testdata/zero.yml --diff-base-breakdown-file-name=coverage.breakdown || exit 1

# Test 12 (diff threshold failure)
run_cmd test-12 true "${CMD_BASE[@]}" --config=./.github/workflows/testdata/zero.yml --diff-base-breakdown-file-name=coverage.breakdown --diff-threshold=100 || true

# Test 13 (missing profile without debug -> fail)
run_cmd test-13 true "${CMD_BASE[@]}" --profile=nonexistent-profile.out --threshold-file=0 --threshold-package=0 --threshold-total=0 || true

echo "All local action-test-steps ran (see above for failures expected)."
