package remote_test

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/device/appproto"
	"github.com/fsuhrau/automationhub/device/remote"
)

type posted struct {
	path, token string
	body        map[string]any
}

func serve(t *testing.T, status int) (host string, got chan posted) {
	t.Helper()
	got = make(chan posted, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		got <- posted{path: r.URL.Path, token: r.Header.Get("X-Auth-Token"), body: body}
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return strings.TrimPrefix(srv.URL, "http://"), got
}

func TestActionHandlerPostsTheResponse(t *testing.T) {
	token := "secret"
	for _, status := range []int{http.StatusOK, http.StatusInternalServerError} {
		host, got := serve(t, status)
		h := remote.NewActionHandler(host, "node1", "dev1", &token)
		h.OnActionResponse(nil, &appproto.Response{ActionID: "x", Success: true})
		p := <-got
		if p.path != "/node/action" || p.token != "secret" || p.body["node"] != "node1" || p.body["deviceId"] != "dev1" {
			t.Errorf("status %d: posted %+v", status, p)
		}
	}
}

func TestActionHandlerWithoutTokenSendsNoHeader(t *testing.T) {
	host, got := serve(t, http.StatusOK)
	remote.NewActionHandler(host, "n", "d", nil).OnActionResponse(nil, &appproto.Response{})
	if p := <-got; p.token != "" {
		t.Errorf("token = %q", p.token)
	}
}

func TestActionHandlerSurvivesErrors(t *testing.T) {
	// an unreachable master, an invalid URL and a response JSON can't encode (NaN)
	remote.NewActionHandler("127.0.0.1:1", "n", "d", nil).OnActionResponse(nil, &appproto.Response{})
	remote.NewActionHandler("bad host", "n", "d", nil).OnActionResponse(nil, &appproto.Response{})
	remote.NewActionHandler("127.0.0.1:1", "n", "d", nil).OnActionResponse(nil, &appproto.Response{
		Payload: appproto.ResponseData{PerformanceData: &appproto.PerformanceData{CPU: math.NaN()}},
	})
}
