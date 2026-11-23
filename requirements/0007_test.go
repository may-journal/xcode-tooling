package requirements

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// stopsEarly wants a run that exits code naming what is missing, before any simulator call.
func stopsEarly(t *testing.T, config string, code int, want string) {
	t.Helper()
	files := map[string]string{"docs/requirements/0001-shots.md": screenshotDoc}
	if config != "" {
		files["fastlane/screenshots.json"] = config
	}
	out, got, calls := shotsCapture(t, appCheckout(t, files), "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0")
	sees(t, out, got, code, want)
	if calls != "" {
		t.Errorf("the run reached the simulators before stopping:\n%s", calls)
	}
}

func Test0007_1(t *testing.T) {
	t.Parallel()
	// A PATH holding only dirname, which the script needs to find itself, so jq is missing.
	bin := t.TempDir()
	dirname, err := exec.LookPath("dirname")
	mustDo(t, err)
	mustDo(t, os.Symlink(dirname, filepath.Join(bin, "dirname")))
	out, code := runScript(t, shotApp(t), "screenshots.sh", "PATH="+bin,
		"SCREENSHOTS_LOCK="+filepath.Join(bin, "lock"))
	sees(t, out, code, 69, "jq is not installed")
	if _, err := os.Stat(filepath.Join(bin, "lock")); err == nil {
		t.Error("the run took the lock before stopping")
	}
}

func Test0007_2(t *testing.T) {
	t.Parallel()
	stopsEarly(t, "", 66, "no config at fastlane/screenshots.json")
}

func Test0007_3(t *testing.T) {
	t.Parallel()
	stopsEarly(t, `{"jobs": 1`, 78, "fastlane/screenshots.json is not valid JSON")
}

func Test0007_4(t *testing.T) {
	t.Parallel()
	stopsEarly(t, strings.Replace(screenshotConfig, `"scheme": "App",`, "", 1), 78, "sets no .scheme")
	stopsEarly(t, strings.Replace(screenshotConfig, `"sizes": [{"folder": "phone", "device": "Phone", "width": 10, "height": 20}]`, `"sizes": []`, 1), 78, "sets no .sizes")
}
