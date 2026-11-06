package requirements

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// threeSizes is an app's config with sizes named second and third, on top of screenshotConfig's.
func threeSizes(t *testing.T, second, third string) string {
	t.Helper()
	var config map[string]any
	mustDo(t, json.Unmarshal([]byte(screenshotConfig), &config))
	sizes := config["sizes"].([]any)
	for _, name := range []string{second, third} {
		sizes = append(sizes, map[string]any{"folder": name, "device": "Phone", "width": 10, "height": 20})
	}
	config["sizes"] = sizes
	body, err := json.Marshal(config)
	mustDo(t, err)
	return string(body)
}

func Test0005_1(t *testing.T) {
	t.Parallel()
	out, code, calls := runCapture(t, sizedApp(t, "tablet"), "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0")
	sees(t, out, code, 0)
	fresh, clones := strings.Count(calls, "simctl create"), strings.Count(calls, "simctl clone")
	if fresh != 1 || clones != 2 {
		t.Errorf("booted the device fresh %d times with %d clones, want once with 2:\n%s", fresh, clones, calls)
	}
}

func Test0005_2(t *testing.T) {
	t.Parallel()
	_, code, calls := runCapture(t, sizedApp(t, "tablet"), "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0", "SHOTS_BOOT_EXIT=1")
	if code == 0 {
		t.Errorf("a first boot that never finished exited 0:\n%s", calls)
	}
	if made, gone := strings.Count(calls, "simctl create"), strings.Count(calls, "simctl delete"); made != 3 || gone != 3 {
		t.Errorf("made %d simulators and deleted %d, want 3 tries, each deleted:\n%s", made, gone, calls)
	}
}

func Test0005_3(t *testing.T) {
	t.Parallel()
	out, code, calls := runCapture(t, shotApp(t), "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0", "SHOTS_CUT_WRITE=1")
	sees(t, out, code, 0)
	if made := strings.Count(calls, "simctl clone"); made != 1 {
		t.Errorf("cloned %d times, want the first clone to start the size:\n%s", made, calls)
	}
}

func Test0005_7(t *testing.T) {
	t.Parallel()
	_, code, calls := runCapture(t, shotApp(t), "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0", "SHOTS_CLONE_EXIT=22")
	if code == 0 {
		t.Errorf("a device that can't be cloned exited 0:\n%s", calls)
	}
	if tries := strings.Count(calls, "simctl clone"); tries != 3 {
		t.Errorf("tried to clone %d times, want 3:\n%s", tries, calls)
	}
	if made, gone := strings.Count(calls, "simctl create"), strings.Count(calls, "simctl delete"); made != 1 || gone != 1 {
		t.Errorf("made %d simulators and deleted %d, want the template, deleted:\n%s", made, gone, calls)
	}
}

func Test0005_4(t *testing.T) {
	t.Parallel()
	app := appCheckout(t, map[string]string{
		"fastlane/screenshots.json":       threeSizes(t, "slow", "last"),
		"docs/requirements/0001-shots.md": screenshotDoc,
	})
	out, code, calls := runCapture(t, app, "SHOTS_BOOTED=", "SHOTS_TEST_EXIT=0", "SCREENSHOTS_JOBS=2")
	sees(t, out, code, 0)
	if made := strings.Count(calls, "simctl clone"); made != 3 {
		t.Errorf("cloned %d simulators, want 3, so the slow size never waited out its try:\n%s", made, calls)
	}
}

// composes runs the real compose tool over one 100 by 200 size named phone, on a Mac.
func composes(t *testing.T, root string, frames ...string) (string, int) {
	t.Helper()
	if _, err := exec.LookPath("xcrun"); err != nil {
		t.Skip("the compose tool draws with CoreGraphics, which only a Mac has")
	}
	tool, err := filepath.Abs(filepath.Join("..", "ci_scripts", "screenshots-compose.swift"))
	mustDo(t, err)
	config := `{"background": [0.1, 0.2, 0.3], "sizes": [{"folder": "phone", "width": 100, "height": 200}]}`
	configPath := filepath.Join(root, "config.json")
	mustDo(t, os.WriteFile(configPath, []byte(config), 0o644))
	framesPath := filepath.Join(root, "frames.tsv")
	mustDo(t, os.WriteFile(framesPath, []byte(strings.Join(frames, "\n")+"\n"), 0o644))
	cmd := exec.Command("xcrun", "swift", tool, configPath, framesPath, filepath.Join(root, "raw"), filepath.Join(root, "out"))
	out, runErr := cmd.CombinedOutput()
	if _, isExit := runErr.(*exec.ExitError); runErr != nil && !isExit {
		t.Fatalf("starting the compose tool: %v", runErr)
	}
	return string(out), cmd.ProcessState.ExitCode()
}

// writesCapture saves a noisy 100 by 200 capture of a frame, so it reads as a real screen.
func writesCapture(t *testing.T, root string, number int) {
	t.Helper()
	picture := image.NewGray(image.Rect(0, 0, 100, 200))
	random := rand.New(rand.NewSource(int64(number)))
	random.Read(picture.Pix)
	var file bytes.Buffer
	mustDo(t, png.Encode(&file, picture))
	path := filepath.Join(root, "raw", "phone", fmt.Sprintf("%d.png", number))
	mustDo(t, os.MkdirAll(filepath.Dir(path), 0o755))
	mustDo(t, os.WriteFile(path, file.Bytes(), 0o644))
}

func Test0005_5(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	frames := []string{"1\t01-day\tSay what you saw.", "2\t02-editor\tYour words stay."}
	writesCapture(t, root, 1)
	writesCapture(t, root, 2)
	out, code := composes(t, root, frames...)
	sees(t, out, code, 0)
	first := map[string][]byte{}
	for _, name := range []string{"01-day.png", "02-editor.png"} {
		image, err := os.ReadFile(filepath.Join(root, "out", "phone", name))
		mustDo(t, err)
		first[name] = image
	}
	out, code = composes(t, root, frames...)
	sees(t, out, code, 0)
	for name, before := range first {
		after, err := os.ReadFile(filepath.Join(root, "out", "phone", name))
		mustDo(t, err)
		if !bytes.Equal(before, after) {
			t.Errorf("%s changed on a rerun over the same captures", name)
		}
	}
}

func Test0005_6(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writesCapture(t, root, 1)
	out, code := composes(t, root, "1\t01-day\tSay what you saw.", "2\t02-editor\tYour words stay.")
	sees(t, out, code, 1, "no capture at", "phone/2.png")
}
