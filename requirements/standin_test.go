package requirements

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// shotsStandIn is a fake xcrun and xcodebuild that logs each call and plays the simulators.
const shotsStandIn = `#!/bin/sh
echo "$(basename "$0") $*" >>"$SHOTS_LOG"
case "$*" in
*"list devices booted"*)
	[ -z "$SHOTS_BOOTED" ] || echo "    Phone (ID) (Booted)"
	for sim in "$SHOTS_SIM"/sim.*; do
		[ -e "$sim" ] && echo "    $(cat "$sim") (ID) (Booted)"
	done
	;;
*"list -j runtimes"*) echo '{"runtimes":[{"isAvailable":true,"identifier":"rt","supportedDeviceTypes":[{"identifier":"com.apple.CoreSimulator.SimDeviceType.Phone"}]}]}' ;;
*"list -j devices"*)
	for sim in "$SHOTS_SIM"/sim.*; do
		[ -e "$sim" ] && printf '{"name":"%s","udid":"%s","dataPath":"%s"}\n' "$(cat "$sim")" "$(basename "$sim")" "$SHOTS_SIM/data.$(basename "$sim")"
	done | jq -s '{devices: {rt: .}}'
	;;
*"simctl create"*) sim=$(mktemp "$SHOTS_SIM/sim.XXXXXX"); echo "$3" >"$sim"; mkdir "$SHOTS_SIM/data.$(basename "$sim")"; basename "$sim" ;;
*"simctl shutdown"*) [ -z "$SHOTS_CUT_WRITE" ] || [ ! -d "$SHOTS_SIM/data.$3" ] || touch "$SHOTS_SIM/data.$3/plist.sb-1" ;;
*"simctl clone"*)
	[ -z "${SHOTS_CLONE_HANG:-}" ] || exec sleep 30
	[ -z "${SHOTS_CLONE_EXIT:-}" ] || { echo "clone refused" >&2; exit "$SHOTS_CLONE_EXIT"; }
	[ -z "$(ls "$SHOTS_SIM/data.$3")" ] || exit 22
	sim=$(mktemp "$SHOTS_SIM/sim.XXXXXX"); echo "$4" >"$sim"; mkdir "$SHOTS_SIM/data.$(basename "$sim")"; basename "$sim"
	;;
*"simctl delete"*) rm -f "$SHOTS_SIM/$3" ;;
*"bootstatus"*) exit "${SHOTS_BOOT_EXIT:-0}" ;;
*"test-without-building"*)
	sim=$(printf '%s' "$*" | sed 's/.*id=\(sim\.[^ ]*\).*/\1/')
	if grep -q bad "$SHOTS_SIM/$sim"; then exit 65; fi
	if grep -q slow "$SHOTS_SIM/$sim"; then
		tries=0
		until grep -q "clone .*-last$" "$SHOTS_LOG"; do
			tries=$((tries + 1))
			[ "$tries" -lt 50 ] || exit 66
			sleep 0.1
		done
	fi
	touch "$SHOTS_SIM/running.$$"
	tries=0
	while [ "$(ls "$SHOTS_SIM" | grep -c '^running\.')" -lt "${SHOTS_AT_ONCE:-1}" ]; do
		tries=$((tries + 1))
		[ "$tries" -lt 50 ] || exit 66
		sleep 0.1
	done
	exit "$SHOTS_TEST_EXIT"
	;;
*swift*) cat "$4" ;;
esac
`

// shotsCapture runs screenshots.sh in app with the stand-ins first on PATH, returning output,
// exit code and calls.
func shotsCapture(t *testing.T, app string, env ...string) (string, int, string) {
	t.Helper()
	return shotsBeside(t, "", app, env...)
}

// shotsBeside is shotsCapture with a simulator of this name already booted, if any.
func shotsBeside(t *testing.T, leftover, app string, env ...string) (string, int, string) {
	t.Helper()
	bin := t.TempDir()
	for _, name := range []string{"xcrun", "xcodebuild", "gh"} {
		mustDo(t, os.WriteFile(filepath.Join(bin, name), []byte(shotsStandIn), 0o755))
	}
	logged := filepath.Join(bin, "calls")
	mustDo(t, os.WriteFile(logged, nil, 0o644))
	sims := filepath.Join(bin, "sims")
	mustDo(t, os.Mkdir(sims, 0o755))
	if leftover != "" {
		mustDo(t, os.WriteFile(filepath.Join(sims, "sim.left"), []byte(leftover), 0o644))
	}
	env = append(env,
		fmt.Sprintf("PATH=%s:%s", bin, os.Getenv("PATH")), "SHOTS_LOG="+logged,
		"SHOTS_SIM="+sims, "SCREENSHOTS_LOCK="+filepath.Join(bin, "lock"))
	out, code := runScript(t, app, "screenshots.sh", env...)
	calls, err := os.ReadFile(logged)
	mustDo(t, err)
	return out, code, string(calls)
}

// failsLeavingNone wants a failed run that made this many simulators and deleted each one.
func failsLeavingNone(t *testing.T, code int, calls string, made int) {
	t.Helper()
	if code == 0 {
		t.Errorf("the run exited 0:\n%s", calls)
	}
	got := strings.Count(calls, "simctl create") + strings.Count(calls, "simctl clone")
	if gone := strings.Count(calls, "simctl delete"); got != made || gone != made {
		t.Errorf("made %d simulators and deleted %d, want %d of each:\n%s", got, gone, made, calls)
	}
}
