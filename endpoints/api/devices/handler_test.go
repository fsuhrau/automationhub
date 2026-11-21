package devices_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/endpoints/api/devices"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// fakeService answers every call with its fields.
type fakeService struct {
	err     error
	dev     device.Device
	params  *app.Parameter
	request svc.TestRequest
}

func (f *fakeService) List(context.Context, *models.Project, *models.PlatformType) ([]models.Device, error) {
	return nil, f.err
}

func (f *fakeService) Get(context.Context, *models.Project, string) (*models.Device, error) {
	return nil, f.err
}

func (f *fakeService) Delete(context.Context, *models.Project, string) error { return f.err }

func (f *fakeService) Unlock(context.Context, *models.Project, string) (*models.Device, error) {
	return nil, f.err
}

func (f *fakeService) Update(context.Context, *models.Project, svc.Update) (*models.Device, error) {
	return nil, f.err
}

func (f *fakeService) RunTests(_ context.Context, _ *models.Project, req svc.TestRequest) error {
	f.request = req
	return f.err
}

func (f *fakeService) Protocols(context.Context, *models.Project, svc.ProtocolQuery) ([]models.TestProtocol, error) {
	return nil, f.err
}

func (f *fakeService) ControlDevice(context.Context, *models.Project, string) (device.Device, error) {
	return f.dev, f.err
}

func (f *fakeService) ControlApp(context.Context, *models.Project, uint) (*app.Parameter, error) {
	return f.params, f.err
}

// serve routes a request through the handlers with the project set (or not).
func serve(h *devices.Handler, withProject bool, method, path, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if withProject {
			c.Set(httpx.ProjectKey, &models.Project{})
		}
	})
	g := r.Group("")
	h.Register(g)
	h.RegisterProtocols(g)
	r.GET("/resolve/:device_id", func(c *gin.Context) {
		if _, ok := h.ResolveControlDevice(c); ok {
			c.Status(http.StatusNoContent)
		}
	})
	r.GET("/app", func(c *gin.Context) {
		if _, err := h.ResolveControlApp(c, 1); err != nil {
			httpx.Fail(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestRunTestsAnswersWhenTheExecutionFinished(t *testing.T) {
	f := &fakeService{}

	rec := serve(devices.New(f), true, "POST", "/device/1/tests", `{"testName":"Suite Method","env":"a=b"}`)

	if rec.Code != http.StatusOK || rec.Body.String() != `{"success":true,"message":"Execution Finished"}` {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
	if f.request != (svc.TestRequest{DeviceID: "1", TestName: "Suite Method", Env: "a=b"}) {
		t.Fatalf("service got %+v", f.request)
	}
}

func TestEveryRouteAnswersWithTheStatusOfTheServiceError(t *testing.T) {
	routes := []struct{ method, path string }{
		{"GET", "/devices"}, {"GET", "/device/1"}, {"DELETE", "/device/1"}, {"POST", "/device/1/unlock"},
		{"PUT", "/device/1"}, {"POST", "/device/1/tests"}, {"GET", "/device/1/protocols"},
		{"GET", "/resolve/1"}, {"GET", "/app"},
	}
	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			rec := serve(devices.New(&fakeService{err: apperr.Busy("taken")}), true, r.method, r.path, `{"testName":"A b"}`)

			if rec.Code != http.StatusConflict {
				t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestResolversNeedAProject(t *testing.T) {
	for _, path := range []string{"/resolve/1", "/app"} {
		rec := serve(devices.New(&fakeService{}), false, "GET", path, "")

		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "project not found") {
			t.Fatalf("%s: %d %s", path, rec.Code, rec.Body)
		}
	}
}

func TestResolversReturnWhatTheServiceFound(t *testing.T) {
	f := &fakeService{params: &app.Parameter{}}
	for _, path := range []string{"/resolve/1", "/app"} {
		if rec := serve(devices.New(f), true, "GET", path, ""); rec.Code != http.StatusNoContent {
			t.Fatalf("%s: %d %s", path, rec.Code, rec.Body)
		}
	}
}

func TestPlatformQueryThatIsNoNumberSelectsPlatformZero(t *testing.T) {
	rec := serve(devices.New(&fakeService{}), true, "GET", "/devices?platform=android", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
