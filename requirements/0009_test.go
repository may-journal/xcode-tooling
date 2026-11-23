package requirements

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// launchdStandIn is a fake launchctl that logs each call and keeps the loaded jobs as files.
// Like the real one, it refuses to unload a job that is not loaded, or to load one twice.
const launchdStandIn = `#!/bin/sh
echo "$*" >>"$AGENT_LOG"
case "$1" in
bootout) rm "$AGENT_LOADED/${2##*/}" 2>/dev/null || exit 3 ;;
bootstrap)
	job=$(basename "$3" .plist)
	[ ! -e "$AGENT_LOADED/$job" ] || exit 5
	cp "$3" "$AGENT_LOADED/$job"
	;;
esac
`

// agentMac is a Mac account with an app's main checkout and a stand-in launchd.
type agentMac struct{ home, app, bin string }

// newAgentMac makes the account, with a LaunchAgents folder only when asked.
func newAgentMac(t *testing.T, launchAgents bool) agentMac {
	t.Helper()
	m := agentMac{t.TempDir(), shotApp(t), t.TempDir()}
	mustDo(t, os.WriteFile(filepath.Join(m.bin, "launchctl"), []byte(launchdStandIn), 0o755))
	mustDo(t, os.Mkdir(filepath.Join(m.bin, "loaded"), 0o755))
	if launchAgents {
		mustDo(t, os.MkdirAll(filepath.Dir(m.plist()), 0o755))
	}
	return m
}

func (m agentMac) plist() string {
	return filepath.Join(m.home, "Library", "LaunchAgents", "org.example.shots.plist")
}

func (m agentMac) watchCopy() string {
	return filepath.Join(m.home, "Library", "Application Support", "org.example.shots", "screenshots-watch.sh")
}

// runs runs screenshots-agent.sh from the checkout and returns its output, exit code and
// the launchctl calls so far.
func (m agentMac) runs(t *testing.T, env ...string) (string, int, []string) {
	t.Helper()
	logged := filepath.Join(m.bin, "calls")
	env = append(env, "HOME="+m.home, "PATH="+m.bin+":"+os.Getenv("PATH"),
		"AGENT_LOG="+logged, "AGENT_LOADED="+filepath.Join(m.bin, "loaded"))
	out, code := runScript(t, m.app, "screenshots-agent.sh", env...)
	calls, err := os.ReadFile(logged)
	mustDo(t, err)
	return out, code, strings.Split(strings.TrimSpace(string(calls)), "\n")
}

// loaded is the label of each job launchd holds.
func (m agentMac) loaded(t *testing.T) []string {
	t.Helper()
	jobs, err := os.ReadDir(filepath.Join(m.bin, "loaded"))
	mustDo(t, err)
	var labels []string
	for _, job := range jobs {
		labels = append(labels, job.Name())
	}
	return labels
}

// holdsOneJob wants launchd to hold the config's job and nothing else.
func (m agentMac) holdsOneJob(t *testing.T) {
	t.Helper()
	if got := m.loaded(t); len(got) != 1 || got[0] != "org.example.shots" {
		t.Errorf("launchd holds %v, want only org.example.shots", got)
	}
}

func Test0009_1(t *testing.T) {
	t.Parallel()
	m := newAgentMac(t, true)
	out, code, _ := m.runs(t)
	sees(t, out, code, 0, "org.example.shots polls main every five minutes")
	m.holdsOneJob(t)
	job, err := os.ReadFile(m.plist())
	mustDo(t, err)
	checkout, err := filepath.EvalSymlinks(m.app)
	mustDo(t, err)
	for _, want := range []string{
		"<key>Label</key>\n\t<string>org.example.shots</string>",
		"<string>" + m.watchCopy() + "</string>",
		"<key>SCREENSHOTS_CHECKOUT</key>\n\t\t<string>" + checkout + "</string>",
		"<key>StartInterval</key>\n\t<integer>300</integer>",
	} {
		if !strings.Contains(string(job), want) {
			t.Errorf("the job file is missing %q:\n%s", want, job)
		}
	}
	copied, err := os.ReadFile(m.watchCopy())
	mustDo(t, err)
	watch, err := os.ReadFile(filepath.Join("..", "ci_scripts", "screenshots-watch.sh"))
	mustDo(t, err)
	if string(copied) != string(watch) {
		t.Error("the job's watch script differs from ci_scripts/screenshots-watch.sh")
	}
}

func Test0009_2(t *testing.T) {
	t.Parallel()
	m := newAgentMac(t, false)
	out, code, _ := m.runs(t)
	sees(t, out, code, 0, "org.example.shots polls main")
	m.holdsOneJob(t)
}

func Test0009_3(t *testing.T) {
	t.Parallel()
	m := newAgentMac(t, true)
	m.runs(t)
	out, code, calls := m.runs(t)
	sees(t, out, code, 0, "org.example.shots polls main")
	m.holdsOneJob(t)
	if len(calls) != 4 || !strings.HasPrefix(calls[2], "bootout ") || !strings.HasPrefix(calls[3], "bootstrap ") {
		t.Errorf("the second run did not unload the job before loading it:\n%s", strings.Join(calls, "\n"))
	}
	jobs, err := os.ReadDir(filepath.Dir(m.plist()))
	mustDo(t, err)
	if len(jobs) != 1 {
		t.Errorf("LaunchAgents holds %d files, want 1", len(jobs))
	}
}

func Test0009_4(t *testing.T) {
	t.Parallel()
	m := newAgentMac(t, true)
	m.runs(t)
	out, code, calls := m.runs(t, "SCREENSHOTS_REMOVE=1")
	sees(t, out, code, 0)
	if got := m.loaded(t); len(got) != 0 {
		t.Errorf("launchd still holds %v", got)
	}
	if last := calls[len(calls)-1]; !strings.HasPrefix(last, "bootout ") {
		t.Errorf("the removal ended with %q, want an unload", last)
	}
	for _, path := range []string{m.plist(), filepath.Dir(m.watchCopy())} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s is still there", path)
		}
	}
}
