package models

import "time"

// Codecs of TestRunRecording.Codec.
const (
	// RecordingMP4 is an H.264 MP4 file that plays in a video element.
	RecordingMP4 = "mp4"
	// RecordingMJPEG is a file of JPEG frames; IndexFile lists for each frame
	// its time in milliseconds since StartedAt, its offset and its length,
	// one "ms offset length" line per frame.
	RecordingMJPEG = "mjpeg"
)

// TestRunRecording is one stretch of video of a device's screen during a run.
// A device has several when its stream restarted, e.g. after the app or the
// device restarted or the screen rotated.
type TestRunRecording struct {
	Model
	TestRunID uint       `json:"testRunId" gorm:"index"`
	DeviceID  uint       `json:"deviceId"`
	Codec     string     `json:"codec"`
	FileName  string     `json:"fileName"`
	IndexFile string     `json:"indexFile,omitempty"`
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	StartedAt time.Time  `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt"`
	Size      int64      `json:"size"`
}
