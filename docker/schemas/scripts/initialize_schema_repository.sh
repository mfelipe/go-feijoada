#!/bin/bash
set -e

SCHEMA_DIR="/schemas"

for file in "$SCHEMA_DIR"/*.json; do
  SCHEMA_ID=$(jq -r '."$id"' "$file")
  echo "Registering schema: $SCHEMA_ID from $file"
  # Wrap the schema file content inside a {"schema": ...} JSON object
  BODY=$(jq -n --argjson schema "$(jq . "$file")" '{schema: $schema}')
  HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$SCHEMA_ID" \
    -H "Content-Type: application/json" \
    --data "$BODY")
  if [ "$HTTP_CODE" -ne 201 ]; then
    echo "Failed to register schema $SCHEMA_ID from $file (HTTP $HTTP_CODE)"
    exit 1
  fi
  echo "Schema $SCHEMA_ID registered successfully (HTTP $HTTP_CODE)"
done