package iosdevice

import (
	"time"

	"github.com/fsuhrau/automationhub/device"
)

// stopGracePeriod is how long xcodebuild gets to end after an interrupt.
const stopGracePeriod = 5 * time.Second

// wdaRunner is a WebDriverAgentRunner started by xcodebuild, reachable on
// local ports forwarded over usbmux.
type wdaRunner struct {
	proc     device.Process // nil when attached to a runner someone else started
	done     chan struct{}
	after    func(time.Duration) <-chan time.Time
	forwards []Forwarder
	// Address and MjpegAddress are the local WebDriverAgent and MJPEG URLs.
	Address      string
	MjpegAddress string
}

// Stop ends xcodebuild and the port forwards.
func (r *wdaRunner) Stop() {
	for _, f := range r.forwards {
		f.Close()
	}
	if r.proc == nil {
		// attached to a runner someone else started
		return
	}
	r.interrupt()
	select {
	case <-r.done:
	case <-r.after(stopGracePeriod):
		_ = r.proc.Kill()
		<-r.done
	}
}

// interrupt asks xcodebuild to end; a process that can't be interrupted is
// killed.
func (r *wdaRunner) interrupt() {
	if p, ok := r.proc.(device.Interrupter); ok {
		_ = p.Interrupt()
		return
	}
	_ = r.proc.Kill()
}
