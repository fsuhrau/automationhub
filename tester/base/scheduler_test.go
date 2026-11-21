package base

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

type started struct {
	device string
	task   action.TestStart
}

// startLog records which test starts which device got, in order.
type startLog struct {
	mu     sync.Mutex
	starts []started
}

func (l *startLog) run(_ context.Context, dev DeviceMap, task action.TestStart) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.starts = append(l.starts, started{dev.Device.DeviceID(), task})
}

func (l *startLog) forDevice(id string) []string {
	var methods []string
	for _, s := range l.starts {
		if s.device == id {
			methods = append(methods, s.task.Method)
		}
	}
	return methods
}

func twoDevices() []DeviceMap {
	return []DeviceMap{{Device: testkit.NewFakeDevice("a")}, {Device: testkit.NewFakeDevice("b")}}
}

func threeTests() []models.UnityTestFunction {
	return []models.UnityTestFunction{
		{Assembly: "As", Class: "A", Method: "m1"},
		{Assembly: "As", Class: "B", Method: "m2"},
		{Assembly: "As", Class: "C", Method: "m3"},
	}
}

// Both execution types hand the run's env to the tests.
func TestScheduleBuildsOneStartPerTestWithEnv(t *testing.T) {
	env := map[string]string{"KEY": "value"}
	tests := []struct {
		name  string
		mode  models.ExecutionType
		wantN int
	}{
		{"simultaneously runs every test on every device", models.SimultaneouslyExecutionType, 6},
		{"concurrent runs every test once", models.ConcurrentExecutionType, 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			log := &startLog{}
			Scheduler{Mode: tc.mode, Env: env, Run: log.run}.Schedule(context.Background(), twoDevices(), threeTests())

			if len(log.starts) != tc.wantN {
				t.Fatalf("got %d test starts, want %d", len(log.starts), tc.wantN)
			}
			for _, s := range log.starts {
				if !reflect.DeepEqual(s.task.Env, env) || s.task.Assembly != "As" {
					t.Fatalf("test start %+v has the wrong env or assembly", s.task)
				}
			}
		})
	}
}

func TestSimultaneousOrdersTheTestsPerDevice(t *testing.T) {
	log := &startLog{}
	Scheduler{Mode: models.SimultaneouslyExecutionType, Run: log.run}.Schedule(context.Background(), twoDevices(), threeTests())

	want := []string{"m1", "m2", "m3"}
	for _, id := range []string{"a", "b"} {
		if got := log.forDevice(id); !reflect.DeepEqual(got, want) {
			t.Fatalf("device %s ran %v, want %v", id, got, want)
		}
	}
}

func TestConcurrentSharesOneQueue(t *testing.T) {
	log := &startLog{}
	one := []DeviceMap{{Device: testkit.NewFakeDevice("a")}}
	Scheduler{Mode: models.ConcurrentExecutionType, Run: log.run}.Schedule(context.Background(), one, threeTests())

	if got, want := log.forDevice("a"), []string{"m1", "m2", "m3"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ran %v, want %v", got, want)
	}
}

func TestScheduleWithoutTestsOrDevices(t *testing.T) {
	log := &startLog{}
	s := Scheduler{Run: log.run}
	s.Schedule(context.Background(), twoDevices(), nil)
	s.Schedule(context.Background(), nil, threeTests())
	if len(log.starts) != 0 {
		t.Fatalf("ran %v", log.starts)
	}
}

// A cancelled run starts no further test.
func TestScheduleStopsWhenTheContextEnds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	log := &startLog{}
	run := func(ctx context.Context, dev DeviceMap, task action.TestStart) {
		log.run(ctx, dev, task)
		cancel()
	}
	one := []DeviceMap{{Device: testkit.NewFakeDevice("a")}}
	Scheduler{Mode: models.ConcurrentExecutionType, Run: run}.Schedule(ctx, one, threeTests())

	if len(log.starts) != 1 {
		t.Fatalf("ran %d tests after the cancel, want 1", len(log.starts))
	}
}

// A panic of a worker reaches the goroutine that scheduled, where the run
// recovers it; the first one wins.
func TestSchedulePanicReachesTheCaller(t *testing.T) {
	run := func(context.Context, DeviceMap, action.TestStart) { panic("boom") }
	defer func() {
		if got := recover(); got != "boom" {
			t.Fatalf("recovered %v, want boom", got)
		}
	}()
	Scheduler{Mode: models.SimultaneouslyExecutionType, Run: run}.Schedule(context.Background(), twoDevices(), threeTests())
	t.Fatal("no panic")
}

func TestMethodName(t *testing.T) {
	tests := map[string]string{"Class.Method": "Class.Method", "Assembly Class.Method": "Class.Method", "": ""}
	for in, want := range tests {
		if got := MethodName(in); got != want {
			t.Errorf("MethodName(%q) = %q, want %q", in, got, want)
		}
	}
}
