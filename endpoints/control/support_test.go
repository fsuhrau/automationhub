package control

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/hub/action"
	hubcontrol "github.com/fsuhrau/automationhub/hub/control"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
)

// plainDevice hides every optional capability of the device it wraps.
type plainDevice struct{ device.Device }

const sceneXML = `<Scene><Button ID="7" Name="ok"/></Scene>`

// fakeApp plays the Unity plugin: it answers every action of a device with
// success, and the scene graph request with sceneXML.
type fakeApp struct{ *testkit.FakeDeviceManager }

func (m *fakeApp) SendAction(d device.Device, a action.Interface) error {
	if err := m.FakeDeviceManager.SendAction(d, a); err != nil {
		return err
	}
	resp := &appproto.Response{ActionType: a.GetActionType(), Success: true}
	if a.GetActionType() == appproto.ActionType_GetSceneGraph {
		resp.Payload.Screenshot = &appproto.Screenshot{ContentType: appproto.ContentType_Xml, Sceengraph: []byte(sceneXML)}
	}
	for _, h := range d.(*testkit.FakeDevice).ActionHandlers() {
		h.OnActionResponse(d, resp)
	}
	return nil
}

// env is a control handler on a router, serving the devices of its map by
// the :device_id of the path.
type env struct {
	router  *gin.Engine
	devices map[string]device.Device
	// full is device "1", it has every capability; "2" has none
	full *testkit.FakeDevice
	// resolveApp is what the handler asks for an app bundle
	resolveApp AppResolver
}

func testPNG(t testing.TB) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	img.Set(0, 0, color.White)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func setup(t *testing.T) *env {
	t.Helper()
	gin.SetMode(gin.TestMode)
	full := testkit.NewFakeDevice("android-1")
	full.PNG, full.Width, full.Height = testPNG(t), 4, 2
	plain := testkit.NewFakeDevice("plain-2")
	plain.PNG, plain.Width, plain.Height = full.PNG, 4, 2
	e := &env{
		devices: map[string]device.Device{"1": full, "2": plainDevice{plain}},
		full:    full,
		resolveApp: func(*gin.Context, uint) (*app.Parameter, error) {
			return &app.Parameter{Identifier: "com.example"}, nil
		},
	}
	resolve := func(c *gin.Context) (device.Device, bool) {
		d, ok := e.devices[c.Param("device_id")]
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
		}
		return d, ok
	}
	owner := func(c *gin.Context) string { return c.GetHeader("X-Test-User") }
	appResolver := func(c *gin.Context, id uint) (*app.Parameter, error) { return e.resolveApp(c, id) }
	e.router = gin.New()
	NewHandler(hubcontrol.New(&fakeApp{FakeDeviceManager: &testkit.FakeDeviceManager{}}), resolve, appResolver, owner, "localhost").Register(e.router.Group("/api/p"))
	return e
}

// as sends the requests of h as the user, with the lease token when given.
func as(h http.Handler, user, token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Test-User", user)
		if token != "" {
			r.Header.Set(LeaseHeader, token)
		}
		h.ServeHTTP(w, r)
	})
}

// lock takes the lease of the device as the user and returns its token.
func (e *env) lock(t *testing.T, deviceID, user string) string {
	t.Helper()
	rec := testkit.DoJSON(t, as(e.router, user, ""), http.MethodPost, "/api/p/device/"+deviceID+"/lock", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("lock: %d %s", rec.Code, rec.Body)
	}
	return testkit.Decode[hubcontrol.Lease](t, rec).Token
}

// call sends the request as alice with the lease token and returns the response.
func (e *env) call(t *testing.T, token, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	return testkit.DoJSON(t, as(e.router, "alice", token), method, "/api/p/device/"+path, body)
}
