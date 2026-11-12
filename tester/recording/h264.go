package recording

import (
	"errors"
)

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
		if data[i] == 0 && data[i+1] == 0 && data[i+2] == 1 {
			if start >= 0 {
				end := i
				// a 4-byte start code leaves a zero at the end of the previous unit
				for end > start && data[end-1] == 0 {
					end--
				}
				nals = append(nals, data[start:end])
			}
			i += 3
			start = i
			continue
		}
		i++
	}
	if start >= 0 && start < len(data) {
		nals = append(nals, data[start:])
	} else if start < 0 && len(data) > 0 {
		// no start code: the data is one unit
		nals = append(nals, data)
	}
	return nals
}

func nalType(nal []byte) byte {
	if len(nal) == 0 {
		return 0
	}
	return nal[0] & 0x1f
}

// bitReader reads the bits of an RBSP (emulation prevention bytes removed).
type bitReader struct {
	data []byte
	pos  int
}

var errShortSPS = errors.New("sps too short")

func (r *bitReader) bit() (uint, error) {
	if r.pos >= len(r.data)*8 {
		return 0, errShortSPS
	}
	b := (r.data[r.pos/8] >> (7 - uint(r.pos%8))) & 1
	r.pos++
	return uint(b), nil
}

func (r *bitReader) bits(n int) (uint, error) {
	var v uint
	for i := 0; i < n; i++ {
		b, err := r.bit()
		if err != nil {
			return 0, err
		}
		v = v<<1 | b
	}
	return v, nil
}

func (r *bitReader) ue() (uint, error) {
	zeros := 0
	for {
		b, err := r.bit()
		if err != nil {
			return 0, err
		}
		if b == 1 {
			break
		}
		zeros++
		if zeros > 31 {
			return 0, errShortSPS
		}
	}
	rest, err := r.bits(zeros)
	if err != nil {
		return 0, err
	}
	return (1<<zeros - 1) + rest, nil
}

func (r *bitReader) se() (int, error) {
	v, err := r.ue()
	if err != nil {
		return 0, err
	}
	if v%2 == 1 {
		return int(v+1) / 2, nil
	}
	return -int(v / 2), nil
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
		if b == 0 {
			zeros++
		} else {
			zeros = 0
		}
		out = append(out, b)
	}
	return out
}

// spsSize reads the picture size from a sequence parameter set.
func spsSize(sps []byte) (width, height int, err error) {
	if len(sps) < 4 {
		return 0, 0, errShortSPS
	}
	r := &bitReader{data: rbsp(sps[1:])}
	profile, _ := r.bits(8)
	if _, err := r.bits(16); err != nil { // constraint flags, level
		return 0, 0, err
	}
	if _, err := r.ue(); err != nil { // seq_parameter_set_id
		return 0, 0, err
	}
	chroma := uint(1)
	switch profile {
	case 100, 110, 122, 244, 44, 83, 86, 118, 128, 138, 139, 134, 135:
		if chroma, err = r.ue(); err != nil {
			return 0, 0, err
		}
		if chroma == 3 {
			if _, err := r.bit(); err != nil { // separate_colour_plane_flag
				return 0, 0, err
			}
		}
		for i := 0; i < 2; i++ { // bit depths
			if _, err := r.ue(); err != nil {
				return 0, 0, err
			}
		}
		if _, err := r.bit(); err != nil { // qpprime_y_zero_transform_bypass_flag
			return 0, 0, err
		}
		present, err := r.bit()
		if err != nil {
			return 0, 0, err
		}
		if present == 1 {
			lists := 8
			if chroma == 3 {
				lists = 12
			}
			for i := 0; i < lists; i++ {
				listPresent, err := r.bit()
				if err != nil {
					return 0, 0, err
				}
				if listPresent == 0 {
					continue
				}
				size := 16
				if i >= 6 {
					size = 64
				}
				last, next := 8, 8
				for j := 0; j < size; j++ {
					if next != 0 {
						delta, err := r.se()
						if err != nil {
							return 0, 0, err
						}
						next = (last + delta + 256) % 256
					}
					if next != 0 {
						last = next
					}
				}
			}
		}
	}
	if _, err := r.ue(); err != nil { // log2_max_frame_num_minus4
		return 0, 0, err
	}
	pocType, err := r.ue()
	if err != nil {
		return 0, 0, err
	}
	switch pocType {
	case 0:
		if _, err := r.ue(); err != nil {
			return 0, 0, err
		}
	case 1:
		if _, err := r.bit(); err != nil {
			return 0, 0, err
		}
		if _, err := r.se(); err != nil {
			return 0, 0, err
		}
		if _, err := r.se(); err != nil {
			return 0, 0, err
		}
		cycle, err := r.ue()
		if err != nil {
			return 0, 0, err
		}
		for i := uint(0); i < cycle; i++ {
			if _, err := r.se(); err != nil {
				return 0, 0, err
			}
		}
	}
	if _, err := r.ue(); err != nil { // max_num_ref_frames
		return 0, 0, err
	}
	if _, err := r.bit(); err != nil { // gaps_in_frame_num_value_allowed_flag
		return 0, 0, err
	}
	wMbs, err := r.ue()
	if err != nil {
		return 0, 0, err
	}
	hMaps, err := r.ue()
	if err != nil {
		return 0, 0, err
	}
	frameMbsOnly, err := r.bit()
	if err != nil {
		return 0, 0, err
	}
	if frameMbsOnly == 0 {
		if _, err := r.bit(); err != nil { // mb_adaptive_frame_field_flag
			return 0, 0, err
		}
	}
	if _, err := r.bit(); err != nil { // direct_8x8_inference_flag
		return 0, 0, err
	}
	width = int(wMbs+1) * 16
	height = int(2-frameMbsOnly) * int(hMaps+1) * 16
	cropping, err := r.bit()
	if err != nil {
		return 0, 0, err
	}
	if cropping == 1 {
		var crop [4]uint
		for i := range crop {
			if crop[i], err = r.ue(); err != nil {
				return 0, 0, err
			}
		}
		unitX, unitY := 1, int(2-frameMbsOnly)
		if chroma == 1 || chroma == 2 {
			unitX = 2
		}
		if chroma == 1 {
			unitY *= 2
		}
		width -= int(crop[0]+crop[1]) * unitX
		height -= int(crop[2]+crop[3]) * unitY
	}
	return width, height, nil
}
