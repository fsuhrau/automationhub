package iosdevice

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/modules/webdriver"
)

// Launchers of WebDriverAgent (config webdriver.launcher).
const (
	// LauncherAuto uses xcodebuild for iOS 17 and later, go-ios before.
	LauncherAuto = "auto"
	// LauncherXcodebuild runs the WebDriverAgentRunner with xcodebuild, which
	// mounts the developer disk image and sets up the iOS 17 tunnel itself.
	LauncherXcodebuild = "xcodebuild"
	// LauncherGoIOS runs it through testmanagerd with go-ios; up to iOS 16.
	LauncherGoIOS = "go-ios"

	defaultWDAScheme = "WebDriverAgentRunner"
	// XcodebuildStartTimeout bounds an xcodebuild start; "xcodebuild test"
	// may build first.
	XcodebuildStartTimeout = 120 * time.Second
	wdaDevicePort          = 8100
)

// useXcodebuild decides the launcher for an iOS version like "18.7.1".
func useXcodebuild(cfg *config.WebDriver, osVersion string) bool {
	launcher := LauncherAuto
	if cfg != nil && cfg.Launcher != "" {
		launcher = strings.ToLower(cfg.Launcher)
	}
	switch launcher {
	case LauncherXcodebuild:
		return true
	case LauncherGoIOS:
		return false
	}
	major, _ := strconv.Atoi(strings.SplitN(osVersion, ".", 2)[0])
	return major >= 17
}

// xcodebuildArgs runs a prebuilt runner (.xctestrun from "xcodebuild
// build-for-testing") or builds and runs the WebDriverAgent project.
func xcodebuildArgs(cfg *config.WebDriver, udid string) ([]string, error) {
	if cfg == nil || (cfg.XCTestRun == "" && cfg.Project == "") {
		return nil, errors.New("iOS 17 and later start WebDriverAgent with xcodebuild: set webdriver.xctestrun (a WebDriverAgentRunner .xctestrun) or webdriver.project (WebDriverAgent.xcodeproj) of the ios_device manager on this host")
	}
	destination := "id=" + udid
	var args []string
	if cfg.XCTestRun != "" {
		args = []string{"test-without-building", "-xctestrun", cfg.XCTestRun, "-destination", destination}
	} else {
		scheme := cfg.Scheme
		if scheme == "" {
			scheme = defaultWDAScheme
		}
		args = []string{"test", "-project", cfg.Project, "-scheme", scheme, "-destination", destination, "-allowProvisioningUpdates"}
	}
	if cfg.DerivedData != "" {
		args = append(args, "-derivedDataPath", cfg.DerivedData)
	}
	return args, nil
}

// wdaRunner is a WebDriverAgentRunner started by xcodebuild, reachable on
// local ports forwarded over usbmux.
type wdaRunner struct {
	cmd      *exec.Cmd
	done     chan struct{}
	forwards []*portForward
	// Address and MjpegAddress are the local WebDriverAgent and MJPEG URLs.
	Address      string
	MjpegAddress string
}

// startXcodebuildRunner starts xcodebuild and waits until WebDriverAgent
// logged its URL. It forwards the WebDriverAgent and MJPEG ports of the
// device to local ports, so the device's network doesn't matter.
func startXcodebuildRunner(cfg *config.WebDriver, udid string, logf func(format string, args ...any)) (*wdaRunner, error) {
	args, err := xcodebuildArgs(cfg, udid)
	if err != nil {
		return nil, err
	}
	logf("Start WebDriverAgent: xcodebuild %s", strings.Join(args, " "))
	cmd := exec.Command("xcodebuild", args...)
	setProcessGroup(cmd)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start xcodebuild: %w", err)
	}
	r := &wdaRunner{cmd: cmd, done: make(chan struct{})}
	go func() {
		_ = cmd.Wait()
		_ = pw.Close()
		close(r.done)
	}()

	found := make(chan string, 1)
	tail := newLineTail(20)
	go func() {
		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 64<<10), 1<<20)
		for scanner.Scan() {
			line := scanner.Text()
			tail.add(line)
			if address, ok := webdriver.ParseServerURL(line); ok {
				select {
				case found <- address:
				default:
				}
			}
		}
		// keep draining, so xcodebuild never blocks on a full pipe
		_, _ = io.Copy(io.Discard, pr)
	}()

	var address string
	select {
	case address = <-found:
	case <-r.done:
		return nil, fmt.Errorf("xcodebuild ended before WebDriverAgent started:\n%s", tail.String())
	case <-time.After(XcodebuildStartTimeout):
		r.Stop()
		return nil, fmt.Errorf("WebDriverAgent did not report its server URL within %v:\n%s", XcodebuildStartTimeout, tail.String())
	}
	logf("WebDriverAgent reported %s", address)

	devicePort := wdaDevicePort
	if u, err := url.Parse(address); err == nil && u.Port() != "" {
		devicePort, _ = strconv.Atoi(u.Port())
	}
	entry, err := ios.GetDevice(udid)
	if err != nil {
		r.Stop()
		return nil, err
	}
	for _, port := range []int{devicePort, webdriver.DefaultMjpegServerPort} {
		f, err := forwardPort(entry, uint16(port))
		if err != nil {
			r.Stop()
			return nil, err
		}
		r.forwards = append(r.forwards, f)
	}
	r.Address = "http://" + r.forwards[0].Addr()
	r.MjpegAddress = "http://" + r.forwards[1].Addr() + "/"
	return r, nil
}

