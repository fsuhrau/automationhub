package recording

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSplitAnnexBEdges(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want [][]byte
	}{
		{"empty", nil, nil},
		{"no start code", []byte{1, 2, 3}, [][]byte{{1, 2, 3}}},
		{"short", []byte{1}, [][]byte{{1}}},
		{"start code only", []byte{0, 0, 1}, nil},
		{"4-byte start codes", []byte{0, 0, 0, 1, 0x67, 5, 0, 0, 0, 1, 0x68}, [][]byte{{0x67, 5}, {0x68}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitAnnexB(tt.data)
			if len(got) != len(tt.want) {
				t.Fatalf("got %x, want %x", got, tt.want)
			}
			for i := range got {
				if !bytes.Equal(got[i], tt.want[i]) {
					t.Errorf("unit %d: %x, want %x", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestNalType(t *testing.T) {
	if nalType(nil) != 0 || nalType([]byte{0x67}) != nalSPS || nalType([]byte{0x65}) != nalIDR {
		t.Error("wrong NAL types")
	}
}

func TestRBSPRemovesEmulationPrevention(t *testing.T) {
	got := rbsp([]byte{1, 0, 0, 3, 1, 0, 0, 3, 3})
	if want := []byte{1, 0, 0, 1, 0, 0, 3}; !bytes.Equal(got, want) {
		t.Errorf("got %x, want %x", got, want)
	}
}

func TestParameterSets(t *testing.T) {
	sps, pps := parameterSets(annexB([]byte{0x67, 1, 2, 3}, []byte{0x68, 4}, []byte{0x65, 5}))
	if !bytes.Equal(sps, []byte{0x67, 1, 2, 3}) || !bytes.Equal(pps, []byte{0x68, 4}) {
		t.Errorf("sps %x pps %x", sps, pps)
	}
	if sps, pps := parameterSets(annexB([]byte{0x65, 5})); sps != nil || pps != nil {
		t.Errorf("sps %x pps %x", sps, pps)
	}
}

func TestBitReader(t *testing.T) {
	// 1 | 010 | 011 | 00100 | 00101: ue 0, 1, 2, 3, 4 (the first bit is 1)
	r := &bitReader{data: []byte{0b10100110, 0b01000010, 0b10000000}}
	for want := uint(0); want < 5; want++ {
		if got := r.ue(); got != want {
			t.Errorf("ue %d, want %d", got, want)
		}
	}
	if r.err != nil {
		t.Fatal(r.err)
	}
}

func TestBitReaderSigned(t *testing.T) {
	// ue 1 -> +1, ue 2 -> -1, ue 3 -> +2, ue 4 -> -2
	r := &bitReader{data: []byte{0b01001100, 0b10000101, 0b00000000}}
	for _, want := range []int{1, -1, 2, -2} {
		if got := r.se(); got != want {
			t.Errorf("se %d, want %d", got, want)
		}
	}
}

func TestBitReaderErrorsStick(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		read func(r *bitReader)
	}{
		{"bit past the end", nil, func(r *bitReader) { r.bit() }},
		{"ue without a one", []byte{0, 0}, func(r *bitReader) { r.ue() }},
		{"ue with too many zeros", make([]byte, 8), func(r *bitReader) { r.ue() }},
		{"ue without its rest", []byte{0b00000100}, func(r *bitReader) { r.ue() }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &bitReader{data: tt.data}
			tt.read(r)
			if !errors.Is(r.err, errShortSPS) || r.bit() != 0 || r.ue() != 0 {
				t.Errorf("err %v", r.err)
			}
		})
	}
}

func TestIsHighProfile(t *testing.T) {
	if !isHighProfile(100) || !isHighProfile(244) || isHighProfile(66) || isHighProfile(77) {
		t.Error("wrong profiles")
	}
}

func writeKeys(t *testing.T, w *mp4Writer, n int) {
	t.Helper()
	key := annexB(testSPS(720, 1280), testPPS, []byte{0x65, 1, 2})
	for i := 0; i < n; i++ {
		if err := w.WriteSample(key, 0, true); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMP4WriterErrors(t *testing.T) {
	sps := testSPS(720, 1280)
	tests := []struct {
		name    string
		prepare func(fs *memFS)
		samples int
	}{
		{"no samples", func(*memFS) {}, 0},
		{"mdat size can't be set", func(fs *memFS) { fs.writeAtErr = errDisk }, 1},
		{"can't seek", func(fs *memFS) { fs.seekErr = errDisk }, 1},
		{"moov can't be written", func(fs *memFS) { fs.failWriteAt = 3 }, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := newMemFS()
			w, err := newMP4Writer(fs, "a.mp4", h264Config{sps: sps, pps: testPPS})
			if err != nil {
				t.Fatal(err)
			}
			writeKeys(t, w, tt.samples)
			tt.prepare(fs)
			if err := w.Close(); err == nil {
				t.Error("Close succeeded")
			}
			if !fs.file("a.mp4").closed {
				t.Error("file not closed")
			}
		})
	}
}

func TestMP4WriterAccessors(t *testing.T) {
	w, err := newMP4Writer(newMemFS(), "a.mp4", h264Config{sps: testSPS(720, 1280), pps: testPPS})
	if err != nil {
		t.Fatal(err)
	}
	if w.Duration() != 0 || w.Samples() != 0 || w.Size() == 0 {
		t.Errorf("duration %d samples %d size %d", w.Duration(), w.Samples(), w.Size())
	}
	// a unit with only parameter sets and an AUD has no sample
	if err := w.WriteSample(annexB(testSPS(720, 1280), testPPS, []byte{0x09, 0xf0}), 0, false); err != nil || w.Samples() != 0 {
		t.Errorf("samples %d (%v)", w.Samples(), err)
	}
	if width, height := w.dimensions(); width != 720 || height != 1280 {
		t.Errorf("size %dx%d", width, height)
	}
	for _, tm := range []int64{0, 0} { // equal times: the last sample lasts a 30th of a second
		if err := w.WriteSample(annexB([]byte{0x65, 1}), tm, true); err != nil {
			t.Fatal(err)
		}
	}
	if w.Duration() != 33333 {
		t.Errorf("duration %d", w.Duration())
	}
}

func TestNewMP4WriterErrors(t *testing.T) {
	sps := testSPS(720, 1280)
	tests := []struct {
		name string
		sps  []byte
		fs   func() *memFS
	}{
		{"invalid sps", []byte{1}, newMemFS},
		{"create fails", sps, func() *memFS {
			fs := newMemFS()
			fs.createErr = func(string) error { return errDisk }
			return fs
		}},
		{"head fails", sps, func() *memFS {
			fs := newMemFS()
			fs.failWriteAt = 1
			return fs
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.fs()
			if _, err := newMP4Writer(fs, "a.mp4", h264Config{sps: tt.sps}); err == nil {
				t.Error("no error")
			}
			for _, f := range fs.files {
				if !f.closed {
					t.Error("file left open")
				}
			}
		})
	}
}

func TestMJPEGWriterClose(t *testing.T) {
	fs := newMemFS()
	w, err := newMJPEGWriter(fs, "a", "a.idx")
	if err != nil {
		t.Fatal(err)
	}
	fs.closeErr = errDisk
	if err := w.Close(); !errors.Is(err, errDisk) {
		t.Errorf("err %v", err)
	}
	if !fs.file("a").closed || !fs.file("a.idx").closed {
		t.Error("files not closed")
	}
	if w.Duration() != 0 {
		t.Errorf("duration %d", w.Duration())
	}
}

func TestMJPEGIndexFailureClosesFile(t *testing.T) {
	fs := newMemFS()
	fs.createErr = func(name string) error {
		if name == "a.idx" {
			return errDisk
		}
		return nil
	}
	if _, err := newMJPEGWriter(fs, "a", "a.idx"); !errors.Is(err, errDisk) {
		t.Fatalf("err %v", err)
	}
	if !fs.file("a").closed {
		t.Error("data file left open")
	}
}

func TestDefaultDeps(t *testing.T) {
	d := Deps{}.withDefaults()
	if _, ok := d.Clock.(systemClock); !ok {
		t.Errorf("clock %T", d.Clock)
	}
	if _, ok := d.FS.(osFS); !ok {
		t.Errorf("fs %T", d.FS)
	}
	clock, fs := newFakeClock(), newMemFS()
	if got := (Deps{Clock: clock, FS: fs}).withDefaults(); got.Clock != Clock(clock) || got.FS != FileSystem(fs) {
		t.Errorf("deps replaced: %+v", got)
	}
}

func TestSystemClock(t *testing.T) {
	c := systemClock{}
	if time.Since(c.Now()) > time.Minute {
		t.Error("Now is not now")
	}
	select {
	case <-c.After(time.Millisecond):
	case <-time.After(2 * time.Second):
		t.Error("After did not fire")
	}
}

func TestOSFS(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a")
	f, err := osFS{}.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if err := (osFS{}).Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file still there: %v", err)
	}
	if _, err := (osFS{}).Create(filepath.Join(path, "x", "y")); err == nil {
		t.Error("Create in a missing directory succeeded")
	}
}
