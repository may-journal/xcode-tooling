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
mkdir -p "$home"
last=$(cat "$captured" 2>/dev/null)

note() { gh issue comment "$(value .runsIssue)" --body "Screenshots, $day, \`${head%"${head#???????}"}\`: $1"; }
remember() { echo "$head" >"$captured"; }
finish() {
	cd "$checkout" || exit 1
	git worktree remove --force "$tree"
	git worktree prune
	exit 0
}

[ "$head" != "$last" ] || exit 0
[ "$head" != "$(cat "$failed" 2>/dev/null)" ] || exit 0
if [ -n "$last" ]; then
	# Word splitting is wanted: each watched path is its own argument, and none holds a space.
	# shellcheck disable=SC2046
	if git diff --quiet "$last" "$head" -- $(value '.requirement, .watch[]?'); then
		remember
		exit 0
	fi
fi
if xcrun simctl list devices booted | grep -q '(Booted)'; then
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
if [ "$result" != 0 ]; then
	echo "$head" >"$failed"
	note "the capture failed with exit $result. The log is \`$log\` on the Mac."
	finish
fi
remember

git add -A "$(value .output)"
if git diff --cached --quiet; then
	note "every image matches \`main\`, so nothing changed."
	finish
fi
if [ -n "${SCREENSHOTS_DRY_RUN:-}" ]; then
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
	note "the hooks refused the refresh commit."
	finish
}
git push -q --force-with-lease="$branch:$old" origin "HEAD:$branch" || {
	note "the push failed, so nothing was published."
	finish
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
	note "\`$branch\` is pushed, but no issue tracks it, so no pull request changed."
	finish
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
if [ -n "$pr" ]; then
	note "images changed, so the refresh is at $pr"
else
	note "\`$branch\` is pushed, but no pull request opened."
fi
finish
