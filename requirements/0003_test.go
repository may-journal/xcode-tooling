package requirements

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// screenshotConfig is an app's fastlane/screenshots.json with one size.
const screenshotConfig = `{"label": "org.example.shots", "project": "App.xcodeproj", "scheme": "App",
"bundle": "org.example.App", "test": "Tests/Shots", "grant": [], "output": "shots",
"requirement": "docs/requirements/0001-shots.md",
"sizes": [{"folder": "phone", "device": "Phone", "width": 10, "height": 20}]}`

// screenshotDoc names two frames.
const screenshotDoc = `- 0001.1
  - Given Guest has photos
    - When I look at the Day view
      - Then the store shows it under "Say what you saw."
- 0001.2
  - Given Guest has photos
    - When I look at the Editor
      - Then the store shows it under "Your words stay."
`

// standIn is a fake xcrun and xcodebuild that logs each call and plays one simulator.
const standIn = `#!/bin/sh
echo "$(basename "$0") $*" >>"$SHOTS_LOG"
case "$*" in
*"list devices booted"*) [ -z "$SHOTS_BOOTED" ] || echo "Phone (ID) (Booted)" ;;
*"list -j runtimes"*) echo '{"runtimes":[{"isAvailable":true,"identifier":"rt","supportedDeviceTypes":[{"identifier":"com.apple.CoreSimulator.SimDeviceType.Phone"}]}]}' ;;
*"list -j devices"*) if [ -f "$SHOTS_SIM" ]; then echo "{\"devices\":{\"rt\":[{\"name\":\"$(cat "$SHOTS_SIM")\",\"udid\":\"ID\"}]}}"; else echo '{"devices":{}}'; fi ;;
*"simctl create"*) echo "$3" >"$SHOTS_SIM"; echo ID ;;
*"simctl delete"*) rm -f "$SHOTS_SIM" ;;
*"test-without-building"*) exit "$SHOTS_TEST_EXIT" ;;
*swift*) cat "$4" ;;
esac
`

// runCapture runs screenshots.sh in app with the stand-ins first on PATH, and returns its output,
// exit code and call log.
func runCapture(t *testing.T, app string, env ...string) (string, int, string) {
	t.Helper()
	bin := t.TempDir()
	for _, name := range []string{"xcrun", "xcodebuild"} {
		mustDo(t, os.WriteFile(filepath.Join(bin, name), []byte(standIn), 0o755))
	}
	logged := filepath.Join(bin, "calls")
	mustDo(t, os.WriteFile(logged, nil, 0o644))
	env = append(env,
		fmt.Sprintf("PATH=%s:%s", bin, os.Getenv("PATH")), "SHOTS_LOG="+logged,
		"SHOTS_SIM="+filepath.Join(bin, "sim"), "SCREENSHOTS_LOCK="+filepath.Join(bin, "lock"))
	out, code := runScript(t, app, "screenshots.sh", env...)
	calls, err := os.ReadFile(logged)
	mustDo(t, err)
	return out, code, string(calls)
}

func shotApp(t *testing.T) string {
	t.Helper()
	return appCheckout(t, map[string]string{
		"fastlane/screenshots.json":       screenshotConfig,
		"docs/requirements/0001-shots.md": screenshotDoc,
	})
}

func Test0003_1(t *testing.T) {
	t.Parallel()
	out, code, calls := runCapture(t, shotApp(t), "SHOTS_BOOTED=1", "SHOTS_TEST_EXIT=0")
	sees(t, out, code, 75, "1 simulators are booted")
	if strings.Contains(calls, "simctl create") {
		t.Errorf("a busy run made a simulator:\n%s", calls)
	}
}

func Test0003_2(t *testing.T) {
	t.Parallel()
	out, code, _ := runCapture(t, shotApp(t), "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0")
	sees(t, out, code, 0, "1\t01-day-view\tSay what you saw.", "2\t02-editor\tYour words stay.")
}

func Test0003_3(t *testing.T) {
	t.Parallel()
	_, code, calls := runCapture(t, shotApp(t), "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=65")
	if code == 0 {
		t.Errorf("a failing test exited 0:\n%s", calls)
	}
	if made, gone := strings.Count(calls, "simctl create"), strings.Count(calls, "simctl delete"); made != 3 || gone != 3 {
		t.Errorf("made %d simulators and deleted %d, want 3 of each:\n%s", made, gone, calls)
	}
}
