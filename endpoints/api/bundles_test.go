package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
)

// send is the JSON request helper of the settings tests; the bundle and app
// tests moved to apitest (bundles_flow_test.go, apps_flow_test.go).
func send(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}
