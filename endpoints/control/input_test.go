package control

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestUnsupportedInput(t *testing.T) {
	e := setup(t)
	token := e.lock(t, "2", "alice")
	rec := e.call(t, token, http.MethodPost, "2/tap", map[string]int{"x": 1, "y": 1})
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}
	if rec := e.call(t, "", http.MethodGet, "2/logs", nil); rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501 for logs, got %d", rec.Code)
	}
}

func TestInputCallsReachTheDevice(t *testing.T) {
	tests := []struct {
		name, path, call string
		body             any
		wantArgs         []any
	}{
		{"swipe", "1/swipe", "Swipe", map[string]int{"fromX": 1, "fromY": 2, "toX": 3, "toY": 4, "durationMs": 5}, []any{1, 2, 3, 4}},
		{"text", "1/text", "InputText", map[string]string{"text": "hi"}, []any{"hi"}},
		{"key", "1/key", "PressKey", map[string]string{"key": "back"}, []any{"back"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setup(t)
			token := e.lock(t, "1", "alice")
			if rec := e.call(t, token, http.MethodPost, tt.path, tt.body); rec.Code != http.StatusNoContent {
				t.Fatalf("expected 204, got %d %s", rec.Code, rec.Body)
			}
			calls := e.full.CallsTo(tt.call)
			if len(calls) != 1 {
				t.Fatalf("calls %+v", e.full.Calls())
			}
			for i, want := range tt.wantArgs {
				if calls[0].Args[i] != want {
					t.Fatalf("arg %d = %v, want %v", i, calls[0].Args[i], want)
				}
			}
		})
	}
}

func TestInputNeedsLeaseAndBody(t *testing.T) {
	element := map[string]string{"xpath": "//Button"}
	tests := []struct {
		name, path string
		body       any
		want       int
	}{
		{"swipe without lease", "1/swipe", map[string]int{}, http.StatusPreconditionRequired},
		{"text without lease", "1/text", map[string]string{"text": "hi"}, http.StatusPreconditionRequired},
		{"text without body", "1/text", map[string]string{}, http.StatusBadRequest},
		{"element touch without lease", "1/element/touch", element, http.StatusPreconditionRequired},
		{"tap on element without lease", "1/tap", element, http.StatusPreconditionRequired},
		{"element value without lease", "1/element/value", map[string]string{"xpath": "//Input", "value": "x"}, http.StatusPreconditionRequired},
		{"element value bad json", "1/element/value", json.RawMessage(`[]`), http.StatusBadRequest},
		{"element touch bad json", "1/element/touch", json.RawMessage(`[]`), http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setup(t)
			if rec := e.call(t, "", http.MethodPost, tt.path, tt.body); rec.Code != tt.want {
				t.Fatalf("expected %d, got %d %s", tt.want, rec.Code, rec.Body)
			}
		})
	}
}

func TestElementInput(t *testing.T) {
	tests := []struct {
		name, path string
		body       any
	}{
		{"touch", "1/element/touch", map[string]string{"xpath": "//Button"}},
		{"touch by id", "1/element/touch", map[string]string{"elementId": "7"}},
		{"tap on element", "1/tap", map[string]string{"xpath": "//Button"}},
		{"set value", "1/element/value", map[string]string{"elementId": "7", "value": "x"}},
		{"set value with attr", "1/element/value", map[string]string{"elementId": "7", "attr": "name", "value": "x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setup(t)
			e.full.AppConnected = true
			token := e.lock(t, "1", "alice")
			if rec := e.call(t, token, http.MethodPost, tt.path, tt.body); rec.Code != http.StatusNoContent {
				t.Fatalf("expected 204, got %d %s", rec.Code, rec.Body)
			}
		})
	}
}

func TestInputBodyErrors(t *testing.T) {
	for _, path := range []string{"1/tap", "1/swipe", "1/key", "1/text"} {
		t.Run(path, func(t *testing.T) {
			e := setup(t)
			if rec := e.call(t, "", http.MethodPost, path, json.RawMessage(`[]`)); rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d %s", rec.Code, rec.Body)
			}
		})
	}
}
