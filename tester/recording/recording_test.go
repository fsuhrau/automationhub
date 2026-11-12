package recording

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// bitWriter builds an SPS for the tests.
type bitWriter struct {
	bits []byte
}

func (w *bitWriter) put(v uint, n int) {
	for i := n - 1; i >= 0; i-- {
		w.bits = append(w.bits, byte(v>>uint(i))&1)
	}
}

func (w *bitWriter) ue(v uint) {
	v++
	n := 0
	for x := v; x > 1; x >>= 1 {
		n++
	}
	w.put(0, n)
	w.put(v, n+1)
}

func (w *bitWriter) bytes() []byte {
	w.put(1, 1) // rbsp stop bit
	for len(w.bits)%8 != 0 {
		w.bits = append(w.bits, 0)
	}
	out := make([]byte, len(w.bits)/8)
	for i, b := range w.bits {
		out[i/8] |= b << (7 - uint(i%8))
	}
	return out
}

// testSPS is a High profile 4:2:0 SPS for a picture of width x height.
func testSPS(width, height int) []byte {
	w := &bitWriter{}
	w.put(100, 8) // profile_idc High
	w.put(0, 8)   // constraint flags
	w.put(31, 8)  // level
	w.ue(0)       // sps id
	w.ue(1)       // chroma_format_idc 4:2:0
	w.ue(0)       // bit depth luma
	w.ue(0)       // bit depth chroma
	w.put(0, 1)   // qpprime
	w.put(0, 1)   // no scaling matrix
	w.ue(0)       // log2_max_frame_num_minus4
	w.ue(2)       // pic_order_cnt_type
	w.ue(1)       // max_num_ref_frames
	w.put(0, 1)   // gaps
	mbW, mbH := (width+15)/16, (height+15)/16
	w.ue(uint(mbW - 1))
	w.ue(uint(mbH - 1))
	w.put(1, 1) // frame_mbs_only
	w.put(1, 1) // direct_8x8
	cropR, cropB := mbW*16-width, mbH*16-height
	if cropR > 0 || cropB > 0 {
		w.put(1, 1)
		w.ue(0)
		w.ue(uint(cropR / 2))
		w.ue(0)
		w.ue(uint(cropB / 2))
	} else {
		w.put(0, 1)
	}
	w.put(0, 1) // no vui
	return append([]byte{0x67}, w.bytes()...)
}

func annexB(nals ...[]byte) []byte {
	var out []byte
	for _, n := range nals {
		out = append(out, 0, 0, 0, 1)
		out = append(out, n...)
	}
	return out
}

var testPPS = []byte{0x68, 0xce, 0x3c, 0x80}

func TestSPSSize(t *testing.T) {
	for _, size := range [][2]int{{1080, 2400}, {720, 1280}, {1280, 720}, {1920, 1080}} {
		w, h, err := spsSize(testSPS(size[0], size[1]))
		if err != nil || w != size[0] || h != size[1] {
			t.Errorf("%v: got %dx%d (%v)", size, w, h, err)
		}
	}
}

func TestSplitAnnexB(t *testing.T) {
	data := append(annexB([]byte{0x67, 1, 2}, []byte{0x68, 3}), 0, 0, 1, 0x65, 9, 0)
	nals := splitAnnexB(data)
	if len(nals) != 3 || nalType(nals[0]) != nalSPS || nalType(nals[1]) != nalPPS || nalType(nals[2]) != nalIDR {
		t.Fatalf("nals %x", nals)
	}
	if !bytes.Equal(nals[0], []byte{0x67, 1, 2}) {
		t.Errorf("sps %x", nals[0])
	}
}

// mp4Boxes finds a box by its path in an MP4 file.
func mp4Box(t *testing.T, data []byte, path ...string) []byte {
	t.Helper()
	for _, name := range path {
		found := false
		for len(data) >= 8 {
			size := int(binary.BigEndian.Uint32(data))
			head := 8
			if size == 1 {
				size = int(binary.BigEndian.Uint64(data[8:]))
				head = 16
			}
			if size < head || size > len(data) {
				t.Fatalf("bad box size %d of %q", size, data[4:8])
			}
			if string(data[4:8]) == name {
				data = data[head:size]
				found = true
				break
			}
			data = data[size:]
		}
		if !found {
			t.Fatalf("no box %s in %v", name, path)
		}
	}
	return data
}

