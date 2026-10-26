#!/bin/sh
# Xcode Cloud runs this after each xcodebuild action. After an archive, it writes the merge's
# changelog entry to TestFlight/WhatToTest.en-US.txt. Xcode Cloud uploads that file as the
# build's What to Test notes, which TestFlight shows in its email and app. An app's
# ci_scripts/ci_post_xcodebuild.sh hands its name to its shim, which clones this release and runs
# this file from the app's root.
#
# Usage:   run by the app's shim; here, cd <app> && CI_ARCHIVE_PATH=x sh ci_scripts/ci_post_xcodebuild.sh
# Inputs:  the current folder, the app's root; CI_ARCHIVE_PATH, set on archive actions; the
#          changelog/*.md entry the built commit added, else the last entry by name, else
#          CHANGELOG.md. With none, it writes nothing.
# Outputs: TestFlight/WhatToTest.en-US.txt, at most 4000 characters, TestFlight's limit
set -eu

# Build and test actions have no archive, so they get no notes.
[ -n "${CI_ARCHIVE_PATH:-}" ] || exit 0

# The entry the built commit added is this merge's. A shallow clone cannot tell, so then the
# newest by name counts: entry names start with their date and time.
entry=$(git diff-tree --no-commit-id --name-only --diff-filter=A -r HEAD -- 'changelog/[0-9]*.md' 2>/dev/null | tail -n 1)
if [ -z "$entry" ]; then
	entry=CHANGELOG.md
	for file in changelog/[0-9]*.md; do
		[ -f "$file" ] && entry=$file
	done
fi
# Notes are optional, so a repo with no changelog still uploads its build.
[ -f "$entry" ] || exit 0

mkdir -p TestFlight
# The first "### <stamp>" section: its "> " summary, then its bullets.
awk '
  /^### / { if (seen) exit; seen = 1; next }
  seen && /^> / { sub(/^> /, ""); print; print ""; next }
  seen && NF { print }
' "$entry" | perl -CSD -0777 -pe '$_ = substr($_, 0, 4000)' >TestFlight/WhatToTest.en-US.txt
