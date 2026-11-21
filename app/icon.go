package app

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
)

//nolint:gochecknoglobals // a constant byte signature, never modified
var pngSignature = []byte("\x89PNG\r\n\x1a\n")

// ErrNoIcon means the bundle has no icon a browser can show
var ErrNoIcon = errors.New("the bundle has no icon")

// Icon reads the icon of a stored bundle as an image a browser can show and
// its content type
func Icon(appPath, iconPath string) ([]byte, string, error) {
	if iconPath == "" {
		return nil, "", ErrNoIcon
	}
	data, err := readStoredIcon(appPath, iconPath)
	if err != nil {
		return nil, "", err
	}
	switch strings.ToLower(iconPath[strings.LastIndex(iconPath, ".")+1:]) {
	case "webp":
		return data, "image/webp", nil
	case "icns":
		return icnsImage(data)
	}
	return pngImage(data)
}

// readStoredIcon takes the icon file out of the stored archive
func readStoredIcon(appPath, iconPath string) ([]byte, error) {
	r, err := zip.OpenReader(appPath)
	if err != nil {
		return nil, err
	}
	defer closeQuietly(r)
	f := findFile(r.File, iconPath)
	if f == nil {
		return nil, ErrNoIcon
	}
	return readZipFile(f, 8<<20)
}

func icnsImage(data []byte) ([]byte, string, error) {
	p := pngFromICNS(data)
	if p == nil {
		return nil, "", ErrNoIcon
	}
	return p, "image/png", nil
}

// pngImage serves a png as it is, after the Xcode optimisation is undone
func pngImage(data []byte) ([]byte, string, error) {
	if !bytes.HasPrefix(data, pngSignature) {
		return nil, "", ErrNoIcon
	}
	if !bytes.Contains(data[:min(len(data), 64)], []byte("CgBI")) {
		return data, "image/png", nil
	}
	fixed, err := uncrushPNG(data)
	if err != nil {
		return nil, "", err
	}
	return fixed, "image/png", nil
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
