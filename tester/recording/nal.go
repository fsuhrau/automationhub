package recording

// NAL unit types of H.264 used here.
const (
	nalIDR = 5
	nalSPS = 7
	nalPPS = 8
	nalAUD = 9
)

// splitAnnexB splits H.264 Annex B data at its start codes into NAL units.
func splitAnnexB(data []byte) [][]byte {
	var nals [][]byte
	start := -1
	for i := 0; i+2 < len(data); {
		if !isStartCode(data[i:]) {
			i++
			continue
		}
		if start >= 0 {
			nals = append(nals, data[start:trimZeros(data, start, i)])
		}
		i += 3
		start = i
	}
	return appendLast(nals, data, start)
}

func isStartCode(b []byte) bool {
	return len(b) >= 3 && b[0] == 0 && b[1] == 0 && b[2] == 1
}

// trimZeros returns the end of the unit data[start:end]; a 4-byte start code
// leaves a zero at the end of the previous unit.
func trimZeros(data []byte, start, end int) int {
	for end > start && data[end-1] == 0 {
		end--
	}
	return end
}

// appendLast adds the unit after the last start code, or all data as one unit
// when it has no start code.
func appendLast(nals [][]byte, data []byte, start int) [][]byte {
	switch {
	case start >= 0 && start < len(data):
		return append(nals, data[start:])
	case start < 0 && len(data) > 0:
		return append(nals, data)
	}
	return nals
}

func nalType(nal []byte) byte {
	if len(nal) == 0 {
		return 0
	}
	return nal[0] & 0x1f
}

// rbsp removes the emulation prevention bytes (00 00 03 -> 00 00).
func rbsp(nal []byte) []byte {
	out := make([]byte, 0, len(nal))
	zeros := 0
	for _, b := range nal {
		if zeros >= 2 && b == 3 {
			zeros = 0
			continue
		}
		zeros = nextZeros(zeros, b)
		out = append(out, b)
	}
	return out
}

func nextZeros(zeros int, b byte) int {
	if b == 0 {
		return zeros + 1
	}
	return 0
}

// parameterSets returns the SPS and PPS in the data, if any.
func parameterSets(data []byte) (sps, pps []byte) {
	for _, nal := range splitAnnexB(data) {
		switch nalType(nal) {
		case nalSPS:
			sps = append([]byte(nil), nal...)
		case nalPPS:
			pps = append([]byte(nil), nal...)
		}
	}
	return sps, pps
}
