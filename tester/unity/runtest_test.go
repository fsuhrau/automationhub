package unity

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/tester/base/basetest"
)

func testStart() action.TestStart {
	return action.TestStart{Assembly: "As", Class: "Cls", Method: "As Cls.Method", Env: map[string]string{"KEY": "value"}}
}

func (f *strategyFixture) runTest() {
	f.strategy.RunTest(context.Background(), base.DeviceMap{Device: f.Devices[0], Model: f.Models[0]}, testStart())
}

func screenFiles(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(apps.TestDataPath, "*.png"))
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestRunTestWritesTheProtocolOfAPassedTest(t *testing.T) {
	f := newStrategyFixture(t)
	f.Manager.TestPassed = true
	f.runTest()

	prots := f.protocols()
	if len(prots) != 1 || prots[0].TestName != "Cls/Cls.Method" {
		t.Fatalf("protocols %+v", prots)
	}
	if !f.logged("log", "finished successful") || len(f.Devices[0].CallsTo("Data")) != 0 {
		t.Fatal("passed test not reported as such")
	}
	sent := f.Manager.CallsTo("SendAction")[0].Args[1].(*action.TestStart)
	if sent.Env["KEY"] != "value" || sent.Method != "As Cls.Method" {
		t.Fatalf("test start %+v", sent)
	}
	if f.Devices[0].GetLogWriter() != nil {
		t.Fatal("log writer not removed")
	}
}

// A test that reports a failure is logged as failed, with no error to show.
func TestRunTestLogsAFailedTest(t *testing.T) {
	f := newStrategyFixture(t)
	f.Manager.TestPassed = false
	f.runTest()

	if !f.logged("log", "failed") {
		t.Fatal("failed test not logged")
	}
}

func TestRunTestTakesAScreenshotOfAnErroredTest(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(f *strategyFixture)
		wantFile  bool
		wantError string
	}{
		{"screenshot of the device", func(f *strategyFixture) { f.Devices[0].PNG = basetest.PNG() }, true, ""},
		{"screenshot of the app", func(f *strategyFixture) { f.Manager.Screenshot = basetest.PNG() }, true, ""},
		{"app has no screenshot", func(f *strategyFixture) {}, false, ""},
		{"file can't be saved", func(f *strategyFixture) {
			f.Devices[0].PNG = basetest.PNG()
			apps.TestDataPath = filepath.Join(apps.TestDataPath, "missing")
		}, false, "Save screenshot failed"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newStrategyFixture(t)
			f.Manager.FailWith("SendAction", nil)
			tc.prepare(f)
			f.strategy.captureScreenShot(context.Background(), testCase{dev: base.DeviceMap{Device: f.Devices[0]}, task: testStart()})

			if got := len(screenFiles(t)) == 1; got != tc.wantFile {
				t.Fatalf("screenshot saved %v, want %v", got, tc.wantFile)
			}
			if got := len(f.Devices[0].CallsTo("Data")) == 1; got != tc.wantFile {
				t.Fatalf("data set %v, want %v", got, tc.wantFile)
			}
			if tc.wantError != "" && !f.logged("error", tc.wantError) {
				t.Fatalf("missing log %q", tc.wantError)
			}
		})
	}
}

// An execution that fails gets its error and a screenshot in the log.
func TestRunTestReportsTheExecutionError(t *testing.T) {
	f := newStrategyFixture(t)
	f.Devices[0].PNG = basetest.PNG()
	f.Manager.FailWith("SendAction", errors.New("app gone"))
	f.runTest()

	if !f.logged("error", "failed with errors: app gone") || len(screenFiles(t)) != 1 {
		t.Fatal("execution error not reported")
	}
}

func TestScreenshotErrorOfTheAppIsLogged(t *testing.T) {
	f := newStrategyFixture(t)
	f.Manager.FailWith("SendAction", errors.New("app gone"))
	f.strategy.captureScreenShot(context.Background(), testCase{dev: base.DeviceMap{Device: f.Devices[0]}, task: testStart()})

	if !f.logged("error", "Take screenshot failed") {
		t.Fatal("screenshot error not logged")
	}
}

// A test whose protocol can't be created is skipped, it used to panic.
func TestRunTestWithoutProtocolIsSkipped(t *testing.T) {
	f := newStrategyFixture(t)
	if err := f.DB.Migrator().DropTable("test_protocols"); err != nil {
		t.Fatal(err)
	}
	f.runTest()

	if f.Devices[0].GetLogWriter() != nil || len(f.Manager.CallsTo("SendAction")) != 0 {
		t.Fatal("test ran without protocol")
	}
	if !f.logged("error", "Unable to create LogWriter for a") {
		t.Fatal("missing log")
	}
}

func TestErrorLines(t *testing.T) {
	got := errorLines(errors.New("a"), []error{errors.New("b"), errors.New("c")})
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("got %v", got)
	}
	if len(errorLines(nil, nil)) != 0 {
		t.Fatal("lines without errors")
	}
}
