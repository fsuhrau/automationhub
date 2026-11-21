package app

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/draw"
	"image/png"
	"io"
)

type pngChunk struct {
	kind string
	data []byte
}

// pngSize is the size of a 32 bit png, which has 4 bytes per pixel and a
// filter byte in front of each row
type pngSize struct {
	width, height int
}

func (s pngSize) stride() int { return 1 + s.width*4 }

func pngChunks(data []byte) ([]pngChunk, error) {
	var chunks []pngChunk
	for off := len(pngSignature); off+12 <= len(data); {
		n := int(binary.BigEndian.Uint32(data[off:]))
		if n < 0 || off+12+n > len(data) {
			return nil, errors.New("truncated png")
		}
		kind := string(data[off+4 : off+8])
		chunks = append(chunks, pngChunk{kind, data[off+8 : off+8+n]})
		off += 12 + n
		if kind == "IEND" {
			break
		}
	}
	return chunks, nil
}

func writeChunk(w *bytes.Buffer, kind string, data []byte) {
	var head [8]byte
	binary.BigEndian.PutUint32(head[:], uint32(len(data)))
	copy(head[4:], kind)
	w.Write(head[:])
	w.Write(data)
	crc := crc32.NewIEEE()
	crc.Write(head[4:])
	crc.Write(data)
	var sum [4]byte
	binary.BigEndian.PutUint32(sum[:], crc.Sum32())
	w.Write(sum[:])
}

// uncrushPNG turns an Xcode optimised png (CgBI: raw deflate, BGRA,
// premultiplied alpha) back into a standard one
func uncrushPNG(data []byte) ([]byte, error) {
	chunks, err := pngChunks(data)
	if err != nil {
		return nil, err
	}
	ihdr, idat := headerAndData(chunks)
	size, err := crushedSize(ihdr)
	if err != nil {
		return nil, err
	}
	raw, err := inflateRows(idat, size)
	if err != nil {
		return nil, err
	}
	swapBlueRed(raw, size)
	standard, err := decodeNRGBA(assemblePNG(ihdr, raw))
	if err != nil {
		return nil, err
	}
	unpremultiply(standard.Pix)
	return encodePNG(standard)
}

// headerAndData takes the header and the joined image data out of the chunks
func headerAndData(chunks []pngChunk) (ihdr []byte, idat *bytes.Buffer) {
	idat = &bytes.Buffer{}
	for _, c := range chunks {
		switch c.kind {
		case "IHDR":
			ihdr = c.data
		case "IDAT":
			idat.Write(c.data)
		}
	}
	return ihdr, idat
}

// crushedSize checks that the header is one of a 8 bit RGBA png of a size
// that is read, and returns the size
func crushedSize(ihdr []byte) (pngSize, error) {
	if len(ihdr) != 13 {
		return pngSize{}, errors.New("png without header")
	}
	size := pngSize{int(binary.BigEndian.Uint32(ihdr)), int(binary.BigEndian.Uint32(ihdr[4:]))}
	if ihdr[8] != 8 || ihdr[9] != 6 || ihdr[12] != 0 || size.width <= 0 || size.height <= 0 || size.width*size.height > 4096*4096 {
		return pngSize{}, errors.New("unsupported crushed png")
	}
	return size, nil
}

// inflateRows reads the filtered rows out of the raw deflate stream
func inflateRows(idat *bytes.Buffer, size pngSize) ([]byte, error) {
	want := size.height * size.stride()
	raw, err := io.ReadAll(io.LimitReader(flate.NewReader(idat), int64(want)+1))
	if err != nil {
		return nil, err
	}
	if len(raw) != want {
		return nil, errors.New("crushed png has the wrong size")
	}
	return raw, nil
}

// swapBlueRed turns BGRA into RGBA. png filters work on each channel alone,
// so swapping B and R in the filtered rows gives the same image as swapping
// after unfiltering.
func swapBlueRed(raw []byte, size pngSize) {
	for y := 0; y < size.height; y++ {
		row := raw[y*size.stride()+1 : (y+1)*size.stride()]
		for x := 0; x+3 < len(row); x += 4 {
			row[x], row[x+2] = row[x+2], row[x]
		}
	}
}

// assemblePNG writes a standard png from the header and the filtered rows
func assemblePNG(ihdr, raw []byte) *bytes.Buffer {
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	_, _ = zw.Write(raw)
	_ = zw.Close()
	var out bytes.Buffer
	out.Write(pngSignature)
	writeChunk(&out, "IHDR", ihdr)
	writeChunk(&out, "IDAT", compressed.Bytes())
	writeChunk(&out, "IEND", nil)
	return &out
}

func decodeNRGBA(r io.Reader) (*image.NRGBA, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}
	return toNRGBA(img), nil
}

func toNRGBA(img image.Image) *image.NRGBA {
	if nrgba, ok := img.(*image.NRGBA); ok {
		return nrgba
	}
	nrgba := image.NewNRGBA(img.Bounds())
	draw.Draw(nrgba, nrgba.Bounds(), img, img.Bounds().Min, draw.Src)
	return nrgba
}

// unpremultiply divides the colours by alpha: the crushed png has them
// premultiplied
func unpremultiply(pix []uint8) {
	for i := 0; i+3 < len(pix); i += 4 {
		a := int(pix[i+3])
		if a == 0 || a == 255 {
			continue
		}
		for c := 0; c < 3; c++ {
			pix[i+c] = uint8(min(255, int(pix[i+c])*255/a))
		}
	}
}

func encodePNG(img image.Image) ([]byte, error) {
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
