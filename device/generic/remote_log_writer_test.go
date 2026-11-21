package generic

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRemoteLogWriterSendLogStatus(t *testing.T) {
	for _, tt := range []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"ok", http.StatusOK, false},
		{"non-OK status", http.StatusInternalServerError, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()
			token := "t"
			w := NewRemoteLogWriter(strings.TrimPrefix(srv.URL, "http://"), "n", "d", &token)
			if err := w.sendLog(&request{}); (err != nil) != tt.wantErr {
				t.Errorf("sendLog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
