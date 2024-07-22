#!/bin/bash

# Fetch the latest tag from git
LATEST_TAG=$(git describe --tags "$(git rev-list --tags --max-count=1)")

# Extract the numeric part of the tag (assuming tags are in the format v<MAJOR>.<MINOR>.<PATCH>)
if [[ $LATEST_TAG =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  MAJOR=${BASH_REMATCH[1]}
  MINOR=${BASH_REMATCH[2]}
  PATCH=${BASH_REMATCH[3]}

  # Increment the patch version by one
  PATCH=$((PATCH + 1))

  # Format the new version as a tag (assuming the tag format is v<MAJOR>.<MINOR>.<PATCH>)
  NEW_TAG="v${MAJOR}.${MINOR}.${PATCH}"
else
  echo "Error: Latest tag does not match the expected format 'v<MAJOR>.<MINOR>.<PATCH>'"
  exit 1
fi

echo $NEW_TAG
