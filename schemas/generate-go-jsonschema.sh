#!/bin/bash
set -e

SCHEMA_DIR="$(dirname "$0")/schemas"
CMD="go-jsonschema --struct-name-from-title  --extra-imports --only-models --resolve-extension .json"
FILES=""

for file in "$SCHEMA_DIR"/*.json; do
  # Extract $id from the JSON file using jq
  SCHEMA_ID=$(jq -r '."$id"' "$file")
  SCHEMA_FILE=$(basename "$file")
  # Extract name and version from $id
  NAME=$(echo "$SCHEMA_ID" | awk -F/ '{print $(NF-1)}')
  RAW_VERSION=$(echo "$SCHEMA_ID" | awk -F/ '{print $NF}')
  VERSION=v$(echo "$RAW_VERSION" | sed 's/\./_/g')

  CMD+=" \
  --schema-package=${SCHEMA_ID}=github.com/mfelipe/go-feijoada/schemas/models/${VERSION} \
  --schema-output=${SCHEMA_ID}=models/${VERSION}/${NAME}.go"
  FILES+=" schemas/${SCHEMA_FILE}"
done

CMD+="$FILES"

# Execute the command
exec $CMD
