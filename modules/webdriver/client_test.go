package webdriver

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// reqLog is one request the WDA fake saw.
type reqLog struct{ Method, Path, Body string }

// wdaFake is an httptest WebDriverAgent. Routes are keyed "METHOD /path" and
// answer status and body; unknown routes answer 404 with a WDA error.
type wdaFake struct {
	*httptest.Server
	mu     sync.Mutex
	routes map[string]fakeReply
	seen   []reqLog
}

type fakeReply struct {
	status int
	body   string
}

func newWDAFake(t *testing.T, routes map[string]fakeReply) *wdaFake {
	t.Helper()
	f := &wdaFake{routes: routes}
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.Close)
	return f
}

func (f *wdaFake) serve(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	f.seen = append(f.seen, reqLog{r.Method, r.URL.Path, string(b)})
	reply, ok := f.routes[r.Method+" "+r.URL.Path]
	f.mu.Unlock()
	if !ok {
		reply = fakeReply{http.StatusNotFound, `{"value":{"error":"unknown command","message":"no such route"}}`}
	}
	if reply.status == 0 {
		reply.status = http.StatusOK
	}
	w.WriteHeader(reply.status)
	_, _ = io.WriteString(w, reply.body)
}

func (f *wdaFake) requests() []reqLog {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]reqLog(nil), f.seen...)
}

// sessionClient returns a client with an open session "s1".
func sessionClient(t *testing.T, routes map[string]fakeReply) (*Client, *wdaFake) {
	t.Helper()
	routes["POST /session"] = fakeReply{0, `{"sessionId":"s1","value":{"sessionId":"s1","capabilities":{}}}`}
	f := newWDAFake(t, routes)
	c := New(f.URL, nil)
	if _, err := c.CreateSession(); err != nil {
		t.Fatal(err)
	}
	return c, f
}

func TestCreateAndCloseSession(t *testing.T) {
	c, f := sessionClient(t, map[string]fakeReply{"DELETE /session": {}})
	if c.sessionId != "s1" {
		t.Fatalf("session id = %q", c.sessionId)
	}
	if got := f.requests()[0]; got.Body != `{"capabilities":{}}` {
		t.Errorf("create body = %s", got.Body)
	}
	if err := c.CloseSession(); err != nil || c.sessionId != "" {
		t.Fatalf("CloseSession = %v, session %q", err, c.sessionId)
	}
}

