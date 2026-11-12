// Package recording records the screens of the devices of a test run, so the
// run can be replayed next to its log.
package recording

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// Logger writes to the run log.
type Logger interface {
	LogInfo(format string, params ...interface{})
	LogError(format string, params ...interface{})
}

// Options of a recording.
type Options struct {
	// Dir the files are written to; they are served from there by name.
	Dir      string
	DB       *gorm.DB
	RunID    uint
	DeviceID uint
	Log      Logger
	// Video asks the device for a smaller stream than the live view does:
	// a recording is kept and downloaded later.
	Video device.VideoOptions
}

// DefaultVideo keeps a recording at about 10 MB a minute.
var DefaultVideo = device.VideoOptions{MaxSize: 1280, MaxFPS: 30, BitRate: 1500000}

// restartDelay is the wait before the stream is opened again after it ended,
// e.g. while the device restarts.
var restartDelay = 2 * time.Second

// Recorder records the screen of one device until Stop.
type Recorder struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// Start records the screen of the device in the background. Devices without
// a video stream are skipped with a note in the run log.
func Start(dev device.Device, o Options) *Recorder {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Recorder{cancel: cancel, done: make(chan struct{})}
	if o.Video == (device.VideoOptions{}) {
		o.Video = DefaultVideo
	}
	go func() {
		defer close(r.done)
		run(ctx, dev, o)
	}()
	return r
}

// Stop ends the recording and waits until its files are complete.
func (r *Recorder) Stop() {
	r.cancel()
	select {
	case <-r.done:
	case <-time.After(30 * time.Second):
	}
}

func run(ctx context.Context, dev device.Device, o Options) {
	streamer, ok := dev.(device.VideoStreamer)
	if !ok {
		o.Log.LogInfo("Session recording: device %s has no video stream, it is not recorded", dev.DeviceID())
		return
	}
	failures := 0
	for ctx.Err() == nil {
		stream, err := streamer.StartVideo(ctx, o.Video)
		if err != nil {
			if errors.Is(err, device.ErrUnsupported) {
				o.Log.LogInfo("Session recording: device %s can't be recorded: %v", dev.DeviceID(), err)
				return
			}
			if failures == 0 && ctx.Err() == nil {
				o.Log.LogError("Session recording: unable to start the video of device %s: %v", dev.DeviceID(), err)
			}
			failures++
		} else {
			if failures > 0 || ctx.Err() == nil {
				o.Log.LogInfo("Session recording: recording device %s", dev.DeviceID())
			}
			failures = 0
			record(ctx, stream, o)
		}
		select {
		case <-ctx.Done():
		case <-time.After(restartDelay):
		}
	}
}

// writer is the file of a segment.
type writer interface {
	Close() error
	Size() int64
	Samples() int
	Duration() int64
}

type segment struct {
	row   models.TestRunRecording
	w     writer
	files []string
}

func newName(o Options, ext string) string {
	return fmt.Sprintf("recording-%d-%d-%d.%s", o.RunID, o.DeviceID, time.Now().UnixNano(), ext)
}

func startSegment(o Options, codec string, open func(row *models.TestRunRecording) (writer, []string, error)) (*segment, error) {
	row := models.TestRunRecording{TestRunID: o.RunID, DeviceID: o.DeviceID, Codec: codec, StartedAt: time.Now()}
	w, files, err := open(&row)
	if err != nil {
		return nil, err
	}
	if err := o.DB.Create(&row).Error; err != nil {
		_ = w.Close()
		removeFiles(o.Dir, files)
		return nil, err
	}
	return &segment{row: row, w: w, files: files}, nil
}

func removeFiles(dir string, files []string) {
	for _, f := range files {
		_ = os.Remove(filepath.Join(dir, f))
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
		if err != nil && s.w.Samples() > 0 {
			o.Log.LogError("Session recording: unable to save the recording: %v", err)
		}
		removeFiles(o.Dir, s.files)
		o.DB.Delete(&s.row)
		return
	}
	ended := s.row.StartedAt.Add(time.Duration(s.w.Duration()) * time.Microsecond)
	updates := map[string]interface{}{"ended_at": ended, "size": s.w.Size()}
	switch w := s.w.(type) {
	case *mp4Writer:
		updates["width"], updates["height"] = w.width, w.height
	case *mjpegWriter:
		updates["width"], updates["height"] = w.width, w.height
	}
	o.DB.Model(&s.row).Updates(updates)
}

