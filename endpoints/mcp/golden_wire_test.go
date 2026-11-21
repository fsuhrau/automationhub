package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// handler is the MCP handler on the services of the world.
func (w *goldenWorld) handler() *Handler {
	return NewHandler(Deps{
		Runs: w.runs, Protocols: w.runs, Lookup: w.runs, Apps: w.apps, Tests: w.tests, Devices: w.devices,
		Pool: w.pool, Control: w.control, Tracker: w.tracker, DataDir: w.dataDir, StoreDir: w.storeDir,
	})
}

// serve serves the handler as the project of the world; the caller is the
// X-Caller header.
func (w *goldenWorld) serve(t *testing.T, h *Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(rw, r, w.project, r.Header.Get("X-Caller"))
	}))
	t.Cleanup(srv.Close)
	return srv
}