func TestCreateSessionErrors(t *testing.T) {
	for name, reply := range map[string]fakeReply{
		"error response": {http.StatusInternalServerError, `{"value":{"error":"session not created","message":"no app"}}`},
		"invalid json":   {0, `not json`},
	} {
		f := newWDAFake(t, map[string]fakeReply{"POST /session": reply})
		if _, err := New(f.URL, nil).CreateSession(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestCloseSessionKeepsTheErrorAndClearsTheSession(t *testing.T) {
	c, f := sessionClient(t, map[string]fakeReply{})
	f.Close()
	if err := c.CloseSession(); err == nil || c.sessionId != "" {
		t.Fatalf("CloseSession = %v, session %q", err, c.sessionId)
	}
}

func TestStatus(t *testing.T) {
	f := newWDAFake(t, map[string]fakeReply{"GET /status": {0, `{"value":{"ready":true,"state":"success"}}`}})
	st, err := New(f.URL, nil).Status()
	if err != nil || !st.Value.Ready {
		t.Fatalf("Status = %+v, %v", st, err)
	}
	if _, err := New(f.URL+"/nowhere", nil).Status(); err == nil {
		t.Fatal("expected an error for an unknown route")
	}
}

func TestFindElementAndClick(t *testing.T) {
	c, f := sessionClient(t, map[string]fakeReply{
		"POST /session/s1/element":          {0, `{"value":{"ELEMENT":"e7"}}`},
		"POST /session/s1/element/e7/click": {0, `{"value":null}`},
	})
	el, err := c.FindElement("name", "Login")
	if err != nil || el.Value.Element != "e7" {
		t.Fatalf("FindElement = %+v, %v", el, err)
	}
	if err := c.TapElement(el.Value.Element); err != nil {
		t.Fatal(err)
	}
	reqs := f.requests()
	if reqs[1].Body != `{"using":"name","value":"Login"}` {
		t.Errorf("find body = %s", reqs[1].Body)
	}
}

func TestFindElementErrors(t *testing.T) {
	for name, reply := range map[string]fakeReply{
		"not found":      {http.StatusNotFound, `{"value":{"error":"no such element","message":"gone"}}`},
		"empty element":  {0, `{"value":{"ELEMENT":""}}`},
		"missing value":  {0, `{"value":null}`},
		"invalid answer": {0, `<html>`},
	} {
		c, _ := sessionClient(t, map[string]fakeReply{"POST /session/s1/element": reply})
		if _, err := c.FindElement("name", "x"); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestTapElementError(t *testing.T) {
	c, _ := sessionClient(t, map[string]fakeReply{})
	err := c.TapElement("e1")
	if err == nil || !strings.Contains(err.Error(), "unable to tap element: no such route") {
		t.Fatalf("err = %v", err)
	}
}

func TestSendTextSendsOneKeyPerCharacter(t *testing.T) {
	c, f := sessionClient(t, map[string]fakeReply{"POST /session/s1/wda/keys": {0, `{"value":null}`}})
	if err := c.SendText("hé"); err != nil {
		t.Fatal(err)
	}
	if got := f.requests()[1].Body; got != `{"value":["h","é"]}` {
		t.Errorf("keys body = %s", got)
	}
	if err := c.PressButton("home"); err == nil {
		t.Error("expected an error for the unknown pressButton route")
	}
}

func TestPressButton(t *testing.T) {
	c, f := sessionClient(t, map[string]fakeReply{"POST /session/s1/wda/pressButton": {}})
	if err := c.PressButton("volumeUp"); err != nil {
		t.Fatal(err)
	}
	if got := f.requests()[1].Body; got != `{"name":"volumeUp"}` {
		t.Errorf("body = %s", got)
	}
}

func TestTapFallsBackToTheLegacyRoute(t *testing.T) {
	c, f := sessionClient(t, map[string]fakeReply{"POST /session/s1/wda/tap/0": {}})
	if err := c.Tap(10, 20); err != nil {
		t.Fatal(err)
	}
	reqs := f.requests()
	if len(reqs) != 3 || reqs[1].Path != "/session/s1/wda/tap" || reqs[2].Path != "/session/s1/wda/tap/0" {
		t.Fatalf("requests = %+v", reqs)
	}
	var body map[string]float64
	_ = json.Unmarshal([]byte(reqs[2].Body), &body)
	if body["x"] != 10 || body["y"] != 20 {
		t.Errorf("body = %v", body)
	}
}

func TestTapDoesNotRetryOtherErrors(t *testing.T) {
	c, f := sessionClient(t, map[string]fakeReply{"POST /session/s1/wda/tap": {http.StatusInternalServerError, `{"value":{"error":"x","message":"boom"}}`}})
	if err := c.Tap(1, 2); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v", err)
	}
	if n := len(f.requests()); n != 2 {
		t.Errorf("%d requests", n)
	}
}

func TestDrag(t *testing.T) {
	c, f := sessionClient(t, map[string]fakeReply{"POST /session/s1/wda/dragfromtoforduration": {}})
	if err := c.Drag(Point{1, 2}, Point{3, 4}, 0.5); err != nil {
		t.Fatal(err)
	}
	if got := f.requests()[1].Body; got != `{"duration":0.5,"fromX":1,"fromY":2,"toX":3,"toY":4}` {
		t.Errorf("body = %s", got)
	}
}

func TestLaunchTerminateAndActiveApp(t *testing.T) {
	c, f := sessionClient(t, map[string]fakeReply{
		"POST /session/s1/wda/apps/launch":    {0, `{"value":null}`},
		"POST /session/s1/wda/apps/terminate": {0, `{"value":true}`},
		"GET /session/s1/wda/activeAppInfo":   {0, `{"value":{"bundleId":"com.app","pid":9}}`},
	})
	if err := c.Launch("com.app", true, []string{"K", "V"}); err != nil {
		t.Fatal(err)
	}
	if got := f.requests()[1].Body; got != `{"arguments":["K","V"],"bundleId":"com.app","shouldWaitForQuiescence":1}` {
		t.Errorf("launch body = %s", got)
	}
	if err := c.Launch("com.app", false, nil); err != nil {
		t.Fatal(err)
	}
	info, err := c.ActiveAppInfo()
	if err != nil || info.Value.BundleId != "com.app" {
		t.Fatalf("ActiveAppInfo = %+v, %v", info, err)
	}
	if err := c.Terminate("com.app"); err != nil {
		t.Fatal(err)
	}
}

func TestAppCallsFailWithoutRoutes(t *testing.T) {
	c, _ := sessionClient(t, map[string]fakeReply{})
	if c.Launch("a", true, nil) == nil || c.Terminate("a") == nil {
		t.Error("launch and terminate should fail")
	}
	if _, err := c.ActiveAppInfo(); err == nil {
		t.Error("ActiveAppInfo should fail")
	}
}

func TestAlerts(t *testing.T) {
	c, _ := sessionClient(t, map[string]fakeReply{
		"GET /session/s1/alert/text":     {0, `{"value":"Allow notifications?"}`},
		"POST /session/s1/alert/accept":  {0, `{"value":null}`},
		"POST /session/s1/alert/dismiss": {0, `{"value":"gone"}`},
	})
	if text, err := c.GetAlertText(); err != nil || text != "Allow notifications?" {
		t.Fatalf("GetAlertText = %q, %v", text, err)
	}
	if _, err := c.AcceptAlert(true); err != nil {
		t.Fatal(err)
	}
	if got, err := c.AcceptAlert(false); err != nil || got != "gone" {
		t.Fatalf("dismiss = %q, %v", got, err)
	}
}

func TestAlertErrors(t *testing.T) {
	c, _ := sessionClient(t, map[string]fakeReply{"GET /session/s1/alert/text": {0, `{"value":""}`}})
	if _, err := c.GetAlertText(); err == nil || !strings.Contains(err.Error(), "no alert text") {
		t.Fatalf("empty alert: %v", err)
	}
	c, _ = sessionClient(t, map[string]fakeReply{"GET /session/s1/alert/text": {http.StatusNotFound, `{"value":{"error":"no such alert","message":"no alert open"}}`}})
	if _, err := c.GetAlertText(); err == nil || !strings.Contains(err.Error(), "no alert open") {
		t.Fatalf("no alert: %v", err)
	}
	if _, err := c.AcceptAlert(true); err == nil {
		t.Fatal("accept without a route should fail")
	}
}

func TestScreenshot(t *testing.T) {
	png := []byte("\x89PNG fake")
	c, _ := sessionClient(t, map[string]fakeReply{"GET /screenshot": {0, `{"value":"` + base64.StdEncoding.EncodeToString(png) + `"}`}})
	got, err := c.Screenshot()
	if err != nil || string(got) != string(png) {
		t.Fatalf("Screenshot = %q, %v", got, err)
	}
}

func TestScreenshotErrors(t *testing.T) {
	c, _ := sessionClient(t, map[string]fakeReply{"GET /screenshot": {0, `{"value":"!!not base64"}`}})
	if _, err := c.Screenshot(); err == nil || !strings.Contains(err.Error(), "invalid image") {
		t.Fatalf("bad base64: %v", err)
	}
	c, _ = sessionClient(t, map[string]fakeReply{})
	if _, err := c.Screenshot(); err == nil {
		t.Fatal("missing route should fail")
	}
}

func TestSettings(t *testing.T) {
	c, f := sessionClient(t, map[string]fakeReply{
		"GET /session/s1/appium/settings":  {0, `{"value":{"mjpegServerFramerate":30}}`},
		"POST /session/s1/appium/settings": {0, `{"value":{}}`},
	})
	got, err := c.GetSettings()
	if err != nil || got.Value["mjpegServerFramerate"] != float64(30) {
		t.Fatalf("GetSettings = %+v, %v", got, err)
	}
	if err := c.SetSettings(Settings{Value: map[string]interface{}{"reduceMotion": true}}); err != nil {
		t.Fatal(err)
	}
	if body := f.requests()[2].Body; !strings.Contains(body, `"reduceMotion":true`) {
		t.Errorf("set body = %s", body)
	}
}

func TestSettingsErrors(t *testing.T) {
	c, _ := sessionClient(t, map[string]fakeReply{})
	if _, err := c.GetSettings(); err == nil {
		t.Error("GetSettings should fail")
	}
	if err := c.SetSettings(Settings{}); err == nil {
		t.Error("SetSettings should fail")
	}
}

func TestScreenErrorsAndScale(t *testing.T) {
	c, _ := sessionClient(t, map[string]fakeReply{})
	if _, err := c.Screen(); err == nil {
		t.Fatal("Screen without route should fail")
	}
	if _, err := c.ScreenScale(); err == nil {
		t.Fatal("ScreenScale without route should fail")
	}
	c, _ = sessionClient(t, map[string]fakeReply{"GET /wda/screen": {0, `{"value":{"scale":0}}`}})
	if scale, err := c.ScreenScale(); err != nil || scale != 1 {
		t.Fatalf("ScreenScale = %v, %v", scale, err)
	}
}

func TestAliveIsFalseWhenTheServerIsGone(t *testing.T) {
	f := newWDAFake(t, map[string]fakeReply{"GET /status": {}})
	c := New(f.URL, nil)
	if !c.Alive(time.Second) {
		t.Fatal("Alive = false")
	}
	f.Close()
	if c.Alive(time.Second) {
		t.Fatal("Alive = true after the server closed")
	}
}

func TestClientUsesTheInjectedHTTPClient(t *testing.T) {
	f := newWDAFake(t, map[string]fakeReply{"GET /status": {0, `{"value":{}}`}})
	rt := &countingTransport{}
	if _, err := New(f.URL, &http.Client{Transport: rt}).Status(); err != nil {
		t.Fatal(err)
	}
	if rt.calls != 1 {
		t.Fatalf("injected transport used %d times", rt.calls)
	}
}

type countingTransport struct{ calls int }

func (c *countingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.calls++
	return http.DefaultTransport.RoundTrip(r)
}

func TestDoReportsRequestBuildingErrors(t *testing.T) {
	c := New("http://x", nil)
	if err := c.call(call{"bad", "BAD METHOD", "http://x", nil}, nil); err == nil {
		t.Error("an invalid method should fail")
	}
	if err := c.call(call{"bad", http.MethodPost, "http://x", func() {}}, nil); err == nil {
		t.Error("an unmarshalable body should fail")
	}
}

func TestDoReportsBodyReadErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "10")
		_, _ = io.WriteString(w, "short")
	}))
	defer srv.Close()
	if err := New(srv.URL, nil).call(call{"read", http.MethodGet, srv.URL, nil}, nil); err == nil {
		t.Error("a truncated body should fail")
	}
}

func TestErrorMessageFallsBackToTheStatus(t *testing.T) {
	for _, tc := range []struct {
		body, want string
	}{
		{`{"value":{"message":"m","error":"e"}}`, "m"},
		{`{"value":{"error":"e"}}`, "e"},
		{`<html>`, "Bad Gateway"},
	} {
		if got := errorMessage([]byte(tc.body), http.StatusBadGateway); got != tc.want {
			t.Errorf("errorMessage(%s) = %q, want %q", tc.body, got, tc.want)
		}
	}
}
