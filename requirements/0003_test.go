package requirements

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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

// runCapture runs screenshots.sh in app with the stand-ins, returning output, exit code and calls.
func runCapture(t *testing.T, app string, env ...string) (string, int, string) {
	t.Helper()
	return runWithStandIns(t, app, "screenshots.sh", env...)
}

// runWithStandIns runs script in app with the stand-ins first on PATH.
func runWithStandIns(t *testing.T, app, script string, env ...string) (string, int, string) {
	t.Helper()
	bin := t.TempDir()
	for _, name := range []string{"xcrun", "xcodebuild", "gh"} {
		mustDo(t, os.WriteFile(filepath.Join(bin, name), []byte(standIn), 0o755))
	}
	logged := filepath.Join(bin, "calls")
	mustDo(t, os.WriteFile(logged, nil, 0o644))
	env = append(env,
		fmt.Sprintf("PATH=%s:%s", bin, os.Getenv("PATH")), "SHOTS_LOG="+logged,
		"SHOTS_SIM="+filepath.Join(bin, "sim"), "SCREENSHOTS_LOCK="+filepath.Join(bin, "lock"))
	out, code := runScript(t, app, script, env...)
	calls, err := os.ReadFile(logged)
	mustDo(t, err)
	return out, code, string(calls)
}

// watchedApp is an app clone with a pushed first commit, a captured state and a log of captures.
type watchedApp struct {
	origin, checkout, state, captures string
}

// newWatchedApp makes the app, marks its first commit captured and returns its paths.
func newWatchedApp(t *testing.T) watchedApp {
	t.Helper()
	files := map[string]string{
		"fastlane/screenshots.json":       strings.Replace(screenshotConfig, `"grant": [],`, `"grant": [], "watch": ["src"],`, 1),
		"docs/requirements/0001-shots.md": screenshotDoc,
		"src/App.swift":                   "first\n",
		"README.md":                       "first\n",
	}
	origin := appCheckout(t, files)
	captures := filepath.Join(t.TempDir(), "captures")
	shim := "#!/bin/sh\necho \"$1\" >>\"$SHOTS_CAPTURES\"\n"
	mustDo(t, os.MkdirAll(filepath.Join(origin, "ci_scripts"), 0o755))
	mustDo(t, os.WriteFile(filepath.Join(origin, "ci_scripts", "xcode-tooling.sh"), []byte(shim), 0o755))
	gitIn(t, origin, "init", "-q", "-b", "main")
	commitAll(t, origin)
	parent := t.TempDir()
	checkout := filepath.Join(parent, "app")
	gitIn(t, parent, "clone", "-q", origin, checkout)
	state := t.TempDir()
	first := strings.TrimSpace(gitOut(t, checkout, "rev-parse", "HEAD"))
	mustDo(t, os.WriteFile(filepath.Join(state, "captured"), []byte(first+"\n"), 0o644))
	return watchedApp{origin, checkout, state, captures}
}

// merges commits a change to path on the origin's main.
func (a watchedApp) merges(t *testing.T, path string) {
	t.Helper()
	mustDo(t, os.WriteFile(filepath.Join(a.origin, filepath.FromSlash(path)), []byte("changed\n"), 0o644))
	commitAll(t, a.origin)
}

// polls runs one agent poll, with SHOTS_BOOTED set when booted, and returns the captures so far.
func (a watchedApp) polls(t *testing.T, booted bool) int {
	t.Helper()
	sim := "SHOTS_BOOTED="
	if booted {
		sim = "SHOTS_BOOTED=1"
	}
	out, code, _ := runWithStandIns(t, a.checkout, "screenshots-watch.sh", sim, "SHOTS_TEST_EXIT=0",
		"SCREENSHOTS_CHECKOUT="+a.checkout, "SCREENSHOTS_STATE="+a.state, "SHOTS_CAPTURES="+a.captures)
	if code != 0 {
		t.Fatalf("screenshots-watch.sh exited %d:\n%s", code, out)
	}
	done, err := os.ReadFile(a.captures)
	if errors.Is(err, os.ErrNotExist) {
		return 0
	}
	mustDo(t, err)
	return strings.Count(string(done), "screenshots.sh")
}

func commitAll(t *testing.T, dir string) {
	t.Helper()
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-q", "-m", "change")
}

// gitOut runs git in dir and returns its output.
func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	mustDo(t, err)
	return string(out)
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

func Test0003_4(t *testing.T) {
	t.Parallel()
	app := newWatchedApp(t)
	app.merges(t, "README.md")
	if got := app.polls(t, false); got != 0 {
		t.Errorf("captured %d times for a change to no watched path", got)
	}
}

func Test0003_5(t *testing.T) {
	t.Parallel()
	app := newWatchedApp(t)
	app.merges(t, "src/App.swift")
	app.polls(t, false)
	if got := app.polls(t, false); got != 1 {
		t.Errorf("two polls captured %d times, want 1", got)
	}
}

func Test0003_6(t *testing.T) {
	t.Parallel()
	app := newWatchedApp(t)
	app.merges(t, "src/App.swift")
	if got := app.polls(t, true); got != 0 {
		t.Errorf("a busy poll captured %d times", got)
	}
	if got := app.polls(t, false); got != 1 {
		t.Errorf("the next poll captured %d times, want 1", got)
	}
}

// releaseApp is a clean release commit of an app.
func releaseApp(t *testing.T) string {
	t.Helper()
	app := shotApp(t)
	gitIn(t, app, "init", "-q")
	commitAll(t, app)
	return app
}

func Test0003_7(t *testing.T) {
	t.Parallel()
	out, code, _ := runWithStandIns(t, releaseApp(t), "screenshots-release.sh", "SHOTS_BOOTED=1", "SHOTS_TEST_EXIT=0")
	sees(t, out, code, 1, "the release stops before any upload")
}

func Test0003_8(t *testing.T) {
	t.Parallel()
	app := releaseApp(t)
	out, code, _ := runWithStandIns(t, app, "screenshots-release.sh", "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0")
	sees(t, out, code, 0, "images come from "+strings.TrimSpace(gitOut(t, app, "rev-parse", "--short", "HEAD")))
}
