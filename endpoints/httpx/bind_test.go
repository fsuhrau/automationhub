package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/gin-gonic/gin"
)

func TestBindJSON(t *testing.T) {
	type body struct {
		Name string `json:"name" binding:"required"`
		N    int    `json:"n"`
	}
	tests := []struct {
		name     string
		body     string
		want     body
		wantKind apperr.Kind
	}{
		{"valid", `{"name":"a","n":2}`, body{Name: "a", N: 2}, apperr.KindNone},
		{"malformed", `{`, body{}, apperr.KindInvalid},
		{"wrong type", `{"name":"a","n":"x"}`, body{Name: "a"}, apperr.KindInvalid},
		{"failed validation", `{}`, body{}, apperr.KindInvalid},
		{"empty body", ``, body{}, apperr.KindInvalid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			var got body

			err := httpx.BindJSON(c, &got)

			if kind := apperr.KindOf(err); kind != tc.wantKind {
				t.Fatalf("kind %v (%v), want %v", kind, err, tc.wantKind)
			}
			if got != tc.want {
				t.Fatalf("bound %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestBindJSONKeepsTheBindersMessage(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`))
	var v struct{}

	err := httpx.BindJSON(c, &v)

	if err == nil || !strings.Contains(err.Error(), "unexpected EOF") {
		t.Fatalf("message %v", err)
	}
}
