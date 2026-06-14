#!/usr/bin/env bash
# PreToolUse(Edit|Write): refuse manual edits to generated code.
# Generated files come from gqlgen / sqlc; hand-editing them creates drift that the next
# 'make generate' silently destroys. Blocks (exit 2) before the edit happens.
set -euo pipefail

input=$(cat)
file=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')

[ -z "$file" ] && exit 0

base=$(basename "$file")

# Generated code (gqlgen / sqlc) — never hand-edited.
gen=0
case "$base" in
  *_gen.go|generated.go|*.sql.go) gen=1 ;;
esac
case "$file" in
  */generated/*|*/sqlc/*) gen=1 ;;
esac

if [ "$gen" -eq 1 ]; then
  {
    echo "Refusing to edit generated file: $file"
    echo "This file is produced by code generation (gqlgen / sqlc)."
    echo "Edit the source instead — the GraphQL schema (*.graphqls) or db/queries/*.sql — and run 'make generate'."
  } >&2
  exit 2
fi

# Atlas migrations — immutable once created.
case "$file" in
  */db/migrations/*)
    {
      echo "Refusing to edit migration file: $file"
      echo "Atlas migrations are versioned and immutable once created."
      echo "Don't edit an existing migration — change the schema in db/schema/ and run 'make migrate-new name=...' to create a new one (see the manage-db-schema skill)."
    } >&2
    exit 2
    ;;
esac

exit 0
