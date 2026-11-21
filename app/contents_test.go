package app

import (
	"fmt"
	"testing"
)

func TestGroupOfEdges(t *testing.T) {
	for in, want := range map[string]string{
		"lib/x":                    "lib/",
		"assets/bin/Data":          "assets/bin/",
		"assets/bin/Data/x":        "assets/bin/Data/",
		"assets/aa/Android/x.json": "assets/aa/",
		"other/dir/file":           "other/",
		"Contents/MacOS/Game":      "Contents/MacOS/",
	} {
		if got := groupOf(in); got != want {
			t.Errorf("groupOf(%s) = %s, want %s", in, got, want)
		}
	}
}

func TestReadContents(t *testing.T) {
	_, files := openZip(t, map[string][]byte{
		"root/lib/arm64/a.so": []byte("12345"), "root/lib/arm64/b.so": []byte("123"), "root/classes.dex": []byte("1234567"),
		"root/": nil, "other/skipped": []byte("x"),
	})
	c := readContents(files, "root/")
	if c.Files != 3 || c.Size != 15 {
		t.Errorf("contents = %+v", c)
	}
	if len(c.Groups) != 2 || c.Groups[0].Path != "lib/arm64/" || c.Groups[0].Files != 2 || c.Groups[0].Size != 8 {
		t.Errorf("groups = %+v", c.Groups)
	}
	if len(c.Largest) != 3 || c.Largest[0].Path != "classes.dex" || c.Largest[0].Files != 0 {
		t.Errorf("largest = %+v", c.Largest)
	}
}

func TestReadContentsOfNothing(t *testing.T) {
	c := readContents(nil, "")
	if c.Files != 0 || c.Groups != nil || len(c.Largest) != 0 {
		t.Errorf("contents = %+v", c)
	}
}

func TestReadContentsSumsUpSmallGroupsAndLimitsTheLargest(t *testing.T) {
	files := map[string][]byte{}
	for i := 0; i < maxGroups+5; i++ {
		files[fmt.Sprintf("file%02d", i)] = make([]byte, i+1)
	}
	_, entries := openZip(t, files)
	c := readContents(entries, "")
	if len(c.Groups) != maxGroups || c.Groups[maxGroups-1].Path != "other" || c.Groups[maxGroups-1].Files != 6 {
		t.Errorf("groups = %+v", c.Groups)
	}
	if c.Groups[0].Size != maxGroups+5 {
		t.Errorf("biggest group = %+v", c.Groups[0])
	}
	if len(c.Largest) != maxLargest {
		t.Errorf("largest = %d files", len(c.Largest))
	}
}
