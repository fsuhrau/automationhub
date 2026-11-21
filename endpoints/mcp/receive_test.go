package mcp

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
)

func TestReceiveWritesTheBodyIntoAFreshFolderThatCleanupRemoves(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "store")

	path, cleanup, err := receiver{dir: dir, limit: 10}.receive("a.apk", strings.NewReader("data"))

	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(path); string(got) != "data" {
		t.Errorf("file content %q", got)
	}
	cleanup()
	if left, _ := os.ReadDir(dir); len(left) != 0 {
		t.Errorf("leftovers %v", left)
	}
}

func TestReceiveRefusesWhatItCannotStore(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	_ = os.WriteFile(file, nil, 0o644)
	tests := []struct {
		name string
		r    receiver
		file string
		body string
		want string
	}{
		{"folder cannot be created", receiver{dir: filepath.Join(file, "x"), limit: 10}, "a", "x", ""},
		{"file cannot be created", receiver{dir: t.TempDir(), limit: 10}, "no/such/dir/a", "x", ""},
		{"empty", receiver{dir: t.TempDir(), limit: 10}, "a", "", "the upload is empty"},
		{"too large", receiver{dir: t.TempDir(), limit: 3}, "a", "abcd", "upload is larger than 3 bytes"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := tc.r.receive(tc.file, strings.NewReader(tc.body))

			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestReceiveReportsAFailingBody(t *testing.T) {
	boom := errors.New("boom")

	_, _, err := receiver{dir: t.TempDir(), limit: 10}.receive("a", iotest.ErrReader(boom))

	if !errors.Is(err, boom) || !strings.HasPrefix(err.Error(), "receive file: ") {
		t.Fatalf("err %v", err)
	}
}
