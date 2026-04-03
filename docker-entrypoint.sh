#!/bin/bash
set -e

# Start building the command
args=(/go-test-coverage)

# Only add arguments if they are specified (not empty/default)
[ -n "$INPUT_CONFIG" ] && args+=("--config=$INPUT_CONFIG")
[ -n "$INPUT_PROFILE" ] && args+=("--profile=$INPUT_PROFILE")
[ -n "$INPUT_SOURCE_DIR" ] && args+=("--source-dir=$INPUT_SOURCE_DIR")
[ "$INPUT_DEBUG" = "true" ] && args+=("--debug=true")

args+=("--github-action-output=true")

# For numeric thresholds, only add if explicitly set and not -1 (the "not set" indicator)
[ -n "$INPUT_THRESHOLD_FILE" ] && [ "$INPUT_THRESHOLD_FILE" != "-1" ] && args+=("--threshold-file=$INPUT_THRESHOLD_FILE")
[ -n "$INPUT_THRESHOLD_PACKAGE" ] && [ "$INPUT_THRESHOLD_PACKAGE" != "-1" ] && args+=("--threshold-package=$INPUT_THRESHOLD_PACKAGE")
[ -n "$INPUT_THRESHOLD_TOTAL" ] && [ "$INPUT_THRESHOLD_TOTAL" != "-1" ] && args+=("--threshold-total=$INPUT_THRESHOLD_TOTAL")

# Badge and CDN/Git configs (only if specified)
[ -n "$INPUT_BREAKDOWN_FILE_NAME" ] && args+=("--breakdown-file-name=$INPUT_BREAKDOWN_FILE_NAME")
[ -n "$INPUT_DIFF_BASE_BREAKDOWN_FILE_NAME" ] && args+=("--diff-base-breakdown-file-name=$INPUT_DIFF_BASE_BREAKDOWN_FILE_NAME")
[ -n "$INPUT_BADGE_FILE_NAME" ] && args+=("--badge-file-name=$INPUT_BADGE_FILE_NAME")

# CDN options
[ -n "$INPUT_CDN_KEY" ] && args+=("--cdn-key=$INPUT_CDN_KEY")
[ -n "$INPUT_CDN_SECRET" ] && args+=("--cdn-secret=$INPUT_CDN_SECRET")
[ -n "$INPUT_CDN_REGION" ] && args+=("--cdn-region=$INPUT_CDN_REGION")
[ -n "$INPUT_CDN_ENDPOINT" ] && args+=("--cdn-endpoint=$INPUT_CDN_ENDPOINT")
[ -n "$INPUT_CDN_FILE_NAME" ] && args+=("--cdn-file-name=$INPUT_CDN_FILE_NAME")
[ -n "$INPUT_CDN_BUCKET_NAME" ] && args+=("--cdn-bucket-name=$INPUT_CDN_BUCKET_NAME")
[ "$INPUT_CDN_FORCE_PATH_STYLE" = "true" ] && args+=("--cdn-force-path-style=true")

# Git options
[ -n "$INPUT_GIT_TOKEN" ] && args+=("--git-token=$INPUT_GIT_TOKEN")
[ -n "$INPUT_GIT_BRANCH" ] && args+=("--git-branch=$INPUT_GIT_BRANCH")
[ -n "$INPUT_GIT_REPOSITORY" ] && args+=("--git-repository=$INPUT_GIT_REPOSITORY")
[ -n "$INPUT_GIT_FILE_NAME" ] && args+=("--git-file-name=$INPUT_GIT_FILE_NAME")

# Execute the command
exec "${args[@]}"