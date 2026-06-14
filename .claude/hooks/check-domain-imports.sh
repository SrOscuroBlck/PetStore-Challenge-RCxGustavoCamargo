#!/usr/bin/env bash
# PostToolUse(Edit|Write): enforce domain-layer purity.
# internal/domain must stay free of infrastructure and framework imports (layer separation).
# Blocks (exit 2) and feeds the violation back to Claude to fix.
set -euo pipefail

input=$(cat)
file=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')

[ -z "$file" ] && exit 0
case "$file" in *.go) ;; *) exit 0 ;; esac
case "$file" in */internal/domain/*) ;; *) exit 0 ;; esac
[ -f "$file" ] || exit 0

forbidden='jackc/pgx|redis/go-redis|minio/minio-go|99designs/gqlgen|graphql-go/graphql|"database/sql"|"net/http"'
hits=$(grep -nE "$forbidden" "$file" || true)

if [ -n "$hits" ]; then
  {
    echo "Domain-layer purity violation in: $file"
    echo "internal/domain is the pure business core and must not import infrastructure or framework packages."
    echo "Offending lines:"
    echo "$hits"
    echo "Fix: depend on a repository interface (declared in the domain/app layer) and move the concrete dependency into internal/adapter."
  } >&2
  exit 2
fi

exit 0
