#!/bin/sh
# The launchd job: it wakes nightly and captures once a week from origin/main in its own worktree.
# Changed images refresh one pull request, and every result lands on the app's runs issue.
set -u
checkout=${SCREENSHOTS_CHECKOUT:?the agent sets the app checkout}
config=fastlane/screenshots.json
ref=${SCREENSHOTS_REF:-origin/main}
tree=$(dirname "$checkout")/wt-$(basename "$checkout")-screenshots-weekly
day=$(date +%Y-%m-%d)

cd "$checkout" || exit 1
git fetch -q origin
settings=$(git show "$ref:$config") || exit 1
value() { printf '%s' "$settings" | jq -r "$1"; }
repo=$(value .repo)
branch=$(value .branch)
stamp=$HOME/Library/Application\ Support/$(value .label)/last-run
log=$HOME/Library/Logs/$(value .label)-capture.log

note() { gh issue comment "$(value .runsIssue)" --body "Screenshots, $day: $1"; }
finish() {
	cd "$checkout" || exit 1
	git worktree remove --force "$tree"
	git worktree prune
	[ -z "${1:-}" ] || touch "$stamp"
	exit 0
}

# 10000 minutes is just under seven days, so a run a little later than the last still counts.
[ -z "$(find "$stamp" -mmin -10000 2>/dev/null)" ] || exit 0

git worktree remove --force "$tree" 2>/dev/null
git worktree add -q --detach "$tree" "$ref" || {
	note "no worktree at \`$tree\`, so the capture never started."
	exit 1
}
cd "$tree" || exit 1

ci_scripts/xcode-tooling.sh screenshots.sh >"$log" 2>&1
result=$?
if [ "$result" = 75 ]; then
	note "a simulator was busy, so this tries again tomorrow night."
	finish
fi
if [ "$result" != 0 ]; then
	note "the capture failed with exit $result. The log is \`$log\` on the Mac."
	finish stamped
fi

git add -A "$(value .output)"
if git diff --cached --quiet; then
	note "every image matches \`main\`, so nothing changed."
	finish stamped
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

- Chore: refresh \`$(value .output)\` from the weekly capture.
- Chore: the capture ran on $day from \`$(git rev-parse --short HEAD)\`.
- Chore: nothing uploads to App Store Connect.
ENTRY
git add "$entry"
old=$(git ls-remote origin "refs/heads/$branch" | cut -f1)
git commit -q -m "chore(screenshots): refresh the App Store screenshots" -m "$(value .trailer)" || {
	note "the hooks refused the refresh commit."
	finish stamped
}
git push -q --force-with-lease="$branch:$old" origin "HEAD:$branch" || {
	note "the push failed, so nothing was published."
	finish stamped
}

pr=$(gh pr list -R "$repo" --head "$branch" --json url --jq '.[0].url // ""')
if [ -n "$pr" ]; then
	bug=$(gh pr view "$pr" --json closingIssuesReferences --jq '.closingIssuesReferences[0].url // ""')
else
	bug=$(gh issue create -R "$repo" --label Bug --title "[Bug] App Store screenshots are out of date" --body "> The screenshots in \`$(value .output)\` no longer match the app.

## Background

The weekly capture on $day drew different images. Agent: $(value .agent).

## Steps to reproduce

1. Run \`make screenshots\`.

## Related requirements

[The screenshots requirement](https://github.com/$repo/blob/main/$(value .requirement)) needs no change.")
fi
body="> The App Store screenshots match the app again.

## Background

Closes #${bug##*/}. The weekly capture found images that differ from \`main\`. Agent: $(value .agent). Compare them in this diff.

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
finish stamped
