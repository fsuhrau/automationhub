package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnzip(t *testing.T) {
	src := filepath.Join(t.TempDir(), "a.zip")
	writeZip(t, src, map[string][]byte{"dir/": nil, "dir/sub/file.txt": []byte("hello"), "top.txt": []byte("top")})
	dest := t.TempDir()
	names, err := Unzip(src, dest)
	if err != nil || len(names) != 3 {
		t.Fatalf("names = %v, err = %v", names, err)
	}
	if data, err := os.ReadFile(filepath.Join(dest, "dir", "sub", "file.txt")); err != nil || string(data) != "hello" {
		t.Errorf("data = %q, err = %v", data, err)
	}
}

func TestUnzipRejectsAZipSlip(t *testing.T) {
	src := filepath.Join(t.TempDir(), "evil.zip")
	writeZip(t, src, map[string][]byte{"../evil.txt": []byte("x")})
	dest := filepath.Join(t.TempDir(), "out")
	if _, err := Unzip(src, dest); err == nil {
		t.Error("no error")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dest), "evil.txt")); err == nil {
		t.Error("the file was written outside of dest")
	}
}

func TestUnzipErrors(t *testing.T) {
	if _, err := Unzip(filepath.Join(t.TempDir(), "none.zip"), t.TempDir()); err == nil {
		t.Error("a missing zip is an error")
	}
	src := filepath.Join(t.TempDir(), "a.zip")
	writeZip(t, src, map[string][]byte{"x/file": []byte("x")})
	blocked := t.TempDir()
	writeFile(t, filepath.Join(blocked, "x"), []byte("a file where the folder should be"))
	if _, err := Unzip(src, blocked); err == nil {
		t.Error("no folder can be made")
	}
	writeFile(t, filepath.Join(blocked, "y"), nil)
	srcDirOverFile := filepath.Join(t.TempDir(), "b.zip")
	writeZip(t, srcDirOverFile, map[string][]byte{"y/": nil})
	if _, err := Unzip(srcDirOverFile, blocked); err == nil {
		t.Error("a folder cannot replace a file")
	}
}

func TestExtractEntryCannotOpenTheEntry(t *testing.T) {
	_, files := openZip(t, nil, "broken")
	if err := extractEntry(files[0], filepath.Join(t.TempDir(), "broken")); err == nil {
		t.Error("no error")
	}
}

func TestExtractEntryCannotCreateTheFile(t *testing.T) {
	_, files := openZip(t, map[string][]byte{"f": []byte("x")})
	if err := extractEntry(files[0], t.TempDir()); err == nil {
		t.Error("a folder cannot be written as file")
	}
}
