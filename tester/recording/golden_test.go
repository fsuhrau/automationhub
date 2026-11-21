package recording

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//nolint:gochecknoglobals // a test flag is registered once, at package level
var update = flag.Bool("update", false, "rewrite the golden files in testdata")

// assertGolden compares got with testdata/<name>.golden, or rewrites the file
// with -update.
func assertGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s differs from its golden file (rerun with -update to accept)\n got: %x\nwant: %x", name, got, want)
	}
}

// spsCase builds one SPS bit by bit.
type spsCase struct {
	name  string
	build func(w *bitWriter)
}

// header writes profile, constraint flags, level and the sps id.
func header(w *bitWriter, profile uint) {
	w.put(profile, 8)
	w.put(0, 8)
	w.put(31, 8)
	w.ue(0)
}

// highFields writes the fields of the high profiles up to the scaling matrix flag.
func highFields(w *bitWriter, chroma uint) {
	w.ue(chroma)
	if chroma == 3 {
		w.put(0, 1)
	}
	w.ue(0)
	w.ue(0)
	w.put(0, 1)
}

// spsPicture is the picture part of a test SPS.
type spsPicture struct {
	mbW, mbH     uint
	frameMbsOnly uint
	crop         []uint
}

// tail writes the fields after the picture order count type.
func tail(w *bitWriter, p spsPicture) {
	w.ue(1)
	w.put(0, 1)
	w.ue(p.mbW - 1)
	w.ue(p.mbH - 1)
	w.put(p.frameMbsOnly, 1)
	if p.frameMbsOnly == 0 {
		w.put(1, 1)
	}
	w.put(1, 1)
	if p.crop == nil {
		w.put(0, 1)
		return
	}
	w.put(1, 1)
	for _, c := range p.crop {
		w.ue(c)
	}
}

// scalingListsSPS has scaling lists of 4x4 and 8x8 blocks; one ends early.
func scalingListsSPS(w *bitWriter) {
	header(w, 100)
	w.ue(1)
	w.ue(0)
	w.ue(0)
	w.put(0, 1)
	w.put(1, 1) // scaling matrix present
	for i := 0; i < 8; i++ {
		scalingListBits(w, i)
	}
	w.ue(0)
	w.ue(2)
	tail(w, spsPicture{20, 15, 1, nil})
}

// scalingListBits writes list i: odd ones are absent, even ones ramp up and
// end early with a delta that brings next to 0.
func scalingListBits(w *bitWriter, i int) {
	if i%2 == 1 {
		w.put(0, 1)
		return
	}
	w.put(1, 1)
	for j := 0; j < 3; j++ {
		w.se(1)
	}
	w.se(-11)
}

func spsCases() []spsCase {
	return append(validCases(), truncatedCases()...)
}

func validCases() []spsCase {
	return []spsCase{
		{"baseline 320x240", func(w *bitWriter) {
			header(w, 66)
			w.ue(0)
			w.ue(2)
			tail(w, spsPicture{20, 15, 1, nil})
		}},
		{"main poc0 640x480", func(w *bitWriter) {
			header(w, 77)
			w.ue(3)
			w.ue(0)
			w.ue(6)
			tail(w, spsPicture{40, 30, 1, nil})
		}},
		{"main poc1 with offsets", func(w *bitWriter) {
			header(w, 77)
			w.ue(0)
			w.ue(1)
			w.put(1, 1)
			w.ue(3)
			w.ue(4)
			w.ue(2)
			w.ue(1)
			w.ue(2)
			tail(w, spsPicture{45, 30, 1, nil})
		}},
		{"main interlaced 720x576", func(w *bitWriter) {
			header(w, 77)
			w.ue(0)
			w.ue(2)
			tail(w, spsPicture{45, 18, 0, nil})
		}},
		{"high 420 cropped 1920x1080", func(w *bitWriter) {
			header(w, 100)
			highFields(w, 1)
			w.put(0, 1)
			w.ue(0)
			w.ue(2)
			tail(w, spsPicture{120, 68, 1, []uint{0, 0, 0, 4}})
		}},
		{"high 422 cropped", func(w *bitWriter) {
			header(w, 122)
			highFields(w, 2)
			w.put(0, 1)
			w.ue(0)
			w.ue(2)
			tail(w, spsPicture{45, 30, 1, []uint{1, 2, 3, 4}})
		}},
		{"high 444 interlaced cropped", func(w *bitWriter) {
			header(w, 244)
			highFields(w, 3)
			w.put(0, 1)
			w.ue(0)
			w.ue(2)
			tail(w, spsPicture{45, 15, 0, []uint{1, 2, 3, 4}})
		}},
		{"high with scaling lists", scalingListsSPS},
		{"high 444 scaling lists", func(w *bitWriter) {
			header(w, 100)
			w.ue(3)
			w.put(0, 1)
			w.ue(0)
			w.ue(0)
			w.put(0, 1)
			w.put(1, 1)
			for i := 0; i < 12; i++ {
				w.put(0, 1)
			}
			w.ue(0)
			w.ue(2)
			tail(w, spsPicture{20, 15, 1, nil})
		}},
	}
}

