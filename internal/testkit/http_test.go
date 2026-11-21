package testkit

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type echo struct {
	Name string `json:"name"`
}

func echoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("X-Method", r.Method+" "+r.URL.Path+" "+r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})
}

func TestDoJSONAndDecodeRoundTrip(t *testing.T) {
	rec := DoJSON(t, echoHandler(), http.MethodPost, "/p", echo{Name: "a"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("code = %d", rec.Code)
	}
	if got := rec.Header().Get("X-Method"); got != "POST /p application/json" {
		t.Errorf("request = %q", got)
	}
	if got := Decode[echo](t, rec); got.Name != "a" {
		t.Errorf("decoded %+v", got)
	}
}

func TestDoJSONWithoutBodySendsNothing(t *testing.T) {
	rec := DoJSON(t, echoHandler(), http.MethodGet, "/", nil)
	if rec.Body.Len() != 0 {
		t.Errorf("body = %q", rec.Body.String())
	}
}

func TestDoJSONFatalOnUnencodableBody(t *testing.T) {
	expectFatal(t, "", func(tb *fatalTB) { DoJSON(tb, echoHandler(), "POST", "/", make(chan int)) })
}

func TestDecodeFatalOnInvalidJSON(t *testing.T) {
	rec := DoJSON(t, echoHandler(), http.MethodPost, "/", json.RawMessage(`"x"`))
	expectFatal(t, "", func(tb *fatalTB) { Decode[echo](tb, rec) })
}
