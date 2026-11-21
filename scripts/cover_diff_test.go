package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const script = "go-cover-diff.sh"

// invocation describes one script run.
type invocation struct {
	dir  string
	env  []string
	args []string
}

// run executes the script in dir and returns its output and exit code.
func run(t *testing.T, in invocation) (string, int) {
	t.Helper()
	abs, err := filepath.Abs(script)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", append([]string{abs}, in.args...)...)
	cmd.Dir = in.dir
	cmd.Env = append(os.Environ(), in.env...)
	out, err := cmd.CombinedOutput()
	if exit, ok := err.(*exec.ExitError); ok {
		return string(out), exit.ExitCode()
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(out), 0
}

func TestCoverDiffFixtures(t *testing.T) {
	profile, _ := filepath.Abs("testdata/profile.cov")
	tests := []struct {
		name     string
		diff     string
		files    map[string]string
		wantOut  string
		wantExit int
	}{
		{"covered line", "covered", nil, "", 0},
		{"uncovered line", "uncovered", nil, "pkg/a.go:15: not covered\n", 1},
		{"comment, blank and brace lines", "comment", nil, "", 0},
		{"test file", "testfile", nil, "", 0},
		{"generated file", "generated", nil, "", 0},
		{"type-only file absent from profile", "typeonly", map[string]string{"pkg/t.go": "package pkg\n\ntype T struct{}\n"}, "", 0},
		{"doc-only file absent from profile", "doconly", map[string]string{"pkg/doc.go": "// Package pkg does things.\npackage pkg\n"}, "", 0},
		{"file with func missing from profile", "missing", map[string]string{"pkg/b.go": "package pkg\n\nfunc b() {\n\tz := 3\n}\n"}, "pkg/b.go:3: file missing from coverage profile\n", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, _ := filepath.Abs("testdata/" + tt.diff + ".diff")
			dir := t.TempDir()
			writeTree(t, dir, tt.files)
			out, exit := run(t, invocation{dir, []string{"GO_COVER_DIFF_FILE=" + diff}, []string{"HEAD", profile}})
			if out != tt.wantOut || exit != tt.wantExit {
				t.Fatalf("got (%q, %d), want (%q, %d)", out, exit, tt.wantOut, tt.wantExit)
			}
		})
	}
}

// writeTree creates the given files (relative path to content) below dir.
func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		write(t, path, content)
	}
}

func TestCoverDiffUsage(t *testing.T) {
	out, exit := run(t, invocation{dir: t.TempDir(), args: []string{"HEAD"}})
	if exit != 2 || !strings.Contains(out, "usage:") {
		t.Fatalf("got (%q, %d), want usage and exit 2", out, exit)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCoverDiffWorkingTree covers the real git path: unstaged edits of tracked
// files and untracked files are both diffed against the base.
func TestCoverDiffWorkingTree(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	git(t, dir, "config", "user.email", "t@example.com")
	git(t, dir, "config", "user.name", "t")
	write(t, filepath.Join(dir, "a.go"), "package p\n\nfunc a() {\n\tprintln(1)\n}\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "base")
	write(t, filepath.Join(dir, "a.go"), "package p\n\nfunc a() {\n\tprintln(1)\n\tprintln(2)\n}\n")
	write(t, filepath.Join(dir, "b.go"), "package p\n\nfunc b() {\n\tprintln(3)\n}\n")
	profile := "mode: set\n" +
		"github.com/fsuhrau/automationhub/a.go:3.10,6.2 2 0\n" +
		"github.com/fsuhrau/automationhub/b.go:3.10,5.2 1 1\n"
	write(t, filepath.Join(dir, "p.cov"), profile)

	tests := []struct {
		name     string
		dirs     []string
		wantOut  string
		wantExit int
	}{
		{"whole repo", nil, "a.go:5: not covered\n", 1},
		{"explicit root dir", []string{"."}, "a.go:5: not covered\n", 1},
		{"other dir", []string{"other/"}, "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"HEAD", "p.cov"}, tt.dirs...)
			out, exit := run(t, invocation{dir: dir, args: args})
			if out != tt.wantOut || exit != tt.wantExit {
				t.Fatalf("got (%q, %d), want (%q, %d)", out, exit, tt.wantOut, tt.wantExit)
			}
		})
	}
}
