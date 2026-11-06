package requirements

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// twoSizes is an app's config with a size named second, on top of screenshotConfig's one.
func twoSizes(second string) string {
	size := `{"folder": "` + second + `", "device": "Phone", "width": 10, "height": 20}`
	return strings.Replace(screenshotConfig, `"height": 20}]`, `"height": 20}, `+size+`]`, 1)
}

func sizedApp(t *testing.T, second string) string {
	t.Helper()
	return appCheckout(t, map[string]string{
		"fastlane/screenshots.json":       twoSizes(second),
		"docs/requirements/0001-shots.md": screenshotDoc,
	})
}

func Test0004_1(t *testing.T) {
	t.Parallel()
	out, code, calls := runCapture(t, sizedApp(t, "tablet"), "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0", "SHOTS_AT_ONCE=2")
	sees(t, out, code, 0)
	if made := strings.Count(calls, "simctl create"); made != 2 {
		t.Errorf("made %d simulators, want one per size, so neither size retried:\n%s", made, calls)
	}
}

func Test0004_2(t *testing.T) {
	t.Parallel()
	_, code, calls := runCapture(t, sizedApp(t, "bad"), "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0")
	if code == 0 {
		t.Errorf("a failing size exited 0:\n%s", calls)
	}
	made, gone := strings.Count(calls, "simctl create"), strings.Count(calls, "simctl delete")
	if made != 4 || gone != 4 {
		t.Errorf("made %d simulators and deleted %d, want 4 of each:\n%s", made, gone, calls)
	}
}

func Test0004_3(t *testing.T) {
	t.Parallel()
	out, code, calls := runBeside(t, "screenshots-1-phone", shotApp(t), "screenshots.sh", "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0")
	sees(t, out, code, 0)
	if !strings.Contains(calls, "simctl delete sim.left") {
		t.Errorf("the dead run's simulator stayed:\n%s", calls)
	}
}

func Test0004_4(t *testing.T) {
	t.Parallel()
	doc := screenshotDoc + "- 0001.3\n  - Given Guest has photos\n    - When I tap Save\n      - Then it says \"Saved\"\n"
	app := appCheckout(t, map[string]string{"fastlane/screenshots.json": screenshotConfig, "docs/requirements/0001-shots.md": doc})
	out, code, _ := runCapture(t, app, "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0")
	sees(t, out, code, 0, "2\t02-editor")
	if strings.Contains(out, "Saved") {
		t.Errorf("an acceptance with no frame became one:\n%s", out)
	}
}

func Test0004_5(t *testing.T) {
	t.Parallel()
	app := newWatchedApp(t)
	app.merges(t, "ci_scripts/xcode-tooling.version")
	if got := app.polls(t, false); got != 1 {
		t.Errorf("captured %d times for a tooling bump, want 1", got)
	}
}

func Test0004_6(t *testing.T) {
	t.Parallel()
	app := releaseApp(t)
	mustDo(t, os.WriteFile(filepath.Join(app, "New.swift"), []byte("new\n"), 0o644))
	out, code, _ := runWithStandIns(t, app, "screenshots-release.sh", "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0")
	sees(t, out, code, 1, "so the images match the release commit")
}

func Test0004_7(t *testing.T) {
	t.Parallel()
	app := newWatchedApp(t)
	app.merges(t, "src/App.swift")
	hooks := t.TempDir()
	mustDo(t, os.WriteFile(filepath.Join(hooks, "pre-commit"), []byte("#!/bin/sh\nexit 1\n"), 0o755))
	gitIn(t, app.checkout, "config", "core.hooksPath", hooks)
	mustDo(t, os.WriteFile(app.captures+".images", nil, 0o644))
	app.polls(t, false)
	app.polls(t, false)
	app.merges(t, "src/App.swift2")
	if got := app.polls(t, false); got != 2 {
		t.Errorf("captured %d times, want the first change once and the later one once", got)
	}
}
