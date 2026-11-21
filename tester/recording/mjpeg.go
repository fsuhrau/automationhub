package recording

import (
	"bytes"
	"fmt"
	"image/jpeg"
)

// mjpegWriter appends JPEG frames to one file and writes a line per frame to
// the index file: its time in milliseconds since the start of the recording,
// its offset and its length. The index is written as frames arrive, so a
// recording stays readable when the master stops in the middle of it.
type mjpegWriter struct {
	f, index      File
	pos           int64
	frames        int
	last          int64
	width, height int
}

func newMJPEGWriter(fs FileSystem, path, indexPath string) (*mjpegWriter, error) {
	f, err := fs.Create(path)
	if err != nil {
		return nil, err
	}
	index, err := fs.Create(indexPath)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return &mjpegWriter{f: f, index: index}, nil
}

// WriteFrame writes a frame at the time in microseconds since the start of
// the recording.
func (w *mjpegWriter) WriteFrame(data []byte, time int64) error {
	if w.frames == 0 {
		if cfg, err := jpeg.DecodeConfig(bytes.NewReader(data)); err == nil {
			w.width, w.height = cfg.Width, cfg.Height
		}
	}
	if _, err := w.f.Write(data); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w.index, "%d %d %d\n", time/1000, w.pos, len(data)); err != nil {
		return err
	}
	w.pos += int64(len(data))
	w.frames++
	w.last = time
	return nil
}

func (w *mjpegWriter) dimensions() (width, height int) { return w.width, w.height }

func (w *mjpegWriter) Size() int64 { return w.pos }

func (w *mjpegWriter) Samples() int { return w.frames }

// Duration until the last frame in microseconds.
func (w *mjpegWriter) Duration() int64 { return w.last }

func (w *mjpegWriter) Close() error {
	err := w.f.Close()
	if ierr := w.index.Close(); err == nil {
		err = ierr
	}
	return err
}
