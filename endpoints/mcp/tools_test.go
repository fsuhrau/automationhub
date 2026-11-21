package mcp

import (
	"image"
	"net/http"
	"strings"
	"testing"
	"time"

	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestPlatformFilter(t *testing.T) {
	tests := []struct {
		arg, want string
		fails     bool
	}{
		{"", "", false},
		{"Android", "android", false},
		{"unity_editor", "unity_editor", false},
		{"toaster", "", true},
	}
	for _, tc := range tests {
		got, err := platformFilter(tc.arg)
		if got != tc.want || (err != nil) != tc.fails {
			t.Errorf("platformFilter(%q) = %q, %v", tc.arg, got, err)
		}
	}
}

func TestBundleLimit(t *testing.T) {
	for asked, want := range map[int]int{-1: 20, 0: 20, 1: 1, 200: 200, 201: 20} {
		if got := bundleLimit(asked); got != want {
			t.Errorf("bundleLimit(%d) = %d, want %d", asked, got, want)
		}
	}
}

func TestWaitTimeout(t *testing.T) {
	for seconds, want := range map[int]time.Duration{-5: defaultWait, 0: defaultWait, 7: 7 * time.Second, 99999: maxWait} {
		if got := waitTimeout(seconds); got != want {
			t.Errorf("waitTimeout(%d) = %v, want %v", seconds, got, want)
		}
	}
}

func TestTruncate(t *testing.T) {
	long := strings.Repeat("x", defaultHierarchyChars+5)
	tests := []struct {
		name     string
		content  string
		maxChars int
		want     string
	}{
		{"short", "abc", 0, "abc"},
		{"exact", "abc", 3, "abc"},
		{"cut", "abcdef", 3, "abc\n... truncated (6 characters total); use xpath to narrow it down"},
		{"default limit", long, 0, long[:defaultHierarchyChars] + "\n... truncated (60005 characters total); use xpath to narrow it down"},
	}
	for _, tc := range tests {
		if got := truncate(tc.content, tc.maxChars); got != tc.want {
			t.Errorf("%s: truncate = %q", tc.name, got)
		}
	}
}

func TestSelectElements(t *testing.T) {
	scene := []byte(`<Root><Button Name="Play"/><Button Name="Quit"/></Root>`)

	got, err := selectElements(scene, "//Button[@Name='Play']")
	if err != nil || !strings.HasPrefix(got, "1 matching elements\n") || !strings.Contains(got, `Name="Play"`) || strings.Contains(got, "Quit") {
		t.Fatalf("matching: %q, %v", got, err)
	}
	if _, err := selectElements(scene, "//["); err == nil || !strings.Contains(err.Error(), "invalid xpath") {
		t.Errorf("invalid xpath: %v", err)
	}
	if _, err := selectElements([]byte("<a"), "//a"); err == nil {
		t.Error("broken xml accepted")
	}
}

func TestScaleToJPEGFailsForAnImageJPEGCannotEncode(t *testing.T) {
	_, _, err := scaleToJPEG(image.NewGray(image.Rect(0, 0, 70000, 1)), 100000)

	if err == nil {
		t.Fatal("an image wider than 65535 pixels was encoded")
	}
}

func TestToTestInfoNamesTheLastRunAndTheDevices(t *testing.T) {
	entry := testsvc.ListEntry{Test: models.Test{Model: models.Model{ID: 3}, Name: "Smoke",
		TestConfig: models.TestConfig{Type: models.TestTypeScenario, Devices: []models.TestConfigDevice{{DeviceID: 5}, {DeviceID: 6}}}}}
	if got := toTestInfo(&entry); got.LastRunID != 0 || got.Type != "scenario" || len(got.DeviceIDs) != 2 {
		t.Errorf("without runs %+v", got)
	}
	entry.RecentRuns = []testsvc.RunSummary{{ID: 9}, {ID: 8}}
	if got := toTestInfo(&entry); got.LastRunID != 9 {
		t.Errorf("last run %d, want the newest", got.LastRunID)
	}
}

func TestRequestHeaderIsEmptyWithoutARequest(t *testing.T) {
	if h := requestHeader(nil); h == nil || len(h) != 0 {
		t.Errorf("header %v", h)
	}
}

func TestUploadAddressUsesTLSAndTheProxyHeaders(t *testing.T) {
	project := &models.Project{Identifier: "p1"}
	if got := (uploadAddress{header: http.Header{}, host: "master:8080", tls: true}).url(project, "id"); got != "https://master:8080/api/p1/mcp/uploads/id" {
		t.Errorf("tls: %q", got)
	}
	proxied := http.Header{"X-Forwarded-Proto": {"https"}, "X-Forwarded-Host": {"hub"}, "X-Forwarded-Prefix": {"/p/mcp/"}}
	if got := (uploadAddress{header: proxied, host: "master"}).url(project, "id"); got != "https://hub/p/mcp/uploads/id" {
		t.Errorf("proxy: %q", got)
	}
}

func TestUploadAddressReadsTheMastersOwnAddressFromTheHeaders(t *testing.T) {
	header := http.Header{originalHostHeader: {"master:1"}, originalTLSHeader: {"1"}}

	got := newUploadAddress(header)

	if got.host != "master:1" || !got.tls {
		t.Errorf("address %+v", got)
	}
}

func TestClaimStatus(t *testing.T) {
	for err, want := range map[error]int{
		errUploadNotFound: http.StatusNotFound, errUploadCaller: http.StatusForbidden, errUploadBusy: http.StatusConflict,
	} {
		if got := claimStatus(err); got != want {
			t.Errorf("claimStatus(%v) = %d, want %d", err, got, want)
		}
	}
}
