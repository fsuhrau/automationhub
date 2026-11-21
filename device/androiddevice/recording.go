package androiddevice

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/sirupsen/logrus"
)

const (
	recordingFile = "/data/local/tmp/automation_hub_record.mp4"
	// recordingFlush gives screenrecord time to finalize the file after SIGINT
	recordingFlush = 500 * time.Millisecond
)

// recording is the running screenrecord of a device.
type recording struct {
	path string
	proc device.Process
}

// StartRecording implements device.Recorder.
func (d *Device) StartRecording(path string) error {
	d.Log("device", "Start Recording Session")
	proc, err := d.adb().Start(context.Background(), "shell", "screenrecord", "--verbose", recordingFile)
	if err != nil {
		return err
	}
	d.recording = recording{path: fmt.Sprintf("./%s.mp4", path), proc: proc}
	return nil
}

// StopRecording implements device.Recorder: it stops screenrecord, pulls the
// file and moves it to the path given at the start.
func (d *Device) StopRecording() error {
	d.Log("device", "Stop Recording Session")
	if err := d.stopRecorder(); err != nil {
		return err
	}
	<-d.deps.Clock.After(recordingFlush)
	if _, err := d.run("pull", recordingFile); err != nil {
		return err
	}
	return os.Rename("automation_hub_record.mp4", d.recording.path)
}

// stopRecorder interrupts screenrecord on the device (SIGINT lets it finish
// the mp4, killing the adb client would not) and waits for the adb session.
func (d *Device) stopRecorder() error {
	if d.recording.proc == nil {
		return nil
	}
	if _, err := d.run("shell", "pkill", "-2", "screenrecord"); err != nil {
		return err
	}
	if err := d.recording.proc.Wait(); err != nil {
		logrus.Infof("stop recording session failed: %v", err)
	}
	d.recording.proc = nil
	return nil
}
