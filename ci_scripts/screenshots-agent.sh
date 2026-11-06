#!/bin/sh
# Installs a launchd agent from the app's main checkout that watches main for app changes.
# With SCREENSHOTS_REMOVE set it takes the agent away, and rerunning it updates the job.
set -eu
tools=$(cd "$(dirname "$0")" && pwd)
config=${SCREENSHOTS_CONFIG:-fastlane/screenshots.json}
label=$(jq -r .label "$config")
plist=$HOME/Library/LaunchAgents/$label.plist
home=$HOME/Library/Application\ Support/$label
domain=gui/$(id -u)

launchctl bootout "$domain/$label" 2>/dev/null || true
if [ -n "${SCREENSHOTS_REMOVE:-}" ]; then
	rm -rf "$plist" "$home"
	exit 0
fi

mkdir -p "$home"
cp "$tools/screenshots-watch.sh" "$home/screenshots-watch.sh"
cat >"$plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>$label</string>
	<key>LimitLoadToSessionType</key>
	<string>Aqua</string>
	<key>ProgramArguments</key>
	<array>
		<string>/bin/sh</string>
		<string>$home/screenshots-watch.sh</string>
	</array>
	<key>EnvironmentVariables</key>
	<dict>
		<key>PATH</key>
		<string>$PATH</string>
		<key>SCREENSHOTS_CHECKOUT</key>
		<string>$(pwd)</string>
	</dict>
	<key>StartInterval</key>
	<integer>300</integer>
	<key>StandardErrorPath</key>
	<string>$HOME/Library/Logs/$label.log</string>
	<key>StandardOutPath</key>
	<string>$HOME/Library/Logs/$label.log</string>
</dict>
</plist>
PLIST
launchctl bootstrap "$domain" "$plist"
echo "$label polls main every five minutes and captures each app change once"
