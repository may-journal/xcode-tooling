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
jobs=${SCREENSHOTS_JOBS:-$(value '.jobs // 1')}
# How long one clone may take before that try gives up.
clone_seconds=${SCREENSHOTS_CLONE_SECONDS:-120}
# One lock for every app, since one agent drives a simulator at a time.
lock=${SCREENSHOTS_LOCK:-$HOME/Library/Caches/org.may-journal.screenshots.lock}

# A run with no slot would wait forever, so a bad jobs value stops it at once.
case $jobs in
"" | *[!0-9]* | 0)
	echo "screenshots: jobs must be a whole number above zero, not '$jobs'" >&2
	exit 64
	;;
esac

skip() {
	echo "screenshots: $1, skipping this run" >&2
	exit 75
}

# A lock held by a live run skips this one.
holds_lock() {
	mkdir "$lock" 2>/dev/null && return 0
	kill -0 "$(cat "$lock/pid" 2>/dev/null)" 2>/dev/null && skip "another capture holds the lock"
	return 0
}

# Writes one line per frame: its acceptance number, file name and headline, from the doc.
frames() {
	awk '
		/^- [0-9]+\.[0-9]+$/ { split($2, id, "."); number = id[2]; name = "" }
		/When I look at the / { name = $0; sub(/.*When I look at the /, "", name); name = tolower(name); gsub(/ /, "-", name) }
		/Then .*"/ && name != "" { split($0, words, "\""); printf "%d\t%02d-%s\t%s\n", number, number, name, words[2] }
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
		sleep 0.05
	done
}

# Deletes the simulators whose names match a pattern; the lock's holder may reap dead runs'.
delete_simulators() {
	xcrun simctl list -j devices | jq -r --arg names "$1" \
		'.devices[][] | select(.name | test($names)) | .udid' | while read -r udid; do
		xcrun simctl shutdown "$udid" 2>/dev/null
		xcrun simctl delete "$udid"
	done
}

# Runs a command, stopping it after a number of seconds, so a hung simulator can't hold the lock.
within() {
	seconds=$1
	shift
	"$@" &
	task=$!
	(
		sleep "$seconds"
		kill "$task" 2>/dev/null
	) >/dev/null 2>&1 &
	timer=$!
	wait "$task"
	finished=$?
	kill "$timer" 2>/dev/null
	return $finished
}

# A shutdown can cut a file write short, and no clone can copy the unreadable file it leaves.
clears_cut_writes() {
	xcrun simctl list -j devices | jq -r --arg udid "$1" \
		'.devices[][] | select(.udid == $udid) | .dataPath' | while read -r path; do
		find "$path" -name '*.sb-*' -delete 2>/dev/null
	done
	# A folder find cannot read holds no cut write, so it never costs a good first boot.
	return 0
}

# Boots a device type once and leaves it shut down, since a clone skips the first boot's setup.
# A first boot can crash too, so it gets three tries.
makes_template() {
	type=com.apple.CoreSimulator.SimDeviceType.$1
	for attempt in 1 2 3; do
		udid=$(xcrun simctl create "$prefix-template-$1" "$type" "$(runtime_for "$type")")
		if within 180 xcrun simctl bootstatus "$udid" -b >/dev/null && xcrun simctl shutdown "$udid" &&
			clears_cut_writes "$udid"; then
			echo "$udid" >"$raw/$1.template"
			return 0
		fi
		xcrun simctl shutdown "$udid" 2>/dev/null
		xcrun simctl delete "$udid"
	done
	echo "screenshots: $1 never finished its first boot" >&2
	return 1
}

# One template per device type, one first boot at a time, so they stay light beside the build.
makes_templates() {
	for device in $(cut -f 2 "$raw/sizes.tsv" | sort -u); do
		makes_template "$device" || return 1
	done
}

# clones_for <device type> <name> <log>: prints a new simulator's id, one clone at a time.
# A clone that hangs gives up after $clone_seconds, so it never holds the others back.
clones_for() {
	until mkdir "$raw/cloning" 2>/dev/null; do
		sleep 0.1
	done
	cloned=0
	within "$clone_seconds" xcrun simctl clone "$(cat "$raw/$1.template")" "$2" 2>>"$3" || cloned=$?
	rmdir "$raw/cloning"
	return $cloned
}

