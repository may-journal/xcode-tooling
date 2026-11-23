package requirements

import (
	"os/exec"
	"testing"
)

func Test0008_1(t *testing.T) {
	t.Parallel()
	// A bound no other test uses, so a timer still asleep on it belongs to this capture.
	out, code, _ := shotsCapture(t, shotApp(t), "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0", "SCREENSHOTS_CLONE_SECONDS=7919")
	sees(t, out, code, 0)
	if left, _ := exec.Command("pgrep", "-f", "sleep 7919").Output(); len(left) != 0 {
		t.Errorf("the clone's timer outlived the capture, as process %s", left)
	}
}
