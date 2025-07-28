#!/bin/sh
set -e

# This script finds the highest version for each subproject, determines the next vX.Y.Z tag,
# and creates/pushes the new tag according to the rules:
# - If only PATCH increases, increment PATCH.
# - If only MINOR increases, increment MINOR.
# - If any MAJOR increases, increment PATCH in the main tag.
ç
# List of subprojects
PROJECTS="utils stream-buffer schemas schema-validator kafka-consumer kafka-producer schema-repository stream-consumer"

# Get all tags for subprojects and the main tag
git fetch --tags

# Find the latest main tag
LATEST_MAIN_TAG=$(git tag | grep '^v' | sort -V | tail -n1)
if [ -z "$LATEST_MAIN_TAG" ]; then
  LATEST_MAIN_TAG="v0.0.0"
fi
MAIN_MAJOR=$(echo "$LATEST_MAIN_TAG" | cut -d'/' -f2 | cut -d'.' -f1 | tr -d 'v')
MAIN_MINOR=$(echo "$LATEST_MAIN_TAG" | cut -d'.' -f2)
MAIN_PATCH=$(echo "$LATEST_MAIN_TAG" | cut -d'.' -f3)

# Find the latest version for each project
LATESTS=""
for p in $PROJECTS; do
  TAG=$(git tag | grep "^$p/v" | sort -V | tail -n1)
  if [ -n "$TAG" ]; then
    LATESTS="$LATESTS $TAG"
  fi
done

# Parse the highest MAJOR, MINOR, PATCH for all subprojects
MAX_MAJOR=$MAIN_MAJOR
MAX_MINOR=$MAIN_MINOR
MAX_PATCH=$MAIN_PATCH
MAJOR_BUMP=0
MINOR_BUMP=0
PATCH_BUMP=0

for t in $LATESTS; do
  V=$(echo $t | cut -d'/' -f2 | tr -d 'v')
  MAJOR=$(echo $V | cut -d'.' -f1)
  MINOR=$(echo $V | cut -d'.' -f2)
  PATCH=$(echo $V | cut -d'.' -f3)
  if [ "$MAJOR" -gt "$MAX_MAJOR" ]; then
    MAJOR_BUMP=1
  fi
  if [ "$MINOR" -gt "$MAX_MINOR" ]; then
    MINOR_BUMP=1
  fi
  if [ "$PATCH" -gt "$MAX_PATCH" ]; then
    PATCH_BUMP=1
  fi
done

# Decide next main tag
if [ "$MAJOR_BUMP" -eq 1 ]; then
  # Any MAJOR bump in subprojects increments PATCH in main tag
  NEXT_MAJOR=$MAIN_MAJOR
  NEXT_MINOR=$MAIN_MINOR
  NEXT_PATCH=$((MAIN_PATCH+1))
elif [ "$MINOR_BUMP" -eq 1 ]; then
  NEXT_MAJOR=$MAIN_MAJOR
  NEXT_MINOR=$((MAIN_MINOR+1))
  NEXT_PATCH=0
elif [ "$PATCH_BUMP" -eq 1 ]; then
  NEXT_MAJOR=$MAIN_MAJOR
  NEXT_MINOR=$MAIN_MINOR
  NEXT_PATCH=$((MAIN_PATCH+1))
else
  # No change, just increment PATCH
  NEXT_MAJOR=$MAIN_MAJOR
  NEXT_MINOR=$MAIN_MINOR
  NEXT_PATCH=$((MAIN_PATCH+1))
fi

NEXT_TAG="v${NEXT_MAJOR}.${NEXT_MINOR}.${NEXT_PATCH}"

echo "Latest main tag: $LATEST_MAIN_TAG"
echo "Next main tag: $NEXT_TAG"

git tag "$NEXT_TAG"

git push origin "$NEXT_TAG"

echo "Pushed $NEXT_TAG"

