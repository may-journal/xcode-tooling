#!/bin/sh
# Captures every App Store screenshot an app's fastlane/screenshots.json and requirement doc name.
# Exit 75 means a simulator is busy, so try again later; the README explains the run.
set -u
config=${SCREENSHOTS_CONFIG:-fastlane/screenshots.json}
tools=$(cd "$(dirname "$0")" && pwd)
root=$(pwd)
prefix=screenshots-$$
data=.derivedData-screenshots
raw=$root/.build/screenshots
value() { jq -r "$1" "$config"; }
lock=${SCREENSHOTS_LOCK:-$HOME/Library/Caches/$(value .label).lock}

skip() {
	echo "screenshots: $1, skipping this run" >&2
	exit 75
}

# One agent drives a simulator at a time, so a lock held by a live run skips this one.
holds_lock() {
	mkdir "$lock" 2>/dev/null && return 0
	kill -0 "$(cat "$lock/pid" 2>/dev/null)" 2>/dev/null && skip "another capture holds the lock"
	return 0
}

# Writes one line per frame: its acceptance number, file name and headline, from the doc.
frames() {
	awk '
		/^- [0-9]+\.[0-9]+$/ { split($2, id, "."); number = id[2] }
		/When I look at the / { name = $0; sub(/.*When I look at the /, "", name); name = tolower(name); gsub(/ /, "-", name) }
		/Then .*"/ { split($0, words, "\""); printf "%d\t%02d-%s\t%s\n", number, number, name, words[2] }
	' "$(value .requirement)" >"$raw/frames.tsv"
	[ -s "$raw/frames.tsv" ] || {
		echo "screenshots: $(value .requirement) names no frames" >&2
		return 1
	}
}

# The newest runtime that can run a device type, since a new runtime can leave one out.
runtime_for() {
	xcrun simctl list -j runtimes | jq -r --arg type "$1" \
		'[.runtimes[] | select(.isAvailable and any(.supportedDeviceTypes[]; .identifier == $type))]
		 | last.identifier // error("no runtime runs \($type)")'
}

# Shoots the simulator each time the test leaves a .wants file; it stops with the lock.
# Display 1 is the lit screen on a device with two.
watches_shutter() {
	while [ -d "$lock" ]; do
		for wants in "$2"/*.wants; do
			[ -e "$wants" ] || continue
			xcrun simctl io "$1" screenshot --display=1 "${wants%.wants}.png" 2>/dev/null
			mv "$wants" "${wants%.wants}.taken"
		done
		sleep 0.2
	done
}

delete_simulators() {
	xcrun simctl list -j devices | jq -r --arg prefix "$prefix-" \
		'.devices[][] | select(.name | startswith($prefix)) | .udid' | while read -r udid; do
		xcrun simctl shutdown "$udid" 2>/dev/null
		xcrun simctl delete "$udid"
	done
}

# shoots_once <folder> <device type> <window size or nothing>
shoots_once() {
	folder=$raw/$1
	type=com.apple.CoreSimulator.SimDeviceType.$2
	mkdir -p "$folder"
	udid=$(xcrun simctl create "$prefix-$1" "$type" "$(runtime_for "$type")")
	xcrun simctl bootstatus "$udid" -b >/dev/null
	xcrun simctl status_bar "$udid" override --time 9:41 --dataNetwork wifi --wifiMode active \
		--wifiBars 3 --cellularMode active --cellularBars 4 --batteryState charged --batteryLevel 100
	xcrun simctl ui "$udid" appearance light
	for service in $(value '.grant[]?'); do
		xcrun simctl privacy "$udid" grant "$service" "$(value .bundle)"
	done
	watches_shutter "$udid" "$folder" &
	watcher=$!
	# SwiftUI's render thread can run a test library's main-actor closure.
	# These two settings stop Swift trapping on that during a capture.
	TEST_RUNNER_SWIFT_IS_CURRENT_EXECUTOR_LEGACY_MODE_OVERRIDE=legacy \
		TEST_RUNNER_SWIFT_UNEXPECTED_EXECUTOR_LOG_LEVEL=0 \
		TEST_RUNNER_SCREENSHOT_SHUTTER=$folder TEST_RUNNER_SCREENSHOT_WINDOW=$3 xcodebuild \
		test-without-building -scheme "$(value .scheme)" -project "$(value .project)" \
		-destination "id=$udid" -derivedDataPath $data -collect-test-diagnostics never \
		-only-testing:"$(value .test)"
	result=$?
	kill "$watcher" 2>/dev/null
	delete_simulators
	return $result
}

# A simulator run can crash, so each size gets three tries.
shoot() {
	for attempt in 1 2 3; do
		shoots_once "$1" "$2" "$3" && return 0
		rm -rf "${raw:?}/$1"
	done
	return 1
}

capture() {
	set -e
	rm -rf "$raw"
	mkdir -p "$raw"
	frames
	[ ! -f project.yml ] || xcodegen generate --quiet
	xcodebuild build-for-testing -scheme "$(value .scheme)" -project "$(value .project)" \
		-destination 'generic/platform=iOS Simulator' -derivedDataPath $data -quiet
	jq -r '.sizes[] | [.folder, .device, .window // "-"] | @tsv' "$config" >"$raw/sizes.tsv"
	while IFS="$(printf '\t')" read -r folder device window; do
		[ "$window" != "-" ] || window=
		shoot "$folder" "$device" "$window"
	done <"$raw/sizes.tsv"
	xcrun swift "$tools/screenshots-compose.swift" "$config" "$raw/frames.tsv" "$raw" "$(value .output)"
}

holds_lock
echo $$ >"$lock/pid"
booted=$(xcrun simctl list devices booted | grep -c '(Booted)')
if [ "$booted" != 0 ]; then
	rm -rf "$lock"
	skip "$booted simulators are booted"
fi

# The capture runs in a child, so this shell cleans up however it ended.
(capture) &
job=$!
trap 'pkill -P $job; kill $job 2>/dev/null' INT TERM
wait $job
result=$?
rm -rf "$lock"
delete_simulators
exit $result
