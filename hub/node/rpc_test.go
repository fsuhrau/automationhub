package node

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
)

func TestMissingDevice_SameErrorOnEveryCall(t *testing.T) {
	p := wireParameter()
	cfg := &device.DeviceConfig{}
	calls := map[string]func(c *RPCClient) error{
		"StartDevice":    func(c *RPCClient) error { return c.StartDevice("ghost") },
		"StopDevice":     func(c *RPCClient) error { return c.StopDevice("ghost") },
		"IsAppInstalled": func(c *RPCClient) error { _, err := c.IsAppInstalled("ghost", p); return err },
		"InstallApp":     func(c *RPCClient) error { return c.InstallApp("ghost", p) },
		"UninstallApp":   func(c *RPCClient) error { return c.UninstallApp("ghost", p) },
		"StartApp":       func(c *RPCClient) error { return c.StartApp("ghost", cfg, p, "s", "n") },
		"StopApp":        func(c *RPCClient) error { return c.StopApp("ghost", p) },
		"ClearAppData":   func(c *RPCClient) error { return c.ClearAppData("ghost", p) },
		"Control":        func(c *RPCClient) error { _, err := c.Control(device.ControlRequest{DeviceID: "ghost"}); return err },
		"StartVideoStream": func(c *RPCClient) error {
			return c.StartVideoStream(device.VideoStreamRequest{DeviceID: "ghost"})
		},
		"GetScreenshot": func(c *RPCClient) error { _, _, _, err := c.GetScreenshot("ghost"); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			err := call(newWire(t).client)
			if err == nil || err.Error() != "device with id ghost not found" {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestMissingDevice_QueriesFallBack(t *testing.T) {
	w := newWire(t)
	if w.client.IsConnected("ghost") || w.client.HasFeature("ghost", "f") {
		t.Error("a missing device is neither connected nor has features")
	}
	if got := w.client.ConnectionTimeout("ghost"); got != defaultConnectionTimeout {
		t.Errorf("timeout = %v", got)
	}
	w.client.Execute("ghost", "x")
	w.client.RunNativeScript("ghost", nil)
	w.client.SendAction("ghost", nil)
}

func TestWithDevice_RunsFnForKnownDevice(t *testing.T) {
	w := newWire(t)
	var got string
	err := w.node.withDevice("dev-1", func(d device.Device) error { got = d.DeviceID(); return nil })
	if err != nil || got != "dev-1" {
		t.Fatalf("got %q, %v", got, err)
	}
	boom := errors.New("boom")
	if err := w.node.withDevice("dev-1", func(device.Device) error { return boom }); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
	if err := w.node.withDevice("nope", func(device.Device) error { t.Error("fn ran"); return nil }); err == nil ||
		err.Error() != "device with id nope not found" {
		t.Fatalf("err = %v", err)
	}
}

func TestDeviceCalls_ReachTheDevice(t *testing.T) {
	w := newWire(t)
	w.dev.Installed, w.dev.AppConnected, w.dev.Timeout = true, true, 42*time.Second
	w.dev.Features = []string{"lua"}
	w.dev.LogLines = []string{"a", "b"}
	c, p := w.client, wireParameter()

	if ok, err := c.IsAppInstalled("dev-1", p); err != nil || !ok {
		t.Errorf("IsAppInstalled = %v, %v", ok, err)
	}
	if !c.IsConnected("dev-1") || !c.HasFeature("dev-1", "lua") || c.HasFeature("dev-1", "x") {
		t.Error("IsConnected/HasFeature")
	}
	if got := c.ConnectionTimeout("dev-1"); got != 42*time.Second {
		t.Errorf("timeout = %v", got)
	}
	resp, err := c.Control(device.ControlRequest{DeviceID: "dev-1", Op: device.ControlReadLogs, Lines: 2})
	if err != nil || len(resp.Lines) != 2 {
		t.Errorf("Control = %+v, %v", resp, err)
	}
	c.Execute("dev-1", "cmd")
	c.RunNativeScript("dev-1", []byte("lua"))
	c.SendAction("dev-1", []byte("act"))
	for _, m := range []string{"Execute", "RunNativeScript", "Send"} {
		if len(w.dev.CallsTo(m)) != 1 {
			t.Errorf("%s not forwarded", m)
		}
	}
	if err := c.StartDevice("dev-1"); err != nil {
		t.Error(err)
	}
	if err := c.StopDevice("dev-1"); err != nil {
		t.Error(err)
	}
	if len(w.dm.CallsTo("Start")) != 1 || len(w.dm.CallsTo("Stop")) != 1 {
		t.Error("Start/Stop not forwarded")
	}
}

func TestAppCalls_ReachTheDeviceAndReportItsErrors(t *testing.T) {
	boom := errors.New("adb failed")
	calls := map[string]func(c *RPCClient) error{
		"InstallApp":   func(c *RPCClient) error { return c.InstallApp("dev-1", wireParameter()) },
		"UninstallApp": func(c *RPCClient) error { return c.UninstallApp("dev-1", wireParameter()) },
		"StartApp": func(c *RPCClient) error {
			return c.StartApp("dev-1", &device.DeviceConfig{}, wireParameter(), "s", "n")
		},
		"StopApp":      func(c *RPCClient) error { return c.StopApp("dev-1", wireParameter()) },
		"ClearAppData": func(c *RPCClient) error { return c.ClearAppData("dev-1", wireParameter()) },
		"IsAppInstalled": func(c *RPCClient) error {
			_, err := c.IsAppInstalled("dev-1", wireParameter())
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			w := newWire(t)
			if err := call(w.client); err != nil {
				t.Fatalf("success: %v", err)
			}
			if len(w.dev.CallsTo(name)) != 1 {
				t.Fatalf("device not called: %v", w.dev.Calls())
			}
			w.dev.FailWith(name, boom)
			if err := call(w.client); err == nil || err.Error() != boom.Error() {
				t.Fatalf("failure: %v", err)
			}
		})
	}
}

func TestStopAndClear_DropTheAppConnection(t *testing.T) {
	for name, call := range map[string]func(c *RPCClient) error{
		"StopApp":      func(c *RPCClient) error { return c.StopApp("dev-1", wireParameter()) },
		"ClearAppData": func(c *RPCClient) error { return c.ClearAppData("dev-1", wireParameter()) },
	} {
		t.Run(name, func(t *testing.T) {
			w := newWire(t)
			w.dev.SetConnection(&device.Connection{})
			if err := call(w.client); err != nil {
				t.Fatal(err)
			}
			if w.dev.Connection() != nil {
				t.Error("connection kept")
			}
		})
	}
}

// bare hides every optional capability of the fake device.
type bare struct{ device.Device }

func TestOptionalCapabilities_UnsupportedDevice(t *testing.T) {
	w := newWire(t)
	w.dm.Add("android_device", bare{testDeviceNamed("bare-1")})
	if err := w.client.ClearAppData("bare-1", wireParameter()); !errors.Is(err, device.ErrUnsupported) && (err == nil || err.Error() != device.ErrUnsupported.Error()) {
		t.Errorf("ClearAppData: %v", err)
	}
	err := w.client.StartVideoStream(device.VideoStreamRequest{DeviceID: "bare-1"})
	if !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("StartVideoStream: %v", err)
	}
}

func TestStartVideoStream_DialsBackToTheMaster(t *testing.T) {
	w := newWire(t)
	err := w.client.StartVideoStream(device.VideoStreamRequest{DeviceID: "dev-1", StreamID: "stream-9"})
	if err != nil {
		t.Fatal(err)
	}
	if got := <-w.videoPaths; got != "/node/video/stream-9" {
		t.Errorf("path = %q", got)
	}
	w.dev.FailWith("StartVideo", errors.New("no scrcpy"))
	if err := w.client.StartVideoStream(device.VideoStreamRequest{DeviceID: "dev-1"}); err == nil || err.Error() != "no scrcpy" {
		t.Errorf("err = %v", err)
	}
}

func TestAppBundle_UnknownHash(t *testing.T) {
	w := newWire(t)
	p := wireParameter()
	p.App.Hash = "unknown"
	if ok, err := w.client.IsAppUploaded(p); err != nil || ok {
		t.Errorf("IsAppUploaded = %v, %v", ok, err)
	}
	if ok, err := w.client.IsAppInstalled("dev-1", p); err != nil || ok {
		t.Errorf("IsAppInstalled = %v, %v", ok, err)
	}
	if err := w.client.InstallApp("dev-1", p); err == nil {
		t.Error("InstallApp must fail without the bundle")
	}
	if ok, err := w.client.IsAppUploaded(wireParameter()); err != nil || !ok {
		t.Errorf("IsAppUploaded(known) = %v, %v", ok, err)
	}
}

func TestGetDevices_ListsAndFails(t *testing.T) {
	w := newWire(t)
	w.dev.Name = "Pixel"
	got, err := w.client.GetDevices()
	if err != nil || len(got["android_device"]) != 1 || got["android_device"][0].DeviceName() != "Pixel" {
		t.Fatalf("GetDevices = %v, %v", got, err)
	}
	w.dm.FailWith("Devices", errors.New("scan failed"))
	if _, err := w.client.GetDevices(); err == nil || err.Error() != "scan failed" {
		t.Errorf("err = %v", err)
	}
}

func TestDeviceStatus(t *testing.T) {
	w := newWire(t)
	var resp DeviceResponse
	if err := w.node.DeviceStatus(&DeviceStatusRequest{DeviceId: "dev-1"}, &resp); err != nil || resp.DeviceID != "dev-1" || resp.Manager != "android_device" {
		t.Fatalf("resp = %+v, %v", &resp, err)
	}
	if err := w.node.DeviceStatus(&DeviceStatusRequest{DeviceId: "x"}, &resp); err == nil {
		t.Error("expected not found")
	}
}

type pingCounter struct{ n int }

func (p *pingCounter) PingReceived() { p.n++ }

func TestPing_NotifiesTheConnectionHandler(t *testing.T) {
	w := newWire(t)
	counter := &pingCounter{}
	w.node.connectionHandler = counter
	if err := w.client.Ping(); err != nil || counter.n != 1 {
		t.Fatalf("ping: %v, %d", err, counter.n)
	}
	w.node.connectionHandler = nil
	if err := w.client.Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestGetScreenshot_FetchesInChunks(t *testing.T) {
	w := newWire(t)
	w.dev.PNG = make([]byte, ScreenshotChunkSize*2+10)
	w.dev.PNG[len(w.dev.PNG)-1] = 7
	w.dev.Width, w.dev.Height = 3, 4
	data, width, height, err := w.client.GetScreenshot("dev-1")
	if err != nil || len(data) != len(w.dev.PNG) || data[len(data)-1] != 7 || width != 3 || height != 4 {
		t.Fatalf("len=%d %dx%d err=%v", len(data), width, height, err)
	}
	if len(w.node.screenshots) != 0 {
		t.Error("screenshot buffer not released")
	}
	w.dev.FailWith("GetScreenshot", errors.New("no display"))
	if _, _, _, err := w.client.GetScreenshot("dev-1"); err == nil {
		t.Error("expected the device error")
	}
}

func TestGetScreenshotData_Ranges(t *testing.T) {
	w := newWire(t)
	w.node.screenshots["d"] = []byte("0123456789")
	var resp ScreenShotDataResponse
	for _, r := range []*ScreenShotDataRequest{
		{Hash: "none", Start: 0, End: 1},
		{Hash: "d", Start: -1, End: 1},
		{Hash: "d", Start: 5, End: 2},
		{Hash: "d", Start: 11, End: 12},
	} {
		if err := w.node.GetScreenshotData(r, &resp); err == nil {
			t.Errorf("%v: expected an error", r)
		}
	}
	if err := w.node.GetScreenshotData(&ScreenShotDataRequest{Hash: "d", Start: 2, End: 4}, &resp); err != nil || string(resp.Data) != "23" {
		t.Fatalf("data = %q, %v", resp.Data, err)
	}
	if err := w.node.GetScreenshotData(&ScreenShotDataRequest{Hash: "d", Start: 8, End: 99}, &resp); err != nil || string(resp.Data) != "89" {
		t.Fatalf("data = %q, %v", resp.Data, err)
	}
	if _, ok := w.node.screenshots["d"]; ok {
		t.Error("buffer kept after the last chunk")
	}
}

func TestClosedConnection_QueriesFallBack(t *testing.T) {
	w := newWire(t)
	_ = w.client.client.Close()
	c := w.client
	if c.IsConnected("dev-1") || c.HasFeature("dev-1", "f") {
		t.Error("closed connection must read as not connected")
	}
	if c.ConnectionTimeout("dev-1") != defaultConnectionTimeout {
		t.Error("timeout fallback")
	}
	if _, err := c.GetDevices(); err == nil {
		t.Error("GetDevices")
	}
	if _, _, _, err := c.GetScreenshot("dev-1"); err == nil {
		t.Error("GetScreenshot")
	}
	if _, err := c.screenshotData(&ScreenShotResponse{Size: 5}); err == nil {
		t.Error("screenshotData")
	}
	if _, err := c.Control(device.ControlRequest{}); err == nil {
		t.Error("Control")
	}
	if err := c.StartVideoStream(device.VideoStreamRequest{}); err == nil {
		t.Error("StartVideoStream")
	}
	if err := c.UploadApp(context.Background(), wireParameter()); err == nil {
		t.Error("UploadApp")
	}
}

func TestSafeCall_RecoversFromPanic(t *testing.T) {
	if err := (&RPCClient{}).Ping(); err != nil {
		t.Fatalf("a panic in the call must be recovered, got %v", err)
	}
}

func TestResponseFailure(t *testing.T) {
	for name, resp := range map[string]interface{ failure() error }{
		"bool":  &BoolResponse{ErrorCode: 1, ErrorMessage: "bad"},
		"error": &ErrorResponse{ErrorCode: 2, ErrorMessage: "bad"},
	} {
		if err := resp.failure(); err == nil || err.Error() != "bad" {
			t.Errorf("%s: %v", name, err)
		}
	}
	if (&BoolResponse{}).failure() != nil || (&ErrorResponse{}).failure() != nil {
		t.Error("code 0 is no failure")
	}
}

// failingNode answers every call with an error code, like a node that
// reports failures inside the response.
type failingNode struct{}

func (failingNode) StartDevice(*DeviceRequest, *ErrorResponse) error {
	return nil
}

func (failingNode) StopDevice(_ *DeviceRequest, resp *ErrorResponse) error {
	resp.ErrorCode, resp.ErrorMessage = 3, "stop refused"
	return nil
}

func (failingNode) InstallApp(_ *AppParameterRequest, resp *BoolResponse) error {
	resp.ErrorCode, resp.ErrorMessage = 1, "install refused"
	return nil
}

func (failingNode) IsAppInstalled(_ *AppParameterRequest, resp *BoolResponse) error {
	resp.Value, resp.ErrorCode, resp.ErrorMessage = true, 1, "no"
	return nil
}

func (failingNode) HasFeature(_ *FeatureRequest, resp *BoolResponse) error {
	resp.Value, resp.ErrorCode = true, 1
	return nil
}

func TestForward_ReportedErrorsBecomeErrors(t *testing.T) {
	c := newWireFor(t, failingNode{}).client
	if err := c.StartDevice("d"); err != nil {
		t.Errorf("StartDevice: %v", err)
	}
	if err := c.StopDevice("d"); err == nil || err.Error() != "stop refused" {
		t.Errorf("StopDevice: %v", err)
	}
	if err := c.InstallApp("d", wireParameter()); err == nil || err.Error() != "install refused" {
		t.Errorf("InstallApp: %v", err)
	}
	if ok, err := c.IsAppInstalled("d", wireParameter()); ok || err == nil {
		t.Errorf("IsAppInstalled = %v, %v", ok, err)
	}
	if c.HasFeature("d", "f") {
		t.Error("a reported error means no feature")
	}
}

func TestControlError(t *testing.T) {
	if err := controlError(netrpcServerError(device.ErrUnsupported.Error())); !errors.Is(err, device.ErrUnsupported) {
		t.Errorf("err = %v", err)
	}
	plain := errors.New("plain")
	if err := controlError(plain); err != plain {
		t.Errorf("err = %v", err)
	}
}

// --- app upload ---

func serveUpload(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// download runs the node's download of url and returns the error it ended with.
func download(t *testing.T, w *wire, url string, size int64) (*UploadProgress, error) {
	t.Helper()
	progress := &UploadProgress{Request: &UploadAppRequest{AppID: 1, Name: "app.apk", Hash: "h2", Size: size, URL: url}}
	w.node.downloadFile(progress)
	return progress, progress.Error()
}

type downloadCase struct {
	handler http.HandlerFunc
	size    int64
	prepare func(w *wire)
}

func sendBody(body string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) { _, _ = rw.Write([]byte(body)) }
}

func TestDownloadFile_StoresTheBundle(t *testing.T) {
	const token = "tok"
	cases := map[string]downloadCase{
		"known size":   {handler: sendBody("abcd"), size: 4},
		"unknown size": {handler: sendBody("abcd"), size: 0},
		"with token": {
			handler: func(rw http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Auth-Token") != token {
					http.Error(rw, "no", http.StatusUnauthorized)
					return
				}
				_, _ = rw.Write([]byte("abcd"))
			},
			size:    4,
			prepare: func(w *wire) { w.node.config.Auth.Token = &config.Token{AuthToken: token} },
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			w := newWire(t)
			if tc.prepare != nil {
				tc.prepare(w)
			}
			progress, err := download(t, w, serveUpload(t, tc.handler).URL, tc.size)
			if err != nil || !progress.done.Load() || progress.received.Load() != 4 {
				t.Fatalf("err=%v done=%v received=%d", err, progress.done.Load(), progress.received.Load())
			}
			if _, err := w.bundle.GetAppParameter("h2"); err != nil {
				t.Errorf("bundle not stored: %v", err)
			}
		})
	}
}

func TestDownloadFile_Failures(t *testing.T) {
	truncated := func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Length", "10")
		_, _ = rw.Write([]byte("ab"))
	}
	cases := map[string]struct {
		downloadCase
		wantErr string
	}{
		"http status":    {downloadCase{handler: http.NotFound, size: 4}, "download: 404 Not Found"},
		"size mismatch":  {downloadCase{handler: sendBody("ab"), size: 4}, "size mismatch: got 2 bytes, expected 4"},
		"truncated body": {downloadCase{handler: truncated, size: 10}, "read file data: unexpected EOF"},
		"store fails": {downloadCase{handler: sendBody("abcd"), size: 4,
			prepare: func(w *wire) { w.bundle.dataDir = "/nonexistent/dir" }}, "store app bundle"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			w := newWire(t)
			if tc.prepare != nil {
				tc.prepare(w)
			}
			progress, err := download(t, w, serveUpload(t, tc.handler).URL, tc.size)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) || progress.done.Load() {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestDownloadFile_Unreachable(t *testing.T) {
	w := newWire(t)
	for url, want := range map[string]string{
		"http://127.0.0.1:1/x": "download:",
		"http://bad host/\x7f": "create request:",
	} {
		if _, err := download(t, w, url, 1); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: err = %v", url, err)
		}
	}
}

func TestReadCounting_ReadError(t *testing.T) {
	progress := &UploadProgress{Request: &UploadAppRequest{}}
	boom := errors.New("disk")
	if _, err := readCounting(iotest.ErrReader(boom), progress); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
	data, err := readCounting(strings.NewReader("xyz"), progress)
	if err != nil || string(data) != "xyz" || progress.received.Load() != 3 {
		t.Fatalf("data=%q err=%v received=%d", data, err, progress.received.Load())
	}
}

func TestUploadAppProgress_States(t *testing.T) {
	w := newWire(t)
	var resp UploadAppProgressResponse
	if err := w.node.UploadAppProgress(&UploadAppProgressRequest{AppID: 5}, &resp); err == nil {
		t.Fatal("unknown upload must fail")
	}

	pending := &UploadProgress{Request: &UploadAppRequest{AppID: 5, Size: 10}}
	w.node.uploadRequests[5] = pending
	pending.received.Store(4)
	_ = w.node.UploadAppProgress(&UploadAppProgressRequest{AppID: 5}, &resp)
	if resp.DataReceived != 4 {
		t.Errorf("received = %d", resp.DataReceived)
	}
	pending.received.Store(10) // downloaded, not stored yet
	_ = w.node.UploadAppProgress(&UploadAppProgressRequest{AppID: 5}, &resp)
	if resp.DataReceived != 9 {
		t.Errorf("received = %d, want size-1 until stored", resp.DataReceived)
	}
	pending.done.Store(true)
	_ = w.node.UploadAppProgress(&UploadAppProgressRequest{AppID: 5}, &resp)
	if _, kept := w.node.uploadRequests[5]; resp.DataReceived != 10 || kept {
		t.Errorf("received = %d, kept = %v", resp.DataReceived, kept)
	}

	failed := &UploadProgress{Request: &UploadAppRequest{AppID: 6}}
	failed.setError(errors.New("broken"))
	w.node.uploadRequests[6] = failed
	if err := w.node.UploadAppProgress(&UploadAppProgressRequest{AppID: 6}, &resp); err == nil || err.Error() != "broken" {
		t.Errorf("err = %v", err)
	}
	if _, kept := w.node.uploadRequests[6]; kept {
		t.Error("failed upload kept")
	}
}

func TestUploadApp_EndToEnd(t *testing.T) {
	w := newWire(t)
	p := wireParameter()
	p.App.Hash = "fresh"
	if err := w.client.UploadApp(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if ok, _ := w.client.IsAppUploaded(p); !ok {
		t.Error("bundle not stored on the node")
	}
}

func TestWaitForUpload(t *testing.T) {
	request := &UploadAppRequest{AppID: 8, Size: 100}

	t.Run("context ends", func(t *testing.T) {
		w := newWire(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := w.client.waitForUpload(ctx, request); !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("progress call fails", func(t *testing.T) {
		w := newWire(t) // the node knows no upload 8
		if err := w.client.waitForUpload(context.Background(), request); err == nil || !strings.Contains(err.Error(), "upload app to node:") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("stalls", func(t *testing.T) {
		w := newWire(t)
		w.client.uploadStall = 30 * time.Millisecond
		stuck := &UploadProgress{Request: request}
		stuck.received.Store(10)
		w.node.uploadRequests[8] = stuck
		err := w.client.waitForUpload(context.Background(), request)
		if err == nil || !strings.Contains(err.Error(), "no progress for 30ms (10/100 bytes)") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("keeps waiting while bytes arrive", func(t *testing.T) {
		w := newWire(t)
		w.client.uploadStall = 200 * time.Millisecond
		moving := &UploadProgress{Request: request}
		w.node.uploadRequests[8] = moving
		go func() {
			for i := int64(1); i <= 100; i += 33 {
				moving.received.Store(i)
				time.Sleep(15 * time.Millisecond)
			}
			moving.received.Store(100)
			moving.done.Store(true)
		}()
		if err := w.client.waitForUpload(context.Background(), request); err != nil {
			t.Fatal(err)
		}
	})
}

// brokenScreenshotNode takes the screenshot but fails to hand out its data.
type brokenScreenshotNode struct{}

func (brokenScreenshotNode) TakeScreenshot(_ *DeviceRequest, resp *ScreenShotResponse) error {
	resp.Size = 10
	return nil
}

func (brokenScreenshotNode) GetScreenshotData(*ScreenShotDataRequest, *ScreenShotDataResponse) error {
	return errors.New("buffer gone")
}

func TestGetScreenshot_ChunkFailure(t *testing.T) {
	c := newWireFor(t, brokenScreenshotNode{}).client
	if _, _, _, err := c.GetScreenshot("d"); err == nil || err.Error() != "buffer gone" {
		t.Fatalf("err = %v", err)
	}
}

func TestOptionalRoles_UnsupportedDevice(t *testing.T) {
	w := newWire(t)
	w.dm.Add("android_device", bare{testDeviceNamed("bare-1")})

	_, _, _, err := w.client.GetScreenshot("bare-1")
	if err == nil || !strings.Contains(err.Error(), device.ErrUnsupported.Error()) {
		t.Errorf("GetScreenshot: %v, want the unsupported error", err)
	}
	var features BoolResponse
	if err := w.node.HasFeature(&FeatureRequest{DeviceID: "bare-1", Feature: "f"}, &features); err != nil || features.Value {
		t.Errorf("HasFeature = %v, %v, want false without an error", features.Value, err)
	}
	if err := w.node.Execute(&ExecuteRequest{DeviceID: "bare-1", Data: "x"}, &Void{}); err != nil {
		t.Errorf("Execute: %v, want it skipped", err)
	}
	if err := w.node.RunNativeScript(&ExecuteRequest{DeviceID: "bare-1", Data: "x"}, &Void{}); err != nil {
		t.Errorf("RunNativeScript: %v, want it skipped", err)
	}
}

func TestOptionalRoles_ErrorsOfTheRoleStillReachTheMaster(t *testing.T) {
	w := newWire(t)
	if err := w.node.Execute(&ExecuteRequest{DeviceID: "ghost"}, &Void{}); err == nil {
		t.Error("an unknown device must still fail")
	}
}