func TestMP4Writer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.mp4")
	sps := testSPS(720, 1280)
	w, err := newMP4Writer(path, sps, testPPS)
	if err != nil {
		t.Fatal(err)
	}
	frames := []struct {
		time int64
		key  bool
	}{{0, true}, {40000, false}, {80000, false}, {1080000, true}}
	for _, f := range frames {
		data := annexB([]byte{0x09, 0xf0}, []byte{0x61, 1, 2, 3})
		if f.key {
			data = annexB(sps, testPPS, []byte{0x65, 1, 2, 3, 4})
		}
		if err := w.WriteSample(data, f.time, f.key); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)

	mdat := mp4Box(t, data, "mdat")
	// samples are length prefixed and carry no parameter sets or AUDs
	if first := mdat[:4+5]; binary.BigEndian.Uint32(first) != 5 || first[4] != 0x65 {
		t.Errorf("first sample %x", first)
	}
	stbl := []string{"moov", "trak", "mdia", "minf", "stbl"}
	tkhd := mp4Box(t, data, "moov", "trak", "tkhd")
	if w, h := binary.BigEndian.Uint32(tkhd[len(tkhd)-8:])>>16, binary.BigEndian.Uint32(tkhd[len(tkhd)-4:])>>16; w != 720 || h != 1280 {
		t.Errorf("tkhd size %dx%d", w, h)
	}
	stsz := mp4Box(t, data, append(stbl, "stsz")...)
	if n := binary.BigEndian.Uint32(stsz[8:]); n != 4 {
		t.Errorf("%d samples", n)
	}
	stss := mp4Box(t, data, append(stbl, "stss")...)
	if n, second := binary.BigEndian.Uint32(stss[4:]), binary.BigEndian.Uint32(stss[12:]); n != 2 || second != 4 {
		t.Errorf("key frames %x", stss)
	}
	co64 := mp4Box(t, data, append(stbl, "co64")...)
	first := binary.BigEndian.Uint64(co64[8:])
	if !bytes.Equal(data[first:first+5], []byte{0, 0, 0, 5, 0x65}) {
		t.Errorf("chunk offset %d points to %x", first, data[first:first+5])
	}
	stts := mp4Box(t, data, append(stbl, "stts")...)
	// 3600 ticks twice, 90000 once, and the last one as long as the one before
	if got := stts[4:]; !bytes.Equal(got, concat(u32(2), u32(2), u32(3600), u32(2), u32(90000))) {
		t.Errorf("stts %x", got)
	}
	avcC := mp4Box(t, data, append(stbl, "stsd")...)
	if !bytes.Contains(avcC, sps) || !bytes.Contains(avcC, testPPS) {
		t.Error("avcC lacks the parameter sets")
	}
}

// fakeStream plays packets and then ends, or blocks until closed.
type fakeStream struct {
	meta    device.VideoMeta
	packets chan device.VideoPacket
	closed  chan struct{}
	once    sync.Once
}

func newFakeStream(codec string, packets ...device.VideoPacket) *fakeStream {
	s := &fakeStream{meta: device.VideoMeta{Codec: codec}, packets: make(chan device.VideoPacket, len(packets)), closed: make(chan struct{})}
	for _, p := range packets {
		s.packets <- p
	}
	return s
}

func (s *fakeStream) Meta() device.VideoMeta { return s.meta }
func (s *fakeStream) ReadPacket() (device.VideoPacket, error) {
	select {
	case p := <-s.packets:
		return p, nil
	case <-s.closed:
		return device.VideoPacket{}, errors.New("closed")
	}
}
func (s *fakeStream) Close() error {
	s.once.Do(func() { close(s.closed) })
	return nil
}

type fakeDevice struct {
	device.Device
	streams chan *fakeStream
}

