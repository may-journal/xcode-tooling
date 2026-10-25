// Package requirements proves docs/requirements the way Xcode Cloud meets them: each test
// builds an app checkout in a temporary folder and runs this repo's scripts from its root
// with sh, as an app's shim does. It asserts only what that acceptance's Then promises.
package requirements

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// appCheckout is an app's root holding files, each path relative to it.
func appCheckout(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		mustDo(t, os.MkdirAll(filepath.Dir(path), 0o755))
		mustDo(t, os.WriteFile(path, []byte(body), 0o644))
	}
	return root
}

// runScript runs this repo's ci_scripts/name from app, with env added to the test's own,
// and returns its trimmed output and exit code.
func runScript(t *testing.T, app, name string, env ...string) (string, int) {
	t.Helper()
	script, err := filepath.Abs(filepath.Join("..", "ci_scripts", name))
	mustDo(t, err)
	cmd := exec.Command("sh", script)
	cmd.Dir = app
	cmd.Env = append(os.Environ(), env...)
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), cmd.ProcessState.ExitCode()
}

// gitIn runs git in dir, a test fixture repo, as a throwaway author.
func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// sees fails the test unless out holds every want and the exit code is code.
func sees(t *testing.T, out string, gotCode, code int, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("output is missing %q:\n%s", w, out)
		}
	}
	if gotCode != code {
		t.Errorf("exit code = %d, want %d:\n%s", gotCode, code, out)
	}
}

func mustDo(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
