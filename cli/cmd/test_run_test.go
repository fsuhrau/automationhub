package cmd

import (
	"bytes"
	"context"
	sse "github.com/r3labs/sse/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeBinary(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "app.apk")
	if err := os.WriteFile(path, []byte("apk"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunAsyncStartsTheTestWithoutWaiting(t *testing.T) {
	hub := newFakeHub(t)

	out, err := execute(t, "test", "run", hub.URL, "demo", "7", "smoke", "--binaryID", "50", "--params", "a=1;b=2", "--async")

	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"execute test 3 with binaryId: 50", "Test Running...", "Check your restults at: " + hub.URL + "/project/demo/app/7/test/3/run/5"} {
		if !strings.Contains(out, want) {
			t.Errorf("out lacks %q:\n%s", want, out)
		}
	}
	if got := hub.seen(); len(got) != 2 || !strings.HasPrefix(got[1], "POST "+appPrefix+"/test/3/run ") || !strings.Contains(got[1], `"appBinaryId":50`) || !strings.Contains(got[1], `"params":"a=1\nb=2"`) {
		t.Errorf("requests = %q", got)
	}
}

func TestRunAsyncFlagIsAcceptedOnTheTestCommand(t *testing.T) {
	hub := newFakeHub(t)

	out, err := execute(t, "test", "--async", "run", hub.URL, "demo", "7", "smoke")

	if err != nil || !strings.Contains(out, "Test Running...") {
		t.Errorf("out = %q, err = %v", out, err)
	}
}

func TestRunUploadsTheBinaryAndTagsIt(t *testing.T) {
	hub := newFakeHub(t)

	out, err := execute(t, "test", "run", hub.URL, "demo", "7", "smoke", "--binary", writeBinary(t), "--tags", "a,b", "-a")

	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "upload finished new binaryID: 11") || !strings.Contains(out, "execute test 3 with binaryId: 11") {
		t.Errorf("out = %s", out)
	}
	got := hub.seen()
	if len(got) != 4 || !strings.HasPrefix(got[0], "POST "+appPrefix+"/upload") || !strings.HasPrefix(got[1], "POST "+appPrefix+"/bundle/11") || !strings.Contains(got[1], `"tags":"a,b"`) {
		t.Errorf("requests = %q", got)
	}
}

func TestRunUploadsWithoutTagsSkipsTheUpdate(t *testing.T) {
	hub := newFakeHub(t)

	if _, err := execute(t, "test", "run", hub.URL, "demo", "7", "smoke", "--binary", writeBinary(t), "-a"); err != nil {
		t.Fatal(err)
	}

	for _, req := range hub.seen() {
		if strings.Contains(req, "/bundle/") {
			t.Errorf("unexpected update: %s", req)
		}
	}
}

func TestRunWaitsForTheResultOverServerSentEvents(t *testing.T) {
	hub := newFakeHub(t)
	hub.publish("test_run_5_log", map[string]any{"testRunId": 5, "entry": "step one"})
	hub.publish("test_run_5_finished", map[string]any{"success": true})

	out, err := execute(t, "test", "run", hub.URL, "demo", "7", "smoke", "--timeout", "10s")

	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Test finished Successfully!") || strings.Contains(out, "Test Running...") {
		t.Errorf("out = %s", out)
	}
}

func TestRunFailures(t *testing.T) {
	tests := map[string]struct {
		args    func(url string) []string
		failOn  string
		wantErr string
	}{
		"missing parameters":  {args: func(u string) []string { return []string{"test", "run", u, "demo", "7"} }, wantErr: "missing parameter"},
		"bad app id":          {args: func(u string) []string { return []string{"test", "run", u, "demo", "x", "smoke"} }, wantErr: "invalid appID"},
		"unknown test":        {args: func(u string) []string { return []string{"test", "run", u, "demo", "7", "other"} }, wantErr: "could not be found"},
		"empty test name":     {args: func(u string) []string { return []string{"test", "run", u, "demo", "7", ""} }, wantErr: "could not be found"},
		"missing binary file": {args: func(u string) []string { return []string{"test", "run", u, "demo", "7", "smoke", "--binary", "/none"} }, wantErr: "open bundle"},
		"upload fails": {args: func(u string) []string {
			return []string{"test", "run", u, "demo", "7", "smoke", "--binary", writeBinary(t)}
		}, failOn: appPrefix + "/upload", wantErr: "hub says no"},
		"tagging fails": {args: func(u string) []string {
			return []string{"test", "run", u, "demo", "7", "smoke", "--binary", writeBinary(t), "--tags", "x"}
		}, failOn: appPrefix + "/bundle/11", wantErr: "hub says no"},
		"test list fails":   {args: func(u string) []string { return []string{"test", "run", u, "demo", "7", "smoke"} }, failOn: appPrefix + "/tests", wantErr: "hub says no"},
		"start fails":       {args: func(u string) []string { return []string{"test", "run", u, "demo", "7", "smoke"} }, failOn: appPrefix + "/test/3/run", wantErr: "hub says no"},
		"no result in time": {args: func(u string) []string { return []string{"test", "run", u, "demo", "7", "smoke", "--timeout", "50ms"} }, wantErr: "waiting for test run 5"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			hub := newFakeHub(t)
			hub.events.CreateStream("test_run_5_finished")
			hub.events.CreateStream("test_run_5_log")
			if tc.failOn != "" {
				hub.failOn(tc.failOn, 500)
			}
			if _, err := execute(t, tc.args(hub.URL)...); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("err = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestRunCommandDoesNotReportSuccessForAFailedRun(t *testing.T) {
	hub := newFakeHub(t)
	var out bytes.Buffer
	stream := fakeStream{events: map[string][]*sse.Event{"test_run_5_finished": {{Data: []byte(`{"success":false}`)}}}}
	r := runCommand{
		out: &out, opts: &runOptions{timeout: time.Second},
		newStream: func(string) eventStream { return stream },
	}

	err := r.run(context.Background(), []string{hub.URL, "demo", "7", "smoke"})

	if err != nil || strings.Contains(out.String(), "Successfully") {
		t.Errorf("err = %v, out = %s", err, &out)
	}
}
