package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/api"
	"github.com/fsuhrau/automationhub/storage/models"
)

type seen struct {
	method, path, contentType, token string
	body                             []byte
}

// hubServer answers every request with status and body and records it.
func hubServer(t *testing.T, status int, body string) (*Client, *seen) {
	t.Helper()
	got := &seen{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method, got.path = r.Method, r.URL.Path
		got.contentType, got.token = r.Header.Get("Content-Type"), r.Header.Get("X-Auth-Token")
		got.body, _ = io.ReadAll(r.Body)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return NewClient(Config{URL: srv.URL, Token: "secret", ProjectID: "demo", AppID: 7}), got
}

func TestNewClientBuildsTheAppURL(t *testing.T) {
	c := NewClient(Config{URL: "http://hub", Token: "tok", ProjectID: "demo", AppID: 7})

	if c.BaseURL != "http://hub/api/demo/app/7" || c.HTTPClient.Timeout == 0 {
		t.Errorf("client = %+v", c)
	}
}

func TestGetTests(t *testing.T) {
	c, got := hubServer(t, http.StatusOK, `[{"ID":1,"name":"smoke"},{"ID":2,"name":"login"}]`)

	tests, err := c.GetTests(context.Background())

	if err != nil || len(tests) != 2 || tests[1].Name != "login" {
		t.Fatalf("tests = %+v, err = %v", tests, err)
	}
	if got.method != "GET" || got.path != "/api/demo/app/7/tests" || got.token != "secret" || got.contentType != jsonContentType {
		t.Errorf("request = %+v", got)
	}
}

func TestExecuteTestSendsTheRequestAndReturnsTheRun(t *testing.T) {
	c, got := hubServer(t, http.StatusOK, `{"ID":42}`)

	run, err := c.ExecuteTest(context.Background(), 3, api.RunTestRequest{AppBinaryID: 50, Params: "a=1\nb=2"})

	if err != nil || run.ID != 42 {
		t.Fatalf("run = %+v, err = %v", run, err)
	}
	if got.method != "POST" || got.path != "/api/demo/app/7/test/3/run" {
		t.Errorf("request = %+v", got)
	}
	var sent map[string]any
	if err := json.Unmarshal(got.body, &sent); err != nil || sent["params"] != "a=1\nb=2" || sent["appBinaryId"] != float64(50) {
		t.Errorf("body = %s (%v)", got.body, err)
	}
}

func TestUpdateAppPostsTheBinary(t *testing.T) {
	c, got := hubServer(t, http.StatusOK, `{}`)
	binary := &models.AppBinary{Tags: "a,b"}
	binary.ID = 9

	if err := c.UpdateApp(context.Background(), binary); err != nil {
		t.Fatal(err)
	}

	if got.path != "/api/demo/app/7/bundle/9" || !strings.Contains(string(got.body), `"tags":"a,b"`) {
		t.Errorf("request = %+v %s", got, got.body)
	}
}

func TestUploadBundleSendsTheFileAsMultipart(t *testing.T) {
	c, got := hubServer(t, http.StatusOK, `{"ID":11}`)
	path := filepath.Join(t.TempDir(), "app.apk")
	if err := os.WriteFile(path, []byte("binary-content"), 0o600); err != nil {
		t.Fatal(err)
	}

	app, err := c.UploadBundle(context.Background(), path)

	if err != nil || app.ID != 11 {
		t.Fatalf("app = %+v, err = %v", app, err)
	}
	if got.path != "/api/demo/app/7/upload" || got.token != "secret" || !strings.HasPrefix(got.contentType, "multipart/form-data; boundary=") {
		t.Errorf("request = %+v", got)
	}
	if !strings.Contains(string(got.body), `name="test_target"; filename="`+path+`"`) || !strings.Contains(string(got.body), "binary-content") {
		t.Errorf("body = %s", got.body)
	}
}

func TestUploadBundleFailures(t *testing.T) {
	c, _ := hubServer(t, http.StatusOK, `{}`)
	tests := map[string]struct{ path, want string }{
		"missing file":     {filepath.Join(t.TempDir(), "none"), "open bundle"},
		"path is a folder": {t.TempDir(), "read bundle"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := c.UploadBundle(context.Background(), tc.path); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestUploadBundleReturnsTheServerError(t *testing.T) {
	c, _ := hubServer(t, http.StatusBadRequest, `{"message":"bad bundle"}`)
	path := filepath.Join(t.TempDir(), "app")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := c.UploadBundle(context.Background(), path); err == nil || err.Error() != "bad bundle" {
		t.Errorf("err = %v", err)
	}
}

func TestRequestFailures(t *testing.T) {
	tests := map[string]struct {
		status int
		body   string
		want   string
	}{
		"error message":      {http.StatusForbidden, `{"message":"no access"}`, "no access"},
		"error without body": {http.StatusInternalServerError, `oops`, "unknown error, status code: 500"},
		"answer is not json": {http.StatusOK, `<html>`, "decode response"},
		"empty answer":       {http.StatusOK, ``, "decode response"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c, _ := hubServer(t, tc.status, tc.body)
			if _, err := c.GetTests(context.Background()); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestExecuteTestReturnsTheError(t *testing.T) {
	c, _ := hubServer(t, http.StatusNotFound, `{"message":"no such test"}`)

	if _, err := c.ExecuteTest(context.Background(), 1, api.RunTestRequest{}); err == nil || err.Error() != "no such test" {
		t.Errorf("err = %v", err)
	}
}

func TestSendJSONFailures(t *testing.T) {
	c, _ := hubServer(t, http.StatusOK, `{}`)
	unreachable := NewClient(Config{URL: "http://127.0.0.1:1", ProjectID: "p", AppID: 1})
	badURL := &Client{BaseURL: ":bad", HTTPClient: http.DefaultClient}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	tests := map[string]struct {
		client *Client
		ctx    context.Context
		in     any
		want   string
	}{
		"unencodable body": {c, context.Background(), make(chan int), "encode request"},
		"bad url":          {badURL, context.Background(), nil, "create request"},
		"unreachable":      {unreachable, context.Background(), nil, "send request"},
		"cancelled":        {c, cancelled, nil, "send request"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.client.sendJSON(tc.ctx, jsonCall{http.MethodPost, "/x", tc.in, &struct{}{}})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want %q", err, tc.want)
			}
		})
	}
}
