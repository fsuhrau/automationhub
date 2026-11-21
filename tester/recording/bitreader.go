package recording

import "errors"

var errShortSPS = errors.New("sps too short")

// bitReader reads the bits of an RBSP (emulation prevention bytes removed).
// The first error sticks: later reads return 0, so a parser checks err once.
type bitReader struct {
	data []byte
	pos  int
	err  error
}

func (r *bitReader) bit() uint {
	if r.err != nil {
		return 0
	}
	if r.pos >= len(r.data)*8 {
		r.err = errShortSPS
		return 0
	}
	b := (r.data[r.pos/8] >> (7 - uint(r.pos%8))) & 1
	r.pos++
	return uint(b)
}

func (r *bitReader) bits(n int) uint {
	var v uint
	for i := 0; i < n; i++ {
		v = v<<1 | r.bit()
	}
	return v
}

// ue reads an unsigned Exp-Golomb value.
func (r *bitReader) ue() uint {
	zeros := 0
	for r.bit() == 0 {
		if r.err != nil {
			return 0
		}
		zeros++
		if zeros > 31 {
			r.err = errShortSPS
			return 0
		}
	}
	rest := r.bits(zeros)
	return (1<<zeros - 1) + rest
}

// se reads a signed Exp-Golomb value.
func (r *bitReader) se() int {
	v := r.ue()
	if v%2 == 1 {
		return int(v+1) / 2
	}
	return -int(v / 2)
}
