#!/bin/bash

# Fetch the latest tag from git
LATEST_TAG=$(git describe --tags "$(git rev-list --tags --max-count=1)")

# Extract the numeric part of the tag (assuming tags are in the format v<NUMBER>)
LATEST_VERSION=${LATEST_TAG//[!0-9]/}

# Increment the latest version by one
NEW_VERSION=$((LATEST_VERSION + 1))

# Format the new version as a tag (assuming the tag format is v<NUMBER>)
NEW_TAG="v${NEW_VERSION}"

echo $NEW_TAG
