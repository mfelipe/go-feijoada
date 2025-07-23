#!/bin/bash

# Default parameters values
ENDPOINT="http://dynamodb:8000"
TIMEOUT=5
RETRIES=10
WAIT=2
FORCE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -e)
      ENDPOINT="$2"
      shift 2
      ;;
    -t)
      TIMEOUT="$2"
      shift 2
      ;;
    -r)
      RETRIES="$2"
      shift 2
      ;;
    -w)
      WAIT="$2"
      shift 2
      ;;
    -force)
      FORCE=true
      shift
      ;;
    *)
      echo "Unknown parameter: $1"
      echo "Usage: $0 [-e endpoint] [-t timeout_seconds] [-r retry_count] [-w retry_wait_seconds] [-force]"
      exit 1
      ;;
  esac
done

echo "Using endpoint: $ENDPOINT"
echo "Timeout: $TIMEOUT s, Retries: $RETRIES, Wait: $WAIT s, Force: $FORCE"

# Wait for DynamoDB to become available
TRY=1
while [ $TRY -le "$RETRIES" ]; do
  echo "Checking DynamoDB availability (Attempt $TRY/$RETRIES)..."
  if timeout "$TIMEOUT"s aws dynamodb list-tables --endpoint-url "$ENDPOINT"; then
    echo "DynamoDB is up!"
    break
  else
    echo "DynamoDB not available. Waiting $WAIT seconds..."
    sleep "$WAIT"
    TRY=$((TRY + 1))
  fi
done

if [ $TRY -gt "$RETRIES" ]; then
  echo "DynamoDB not reachable after $RETRIES attempts."
  exit 1
fi

echo "Validating table definition files..."

FAILED=0
declare -A TABLE_FILES

#shopt -s nullglob
for FILE in /scripts/table_*.json; do
  echo  "🔍 Validating $FILE..."
  [ -f "$FILE" ] || continue
  TABLE_NAME=$(jq -r '.TableName // empty' "$FILE" 2>/dev/null)

  if [ -z "$TABLE_NAME" ]; then
    echo "❌ ERROR: Could not extract valid TableName from $FILE"
    FAILED=1
  else
    echo "✅ Valid: $FILE → TableName: $TABLE_NAME"
    TABLE_FILES["$TABLE_NAME"]="$FILE"
  fi
done

if [ $FAILED -ne 0 ]; then
  echo "🚫 Aborting due to validation errors."
  exit 1
fi

if [ $TABLE_FILES -e 0 ]; then
  echo "🚫 Aborting due to no table files found."
  exit 1
fi

# Proceed with deletion (if force) and creation
for TABLE_NAME in "${!TABLE_FILES[@]}"; do
  FILE="${TABLE_FILES[$TABLE_NAME]}"
  echo "🔍 Processing table: $TABLE_NAME from $FILE"

  # Check if the table exists
  EXISTS=$(aws dynamodb list-tables --endpoint-url "$ENDPOINT" | jq -r '.TableNames[]' | grep -Fx "$TABLE_NAME")

  if [ -n "$EXISTS" ]; then
    if [ "$FORCE" = true ]; then
      echo "⚠️ Table '$TABLE_NAME' exists. Deleting (force mode)..."
      aws dynamodb delete-table --table-name "$TABLE_NAME" --endpoint-url "$ENDPOINT"
      echo "⏳ Waiting for table '$TABLE_NAME' to be deleted..."
      aws dynamodb wait table-not-exists --table-name "$TABLE_NAME" --endpoint-url "$ENDPOINT"
    else
      echo "⏩ Table '$TABLE_NAME' already exists. Skipping creation."
      continue
    fi
  fi

  echo "📦 Creating table '$TABLE_NAME'..."
  if aws dynamodb create-table --cli-input-json "file://$FILE" --endpoint-url "$ENDPOINT"; then
    echo "✅ Table '$TABLE_NAME' creation request sent."
  else
    echo "❌ Failed to create table '$TABLE_NAME'."
    exit 2
  fi
done

aws dynamodb list-tables --endpoint-url "$ENDPOINT"

echo "🎉 All table operations completed successfully."
