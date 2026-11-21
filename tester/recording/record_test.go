package recording

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// sliceStream plays its packets and then ends.
type sliceStream struct {
	codec   string
	packets []device.VideoPacket
	closed  bool
}

func (s *sliceStream) Meta() device.VideoMeta { return device.VideoMeta{Codec: s.codec} }
func (s *sliceStream) Close() error           { s.closed = true; return nil }
func (s *sliceStream) ReadPacket() (device.VideoPacket, error) {
	if len(s.packets) == 0 {
		return device.VideoPacket{}, io.EOF
	}
	p := s.packets[0]
	s.packets = s.packets[1:]
	return p, nil
}

// session is one recording over fakes.
type session struct {
	t     *testing.T
	fs    *memFS
	log   *msgLog
	clock *fakeClock
	db    *gorm.DB
}

func newSession(t *testing.T) *session {
	return &session{t: t, fs: newMemFS(), log: &msgLog{}, clock: newFakeClock(), db: testDB(t)}
}

func (s *session) options() Options {
	return Options{Deps: Deps{Clock: s.clock, FS: s.fs}, Dir: "rec", DB: s.db, RunID: 4, DeviceID: 9, Log: s.log}
}

// record plays the packets until the stream ends.
func (s *session) record(codec string, packets ...device.VideoPacket) {
	record(context.Background(), &sliceStream{codec: codec, packets: packets}, s.options())
}

func (s *session) rows() []models.TestRunRecording {
	var rows []models.TestRunRecording
	s.db.Order("id").Find(&rows)
	return rows
}

func (s *session) data(row models.TestRunRecording) []byte {
	f := s.fs.file(filepath.Join("rec", row.FileName))
	if f == nil {
		s.t.Fatalf("no file %s", row.FileName)
	}
	return f.data
}

func h264Key(pts int64) device.VideoPacket {
	return device.VideoPacket{KeyFrame: true, PTS: pts, Data: annexB([]byte{0x65, 1, 2})}
}

func h264Delta(pts int64) device.VideoPacket {
	return device.VideoPacket{PTS: pts, Data: annexB([]byte{0x41, 1})}
}

func configPacket(sps []byte) device.VideoPacket {
	return device.VideoPacket{Config: true, Data: annexB(sps, testPPS)}
}

func sampleCount(t *testing.T, mp4 []byte) uint32 {
	stsz := mp4Box(t, mp4, "moov", "trak", "mdia", "minf", "stbl", "stsz")
	return binary.BigEndian.Uint32(stsz[8:])
}

// A rotation starts the next segment; both files are complete in the file
// system and their rows carry the clock's time.
func TestRecordH264Segments(t *testing.T) {
	s := newSession(t)
	portrait, landscape := testSPS(720, 1280), testSPS(1280, 720)
	s.record(device.CodecH264,
		h264Delta(1), configPacket(portrait), h264Delta(2),
		h264Key(1000000), h264Delta(1033333),
		configPacket(landscape), h264Key(2000000), h264Delta(2033333), h264Delta(2066666))

	rows := s.rows()
	if len(rows) != 2 {
		t.Fatalf("%d rows: %+v", len(rows), rows)
	}
	want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if !rows[0].StartedAt.UTC().Equal(want) || rows[0].EndedAt == nil || rows[0].Width != 720 || rows[1].Width != 1280 {
		t.Errorf("rows %+v", rows)
	}
	if n := sampleCount(t, s.data(rows[0])); n != 2 {
		t.Errorf("first segment: %d samples", n)
	}
	if n := sampleCount(t, s.data(rows[1])); n != 3 {
		t.Errorf("second segment: %d samples", n)
	}
	if s.log.text() != "" {
		t.Errorf("log %q", s.log.text())
	}
}

