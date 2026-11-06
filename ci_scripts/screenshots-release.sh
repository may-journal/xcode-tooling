#!/bin/sh
# Captures from the release commit, so a release never carries stale images.
# A release lane runs this before any upload, and any failure here stops the release.
set -u
tools=$(cd "$(dirname "$0")" && pwd)

output=$(jq -r .output "${SCREENSHOTS_CONFIG:-fastlane/screenshots.json}")
if [ -n "$(git status --porcelain -- . ":(exclude)$output")" ]; then
	echo "screenshots-release: commit or stash first, so the images match the release commit" >&2
	exit 1
fi
sh "$tools/screenshots.sh"
result=$?
if [ "$result" = 75 ]; then
	echo "screenshots-release: a simulator is busy, so the release stops before any upload" >&2
	exit 1
fi
[ "$result" = 0 ] || exit "$result"
echo "screenshots-release: images come from $(git rev-parse --short HEAD)"
