package control

import (
	"reflect"
	"testing"
)

func TestLogCursor(t *testing.T) {
	c := &logCursor{}
	first := c.next([]string{
		"--------- beginning of main",
		"11-09 12:00:00.100 I/A( 1): one",
		"11-09 12:00:00.200 I/A( 1): two",
		"11-09 12:00:00.200 I/A( 1): two",
		"    at continued",
	})
	want := []string{
		"11-09 12:00:00.100 I/A( 1): one",
		"11-09 12:00:00.200 I/A( 1): two",
		"11-09 12:00:00.200 I/A( 1): two",
		"    at continued",
	}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("first read %q", first)
	}
	if c.last != "11-09 12:00:00.200" {
		t.Fatalf("cursor at %q", c.last)
	}

	// a read since the last time repeats its lines; a third identical line
	// is new, and so are later ones
	second := c.next([]string{
		"11-09 12:00:00.200 I/A( 1): two",
		"11-09 12:00:00.200 I/A( 1): two",
		"    at continued",
		"11-09 12:00:00.200 I/A( 1): two",
		"11-09 12:00:00.300 W/B( 2): three",
	})
	want = []string{
		"11-09 12:00:00.200 I/A( 1): two",
		"11-09 12:00:00.300 W/B( 2): three",
	}
	if !reflect.DeepEqual(second, want) {
		t.Fatalf("second read %q", second)
	}

	// an older node returns the recent lines instead: old ones are dropped
	third := c.next([]string{
		"11-09 12:00:00.100 I/A( 1): one",
		"11-09 12:00:00.300 W/B( 2): three",
		"11-09 12:00:01.000 E/C( 3): four",
	})
	if !reflect.DeepEqual(third, []string{"11-09 12:00:01.000 E/C( 3): four"}) {
		t.Fatalf("third read %q", third)
	}

	if got := c.next(nil); got != nil {
		t.Fatalf("empty read %q", got)
	}
}