// Parameter sets without a PPS are not used, and the same ones again do not
// end the segment.
func TestRecordH264ConfigHandling(t *testing.T) {
	s := newSession(t)
	sps := testSPS(720, 1280)
	onlySPS := device.VideoPacket{Config: true, Data: annexB(sps)}
	s.record(device.CodecH264, onlySPS, h264Key(1), configPacket(sps), h264Key(100), configPacket(sps), h264Delta(200))

	rows := s.rows()
	if len(rows) != 1 || sampleCount(t, s.data(rows[0])) != 2 {
		t.Fatalf("rows %+v", rows)
	}
}

// A packet without a usable PTS gets the time on the clock since the segment
// started.
func TestRecordH264TimeFromClock(t *testing.T) {
	s := newSession(t)
	s.record(device.CodecH264, configPacket(testSPS(720, 1280)), h264Key(0))

	row := s.rows()[0]
	// the clock moves 10 ms per reading: the sample is read 20 ms after the
	// start and lasts a 30th of a second
	if got, want := row.EndedAt.Sub(row.StartedAt), 20*time.Millisecond+33333*time.Microsecond; got != want {
		t.Errorf("duration %v, want %v", got, want)
	}
}

func TestRecordH264Failures(t *testing.T) {
	sps := testSPS(720, 1280)
	tests := []struct {
		name    string
		prepare func(s *session)
		packets []device.VideoPacket
		logged  string
		files   int // files left in the file system
	}{
		{"invalid sps", func(*session) {}, []device.VideoPacket{configPacket([]byte{0x67, 1}), {KeyFrame: true, Data: annexB([]byte{0x67, 1}, testPPS, []byte{0x65, 1})}}, "ERROR Session recording: sps too short", 0},
		{"file can't be created", func(s *session) { s.fs.createErr = func(string) error { return errDisk } }, []device.VideoPacket{configPacket(sps), h264Key(1)}, "ERROR Session recording: disk full", 0},
		{"head can't be written", func(s *session) { s.fs.failWriteAt = 1 }, []device.VideoPacket{configPacket(sps), h264Key(1)}, "ERROR Session recording: disk full", 1},
		{"sample can't be written", func(s *session) { s.fs.failWriteAt = 2 }, []device.VideoPacket{configPacket(sps), h264Key(1), h264Delta(2)}, "ERROR Session recording: disk full", 0},
		{"row can't be stored", func(s *session) { s.db = unmigratedDB(t) }, []device.VideoPacket{configPacket(sps), h264Key(1)}, "ERROR Session recording: SQL logic error: no such table", 0},
		{"file can't be completed", func(s *session) { s.fs.writeAtErr = errDisk }, []device.VideoPacket{configPacket(sps), h264Key(1)}, "ERROR Session recording: unable to save the recording: disk full", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSession(t)
			tt.prepare(s)
			s.record(device.CodecH264, tt.packets...)
			if !strings.Contains(s.log.text(), tt.logged) {
				t.Errorf("log %q lacks %q", s.log.text(), tt.logged)
			}
			if len(s.fs.names()) != tt.files {
				t.Errorf("files left: %v", s.fs.names())
			}
		})
	}
}

func unmigratedDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestRecordMJPEG(t *testing.T) {
	s := newSession(t)
	frame := testJPEG(t)
	s.record(device.CodecMJPEG, device.VideoPacket{Data: frame}, device.VideoPacket{Data: frame})

	rows := s.rows()
	if len(rows) != 1 || rows[0].Width != 40 || rows[0].Height != 80 || rows[0].Size != int64(2*len(frame)) {
		t.Fatalf("rows %+v", rows)
	}
	index := strings.Fields(string(s.fs.file(filepath.Join("rec", rows[0].IndexFile)).data))
	// the clock moves 10 ms per reading: 20 ms and 30 ms after the start
	if strings.Join(index, " ") != fmt.Sprintf("20 0 %d 30 %d %d", len(frame), len(frame), len(frame)) {
		t.Errorf("index %v", index)
	}
}