func (d *fakeDevice) DeviceID() string { return "phone" }
func (d *fakeDevice) StartVideo(ctx context.Context, opts device.VideoOptions) (device.VideoStream, error) {
	select {
	case s := <-d.streams:
		return s, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type testLog struct {
	mu    sync.Mutex
	lines []string
}

func (l *testLog) LogInfo(format string, params ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, format)
}
func (l *testLog) LogError(format string, params ...interface{}) {
	l.LogInfo("ERROR "+format, params...)
}

func testDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.TestRunRecording{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// The recorder writes a segment per H.264 configuration; a rotation (new SPS)
// starts the next one, and Stop completes both files.
func TestRecorderH264Segments(t *testing.T) {
	dir := t.TempDir()
	db := testDB(t)
	portrait, landscape := testSPS(720, 1280), testSPS(1280, 720)
	key := func(pts int64) device.VideoPacket {
		return device.VideoPacket{KeyFrame: true, PTS: pts, Data: annexB([]byte{0x65, 1, 2})}
	}
	delta := func(pts int64) device.VideoPacket {
		return device.VideoPacket{PTS: pts, Data: annexB([]byte{0x41, 1})}
	}
	stream := newFakeStream(device.CodecH264,
		delta(1), // before any configuration: dropped
		device.VideoPacket{Config: true, Data: annexB(portrait, testPPS)},
		delta(2), // before the first key frame: dropped
		key(1000000), delta(1033333),
		device.VideoPacket{Config: true, Data: annexB(landscape, testPPS)},
		key(2000000), delta(2033333), delta(2066666),
	)
	dev := &fakeDevice{streams: make(chan *fakeStream, 1)}
	dev.streams <- stream
	log := &testLog{}
	r := Start(dev, Options{Dir: dir, DB: db, RunID: 4, DeviceID: 9, Log: log})

	waitFor(t, func() bool { return len(stream.packets) == 0 })
	time.Sleep(20 * time.Millisecond)
	r.Stop()

	var rows []models.TestRunRecording
	db.Order("id").Find(&rows)
	if len(rows) != 2 {
		t.Fatalf("%d recordings, want 2: %+v", len(rows), rows)
	}
	for i, want := range [][2]int{{720, 1280}, {1280, 720}} {
		row := rows[i]
		if row.Codec != models.RecordingMP4 || row.TestRunID != 4 || row.DeviceID != 9 || row.EndedAt == nil || row.Width != want[0] || row.Height != want[1] {
			t.Errorf("recording %d: %+v", i, row)
		}
		data, err := os.ReadFile(filepath.Join(dir, row.FileName))
		if err != nil || int64(len(data)) <= row.Size || !bytes.Contains(data, []byte("moov")) {
			t.Errorf("recording %d file: %d bytes, size %d (%v)", i, len(data), row.Size, err)
		}
	}
	stsz := mp4Box(t, mustRead(t, filepath.Join(dir, rows[1].FileName)), "moov", "trak", "mdia", "minf", "stbl", "stsz")
	if n := binary.BigEndian.Uint32(stsz[8:]); n != 3 {
		t.Errorf("second segment has %d samples, want 3", n)
	}
}

func mustRead(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func testJPEG(t *testing.T) []byte {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 40, 80)), nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// MJPEG frames go into one file with an index line per frame; when the
// stream ends the recorder opens it again and starts the next segment.
func TestRecorderMJPEG(t *testing.T) {
	restartDelay = time.Millisecond
	dir := t.TempDir()
	db := testDB(t)
	frame := testJPEG(t)
	first := newFakeStream(device.CodecMJPEG, device.VideoPacket{KeyFrame: true, Data: frame}, device.VideoPacket{KeyFrame: true, Data: frame})
	second := newFakeStream(device.CodecMJPEG, device.VideoPacket{KeyFrame: true, Data: frame})
	dev := &fakeDevice{streams: make(chan *fakeStream, 2)}
	dev.streams <- first
	dev.streams <- second
	r := Start(dev, Options{Dir: dir, DB: db, RunID: 1, DeviceID: 2, Log: &testLog{}})

	waitFor(t, func() bool { return len(first.packets) == 0 })
	time.Sleep(20 * time.Millisecond)
	_ = first.Close() // the device restarted
	waitFor(t, func() bool { return len(second.packets) == 0 })
	time.Sleep(20 * time.Millisecond)
	r.Stop()

	var rows []models.TestRunRecording
	db.Order("id").Find(&rows)
	if len(rows) != 2 {
		t.Fatalf("%d recordings, want 2", len(rows))
	}
	row := rows[0]
	if row.Codec != models.RecordingMJPEG || row.Width != 40 || row.Height != 80 || row.Size != int64(2*len(frame)) || row.EndedAt == nil {
		t.Errorf("recording %+v", row)
	}
	index := strings.Fields(string(mustRead(t, filepath.Join(dir, row.IndexFile))))
	if len(index) != 6 || index[1] != "0" || index[4] != strconv.Itoa(len(frame)) {
		t.Errorf("index %v", index)
	}
}

// Devices without a video stream are skipped with a note in the run log.
func TestRecorderWithoutVideo(t *testing.T) {
	log := &testLog{}
	r := Start(struct{ device.Device }{&fakeDevice{}}, Options{Log: log, DB: testDB(t)})
	r.Stop()
	if len(log.lines) != 1 || !strings.Contains(log.lines[0], "no video stream") {
		t.Errorf("log %v", log.lines)
	}
}
