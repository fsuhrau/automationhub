package unityeditor

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"time"
)

const (
	logStartDelay   = time.Second
	logPollInterval = 100 * time.Millisecond
)

// UnityLogStartListening follows the editor's log file until
// UnityLogStopListening.
func (d *Device) UnityLogStartListening() {
	ctx, cancel := context.WithCancel(context.Background())
	d.mu.Lock()
	d.logCancel = cancel
	file := d.instanceLogFile
	d.mu.Unlock()
	go d.processUnityLog(ctx, file)
}

func (d *Device) UnityLogStopListening() {
	d.mu.Lock()
	cancel := d.logCancel
	d.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (d *Device) processUnityLog(ctx context.Context, path string) {
	if !d.sleep(ctx, logStartDelay) {
		return
	}
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()
	d.followLog(ctx, bufio.NewReader(file))
}

// followLog reads the log until ctx ends or reading fails. The lines are not
// used yet.
func (d *Device) followLog(ctx context.Context, reader *bufio.Reader) {
	for ctx.Err() == nil {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return
		}
		if line == "" && !d.sleep(ctx, logPollInterval) {
			return
		}
	}
}

// sleep waits for the clock and reports false when ctx ended first.
func (d *Device) sleep(ctx context.Context, wait time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-d.clock.After(wait):
		return true
	}
}
