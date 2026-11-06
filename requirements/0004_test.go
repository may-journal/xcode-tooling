package requirements

import (
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
