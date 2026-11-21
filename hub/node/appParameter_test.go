package node

import (
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
)

func TestAppParameter_RoundTrip(t *testing.T) {
	minimal := &app.Parameter{Platform: "web", Name: "n", Identifier: "i", Version: "v"}
	full := wireParameter()
	for name, p := range map[string]*app.Parameter{"minimal": minimal, "full": full} {
		t.Run(name, func(t *testing.T) {
			req := getAppParameterRequest("dev", p)
			if req.DeviceID != "dev" || (p.App == nil) != (req.App == nil) || (p.Web == nil) != (req.Web == nil) {
				t.Fatalf("request = %+v", req)
			}
			got := getAppParameter(req, &AppBundleMetaData{FilePath: "/stored/app.apk"})
			want := *p
			if want.App != nil {
				app := *want.App
				app.AppBinaryID, app.AppPath = 0, "/stored/app.apk" // not part of the wire format
				want.App = &app
			}
			if !reflect.DeepEqual(*got, want) {
				t.Fatalf("got %+v want %+v", got, want)
			}
		})
	}
}

func TestGetAppParameter_WithoutBundleHasNoPath(t *testing.T) {
	got := getAppParameter(getAppParameterRequest("d", wireParameter()), nil)
	if got.App.AppPath != "" {
		t.Errorf("path = %q", got.App.AppPath)
	}
}

func TestStartAppRequest_AndDeviceConfig(t *testing.T) {
	cfg := &device.DeviceConfig{Connection: 1, IP: "1.2.3.4", Port: 99, DeviceParameter: map[string]string{"a": "b"}}
	req := startAppRequest(getAppParameterRequest("d", wireParameter()), cfg)
	if len(req.CustomParameter) != 1 || req.ConnectionParams.IP != "1.2.3.4" {
		t.Fatalf("req = %+v", req)
	}
	if got := getDeviceConfig(req); !reflect.DeepEqual(got, cfg) {
		t.Fatalf("config = %+v want %+v", got, cfg)
	}
	req.ConnectionParams = nil // old masters may omit it
	if got := getDeviceConfig(req); got.IP != "" || got.DeviceParameter["a"] != "b" {
		t.Fatalf("config = %+v", got)
	}
}
