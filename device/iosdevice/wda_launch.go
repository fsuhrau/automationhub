package iosdevice

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/modules/webdriver"
)

// attachAliveTimeout bounds the /status probe of a running WebDriverAgent.
const attachAliveTimeout = 5 * time.Second

// launcher starts or attaches the WebDriverAgent of one device.
type launcher struct {
	cfg  *config.WebDriver
	deps Deps
	logf func(format string, args ...any)
}

// startXcodebuild starts xcodebuild and waits until WebDriverAgent logged its
// URL. It forwards the WebDriverAgent and MJPEG ports of the device to local
// ports, so the device's network doesn't matter.
func (l *launcher) startXcodebuild(udid string) (*wdaRunner, error) {
	args, err := xcodebuildArgs(l.cfg, udid)
	if err != nil {
		return nil, err
	}
	l.logf("Start WebDriverAgent: xcodebuild %s", strings.Join(args, " "))
	r, out, err := l.spawn(args)
	if err != nil {
		return nil, err
	}
	address, err := l.awaitURL(r, out)
	if err != nil {
		return nil, err
	}
	l.logf("WebDriverAgent reported %s", address)
	if err := l.forwardPorts(r, udid, devicePortOf(address)); err != nil {
		r.Stop()
		return nil, err
	}
	return r, nil
}

// xcodebuildOutput is what the xcodebuild output scanner found.
type xcodebuildOutput struct {
	found   chan string
	tail    *lineTail
	scanned chan struct{} // closed when the output ended
}

// spawn starts xcodebuild and scans its output for the WebDriverAgent URL.
func (l *launcher) spawn(args []string) (*wdaRunner, *xcodebuildOutput, error) {
	pr, pw := io.Pipe()
	proc, err := l.deps.Runner.StartStreaming(context.Background(), pw, "xcodebuild", args...)
	if err != nil {
		return nil, nil, fmt.Errorf("start xcodebuild: %w", err)
	}
	r := &wdaRunner{proc: proc, done: make(chan struct{}), after: l.deps.Clock.After}
	go func() {
		_ = proc.Wait()
		_ = pw.Close()
		close(r.done)
	}()
	out := &xcodebuildOutput{found: make(chan string, 1), tail: newLineTail(20), scanned: make(chan struct{})}
	go out.scan(pr)
	return r, out, nil
}

func (o *xcodebuildOutput) scan(pr io.Reader) {
	scanner := bufio.NewScanner(pr)
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		o.tail.add(line)
		if address, ok := webdriver.ParseServerURL(line); ok {
			select {
			case o.found <- address:
			default:
			}
		}
	}
	close(o.scanned)
	// keep draining, so xcodebuild never blocks on a full pipe
	_, _ = io.Copy(io.Discard, pr)
}

// awaitURL waits for the URL xcodebuild logs; a runner that ends or stays
// silent is stopped.
func (l *launcher) awaitURL(r *wdaRunner, out *xcodebuildOutput) (string, error) {
	select {
	case address := <-out.found:
		return address, nil
	case <-r.done:
		<-out.scanned // the last lines say why it ended
		return "", fmt.Errorf("xcodebuild ended before WebDriverAgent started:\n%s", out.tail.String())
	case <-l.deps.Clock.After(XcodebuildStartTimeout):
		r.Stop()
		return "", fmt.Errorf("WebDriverAgent did not report its server URL within %v:\n%s", XcodebuildStartTimeout, out.tail.String())
	}
}

// devicePortOf is the port in the URL WebDriverAgent logged.
func devicePortOf(address string) int {
	if u, err := url.Parse(address); err == nil && u.Port() != "" {
		if port, err := strconv.Atoi(u.Port()); err == nil && port > 0 && port <= math.MaxUint16 {
			return port
		}
	}
	return wdaDevicePort
}

// forwardPorts finds the device and forwards its ports, see forward.
func (l *launcher) forwardPorts(r *wdaRunner, udid string, devicePort int) error {
	entry, err := l.deps.USB.GetDevice(udid)
	if err != nil {
		return err
	}
	return l.forward(r, entry, devicePort)
}

// forward opens local ports for the WebDriverAgent and MJPEG ports of the
// device and records them in the runner.
func (l *launcher) forward(r *wdaRunner, entry ios.DeviceEntry, devicePort int) error {
	for _, port := range []int{devicePort, webdriver.DefaultMjpegServerPort} {
		f, err := l.deps.USB.Forward(entry, uint16(port))
		if err != nil {
			return err
		}
		r.forwards = append(r.forwards, f)
	}
	r.Address = "http://" + r.forwards[0].Addr()
	r.MjpegAddress = "http://" + r.forwards[1].Addr() + "/"
	return nil
}

// attach connects to a WebDriverAgent that already runs on the device
// (started from Xcode, Appium or an earlier hub), over usbmux. The error says
// why none answers on the default port. Stopping such a runner only ends the
// forwards.
func (l *launcher) attach(udid string) (*wdaRunner, error) {
	entry, err := l.deps.USB.GetDevice(udid)
	if err != nil {
		return nil, fmt.Errorf("usbmux: %w", err)
	}
	if err := l.deps.USB.Probe(entry, wdaDevicePort); err != nil {
		return nil, err
	}
	r := &wdaRunner{after: l.deps.Clock.After}
	if err := l.forward(r, entry, wdaDevicePort); err != nil {
		r.Stop()
		return nil, err
	}
	if !webdriver.New(r.Address, l.deps.HTTP).Alive(attachAliveTimeout) {
		r.Stop()
		return nil, fmt.Errorf("port %d on the device accepts connections, but /status doesn't answer like WebDriverAgent", wdaDevicePort)
	}
	return r, nil
}