// record writes the stream into segments until it ends or ctx is done. A new
// segment starts when the H.264 configuration changes (e.g. the screen
// rotated), since an MP4 track keeps one.
func record(ctx context.Context, stream device.VideoStream, o Options) {
	streamCtx, stop := context.WithCancel(ctx)
	defer stop()
	var once sync.Once
	closeStream := func() { once.Do(func() { _ = stream.Close() }) }
	defer closeStream()
	go func() {
		<-streamCtx.Done()
		closeStream()
	}()

	codec := stream.Meta().Codec
	var (
		seg      *segment
		sps, pps []byte
		firstPTS int64
	)
	defer func() { seg.finish(o) }()

	for {
		pkt, err := stream.ReadPacket()
		if err != nil {
			return
		}
		switch codec {
		case device.CodecH264:
			newSPS, newPPS := parameterSets(pkt.Data)
			if pkt.Config || newSPS != nil {
				if newSPS == nil || newPPS == nil {
					continue
				}
				if seg != nil && (!bytes.Equal(newSPS, sps) || !bytes.Equal(newPPS, pps)) {
					seg.finish(o)
					seg = nil
				}
				sps, pps = newSPS, newPPS
				if pkt.Config {
					continue
				}
			}
			if seg == nil {
				if !pkt.KeyFrame || sps == nil {
					continue
				}
				seg, err = startSegment(o, models.RecordingMP4, func(row *models.TestRunRecording) (writer, []string, error) {
					row.FileName = newName(o, "mp4")
					w, err := newMP4Writer(filepath.Join(o.Dir, row.FileName), sps, pps)
					if err != nil {
						return nil, nil, err
					}
					row.Width, row.Height = w.width, w.height
					return w, []string{row.FileName}, nil
				})
				if err != nil {
					o.Log.LogError("Session recording: %v", err)
					return
				}
				firstPTS = pkt.PTS
			}
			t := pkt.PTS - firstPTS
			if pkt.PTS <= 0 || t < 0 {
				t = time.Since(seg.row.StartedAt).Microseconds()
			}
			if err := seg.w.(*mp4Writer).WriteSample(pkt.Data, t, pkt.KeyFrame); err != nil {
				o.Log.LogError("Session recording: %v", err)
				return
			}
		case device.CodecMJPEG:
			if seg == nil {
				seg, err = startSegment(o, models.RecordingMJPEG, func(row *models.TestRunRecording) (writer, []string, error) {
					row.FileName = newName(o, "mjpeg")
					row.IndexFile = row.FileName + ".idx"
					w, err := newMJPEGWriter(filepath.Join(o.Dir, row.FileName), filepath.Join(o.Dir, row.IndexFile))
					if err != nil {
						return nil, nil, err
					}
					return w, []string{row.FileName, row.IndexFile}, nil
				})
				if err != nil {
					o.Log.LogError("Session recording: %v", err)
					return
				}
			}
			// MJPEG frames carry no reliable time; the time they arrive is
			// close enough and on the clock of the run log
			if err := seg.w.(*mjpegWriter).WriteFrame(pkt.Data, time.Since(seg.row.StartedAt).Microseconds()); err != nil {
				o.Log.LogError("Session recording: %v", err)
				return
			}
		default:
			o.Log.LogInfo("Session recording: video codec %q can't be recorded", codec)
			return
		}
	}
}

// parameterSets returns the SPS and PPS in the data, if any.
func parameterSets(data []byte) (sps, pps []byte) {
	for _, nal := range splitAnnexB(data) {
		switch nalType(nal) {
		case nalSPS:
			sps = append([]byte(nil), nal...)
		case nalPPS:
			pps = append([]byte(nil), nal...)
		}
	}
	return sps, pps
}
