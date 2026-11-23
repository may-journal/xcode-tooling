#!/bin/sh
# The launchd job: it polls main and exits at once unless a commit changed the app.
# It captures each change once in its own worktree, and a busy simulator means retry next poll.
set -u
checkout=${SCREENSHOTS_CHECKOUT:?the agent sets the app checkout}
config=fastlane/screenshots.json
tree=$(dirname "$checkout")/wt-$(basename "$checkout")-screenshots-watch
day=$(date +%Y-%m-%d)

cd "$checkout" || exit 1
git fetch -q origin
head=$(git rev-parse "${SCREENSHOTS_HEAD:-origin/main}") || exit 1
settings=$(git show "$head:$config" 2>/dev/null) || exit 0
value() { printf '%s' "$settings" | jq -r "$1"; }
repo=$(value .repo)
branch=$(value .branch)
home=${SCREENSHOTS_STATE:-$HOME/Library/Application Support/$(value .label)}
captured=$home/captured
failed=$home/failed
log=$HOME/Library/Logs/$(value .label)-capture.log
mkdir -p "$home" "$(dirname "$log")"
last=$(cat "$captured" 2>/dev/null)
# The commit the app's tooling tag names now, or nothing when the lookup fails.
tooling_url=${XCODE_TOOLING_URL:-https://github.com/may-journal/xcode-tooling.git}
tag=$(git show "$head:ci_scripts/xcode-tooling.version" 2>/dev/null | head -n 1)
tooling=
# Only a release tag, such as v1 or v1.2.0, is worth a lookup.
case $tag in
v[0-9]*) tooling=$(git ls-remote "$tooling_url" "refs/tags/$tag" 2>/dev/null | cut -f1 | head -n 1) ;;
esac
known=$(cat "$home/tooling" 2>/dev/null)
# A failed lookup or a first record is no move, so neither one starts a capture.
moved=
if [ -n "$tooling" ] && [ -n "$known" ] && [ "$tooling" != "$known" ]; then
	moved=1
fi
[ -z "$tooling" ] || [ -n "$known" ] || echo "$tooling" >"$home/tooling"

note() { gh issue comment "$(value .runsIssue)" --body "Screenshots, $day, \`${head%"${head#???????}"}\`: $1"; }
# Keeps the captured commit and the tooling commit it ran with.
remember() {
	echo "$head" >"$captured"
	[ -z "$tooling" ] || echo "$tooling" >"$home/tooling"
}
finish() {
	cd "$checkout" || exit 1
	git worktree remove --force "$tree"
	git worktree prune
	exit 0
}

# Stops trying this commit until an app change or a moved tag; the last captured one stays.
gives_up() {
	echo "$head" >"$failed"
	[ -z "$tooling" ] || echo "$tooling" >"$home/tooling"
	note "$1"
	finish
}

# A moved tag is an app change by itself, so it skips the three ways out below.
[ -n "$moved" ] || [ "$head" != "$last" ] || exit 0
[ -n "$moved" ] || [ "$head" != "$(cat "$failed" 2>/dev/null)" ] || exit 0
if [ -z "$moved" ] && [ -n "$last" ]; then
	# Word splitting is wanted: each watched path is its own argument, and none holds a space.
	# shellcheck disable=SC2046
	if git diff --quiet "$last" "$head" -- $config ci_scripts/xcode-tooling.version \
		$(value '.requirement, .watch[]?'); then
		remember
		exit 0
	fi
fi
# A leftover from a dead capture doesn't count: the capture reaps those itself.
if xcrun simctl list devices booted | grep '(Booted)' | grep -qvE ' screenshots-[0-9]+-'; then
	exit 0
fi

git worktree remove --force "$tree" 2>/dev/null
git worktree add -q --detach "$tree" "$head" || {
	note "no worktree at \`$tree\`, so the capture never started."
	exit 1
}
cd "$tree" || exit 1

ci_scripts/xcode-tooling.sh screenshots.sh >"$log" 2>&1
result=$?
[ "$result" != 75 ] || finish
[ "$result" = 0 ] || gives_up "the capture failed with exit $result. The log is \`$log\` on the Mac."

git add -A "$(value .output)"
if git diff --cached --quiet; then
	stale=$(gh pr list -R "$repo" --head "$branch" --json url --jq '.[0].url // ""')
	[ -z "$stale" ] || gh pr close "$stale" --comment "This commit's images match \`main\`, so this refresh is stale."
	remember
	note "every image matches \`main\`, so nothing changed."
	finish
fi
if [ -n "${SCREENSHOTS_DRY_RUN:-}" ]; then
	remember
	note "a dry run captured changed images and published nothing."
	finish
fi

git switch -q -C "$branch"
entry=changelog/$(date +%Y-%m-%d-%H%M)-screenshots-refresh.md
cat >"$entry" <<ENTRY
### $(date +%Y.%m.%d.%H%M)

> The App Store screenshots match the app again.

- Chore: refresh \`$(value .output)\` from the capture of this commit.
- Chore: the capture ran on $day from \`$(git rev-parse --short "$head")\`.
- Chore: nothing uploads to App Store Connect.
ENTRY
git add "$entry"
old=$(git ls-remote origin "refs/heads/$branch" | cut -f1)
git commit -q -m "chore(screenshots): refresh the App Store screenshots" -m "$(value .trailer)" || {
	gives_up "the hooks refused the refresh commit."
}
git push -q --force-with-lease="$branch:$old" origin "HEAD:$branch" || {
	gives_up "the push failed, so nothing was published."
}

pr=$(gh pr list -R "$repo" --head "$branch" --json url --jq '.[0].url // ""')
if [ -n "$pr" ]; then
	bug=$(gh pr view "$pr" --json closingIssuesReferences --jq '.closingIssuesReferences[0].url // ""')
else
	bug=$(gh issue create -R "$repo" --label Bug --title "[Bug] App Store screenshots are out of date" --body "> The screenshots in \`$(value .output)\` no longer match the app.

## Background

The capture on $day drew different images. Agent: $(value .agent).

## Steps to reproduce

1. Run \`make screenshots\`.

## Related requirements

[The screenshots requirement](https://github.com/$repo/blob/main/$(value .requirement)) needs no change.")
fi
if [ -z "$bug" ]; then
	gives_up "\`$branch\` is pushed, but no issue tracks it, so no pull request changed."
fi
body="> The App Store screenshots match the app again.

## Background

Closes #${bug##*/}. A capture found images that differ from \`main\`. Agent: $(value .agent). Compare them in this diff.

## Changelog

- [Changelog entry](https://github.com/$repo/blob/$branch/$entry)"
if [ -n "$pr" ]; then
	gh pr edit "$pr" --body "$body" >/dev/null
else
	pr=$(gh pr create -R "$repo" --head "$branch" --body "$body" \
		--title "chore(screenshots): refresh the App Store screenshots")
fi
[ -n "$pr" ] || gives_up "\`$branch\` is pushed, but no pull request opened."
remember
note "images changed, so the refresh is at $pr"
finish