// attachRunningWDA connects to a WebDriverAgent that already runs on the
// device (started from Xcode, Appium or an earlier hub), over usbmux. The
// error says why none answers on the default port. Stopping such a runner
// only ends the forwards.
func attachRunningWDA(udid string) (*wdaRunner, error) {
	entry, err := ios.GetDevice(udid)
	if err != nil {
		return nil, fmt.Errorf("usbmux: %w", err)
	}
	if err := probeDevicePort(entry, wdaDevicePort); err != nil {
		return nil, err
	}
	r := &wdaRunner{}
	for _, port := range []int{wdaDevicePort, webdriver.DefaultMjpegServerPort} {
		f, err := forwardPort(entry, uint16(port))
		if err != nil {
			r.Stop()
			return nil, err
		}
		r.forwards = append(r.forwards, f)
	}
	r.Address = "http://" + r.forwards[0].Addr()
	r.MjpegAddress = "http://" + r.forwards[1].Addr() + "/"
	if !webdriver.New(r.Address).Alive(5 * time.Second) {
		r.Stop()
		return nil, fmt.Errorf("port %d on the device accepts connections, but /status doesn't answer like WebDriverAgent", wdaDevicePort)
	}
	return r, nil
}

// probeDevicePort checks that something listens on a port of the device.
func probeDevicePort(entry ios.DeviceEntry, port uint16) error {
	mux, err := ios.NewUsbMuxConnectionSimple()
	if err != nil {
		return fmt.Errorf("usbmux: %w", err)
	}
	defer mux.Close()
	if err := mux.Connect(entry.DeviceID, port); err != nil {
		return fmt.Errorf("nothing listens on port %d of the device (%v)", port, err)
	}
	return nil
}

// Stop ends xcodebuild and the port forwards.
func (r *wdaRunner) Stop() {
	for _, f := range r.forwards {
		f.Close()
	}
	if r.cmd == nil || r.cmd.Process == nil {
		// attached to a runner someone else started
		return
	}
	interruptProcessGroup(r.cmd)
	select {
	case <-r.done:
	case <-time.After(5 * time.Second):
		killProcessGroup(r.cmd)
		<-r.done
	}
}

// lineTail keeps the last lines of the xcodebuild output for errors.
type lineTail struct {
	mu    sync.Mutex
	max   int
	lines []string
}

func newLineTail(n int) *lineTail { return &lineTail{max: n} }

func (t *lineTail) add(line string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lines = append(t.lines, line)
	if len(t.lines) > t.max {
		t.lines = t.lines[len(t.lines)-t.max:]
	}
}

func (t *lineTail) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return strings.Join(t.lines, "\n")
}

// portForward forwards connections to a local port to a port on the device
// over usbmux. Unlike go-ios' forward package it can be closed.
type portForward struct {
	ln     net.Listener
	ctx    context.Context
	cancel context.CancelFunc
}

func forwardPort(entry ios.DeviceEntry, phonePort uint16) (*portForward, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	f := &portForward{ln: ln, ctx: ctx, cancel: cancel}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go f.proxy(conn, entry.DeviceID, phonePort)
		}
	}()
	return f, nil
}

func (f *portForward) Addr() string { return f.ln.Addr().String() }

func (f *portForward) Close() {
	f.cancel()
	_ = f.ln.Close()
}

func (f *portForward) proxy(conn net.Conn, deviceID int, phonePort uint16) {
	defer conn.Close()
	mux, err := ios.NewUsbMuxConnectionSimple()
	if err != nil {
		return
	}
	if err := mux.Connect(deviceID, phonePort); err != nil {
		mux.Close()
		return
	}
	dc := mux.ReleaseDeviceConnection()
	defer dc.Close()
	ctx, cancel := context.WithCancel(f.ctx)
	go func() {
		_, _ = io.Copy(conn, dc.Reader())
		cancel()
	}()
	go func() {
		_, _ = io.Copy(dc.Writer(), conn)
		cancel()
	}()
	<-ctx.Done()
}
