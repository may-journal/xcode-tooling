package requirements

import (
	"os"
	"path/filepath"
	"testing"
)

func Test0001_1(t *testing.T) {
	t.Parallel()
	app := appCheckout(t, map[string]string{"project.yml": "name: App\n", "ci_scripts/app_post_clone.sh": "echo app step ran\n"})
	fake := filepath.Join(t.TempDir(), "xcodegen.zip")
	mustDo(t, os.WriteFile(fake, []byte("not the release"), 0o644))
	// No XcodeGen on PATH, as on a fresh Xcode Cloud machine, so the script downloads one.
	out, code := runScript(t, app, "ci_post_clone.sh", "PATH=/usr/bin:/bin", "XCODEGEN_URL=file://"+fake)
	sees(t, out, code, 1, "app step ran", "xcodegen.zip does not match XcodeGen 2.46.0's SHA-256")
	if _, err := os.Stat(filepath.Join(app, "App.xcodeproj")); !os.IsNotExist(err) {
		t.Errorf("a project was generated: %v", err)
	}

	committed := appCheckout(t, nil)
	out, code = runScript(t, committed, "ci_post_clone.sh", "XCODEGEN_URL=file:///nowhere")
	sees(t, out, code, 0)

	bin := t.TempDir()
	mustDo(t, os.WriteFile(filepath.Join(bin, "xcodegen"), []byte("#!/bin/sh\n[ \"$1\" = --version ] && echo 'Version: 2.46.0' && exit 0\necho generated in \"$PWD\"\n"), 0o755))
	out, code = runScript(t, app, "ci_post_clone.sh", "PATH="+bin+":/usr/bin:/bin", "XCODEGEN_URL=file:///nowhere")
	real, err := filepath.EvalSymlinks(app)
	mustDo(t, err)
	sees(t, out, code, 0, "app step ran", "generated in "+real)
}