func truncatedCases() []spsCase {
	return []spsCase{
		{"truncated after the level", func(w *bitWriter) {
			w.put(66, 8)
			w.put(0, 8)
			w.put(31, 3)
		}},
		{"truncated in the scaling matrix", func(w *bitWriter) {
			header(w, 100)
			w.ue(1)
			w.ue(0)
			w.ue(0)
			w.put(0, 1)
			w.put(1, 1)
			w.put(1, 1)
			w.put(0, 3)
		}},
		{"truncated in the cropping", func(w *bitWriter) {
			header(w, 66)
			w.ue(0)
			w.ue(2)
			tail(w, spsPicture{20, 15, 1, []uint{1, 1}})
		}},
		{"truncated in the offsets cycle", func(w *bitWriter) {
			header(w, 66)
			w.ue(0)
			w.ue(1)
			w.put(0, 1)
			w.ue(0)
			w.ue(0)
			w.ue(200)
		}},
	}
}

// Parsing SPS bytes gives the same sizes, or the same errors, as before the
// split of spsSize.
func TestSPSSizeGolden(t *testing.T) {
	var out strings.Builder
	for _, c := range spsCases() {
		w := &bitWriter{}
		c.build(w)
		sps := append([]byte{0x67}, w.bytes()...)
		width, height, err := spsSize(sps)
		fmt.Fprintf(&out, "%s\n  %s\n  %dx%d err=%v\n", c.name, hex.EncodeToString(sps), width, height, err)
	}
	for _, short := range [][]byte{nil, {0x67, 1, 2}} {
		_, _, err := spsSize(short)
		fmt.Fprintf(&out, "short %x err=%v\n", short, err)
	}
	// emulation prevention bytes are removed before parsing
	e := &bitWriter{}
	e.put(66, 8)
	e.put(0, 16) // constraint flags and level 0: two zero bytes
	e.ue(0)
	e.ue(0)
	e.ue(2)
	tail(e, spsPicture{20, 15, 1, nil})
	raw := e.bytes()
	emu := append(append([]byte{0x67}, raw[:3]...), append([]byte{0x03}, raw[3:]...)...)
	w, h, err := spsSize(emu)
	fmt.Fprintf(&out, "emulation %x %dx%d err=%v\n", emu, w, h, err)
	assertGolden(t, "sps", []byte(out.String()))
}

// The whole MP4 file of a fixed input stays byte for byte the same.
func TestMP4WriterGolden(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.mp4")
	sps := testSPS(720, 1280)
	w, err := newMP4Writer(osFS{}, path, h264Config{sps: sps, pps: testPPS})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []struct {
		time int64
		key  bool
	}{{0, true}, {40000, false}, {40000, false}, {80000, false}, {1080000, true}, {1100000, false}} {
		data := annexB([]byte{0x09, 0xf0}, []byte{0x61, 1, 2, 3})
		if f.key {
			data = annexB(sps, testPPS, []byte{0x65, 1, 2, 3, 4})
		}
		if err := w.WriteSample(data, f.time, f.key); err != nil {
			t.Fatal(err)
		}
	}
	moov := w.moov()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	assertGolden(t, "mp4", data)
	assertGolden(t, "moov", moov)
}

// A single sample has no successor, so it lasts a 30th of a second.
func TestMP4SingleSampleGolden(t *testing.T) {
	w := &mp4Writer{sps: testSPS(720, 1280)[:], pps: testPPS, width: 720, height: 1280,
		samples: []mp4Sample{{offset: 48, size: 9, time: 0, key: true}}}
	assertGolden(t, "moov_single", w.moov())
}

// se writes a signed Exp-Golomb value.
func (w *bitWriter) se(v int) {
	if v > 0 {
		w.ue(uint(2*v - 1))
		return
	}
	w.ue(uint(-2 * v))
}
