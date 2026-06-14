#!/usr/bin/env bash
# PostToolUse(Edit|Write): keep every Go file gofmt-clean (and goimports-clean when available).
# Non-blocking: formatting never fails a turn.
set -euo pipefail

input=$(cat)
file=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')

[ -z "$file" ] && exit 0
case "$file" in *.go) ;; *) exit 0 ;; esac
[ -f "$file" ] || exit 0

gofmt -w "$file" 2>/dev/null || true
if command -v goimports >/dev/null 2>&1; then
  goimports -w "$file" 2>/dev/null || true
fi

exit 0
