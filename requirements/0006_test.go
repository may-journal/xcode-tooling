package requirements

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// taggedApp is a watched app that follows the tag v1 of a tooling repo on this disk.
type taggedApp struct {
	watchedApp
	tooling string
}

// newTaggedApp makes the app and its tooling repo, with the app's head already captured.
// With recorded, the agent also holds the commit v1 names now.
func newTaggedApp(t *testing.T, recorded bool) taggedApp {
	t.Helper()
	tooling := appCheckout(t, map[string]string{"version.txt": "1.0.0\n"})
	gitIn(t, tooling, "init", "-q", "-b", "main")
	commitAll(t, tooling)
	gitIn(t, tooling, "-c", "tag.gpgSign=false", "tag", "v1")
	// Every poll of this test asks the repo on this disk, never the network.
	t.Setenv("XCODE_TOOLING_URL", tooling)
	app := taggedApp{newWatchedApp(t), tooling}
	mustDo(t, os.WriteFile(filepath.Join(app.origin, "ci_scripts", "xcode-tooling.version"), []byte("v1\n"), 0o644))
	commitAll(t, app.origin)
	mustDo(t, os.WriteFile(filepath.Join(app.state, "captured"), []byte(gitOut(t, app.origin, "rev-parse", "HEAD")), 0o644))
	if recorded {
		mustDo(t, os.WriteFile(app.record(), []byte(app.tagged(t)+"\n"), 0o644))
	}
	return app
}

// record is the file where the agent keeps the tooling commit of its last capture.
func (a taggedApp) record() string { return filepath.Join(a.state, "tooling") }

// tagged is the commit v1 names in the tooling repo.
func (a taggedApp) tagged(t *testing.T) string {
	t.Helper()
	return strings.TrimSpace(gitOut(t, a.tooling, "rev-parse", "v1"))
}

// releases moves v1 to a new tooling commit.
func (a taggedApp) releases(t *testing.T) {
	t.Helper()
	mustDo(t, os.WriteFile(filepath.Join(a.tooling, "version.txt"), []byte(a.tagged(t)+"\n"), 0o644))
	commitAll(t, a.tooling)
	gitIn(t, a.tooling, "-c", "tag.gpgSign=false", "tag", "-f", "v1")
}

// asks runs one poll and returns the captures so far.
func (a taggedApp) asks(t *testing.T) int {
	t.Helper()
	return a.polls(t, false)
}

// recorded is the tooling commit the agent holds.
func (a taggedApp) recorded(t *testing.T) string {
	t.Helper()
	held, err := os.ReadFile(a.record())
	mustDo(t, err)
	return strings.TrimSpace(string(held))
}

func Test0006_1(t *testing.T) {
	app := newTaggedApp(t, true)
	app.releases(t)
	app.asks(t)
	if got := app.asks(t); got != 1 {
		t.Errorf("two polls after a release captured %d times, want 1", got)
	}
}

func Test0006_2(t *testing.T) {
	app := newTaggedApp(t, false)
	if got := app.asks(t); got != 0 {
		t.Errorf("the first poll with no record captured %d times", got)
	}
	if got := app.recorded(t); got != app.tagged(t) {
		t.Errorf("the agent recorded %q, want the commit v1 names, %q", got, app.tagged(t))
	}
}

func Test0006_3(t *testing.T) {
	app := newTaggedApp(t, true)
	before := app.recorded(t)
	app.releases(t)
	t.Setenv("XCODE_TOOLING_URL", filepath.Join(app.state, "gone"))
	if got := app.asks(t); got != 0 {
		t.Errorf("a poll that could not reach the tooling repo captured %d times", got)
	}
	if got := app.recorded(t); got != before {
		t.Errorf("a failed lookup changed the record from %q to %q", before, got)
	}
}

func Test0006_4(t *testing.T) {
	app := newTaggedApp(t, true)
	// A shim that exits 3 stands for a capture that fails.
	mustDo(t, os.WriteFile(filepath.Join(app.origin, "ci_scripts", "xcode-tooling.sh"), []byte("#!/bin/sh\necho \"$1\" >>\"$SHOTS_CAPTURES\"\nexit 3\n"), 0o755))
	app.merges(t, "src/App.swift")
	t.Run("0006.4.1 it skips the failed commit", func(t *testing.T) {
		app.asks(t)
		if got := app.asks(t); got != 1 {
			t.Fatalf("a failed commit was captured %d times before any release, want 1", got)
		}
	})
	t.Run("0006.4.2 it tries the commit again after the release", func(t *testing.T) {
		app.releases(t)
		if got := app.asks(t); got != 2 {
			t.Errorf("captured %d times, want the failed commit tried again after the release", got)
		}
	})
}
