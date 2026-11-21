package mcp

import (
	"bytes"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	_ "image/png" // registers the PNG decoder for screenshots
)

const defaultScreenshotWidth = 1024

// scaleToJPEG shrinks img to at most maxWidth (default defaultScreenshotWidth)
// and encodes it as JPEG, small enough to hand to a model.
func scaleToJPEG(img image.Image, maxWidth int) (image.Image, []byte, error) {
	if maxWidth <= 0 {
		maxWidth = defaultScreenshotWidth
	}
	scaled := img
	if b := img.Bounds(); b.Dx() > maxWidth {
		h := max(1, (b.Dy()*maxWidth+b.Dx()/2)/b.Dx())
		dst := image.NewRGBA(image.Rect(0, 0, maxWidth, h))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Src, nil)
		scaled = dst
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, scaled, &jpeg.Options{Quality: 80}); err != nil {
		return nil, nil, err
	}
	return scaled, buf.Bytes(), nil
}
