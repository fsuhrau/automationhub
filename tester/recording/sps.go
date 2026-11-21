package recording

// isHighProfile reports whether the SPS of the profile_idc has the chroma
// format, bit depth and scaling matrix fields.
func isHighProfile(profile uint) bool {
	switch profile {
	case 100, 110, 122, 244, 44, 83, 86, 118, 128, 138, 139, 134, 135:
		return true
	}
	return false
}

// picture is the coded size of an SPS, before cropping.
type picture struct {
	width, height int
	frameMbsOnly  uint
}

// spsSize reads the picture size from a sequence parameter set.
func spsSize(sps []byte) (width, height int, err error) {
	if len(sps) < 4 {
		return 0, 0, errShortSPS
	}
	r := &bitReader{data: rbsp(sps[1:])}
	profile := r.bits(8)
	r.bits(16) // constraint flags, level
	r.ue()     // seq_parameter_set_id
	chroma := uint(1)
	if isHighProfile(profile) {
		chroma = readHighFields(r)
	}
	skipPicOrder(r)
	r.ue()  // max_num_ref_frames
	r.bit() // gaps_in_frame_num_value_allowed_flag
	pic := readPicture(r)
	cropX, cropY := readCrop(r, chroma, pic.frameMbsOnly)
	if r.err != nil {
		return 0, 0, r.err
	}
	return pic.width - cropX, pic.height - cropY, nil
}

// readHighFields reads the chroma format, bit depths and scaling matrix of
// the high profiles and returns the chroma format.
func readHighFields(r *bitReader) uint {
	chroma := r.ue()
	if chroma == 3 {
		r.bit() // separate_colour_plane_flag
	}
	r.ue()  // bit depth luma
	r.ue()  // bit depth chroma
	r.bit() // qpprime_y_zero_transform_bypass_flag
	if r.bit() == 1 {
		skipScalingLists(r, chroma)
	}
	return chroma
}

func skipScalingLists(r *bitReader, chroma uint) {
	lists := 8
	if chroma == 3 {
		lists = 12
	}
	for i := 0; i < lists && r.err == nil; i++ {
		if r.bit() == 1 {
			skipScalingList(r, listSize(i))
		}
	}
}

func listSize(i int) int {
	if i >= 6 {
		return 64
	}
	return 16
}

func skipScalingList(r *bitReader, size int) {
	last, next := 8, 8
	for j := 0; j < size; j++ {
		if next != 0 {
			next = (last + r.se() + 256) % 256
		}
		if next != 0 {
			last = next
		}
	}
}

// skipPicOrder skips log2_max_frame_num and the picture order count fields.
func skipPicOrder(r *bitReader) {
	r.ue() // log2_max_frame_num_minus4
	switch r.ue() {
	case 0:
		r.ue()
	case 1:
		r.bit()
		r.se()
		r.se()
		cycle := r.ue()
		for i := uint(0); i < cycle && r.err == nil; i++ {
			r.se()
		}
	}
}

// readPicture reads the size in macroblocks and the frame coding fields.
func readPicture(r *bitReader) picture {
	wMbs := r.ue()
	hMbs := r.ue()
	frameMbsOnly := r.bit()
	if frameMbsOnly == 0 {
		r.bit() // mb_adaptive_frame_field_flag
	}
	r.bit() // direct_8x8_inference_flag
	return picture{
		width:        int(wMbs+1) * 16,
		height:       int(2-frameMbsOnly) * int(hMbs+1) * 16,
		frameMbsOnly: frameMbsOnly,
	}
}

// readCrop reads the cropping rectangle and returns what it takes off the
// width and the height.
func readCrop(r *bitReader, chroma, frameMbsOnly uint) (x, y int) {
	if r.bit() == 0 {
		return 0, 0
	}
	var crop [4]uint
	for i := range crop {
		crop[i] = r.ue()
	}
	unitX, unitY := 1, int(2-frameMbsOnly)
	if chroma == 1 || chroma == 2 {
		unitX = 2
	}
	if chroma == 1 {
		unitY *= 2
	}
	return int(crop[0]+crop[1]) * unitX, int(crop[2]+crop[3]) * unitY
}
