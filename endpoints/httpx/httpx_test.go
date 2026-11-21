package httpx_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

type fakeLogger struct{ lines []string }

func (l *fakeLogger) Errorf(format string, args ...any) {
	l.lines = append(l.lines, fmt.Sprintf(format, args...))
}

// serve runs handler behind a router with the middlewares and returns the
// recorded response of a request to path.
func serve(method, route, path string, handler gin.HandlerFunc, mw ...gin.HandlerFunc) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw...)
	r.Handle(method, route, handler)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestStatusMapsEveryKind(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, http.StatusOK},
		{"not found", apperr.NotFound("x"), http.StatusNotFound},
		{"invalid", apperr.Invalid("x"), http.StatusBadRequest},
		{"conflict", apperr.Conflict("x"), http.StatusConflict},
		{"busy", apperr.Busy("x"), http.StatusConflict},
		{"precondition required", apperr.PreconditionRequired("x"), http.StatusPreconditionRequired},
		{"unsupported", apperr.Unsupported("x"), http.StatusNotImplemented},
		{"timeout", apperr.Timeout("x"), http.StatusGatewayTimeout},
		{"forbidden", apperr.Forbidden("x"), http.StatusForbidden},
		{"internal", apperr.Internal("x"), http.StatusInternalServerError},
		{"unclassified", errors.New("x"), http.StatusInternalServerError},
		{"wrapped", fmt.Errorf("ctx: %w", apperr.NotFound("x")), http.StatusNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := httpx.Status(tc.err); got != tc.want {
				t.Fatalf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func failWith(err error) gin.HandlerFunc {
	return func(c *gin.Context) { httpx.Fail(c, err) }
}

func TestFailWritesStatusAndMessage(t *testing.T) {
	rec := serve("GET", "/x", "/x", failWith(apperr.NotFound("test run %d", 4)))

	var body httpx.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNotFound || body.Message != "test run 4" {
		t.Fatalf("got %d %+v", rec.Code, body)
	}
	if rec.Body.String() != `{"message":"test run 4"}` {
		t.Fatalf("body = %s", rec.Body)
	}
}

func TestFailLogsServerErrorsOnly(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantLines []string
	}{
		{"server error", apperr.Internal("disk full"), []string{"GET /x: disk full"}},
		{"timeout", apperr.Timeout("slow"), []string{"GET /x: slow"}},
		{"client error", apperr.Invalid("bad"), nil},
		{"conflict", apperr.Busy("device"), nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			log := &fakeLogger{}
			serve("GET", "/x", "/x", failWith(tc.err), httpx.WithLogger(log))

			if fmt.Sprint(log.lines) != fmt.Sprint(tc.wantLines) {
				t.Fatalf("logged %q, want %q", log.lines, tc.wantLines)
			}
		})
	}
}

func TestFailWithoutLoggerStillAnswers(t *testing.T) {
	rec := serve("GET", "/x", "/x", failWith(apperr.Internal("boom")))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestFailWithNilErrorIsInternal(t *testing.T) {
	rec := serve("GET", "/x", "/x", failWith(nil))

	if rec.Code != http.StatusInternalServerError || rec.Body.String() != `{"message":"unknown error"}` {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestFailAbortsTheChain(t *testing.T) {
	reached := false
	serve("GET", "/x", "/x", func(*gin.Context) { reached = true }, failWith(apperr.Invalid("x")))

	if reached {
		t.Fatal("handler ran after Fail")
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    uint
		wantErr string
	}{
		{"valid", "/runs/42", 42, ""},
		{"zero", "/runs/0", 0, ""},
		{"not a number", "/runs/abc", 0, `invalid run_id "abc"`},
		{"negative", "/runs/-1", 0, `invalid run_id "-1"`},
		{"too large", "/runs/99999999999999999999999", 0, `invalid run_id "99999999999999999999999"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got uint
			var err error
			serve("GET", "/runs/:run_id", tc.path, func(c *gin.Context) { got, err = httpx.ParseID(c, "run_id") })

			if got != tc.want {
				t.Fatalf("id = %d, want %d", got, tc.want)
			}
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if err == nil || err.Error() != tc.wantErr || apperr.KindOf(err) != apperr.KindInvalid {
				t.Fatalf("err = %v (%v), want invalid %q", err, apperr.KindOf(err), tc.wantErr)
			}
		})
	}
}

func TestParseIDOfAnUnknownParameterIsInvalid(t *testing.T) {
	var err error
	serve("GET", "/x", "/x", func(c *gin.Context) { _, err = httpx.ParseID(c, "run_id") })

	if apperr.KindOf(err) != apperr.KindInvalid {
		t.Fatalf("err = %v", err)
	}
}

func TestLimit(t *testing.T) {
	tests := []struct {
		query string
		want  int
	}{
		{"", 25},
		{"?limit=10", 10},
		{"?limit=100", 100},
		{"?limit=500", 100},
		{"?limit=0", 25},
		{"?limit=-3", 25},
		{"?limit=abc", 25},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			var got int
			serve("GET", "/x", "/x"+tc.query, func(c *gin.Context) { got = httpx.Limit(c, 25, 100) })

			if got != tc.want {
				t.Fatalf("limit = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestProjectAndApp(t *testing.T) {
	project := &models.Project{Identifier: "proj"}
	application := &models.App{Name: "app"}
	set := func(c *gin.Context) {
		c.Set(httpx.ProjectKey, project)
		c.Set(httpx.AppKey, application)
	}
	tests := []struct {
		name        string
		mw          []gin.HandlerFunc
		wantProject *models.Project
		wantApp     *models.App
	}{
		{"present", []gin.HandlerFunc{set}, project, application},
		{"absent", nil, nil, nil},
		{"wrong type", []gin.HandlerFunc{func(c *gin.Context) { c.Set(httpx.ProjectKey, "x"); c.Set(httpx.AppKey, 1) }}, nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotProject *models.Project
			var gotApp *models.App
			serve("GET", "/x", "/x", func(c *gin.Context) {
				gotProject, gotApp = httpx.Project(c), httpx.App(c)
			}, tc.mw...)

			if gotProject != tc.wantProject || gotApp != tc.wantApp {
				t.Fatalf("got %v %v, want %v %v", gotProject, gotApp, tc.wantProject, tc.wantApp)
			}
		})
	}
}
