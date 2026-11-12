package recording

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
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

// mp4Writer writes H.264 into an MP4 file that a browser plays in a video
// element. Samples go into one mdat box as they arrive; the moov box with the
// sample tables is written by Close, at the end of the file.
type mp4Writer struct {
	f             *os.File
	sps, pps      []byte
	width, height int
	mdatStart     int64
	pos           int64
	samples       []mp4Sample
}

func newMP4Writer(path string, sps, pps []byte) (*mp4Writer, error) {
	width, height, err := spsSize(sps)
	if err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := &mp4Writer{f: f, sps: sps, pps: pps, width: width, height: height}
	ftyp := box("ftyp", []byte("isom"), u32(0x200), []byte("isomiso2avc1mp41"))
	// mdat with a 64-bit size, set by Close
	mdat := append(u32(1), []byte("mdat")...)
	mdat = append(mdat, make([]byte, 8)...)
	if _, err := f.Write(append(ftyp, mdat...)); err != nil {
		_ = f.Close()
		return nil, err
	}
	w.mdatStart = int64(len(ftyp))
	w.pos = int64(len(ftyp) + len(mdat))
	return w, nil
}

// WriteSample writes one access unit, given as Annex B data, at the time in
// microseconds since the start of the recording.
func (w *mp4Writer) WriteSample(data []byte, time int64, key bool) error {
	var buf []byte
	for _, nal := range splitAnnexB(data) {
		switch nalType(nal) {
		case nalSPS, nalPPS, nalAUD:
			// the parameter sets are in the sample description
			continue
		}
		buf = append(buf, u32(uint32(len(nal)))...)
		buf = append(buf, nal...)
	}
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
	defer w.f.Close()
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

func (w *mp4Writer) moov() []byte {
	// sample durations, run length encoded
	var stts [][2]uint32
	var total int64
	for i := range w.samples {
		d := w.lastDuration()
		if i+1 < len(w.samples) {
			d = ticks(w.samples[i+1].time) - ticks(w.samples[i].time)
		}
		if d <= 0 {
			d = 1
		}
		total += d
		if n := len(stts); n > 0 && stts[n-1][1] == uint32(d) {
			stts[n-1][0]++
		} else {
			stts = append(stts, [2]uint32{1, uint32(d)})
		}
	}
	sttsData := u32(uint32(len(stts)))
	for _, e := range stts {
		sttsData = append(sttsData, u32(e[0])...)
		sttsData = append(sttsData, u32(e[1])...)
	}

	var keys []byte
	keyCount := 0
	sizes := u32(0) // sample_size 0: sizes follow
	sizes = append(sizes, u32(uint32(len(w.samples)))...)
	offsets := u32(uint32(len(w.samples)))
	for i, s := range w.samples {
		if s.key {
			keys = append(keys, u32(uint32(i+1))...)
			keyCount++
		}
		sizes = append(sizes, u32(s.size)...)
		offsets = append(offsets, u64(uint64(s.offset))...)
	}

	durationMs := uint32(total * 1000 / mp4Timescale)
	matrix := concat(u32(0x10000), u32(0), u32(0), u32(0), u32(0x10000), u32(0), u32(0), u32(0), u32(0x40000000))

	mvhd := fullBox("mvhd", 0, 0, u32(0), u32(0), u32(1000), u32(durationMs),
		u32(0x10000), u16(0x100), make([]byte, 10), matrix, make([]byte, 24), u32(2))
	tkhd := fullBox("tkhd", 0, 3, u32(0), u32(0), u32(1), u32(0), u32(durationMs),
		make([]byte, 8), u16(0), u16(0), u16(0), u16(0), matrix,
		u32(uint32(w.width)<<16), u32(uint32(w.height)<<16))
	mdhd := fullBox("mdhd", 0, 0, u32(0), u32(0), u32(mp4Timescale), u32(uint32(total)), u16(0x55c4), u16(0))
	hdlr := fullBox("hdlr", 0, 0, u32(0), []byte("vide"), make([]byte, 12), []byte("VideoHandler\x00"))

	avcC := box("avcC", []byte{1, w.sps[1], w.sps[2], w.sps[3], 0xff, 0xe1},
		u16(uint16(len(w.sps))), w.sps, []byte{1}, u16(uint16(len(w.pps))), w.pps)
	compressor := make([]byte, 32)
	avc1 := box("avc1", make([]byte, 6), u16(1), u16(0), u16(0), make([]byte, 12),
		u16(uint16(w.width)), u16(uint16(w.height)), u32(0x480000), u32(0x480000), u32(0),
		u16(1), compressor, u16(0x18), u16(0xffff), avcC)

	stbl := box("stbl",
		fullBox("stsd", 0, 0, u32(1), avc1),
		fullBox("stts", 0, 0, sttsData),
		fullBox("stss", 0, 0, u32(uint32(keyCount)), keys),
		fullBox("stsc", 0, 0, u32(1), u32(1), u32(1), u32(1)),
		fullBox("stsz", 0, 0, sizes),
		fullBox("co64", 0, 0, offsets),
	)
	minf := box("minf",
		fullBox("vmhd", 0, 1, make([]byte, 8)),
		box("dinf", fullBox("dref", 0, 0, u32(1), fullBox("url ", 0, 1))),
		stbl)
	trak := box("trak", tkhd, box("mdia", mdhd, hdlr, minf))
	return box("moov", mvhd, trak)
}

func concat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func box(typ string, payload ...[]byte) []byte {
	body := concat(payload...)
	return concat(u32(uint32(8+len(body))), []byte(typ), body)
}

func fullBox(typ string, version byte, flags uint32, payload ...[]byte) []byte {
	head := u32(uint32(version)<<24 | flags&0xffffff)
	return box(typ, append([][]byte{head}, payload...)...)
}

func u16(v uint16) []byte { return binary.BigEndian.AppendUint16(nil, v) }
func u32(v uint32) []byte { return binary.BigEndian.AppendUint32(nil, v) }
func u64(v uint64) []byte { return binary.BigEndian.AppendUint64(nil, v) }
