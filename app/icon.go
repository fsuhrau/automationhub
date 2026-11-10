package app

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"io"
	"strings"
)

var pngSignature = []byte("\x89PNG\r\n\x1a\n")

// ErrNoIcon means the bundle has no icon a browser can show
var ErrNoIcon = errors.New("the bundle has no icon")

// Icon reads the icon of a stored bundle as an image a browser can show and
// its content type
func Icon(appPath, iconPath string) ([]byte, string, error) {
	if iconPath == "" {
		return nil, "", ErrNoIcon
	}
	r, err := zip.OpenReader(appPath)
	if err != nil {
		return nil, "", err
	}
	defer r.Close()
	f := findFile(r.File, iconPath)
	if f == nil {
		return nil, "", ErrNoIcon
	}
	data, err := readZipFile(f, 8<<20)
	if err != nil {
		return nil, "", err
	}
	switch strings.ToLower(iconPath[strings.LastIndex(iconPath, ".")+1:]) {
	case "webp":
		return data, "image/webp", nil
	case "icns":
		if p := pngFromICNS(data); p != nil {
			return p, "image/png", nil
		}
		return nil, "", ErrNoIcon
	}
	if !bytes.HasPrefix(data, pngSignature) {
		return nil, "", ErrNoIcon
	}
	if bytes.Contains(data[:min(len(data), 64)], []byte("CgBI")) {
		fixed, err := uncrushPNG(data)
		if err != nil {
			return nil, "", err
		}
		return fixed, "image/png", nil
	}
	return data, "image/png", nil
}

// pngFromICNS takes the largest png image out of a Mac .icns file
func pngFromICNS(data []byte) []byte {
	if len(data) < 8 || string(data[:4]) != "icns" {
		return nil
	}
	var best []byte
	for off := 8; off+8 <= len(data); {
		n := int(binary.BigEndian.Uint32(data[off+4:]))
		if n < 8 || off+n > len(data) {
			break
		}
		body := data[off+8 : off+n]
		if bytes.HasPrefix(body, pngSignature) && len(body) > len(best) {
			best = body
		}
		off += n
	}
	return best
}

type pngChunk struct {
	kind string
	data []byte
}

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
	var ihdr []byte
	var idat bytes.Buffer
	for _, c := range chunks {
		switch c.kind {
		case "IHDR":
			ihdr = c.data
		case "IDAT":
			idat.Write(c.data)
		}
	}
	if len(ihdr) != 13 {
		return nil, errors.New("png without header")
	}
	width, height := int(binary.BigEndian.Uint32(ihdr)), int(binary.BigEndian.Uint32(ihdr[4:]))
	if ihdr[8] != 8 || ihdr[9] != 6 || ihdr[12] != 0 || width <= 0 || height <= 0 || width*height > 4096*4096 {
		return nil, errors.New("unsupported crushed png")
	}
	raw, err := io.ReadAll(io.LimitReader(flate.NewReader(&idat), int64(height*(1+width*4))+1))
	if err != nil {
		return nil, err
	}
	stride := 1 + width*4
	if len(raw) != height*stride {
		return nil, errors.New("crushed png has the wrong size")
	}
	// png filters work on each channel alone, so swapping B and R in the
	// filtered rows gives the same image as swapping after unfiltering
	for y := 0; y < height; y++ {
		row := raw[y*stride+1 : (y+1)*stride]
		for x := 0; x+3 < len(row); x += 4 {
			row[x], row[x+2] = row[x+2], row[x]
		}
	}
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	zw.Write(raw)
	zw.Close()
	var out bytes.Buffer
	out.Write(pngSignature)
	writeChunk(&out, "IHDR", ihdr)
	writeChunk(&out, "IDAT", compressed.Bytes())
	writeChunk(&out, "IEND", nil)

	img, err := png.Decode(&out)
	if err != nil {
		return nil, err
	}
	nrgba, ok := img.(*image.NRGBA)
	if !ok {
		return nil, errors.New("unexpected png format")
	}
	// the colours are premultiplied by alpha
	for i := 0; i+3 < len(nrgba.Pix); i += 4 {
		a := int(nrgba.Pix[i+3])
		if a > 0 && a < 255 {
			for c := 0; c < 3; c++ {
				nrgba.Pix[i+c] = uint8(min(255, int(nrgba.Pix[i+c])*255/a))
			}
		}
	}
	var result bytes.Buffer
	if err := png.Encode(&result, nrgba); err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}