func TestRecordMJPEGFailures(t *testing.T) {
	frame := device.VideoPacket{Data: []byte{1, 2, 3}}
	tests := []struct {
		name    string
		prepare func(s *session)
		logged  string
	}{
		{"file can't be created", func(s *session) { s.fs.createErr = func(string) error { return errDisk } }, "ERROR Session recording: disk full"},
		{"index can't be created", func(s *session) {
			s.fs.createErr = func(name string) error {
				if strings.HasSuffix(name, ".idx") {
					return errDisk
				}
				return nil
			}
		}, "ERROR Session recording: disk full"},
		{"frame can't be written", func(s *session) { s.fs.failWriteAt = 1 }, "ERROR Session recording: disk full"},
		{"index line can't be written", func(s *session) { s.fs.failWriteAt = 2 }, "ERROR Session recording: disk full"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSession(t)
			tt.prepare(s)
			s.record(device.CodecMJPEG, frame)
			if !strings.Contains(s.log.text(), tt.logged) {
				t.Errorf("log %q", s.log.text())
			}
			if len(s.rows()) != 0 {
				t.Errorf("rows %+v", s.rows())
			}
		})
	}
}

func TestRecordUnknownCodec(t *testing.T) {
	s := newSession(t)
	s.record("vp9", device.VideoPacket{Data: []byte{1}})
	if !strings.Contains(s.log.text(), `INFO Session recording: video codec "vp9" can't be recorded`) {
		t.Errorf("log %q", s.log.text())
	}
}

// Ending the context closes the stream, which ends record.
func TestRecordStopsWithContext(t *testing.T) {
	s := newSession(t)
	stream := newFakeStream(device.CodecH264)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		record(ctx, stream, s.options())
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("record did not end")
	}
}

// scriptDevice answers StartVideo from a script, then blocks until ctx is done.
type scriptDevice struct {
	device.Device
	script []scriptStep
}

type scriptStep struct {
	stream device.VideoStream
	err    error
}

func (d *scriptDevice) DeviceID() string { return "phone" }
func (d *scriptDevice) StartVideo(ctx context.Context, _ device.VideoOptions) (device.VideoStream, error) {
	if len(d.script) == 0 {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	step := d.script[0]
	d.script = d.script[1:]
	return step.stream, step.err
}

func TestRunRetriesAndGivesUp(t *testing.T) {
	s := newSession(t)
	dev := &scriptDevice{script: []scriptStep{
		{err: errDisk}, {err: errDisk}, {err: device.ErrUnsupported},
	}}
	run(context.Background(), dev, s.options())

	want := "ERROR Session recording: unable to start the video of device phone: disk full\n" +
		"INFO Session recording: device phone can't be recorded: " + device.ErrUnsupported.Error() + "\n"
	if s.log.text() != want {
		t.Errorf("log %q, want %q", s.log.text(), want)
	}
	if s.clock.afters != 2 {
		t.Errorf("%d waits, want 2", s.clock.afters)
	}
}

func TestRunRecordsAfterFailure(t *testing.T) {
	s := newSession(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.clock.onAfter = func(n int) {
		if n == 2 {
			cancel()
		}
	}
	dev := &scriptDevice{script: []scriptStep{
		{err: errDisk}, {stream: &sliceStream{codec: device.CodecH264}},
	}}
	run(ctx, dev, s.options())

	if !strings.Contains(s.log.text(), "ERROR Session recording: unable to start") || !strings.Contains(s.log.text(), "INFO Session recording: recording device phone") {
		t.Errorf("log %q", s.log.text())
	}
}

// blockedDevice does not answer StartVideo until released, whatever ctx says.
type blockedDevice struct {
	device.Device
	release chan struct{}
}

func (d *blockedDevice) DeviceID() string { return "phone" }
func (d *blockedDevice) StartVideo(context.Context, device.VideoOptions) (device.VideoStream, error) {
	<-d.release
	return nil, device.ErrUnsupported
}

func TestStopGivesUpAfterTimeout(t *testing.T) {
	s := newSession(t)
	s.clock.fireStop = true
	dev := &blockedDevice{release: make(chan struct{})}
	r := Start(dev, s.options())
	r.Stop() // the device never answers; the timeout ends the wait
	select {
	case <-r.done:
		t.Fatal("recorder ended before the device answered")
	default:
	}
	close(dev.release)
	<-r.done
}
