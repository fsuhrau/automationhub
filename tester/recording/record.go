package recording

import (
	"bytes"
	"context"
	"path/filepath"
	"sync"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

// streamRecorder writes the packets of one stream into segments. A new
// segment starts when the H.264 configuration changes (e.g. the screen
// rotated), since an MP4 track keeps one.
type streamRecorder struct {
	o        Options
	seg      *segment
	sps, pps []byte
	firstPTS int64
}

// record writes the stream into segments until it ends or ctx is done.
func record(ctx context.Context, stream device.VideoStream, o Options) {
	defer closeWhenDone(ctx, stream)()
	s := &streamRecorder{o: o}
	defer func() { s.seg.finish(o) }()
	codec := stream.Meta().Codec
	for {
		pkt, err := stream.ReadPacket()
		if err != nil || !s.handle(codec, pkt) {
			return
		}
	}
}

// closeWhenDone closes the stream once ctx is done; the returned func stops
// the watching and closes the stream.
func closeWhenDone(ctx context.Context, stream device.VideoStream) (cleanup func()) {
	streamCtx, stop := context.WithCancel(ctx)
	var once sync.Once
	closeStream := func() { once.Do(func() { _ = stream.Close() }) }
	go func() {
		<-streamCtx.Done()
		closeStream()
	}()
	return func() {
		stop()
		closeStream()
	}
}

// handle records one packet; it returns false when recording must end.
func (s *streamRecorder) handle(codec string, pkt device.VideoPacket) bool {
	switch codec {
	case device.CodecH264:
		return s.h264(pkt)
	case device.CodecMJPEG:
		return s.mjpeg(pkt)
	}
	s.o.Log.LogInfo("Session recording: video codec %q can't be recorded", codec)
	return false
}

func (s *streamRecorder) h264(pkt device.VideoPacket) bool {
	if s.applyConfig(pkt) {
		return true
	}
	if s.seg == nil {
		if !pkt.KeyFrame || s.sps == nil {
			return true
		}
		if !s.openSegment(models.RecordingMP4, s.openMP4) {
			return false
		}
		s.firstPTS = pkt.PTS
	}
	return s.ok(s.seg.w.(*mp4Writer).WriteSample(pkt.Data, s.h264Time(pkt), pkt.KeyFrame))
}

// applyConfig takes the parameter sets of the packet and ends the segment when
// they changed. It returns true when the packet has nothing to write.
func (s *streamRecorder) applyConfig(pkt device.VideoPacket) (skip bool) {
	newSPS, newPPS := parameterSets(pkt.Data)
	if !pkt.Config && newSPS == nil {
		return false
	}
	if newSPS == nil || newPPS == nil {
		return true
	}
	if s.seg != nil && (!bytes.Equal(newSPS, s.sps) || !bytes.Equal(newPPS, s.pps)) {
		s.seg.finish(s.o)
		s.seg = nil
	}
	s.sps, s.pps = newSPS, newPPS
	return pkt.Config
}

// h264Time is the time of the packet since the start of the segment; a packet
// without a usable PTS gets the time it arrived.
func (s *streamRecorder) h264Time(pkt device.VideoPacket) int64 {
	t := pkt.PTS - s.firstPTS
	if pkt.PTS <= 0 || t < 0 {
		return s.elapsed()
	}
	return t
}

func (s *streamRecorder) mjpeg(pkt device.VideoPacket) bool {
	if s.seg == nil && !s.openSegment(models.RecordingMJPEG, s.openMJPEG) {
		return false
	}
	// MJPEG frames carry no reliable time; the time they arrive is close
	// enough and on the clock of the run log
	return s.ok(s.seg.w.(*mjpegWriter).WriteFrame(pkt.Data, s.elapsed()))
}

// elapsed is the time since the start of the segment in microseconds.
func (s *streamRecorder) elapsed() int64 {
	return s.o.Clock.Now().Sub(s.seg.row.StartedAt).Microseconds()
}

// ok logs the error of a write and returns false for it.
func (s *streamRecorder) ok(err error) bool {
	if err != nil {
		s.o.Log.LogError("Session recording: %v", err)
	}
	return err == nil
}

func (s *streamRecorder) openSegment(codec string, open openFunc) bool {
	seg, err := startSegment(s.o, codec, open)
	s.seg = seg
	return s.ok(err)
}

func (s *streamRecorder) openMP4(row *models.TestRunRecording) (writer, []string, error) {
	row.FileName = newName(s.o, "mp4")
	w, err := newMP4Writer(s.o.FS, filepath.Join(s.o.Dir, row.FileName), h264Config{sps: s.sps, pps: s.pps})
	if err != nil {
		return nil, nil, err
	}
	row.Width, row.Height = w.width, w.height
	return w, []string{row.FileName}, nil
}

func (s *streamRecorder) openMJPEG(row *models.TestRunRecording) (writer, []string, error) {
	row.FileName = newName(s.o, "mjpeg")
	row.IndexFile = row.FileName + ".idx"
	w, err := newMJPEGWriter(s.o.FS, filepath.Join(s.o.Dir, row.FileName), filepath.Join(s.o.Dir, row.IndexFile))
	if err != nil {
		return nil, nil, err
	}
	return w, []string{row.FileName, row.IndexFile}, nil
}
