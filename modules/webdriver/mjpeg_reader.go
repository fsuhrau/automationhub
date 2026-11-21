package webdriver

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// MjpegReader splits a multipart/x-mixed-replace body into JPEG frames.
//
// It doesn't use mime/multipart: WebDriverAgent declares the boundary as
// "--BoundaryString" but writes "--BoundaryString" as the delimiter, which a
// strict parser rejects. Parts are found by their Content-Length header;
// parts without one are read up to the JPEG end marker.
type MjpegReader struct {
	r *bufio.Reader
}

func NewMjpegReader(r io.Reader) *MjpegReader {
	return &MjpegReader{r: bufio.NewReaderSize(r, 64<<10)}
}

// ReadFrame returns the next frame. The slice belongs to the caller.
func (m *MjpegReader) ReadFrame() ([]byte, error) {
	for {
		length, err := m.readPartHeader()
		if err != nil {
			return nil, err
		}
		var frame []byte
		if length >= 0 {
			if length > maxMjpegFrame {
				return nil, fmt.Errorf("MJPEG frame of %d bytes is too large", length)
			}
			frame = make([]byte, length)
			if _, err := io.ReadFull(m.r, frame); err != nil {
				return nil, err
			}
		} else if frame, err = m.readJPEG(); err != nil {
			return nil, err
		}
		if len(frame) > 0 {
			return frame, nil
		}
	}
}

// readPartHeader skips to the next part and reads its headers. It returns the
// Content-Length, or -1 when the part has none.
func (m *MjpegReader) readPartHeader() (int, error) {
	line, err := m.skipToHeader()
	if err != nil {
		return 0, err
	}
	length := -1
	for line != "" {
		n, ok, err := parseContentLength(line)
		if err != nil {
			return 0, err
		}
		if ok {
			length = n
		}
		if line, err = m.readLine(); err != nil {
			return 0, err
		}
	}
	return length, nil
}

// skipToHeader skips blank lines and the boundary up to the first header line.
func (m *MjpegReader) skipToHeader() (string, error) {
	for {
		line, err := m.readLine()
		if err != nil || (line != "" && !strings.HasPrefix(line, "--")) {
			return line, err
		}
	}
}

// parseContentLength reads the length of a Content-Length header line; ok is
// false for any other header.
func parseContentLength(line string) (length int, ok bool, err error) {
	name, value, found := strings.Cut(line, ":")
	if !found || !strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
		return 0, false, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < 0 {
		return 0, false, fmt.Errorf("invalid MJPEG Content-Length %q", value)
	}
	return n, true, nil
}

func (m *MjpegReader) readLine() (string, error) {
	line, err := m.r.ReadSlice('\n')
	if err == bufio.ErrBufferFull {
		return "", fmt.Errorf("MJPEG header line too long")
	}
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(line), "\r\n"), nil
}

// readJPEG reads from the start of image (FF D8) to the end of image (FF D9)
// marker.
func (m *MjpegReader) readJPEG() ([]byte, error) {
	var buf bytes.Buffer
	var prev byte
	started := false
	for {
		b, err := m.r.ReadByte()
		if err != nil {
			return nil, err
		}
		if !started {
			if prev == 0xFF && b == 0xD8 {
				started = true
				buf.Write([]byte{0xFF, 0xD8})
			}
			prev = b
			continue
		}
		buf.WriteByte(b)
		if buf.Len() > maxMjpegFrame {
			return nil, fmt.Errorf("MJPEG frame is too large")
		}
		if prev == 0xFF && b == 0xD9 {
			return buf.Bytes(), nil
		}
		prev = b
	}
}
