package recording

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsuhrau/automationhub/storage/models"
)

// writer is the file of a segment.
type writer interface {
	Close() error
	Size() int64
	Samples() int
	Duration() int64
}

// sized is a writer that knows the picture size.
type sized interface {
	dimensions() (width, height int)
}

type segment struct {
	row   models.TestRunRecording
	w     writer
	files []string
}

// openFunc creates the files of a segment and fills in its row.
type openFunc func(row *models.TestRunRecording) (writer, []string, error)

func newName(o Options, ext string) string {
	return fmt.Sprintf("recording-%d-%d-%d.%s", o.RunID, o.DeviceID, o.Clock.Now().UnixNano(), ext)
}

func startSegment(o Options, codec string, open openFunc) (*segment, error) {
	row := models.TestRunRecording{TestRunID: o.RunID, DeviceID: o.DeviceID, Codec: codec, StartedAt: o.Clock.Now()}
	w, files, err := open(&row)
	if err != nil {
		return nil, err
	}
	if err := o.DB.Create(&row).Error; err != nil {
		_ = w.Close()
		removeFiles(o, files)
		return nil, err
	}
	return &segment{row: row, w: w, files: files}, nil
}

func removeFiles(o Options, files []string) {
	for _, f := range files {
		_ = o.FS.Remove(filepath.Join(o.Dir, f))
	}
}

// finish completes the files of the segment and stores its end; a segment
// without a picture is removed.
func (s *segment) finish(o Options) {
	if s == nil {
		return
	}
	err := s.w.Close()
	if s.w.Samples() == 0 || err != nil {
		s.discard(o, err)
		return
	}
	s.store(o)
}

// discard removes the files and the row of a segment that can't be kept.
func (s *segment) discard(o Options, closeErr error) {
	if closeErr != nil && s.w.Samples() > 0 {
		o.Log.LogError("Session recording: unable to save the recording: %v", closeErr)
	}
	removeFiles(o, s.files)
	o.DB.Delete(&s.row)
}

// store saves the end, size and picture size of a complete segment.
func (s *segment) store(o Options) {
	ended := s.row.StartedAt.Add(time.Duration(s.w.Duration()) * time.Microsecond)
	updates := map[string]interface{}{"ended_at": ended, "size": s.w.Size()}
	if d, ok := s.w.(sized); ok {
		updates["width"], updates["height"] = d.dimensions()
	}
	o.DB.Model(&s.row).Updates(updates)
}
