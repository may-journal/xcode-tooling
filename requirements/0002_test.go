package requirements

import (
	"os"
	"path/filepath"
	"testing"
)

// archive is the environment Xcode Cloud sets after an archive action.
const archive = "CI_ARCHIVE_PATH=/tmp/App.xcarchive"

func Test0002_1(t *testing.T) {
	t.Parallel()
	app := appCheckout(t, map[string]string{
		"changelog/2026-10-20-0900-older.md": "---\nrelatedConfigurations: []\n---\n\n### 2026.10.20.0900\n\n> Older change.\n\n- Fix: old\n",
		"changelog/2026-10-24-1200-newer.md": "---\nrelatedConfigurations: []\n---\n\n### 2026.10.24.1200\n\n> Photos load faster.\n\n- Fix: the grid scrolls smoothly\n- Feat: a share button\n",
	})
	notesAre(t, app, "CI_ARCHIVE_PATH=", "")
	notesAre(t, app, archive, "Photos load faster.\n\n- Fix: the grid scrolls smoothly\n- Feat: a share button\n")

	gitIn(t, app, "init", "-q")
	gitIn(t, app, "add", ".")
	gitIn(t, app, "commit", "-q", "-m", "first")
	mustDo(t, os.WriteFile(filepath.Join(app, "changelog", "2026-10-19-0800-merged-last.md"), []byte("### 2026.10.19.0800\n\n> Merged last.\n\n- Fix: late\n"), 0o644))
	gitIn(t, app, "add", ".")
	gitIn(t, app, "commit", "-q", "-m", "second")
	notesAre(t, app, archive, "Merged last.\n\n- Fix: late\n")

	notesAre(t, appCheckout(t, nil), archive, "")
}

// notesAre runs ci_post_xcodebuild.sh in app with env, then fails the test unless the notes
// it wrote are want, or it wrote none when want is empty. It removes the notes after.
func notesAre(t *testing.T, app, env, want string) {
	t.Helper()
	out, code := runScript(t, app, "ci_post_xcodebuild.sh", env)
	if code != 0 {
		t.Fatalf("ci_post_xcodebuild.sh exited %d:\n%s", code, out)
	}
	dir := filepath.Join(app, "TestFlight")
	got, err := os.ReadFile(filepath.Join(dir, "WhatToTest.en-US.txt"))
	if os.IsNotExist(err) && want == "" {
		return
	}
	mustDo(t, err)
	if string(got) != want {
		t.Errorf("notes = %q, want %q", got, want)
	}
	mustDo(t, os.RemoveAll(dir))
}
