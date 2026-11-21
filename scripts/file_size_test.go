package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeLines stores a file of n lines below root.
func writeLines(t *testing.T, root, name string, n int) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("x\n", n)), 0o600); err != nil {
		t.Fatal(err)
	}
}

// runFileSize runs go-file-size.sh with the args and returns its output and
// exit code.
func runFileSize(t *testing.T, args ...string) (string, int) {
	t.Helper()
	abs, err := filepath.Abs("go-file-size.sh")
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("bash", append([]string{abs}, args...)...).CombinedOutput()
	if exit, ok := err.(*exec.ExitError); ok {
		return string(out), exit.ExitCode()
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(out), 0
}

func TestFileSizeReportsFilesOverTheLimit(t *testing.T) {
	root := t.TempDir()
	writeLines(t, root, "pkg/big.go", 201)
	writeLines(t, root, "pkg/exact.go", 200)
	writeLines(t, root, "pkg/small.go", 3)

	out, exit := runFileSize(t, root)

	if want := "pkg/big.go: 201 lines (max 200)\n"; out != want || exit != 1 {
		t.Fatalf("out %q, exit %d, want %q and 1", out, exit, want)
	}
}

func TestFileSizeIgnoresExemptFiles(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{
		"a_test.go", "model.pb.go", "keys/keycodes.go", "out/probe/main.go",
		".hidden/x.go", "frontend/gen.go", "node_modules/m/x.go", "notes.txt",
	} {
		writeLines(t, root, name, 500)
	}

	out, exit := runFileSize(t, root)

	if out != "" || exit != 0 {
		t.Fatalf("out %q, exit %d", out, exit)
	}
}

func TestFileSizeTakesTheLimitAsArgument(t *testing.T) {
	root := t.TempDir()
	writeLines(t, root, "a.go", 11)

	out, exit := runFileSize(t, root, "10")

	if out != "a.go: 11 lines (max 10)\n" || exit != 1 {
		t.Fatalf("out %q, exit %d", out, exit)
	}
}

func TestFileSizeRejectsBadArguments(t *testing.T) {
	root := t.TempDir()
	tests := []struct {
		name string
		args []string
	}{
		{"limit is no number", []string{root, "many"}},
		{"dir is missing", []string{filepath.Join(root, "nope")}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, exit := runFileSize(t, tc.args...); exit != 2 {
				t.Fatalf("exit %d, want 2", exit)
			}
		})
	}
}

// The repository itself stays within the limit.
func TestFileSizeOfTheRepository(t *testing.T) {
	out, exit := runFileSize(t, "..")

	if exit != 0 {
		t.Fatalf("files over the limit:\n%s", out)
	}
}
