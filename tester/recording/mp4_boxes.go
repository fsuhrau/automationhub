package recording

import "encoding/binary"

// moov is the movie box with the sample tables of the written samples.
func (w *mp4Writer) moov() []byte {
	durs := w.sampleDurations()
	var total int64
	for _, d := range durs {
		total += d
	}
	durationMs := uint32(total * 1000 / mp4Timescale)
	return box("moov", mvhdBox(durationMs), w.trakBox(durs, total))
}

// sampleDurations are the durations of the samples in ticks.
func (w *mp4Writer) sampleDurations() []int64 {
	durs := make([]int64, len(w.samples))
	for i := range w.samples {
		d := w.lastDuration()
		if i+1 < len(w.samples) {
			d = ticks(w.samples[i+1].time) - ticks(w.samples[i].time)
		}
		durs[i] = max(d, 1)
	}
	return durs
}

func matrixBytes() []byte {
	return concat(u32(0x10000), u32(0), u32(0), u32(0), u32(0x10000), u32(0), u32(0), u32(0), u32(0x40000000))
}

func mvhdBox(durationMs uint32) []byte {
	return fullBox("mvhd", 0, u32(0), u32(0), u32(1000), u32(durationMs),
		u32(0x10000), u16(0x100), make([]byte, 10), matrixBytes(), make([]byte, 24), u32(2))
}

func tkhdBox(durationMs uint32, width, height int) []byte {
	return fullBox("tkhd", 3, u32(0), u32(0), u32(1), u32(0), u32(durationMs),
		make([]byte, 8), u16(0), u16(0), u16(0), u16(0), matrixBytes(),
		u32(uint32(width)<<16), u32(uint32(height)<<16))
}

func (w *mp4Writer) trakBox(durs []int64, total int64) []byte {
	tkhd := tkhdBox(uint32(total*1000/mp4Timescale), w.width, w.height)
	mdhd := fullBox("mdhd", 0, u32(0), u32(0), u32(mp4Timescale), u32(uint32(total)), u16(0x55c4), u16(0))
	hdlr := fullBox("hdlr", 0, u32(0), []byte("vide"), make([]byte, 12), []byte("VideoHandler\x00"))
	return box("trak", tkhd, box("mdia", mdhd, hdlr, w.minfBox(durs)))
}

func (w *mp4Writer) minfBox(durs []int64) []byte {
	return box("minf",
		fullBox("vmhd", 1, make([]byte, 8)),
		box("dinf", fullBox("dref", 0, u32(1), fullBox("url ", 1))),
		w.stblBox(durs))
}

func (w *mp4Writer) avc1Box() []byte {
	avcC := box("avcC", []byte{1, w.sps[1], w.sps[2], w.sps[3], 0xff, 0xe1},
		u16(uint16(len(w.sps))), w.sps, []byte{1}, u16(uint16(len(w.pps))), w.pps)
	return box("avc1", make([]byte, 6), u16(1), u16(0), u16(0), make([]byte, 12),
		u16(uint16(w.width)), u16(uint16(w.height)), u32(0x480000), u32(0x480000), u32(0),
		u16(1), make([]byte, 32), u16(0x18), u16(0xffff), avcC)
}

func (w *mp4Writer) stblBox(durs []int64) []byte {
	keys, keyCount, sizes, offsets := w.sampleTables()
	return box("stbl",
		fullBox("stsd", 0, u32(1), w.avc1Box()),
		fullBox("stts", 0, sttsData(durs)),
		fullBox("stss", 0, u32(uint32(keyCount)), keys),
		fullBox("stsc", 0, u32(1), u32(1), u32(1), u32(1)),
		fullBox("stsz", 0, sizes),
		fullBox("co64", 0, offsets),
	)
}

// sampleTables are the payloads of the key frame, size and offset boxes.
func (w *mp4Writer) sampleTables() (keys []byte, keyCount int, sizes, offsets []byte) {
	sizes = u32(0) // sample_size 0: sizes follow
	sizes = append(sizes, u32(uint32(len(w.samples)))...)
	offsets = u32(uint32(len(w.samples)))
	for i, s := range w.samples {
		if s.key {
			keys = append(keys, u32(uint32(i+1))...)
			keyCount++
		}
		sizes = append(sizes, u32(s.size)...)
		offsets = append(offsets, u64(uint64(s.offset))...)
	}
	return keys, keyCount, sizes, offsets
}

// sttsData is the payload of the time-to-sample box: the durations, run
// length encoded.
func sttsData(durs []int64) []byte {
	var runs [][2]uint32
	for _, d := range durs {
		if n := len(runs); n > 0 && runs[n-1][1] == uint32(d) {
			runs[n-1][0]++
		} else {
			runs = append(runs, [2]uint32{1, uint32(d)})
		}
	}
	data := u32(uint32(len(runs)))
	for _, e := range runs {
		data = append(data, u32(e[0])...)
		data = append(data, u32(e[1])...)
	}
	return data
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

// fullBox is a box of version 0 with flags.
func fullBox(typ string, flags uint32, payload ...[]byte) []byte {
	head := u32(flags & 0xffffff)
	return box(typ, append([][]byte{head}, payload...)...)
}

func u16(v uint16) []byte { return binary.BigEndian.AppendUint16(nil, v) }
func u32(v uint32) []byte { return binary.BigEndian.AppendUint32(nil, v) }
func u64(v uint64) []byte { return binary.BigEndian.AppendUint64(nil, v) }
