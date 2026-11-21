package scenario

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/fsuhrau/automationhub/tester/protocol"
	hubsync "github.com/fsuhrau/automationhub/utils/sync"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type finishedRecorder chan events.TestRunFinishedPayload

func (r finishedRecorder) Handle(p events.TestRunFinishedPayload) { r <- p }

// newRunner returns a runner with an in-memory database and a protocol writer.
func newRunner(t *testing.T) *testsRunner {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.TestRun{}, &models.TestRunLogEntry{}); err != nil {
		t.Fatal(err)
	}
	run := models.TestRun{}
	db.Create(&run)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	tr := &testsRunner{ctx: ctx, cancelFunc: cancel}
	tr.Init(nil, "", db, "p", 1)
	tr.ProtocolWriter = protocol.NewProtocolWriter(db, "p", 1, "t", &run)
	return tr
}

// runSucceeded closes the run's protocol writer and returns whether the run
// is reported as a success.
func runSucceeded(t *testing.T, tr *testsRunner) bool {
	t.Helper()
	r := make(finishedRecorder, 4)
	events.TestRunFinished.Register(r)
	tr.ProtocolWriter.Close()
	for {
		select {
		case e := <-r:
			if e.TestRunID == tr.ProtocolWriter.RunID() {
				return e.Success
			}
		case <-time.After(time.Second):
			t.Fatal("no finished event")
		}
	}
}

func logCount(tr *testsRunner, level string) int64 {
	var n int64
	tr.DB.Model(&models.TestRunLogEntry{}).Where("level = ?", level).Count(&n)
	return n
}

func TestCheckStarted(t *testing.T) {
	one := []base.DeviceMap{{}}
	tests := []struct {
		name        string
		connected   []base.DeviceMap
		err         error
		want        bool
		wantErrLogs int64
	}{
		{"connected", one, nil, true, 0},
		{"no device connected", nil, nil, false, 1},
		{"wrapped timeout is recognized", one, fmt.Errorf("start: %w", hubsync.TimeoutError), true, 1},
		{"timeout without device", nil, hubsync.TimeoutError, false, 2},
		{"other error is no timeout", one, errors.New("boom"), true, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := newRunner(t)
			if got := tr.checkStarted(tc.connected, tc.err); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			if n := logCount(tr, "error"); n != tc.wantErrLogs {
				t.Fatalf("got %d error logs, want %d", n, tc.wantErrLogs)
			}
			if got := runSucceeded(t, tr); got != tc.want {
				t.Fatalf("run success=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestRecoverRunFailsTheRun(t *testing.T) {
	tr := newRunner(t)
	func() {
		defer tr.recoverRun()
		panic("boom")
	}()
	if logCount(tr, "error") != 1 {
		t.Fatal("panic not logged as error")
	}
	if runSucceeded(t, tr) {
		t.Fatal("run with a panic reported as success")
	}
}

func TestRecoverRunWithoutPanic(t *testing.T) {
	tr := newRunner(t)
	func() { defer tr.recoverRun() }()
	if logCount(tr, "error") != 0 {
		t.Fatal("error logged without a panic")
	}
}

type startRecorder struct {
	mu    sync.Mutex
	tasks []action.TestStart
}

func (r *startRecorder) run(_ base.DeviceMap, task action.TestStart, _ string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks = append(r.tasks, task)
}

// Every execution type hands the run's env to the tests; Simultaneously
// forgot it.
func TestScheduleTestsPassesEnv(t *testing.T) {
	env := map[string]string{"KEY": "value"}
	tests := []struct {
		name string
		typ  models.ExecutionType
		want int
	}{
		{"simultaneously", models.SimultaneouslyExecutionType, 4},
		{"concurrent", models.ConcurrentExecutionType, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := newRunner(t)
			rec := &startRecorder{}
			tr.execTest = rec.run
			tr.env = env
			tr.Config.ExecutionType = tc.typ
			tests := []models.UnityTestFunction{{Class: "A", Method: "m1"}, {Class: "B", Method: "m2"}}
			tr.scheduleTests([]base.DeviceMap{{}, {}}, tests)

			if len(rec.tasks) != tc.want {
				t.Fatalf("got %d test starts, want %d", len(rec.tasks), tc.want)
			}
			for _, task := range rec.tasks {
				if !reflect.DeepEqual(task.Env, env) {
					t.Fatalf("test start %s has env %v, want %v", task.Class, task.Env, env)
				}
			}
		})
	}
}

// A run without devices fails and finishes, and the runner's default test
// executor is wired.
func TestExecWithoutDevicesFailsTheRun(t *testing.T) {
	tr := New(nil, "", nil, nil, "p", 1)
	if tr.execTest == nil {
		t.Fatal("default test executor missing")
	}
	tr2 := newRunner(t)
	tr2.exec(nil, nil, "")
	if !tr2.WaitFinished(time.Second) {
		t.Fatal("run did not finish")
	}
	if runSucceeded(t, tr2) {
		t.Fatal("run without devices reported as success")
	}
}
