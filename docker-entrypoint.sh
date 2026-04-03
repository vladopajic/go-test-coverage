#!/bin/bash
set -e

# Start building the command
cmd="go-test-coverage"

# Only add arguments if they are specified (not empty/default)
[ -n "$INPUT_CONFIG" ] && cmd="$cmd --config=$INPUT_CONFIG"
[ -n "$INPUT_PROFILE" ] && cmd="$cmd --profile=$INPUT_PROFILE"
[ -n "$INPUT_SOURCE_DIR" ] && cmd="$cmd --source-dir=$INPUT_SOURCE_DIR"
[ "$INPUT_DEBUG" = "true" ] && cmd="$cmd --debug=true"

cmd="$cmd --github-action-output=true"

# For numeric thresholds, only add if not -1 (the "not set" indicator)
[ "$INPUT_THRESHOLD_FILE" != "-1" ] && cmd="$cmd --threshold-file=$INPUT_THRESHOLD_FILE"
[ "$INPUT_THRESHOLD_PACKAGE" != "-1" ] && cmd="$cmd --threshold-package=$INPUT_THRESHOLD_PACKAGE"
[ "$INPUT_THRESHOLD_TOTAL" != "-1" ] && cmd="$cmd --threshold-total=$INPUT_THRESHOLD_TOTAL"

# Badge and CDN/Git configs (only if specified)
[ -n "$INPUT_BREAKDOWN_FILE_NAME" ] && cmd="$cmd --breakdown-file-name=$INPUT_BREAKDOWN_FILE_NAME"
[ -n "$INPUT_DIFF_BASE_BREAKDOWN_FILE_NAME" ] && cmd="$cmd --diff-base-breakdown-file-name=$INPUT_DIFF_BASE_BREAKDOWN_FILE_NAME"
[ -n "$INPUT_BADGE_FILE_NAME" ] && cmd="$cmd --badge-file-name=$INPUT_BADGE_FILE_NAME"

# CDN options
[ -n "$INPUT_CDN_KEY" ] && cmd="$cmd --cdn-key=$INPUT_CDN_KEY"
[ -n "$INPUT_CDN_SECRET" ] && cmd="$cmd --cdn-secret=$INPUT_CDN_SECRET"
[ -n "$INPUT_CDN_REGION" ] && cmd="$cmd --cdn-region=$INPUT_CDN_REGION"
[ -n "$INPUT_CDN_ENDPOINT" ] && cmd="$cmd --cdn-endpoint=$INPUT_CDN_ENDPOINT"
[ -n "$INPUT_CDN_FILE_NAME" ] && cmd="$cmd --cdn-file-name=$INPUT_CDN_FILE_NAME"
[ -n "$INPUT_CDN_BUCKET_NAME" ] && cmd="$cmd --cdn-bucket-name=$INPUT_CDN_BUCKET_NAME"
[ "$INPUT_CDN_FORCE_PATH_STYLE" = "true" ] && cmd="$cmd --cdn-force-path-style=true"

# Git options
[ -n "$INPUT_GIT_TOKEN" ] && cmd="$cmd --git-token=$INPUT_GIT_TOKEN"
[ -n "$INPUT_GIT_BRANCH" ] && cmd="$cmd --git-branch=$INPUT_GIT_BRANCH"
[ -n "$INPUT_GIT_REPOSITORY" ] && cmd="$cmd --git-repository=$INPUT_GIT_REPOSITORY"
[ -n "$INPUT_GIT_FILE_NAME" ] && cmd="$cmd --git-file-name=$INPUT_GIT_FILE_NAME"

# Execute the command
exec $cmd