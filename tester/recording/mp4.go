package recording

import (
	"errors"
	"io"
)

// mp4Timescale is the media timescale of the video track (90 kHz, as usual
// for video).
const mp4Timescale = 90000

type mp4Sample struct {
	offset int64
	size   uint32
	// time since the start of the recording in microseconds
	time int64
	key  bool
}

// h264Config are the parameter sets of an H.264 stream.
type h264Config struct {
	sps, pps []byte
}

// mp4Writer writes H.264 into an MP4 file that a browser plays in a video
// element. Samples go into one mdat box as they arrive; the moov box with the
// sample tables is written by Close, at the end of the file.
type mp4Writer struct {
	f             File
	sps, pps      []byte
	width, height int
	mdatStart     int64
	pos           int64
	samples       []mp4Sample
}

func newMP4Writer(fs FileSystem, path string, cfg h264Config) (*mp4Writer, error) {
	width, height, err := spsSize(cfg.sps)
	if err != nil {
		return nil, err
	}
	f, err := fs.Create(path)
	if err != nil {
		return nil, err
	}
	w := &mp4Writer{f: f, sps: cfg.sps, pps: cfg.pps, width: width, height: height}
	if err := w.writeHead(); err != nil {
		_ = f.Close()
		return nil, err
	}
	return w, nil
}

// writeHead writes the ftyp box and the start of the mdat box.
func (w *mp4Writer) writeHead() error {
	ftyp := box("ftyp", []byte("isom"), u32(0x200), []byte("isomiso2avc1mp41"))
	// mdat with a 64-bit size, set by Close
	mdat := append(u32(1), []byte("mdat")...)
	mdat = append(mdat, make([]byte, 8)...)
	if _, err := w.f.Write(append(ftyp, mdat...)); err != nil {
		return err
	}
	w.mdatStart = int64(len(ftyp))
	w.pos = int64(len(ftyp) + len(mdat))
	return nil
}

// sampleData converts the Annex B data of an access unit to length prefixed
// NAL units without the parameter sets and AUDs; they are in the sample
// description.
func sampleData(data []byte) []byte {
	var buf []byte
	for _, nal := range splitAnnexB(data) {
		switch nalType(nal) {
		case nalSPS, nalPPS, nalAUD:
			continue
		}
		buf = append(buf, u32(uint32(len(nal)))...)
		buf = append(buf, nal...)
	}
	return buf
}

// WriteSample writes one access unit, given as Annex B data, at the time in
// microseconds since the start of the recording.
func (w *mp4Writer) WriteSample(data []byte, time int64, key bool) error {
	buf := sampleData(data)
	if len(buf) == 0 {
		return nil
	}
	if _, err := w.f.Write(buf); err != nil {
		return err
	}
	w.samples = append(w.samples, mp4Sample{offset: w.pos, size: uint32(len(buf)), time: time, key: key})
	w.pos += int64(len(buf))
	return nil
}

// Duration of the samples written so far in microseconds.
func (w *mp4Writer) Duration() int64 {
	if len(w.samples) == 0 {
		return 0
	}
	return w.samples[len(w.samples)-1].time + w.lastDuration()*1000000/mp4Timescale
}

func (w *mp4Writer) Size() int64 { return w.pos }

func (w *mp4Writer) Samples() int { return len(w.samples) }

func (w *mp4Writer) dimensions() (width, height int) { return w.width, w.height }

func ticks(us int64) int64 { return us * mp4Timescale / 1000000 }

// lastDuration is the duration of the last sample in ticks; it has no next
// sample, so it lasts as long as the one before it.
func (w *mp4Writer) lastDuration() int64 {
	n := len(w.samples)
	if n >= 2 {
		if d := ticks(w.samples[n-1].time) - ticks(w.samples[n-2].time); d > 0 {
			return d
		}
	}
	return mp4Timescale / 30
}

// Close writes the sample tables and closes the file.
func (w *mp4Writer) Close() error {
	defer func() { _ = w.f.Close() }()
	if len(w.samples) == 0 {
		return errors.New("no samples")
	}
	if _, err := w.f.WriteAt(u64(uint64(w.pos-w.mdatStart)), w.mdatStart+8); err != nil {
		return err
	}
	if _, err := w.f.Seek(w.pos, io.SeekStart); err != nil {
		return err
	}
	_, err := w.f.Write(w.moov())
	return err
}