# shoots_once <folder> <device type> <window size or nothing>, in the size's own child.
shoots_once() {
	folder=$raw/$1
	rm -rf "$raw/$1.xcresult"
	mkdir -p "$folder"
	udid=$(clones_for "$2" "$prefix-$1" "$raw/$1.log") || return 1
	within 180 xcrun simctl bootstatus "$udid" -b >/dev/null
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
	within 600 env TEST_RUNNER_SWIFT_IS_CURRENT_EXECUTOR_LEGACY_MODE_OVERRIDE=legacy \
		TEST_RUNNER_SWIFT_UNEXPECTED_EXECUTOR_LOG_LEVEL=0 \
		TEST_RUNNER_SCREENSHOT_SHUTTER="$folder" TEST_RUNNER_SCREENSHOT_WINDOW="$3" xcodebuild \
		test-without-building -scheme "$(value .scheme)" -project "$(value .project)" \
		-destination "id=$udid" -derivedDataPath $data -collect-test-diagnostics never \
		-resultBundlePath "$raw/$1.xcresult" -only-testing:"$(value .test)" >>"$raw/$1.log" 2>&1
	result=$?
	kill "$watcher" 2>/dev/null
	xcrun simctl shutdown "$udid" 2>/dev/null
	xcrun simctl delete "$udid"
	return $result
}

# A simulator run can crash, so each size gets three tries.
shoot() {
	for attempt in 1 2 3; do
		shoots_once "$1" "$2" "$3" && return 0
		rm -rf "${raw:?}/$1"
	done
	tail -n 20 "$raw/$1.log" >&2
	return 1
}

# Runs the sizes on one queue, at most $jobs together, each on its own simulator.
# Each size starts as soon as another hands back its slot on the slots pipe.
shoot_sizes() {
	mkfifo "$raw/slots"
	exec 3<>"$raw/slots"
	slot=0
	while [ "$slot" -lt "$jobs" ]; do
		echo >&3
		slot=$((slot + 1))
	done
	running=
	while IFS="$(printf '\t')" read -r folder device window; do
		[ "$window" != "-" ] || window=
		read -r _ <&3
		(
			shot=0
			shoot "$folder" "$device" "$window" || shot=1
			echo >&3
			exit "$shot"
		) &
		running="$running $!"
	done <"$raw/sizes.tsv"
	failed=0
	for pid in $running; do
		wait "$pid" || failed=1
	done
	return $failed
}

capture() {
	set -e
	rm -rf "$raw"
	mkdir -p "$raw"
	frames
	sh "$tools/ci_post_clone.sh"
	jq -r '.sizes[] | [.folder, .device, .window // "-"] | @tsv' "$config" >"$raw/sizes.tsv"
	# The first boots run beside the build, so a fresh build pays for them.
	makes_templates &
	templates=$!
	xcodebuild build-for-testing -scheme "$(value .scheme)" -project "$(value .project)" \
		-destination 'generic/platform=iOS Simulator' -derivedDataPath $data -quiet \
		ARCHS="$(uname -m)" || {
		stop_tree "$templates"
		return 1
	}
	wait "$templates"
	shoot_sizes
	xcrun swift "$tools/screenshots-compose.swift" "$config" "$raw/frames.tsv" "$raw" "$(value .output)"
}

holds_lock
echo $$ >"$lock/pid"
delete_simulators '^screenshots-[0-9]+-'
booted=$(xcrun simctl list devices booted | grep -c '(Booted)')
if [ "$booted" != 0 ]; then
	rm -rf "$lock"
	skip "$booted simulators are booted"
fi

# The capture runs in a child, so this shell cleans up however it ended.
stop_tree() {
	for child in $(pgrep -P "$1"); do
		(stop_tree "$child")
	done
	kill "$1" 2>/dev/null
}
(capture) &
job=$!
trap 'stop_tree $job' INT TERM
wait $job
result=$?
delete_simulators "^$prefix-"
rm -rf "$lock"
exit $result
