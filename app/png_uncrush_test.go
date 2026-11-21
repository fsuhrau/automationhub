package app

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// crushedHeader is the IHDR of a 8 bit RGBA png
func crushedHeader(width, height uint32) []byte {
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr, width)
	binary.BigEndian.PutUint32(ihdr[4:], height)
	ihdr[8], ihdr[9] = 8, 6
	return ihdr
}

// crushedFromRaw is a CgBI png of the filtered rows
func crushedFromRaw(ihdr, raw []byte) []byte {
	var deflated bytes.Buffer
	fw, _ := flate.NewWriter(&deflated, flate.BestCompression)
	_, _ = fw.Write(raw)
	_ = fw.Close()
	var out bytes.Buffer
	out.Write(pngSignature)
	writeChunk(&out, "CgBI", []byte{0x50, 0x00, 0x20, 0x06})
	writeChunk(&out, "IHDR", ihdr)
	writeChunk(&out, "IDAT", deflated.Bytes())
	writeChunk(&out, "IEND", nil)
	return out.Bytes()
}

func TestUncrushPNGRestoresTheColours(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 200, G: 100, B: 50, A: 128})
	img.SetNRGBA(1, 0, color.NRGBA{R: 1, G: 2, B: 3, A: 0})
	fixed, err := uncrushPNG(crush(t, img))
	if err != nil {
		t.Fatal(err)
	}
	got, err := png.Decode(bytes.NewReader(fixed))
	if err != nil {
		t.Fatal(err)
	}
	if px := color.NRGBAModel.Convert(got.At(0, 0)).(color.NRGBA); px.A != 128 || px.R < 198 || px.R > 202 {
		t.Errorf("pixel = %v", px)
	}
	if px := color.NRGBAModel.Convert(got.At(1, 0)).(color.NRGBA); px.A != 0 {
		t.Errorf("transparent pixel = %v", px)
	}
}

func TestUncrushPNGErrors(t *testing.T) {
	good := crushedHeader(1, 1)
	tests := map[string][]byte{
		"truncated":          append(append([]byte{}, pngSignature...), 0, 0, 1, 0, 'I', 'H', 'D', 'R', 0, 0, 0, 0),
		"no header":          append(append([]byte{}, pngSignature...), []byte{0, 0, 0, 0, 'I', 'E', 'N', 'D', 0, 0, 0, 0}...),
		"unsupported colors": crushedFromRaw(crushedHeader(0, 1), nil),
		"not deflate":        corruptIDAT(crushedFromRaw(good, make([]byte, 5))),
		"wrong size":         crushedFromRaw(good, make([]byte, 6)),
		"bad filter":         crushedFromRaw(good, []byte{9, 0, 0, 0, 0}),
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if out, err := uncrushPNG(data); err == nil || out != nil {
				t.Errorf("out = %v, err = %v", out, err)
			}
		})
	}
}

// corruptIDAT replaces the image data of a png with bytes that are no deflate stream
func corruptIDAT(data []byte) []byte {
	chunks, _ := pngChunks(data)
	var out bytes.Buffer
	out.Write(pngSignature)
	for _, c := range chunks {
		if c.kind == "IDAT" {
			c.data = bytes.Repeat([]byte{0xff}, 8)
		}
		writeChunk(&out, c.kind, c.data)
	}
	return out.Bytes()
}

func TestPNGChunks(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(pngSignature)
	writeChunk(&buf, "IHDR", []byte{1})
	writeChunk(&buf, "IEND", nil)
	writeChunk(&buf, "late", []byte{2})
	chunks, err := pngChunks(buf.Bytes())
	if err != nil || len(chunks) != 2 || chunks[0].kind != "IHDR" || chunks[1].kind != "IEND" {
		t.Errorf("chunks = %v, err = %v", chunks, err)
	}
}

func TestCrushedSize(t *testing.T) {
	tests := []struct {
		name string
		ihdr []byte
		ok   bool
	}{
		{"valid", crushedHeader(3, 2), true},
		{"short header", []byte{1, 2}, false},
		{"zero height", crushedHeader(3, 0), false},
		{"too big", crushedHeader(5000, 5000), false},
		{"16 bit", func() []byte { h := crushedHeader(1, 1); h[8] = 16; return h }(), false},
		{"palette", func() []byte { h := crushedHeader(1, 1); h[9] = 3; return h }(), false},
		{"interlaced", func() []byte { h := crushedHeader(1, 1); h[12] = 1; return h }(), false},
	}
	for _, tt := range tests {
		size, err := crushedSize(tt.ihdr)
		if (err == nil) != tt.ok || (tt.ok && (size.width != 3 || size.height != 2 || size.stride() != 13)) {
			t.Errorf("%s: size = %+v, err = %v", tt.name, size, err)
		}
	}
}

func TestSwapBlueRed(t *testing.T) {
	raw := []byte{0, 1, 2, 3, 4, 0, 5, 6, 7, 8}
	swapBlueRed(raw, pngSize{width: 1, height: 2})
	if want := []byte{0, 3, 2, 1, 4, 0, 7, 6, 5, 8}; !bytes.Equal(raw, want) {
		t.Errorf("raw = %v, want %v", raw, want)
	}
}

func TestUnpremultiply(t *testing.T) {
	pix := []uint8{50, 100, 200, 100, 9, 9, 9, 0, 7, 8, 9, 255, 250, 1, 2, 10}
	unpremultiply(pix)
	want := []uint8{127, 255, 255, 100, 9, 9, 9, 0, 7, 8, 9, 255, 255, 25, 51, 10}
	if !bytes.Equal(pix, want) {
		t.Errorf("pix = %v, want %v", pix, want)
	}
}

func TestToNRGBAConvertsOtherImages(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.RGBA{R: 255, A: 255})
	if got := toNRGBA(src).NRGBAAt(0, 0); got != (color.NRGBA{R: 255, A: 255}) {
		t.Errorf("pixel = %v", got)
	}
	same := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	if toNRGBA(same) != same {
		t.Error("an NRGBA image is taken as it is")
	}
}

func TestDecodeNRGBARejectsGarbage(t *testing.T) {
	if _, err := decodeNRGBA(bytes.NewReader([]byte("garbage"))); err == nil {
		t.Error("no error")
	}
}

func TestEncodePNGRejectsEmptyImage(t *testing.T) {
	if _, err := encodePNG(image.NewNRGBA(image.Rect(0, 0, 0, 0))); err == nil {
		t.Error("no error")
	}
}
