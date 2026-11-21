package base

import (
	"context"
	"strings"
	"sync"

	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Scheduler hands the tests to the connected devices as the execution type
// of the config says, and waits until all are done.
type Scheduler struct {
	Mode models.ExecutionType
	// Env is passed to every test.
	Env map[string]string
	// Run runs one test on one device.
	Run func(ctx context.Context, dev DeviceMap, task action.TestStart)
}

// MethodName is the method of a test start without its assembly prefix
// ("Assembly Class.Method" becomes "Class.Method").
func MethodName(method string) string {
	if parts := strings.Split(method, " "); len(parts) > 1 {
		return parts[1]
	}
	return method
}

// panicBox carries the first panic of a worker to the goroutine that
// scheduled it, where the run recovers it.
type panicBox struct {
	mu    sync.Mutex
	value any
	set   bool
}

// capture recovers a panic of the calling goroutine; defer it directly.
func (b *panicBox) capture() {
	if r := recover(); r != nil {
		b.mu.Lock()
		defer b.mu.Unlock()
		if !b.set {
			b.value, b.set = r, true
		}
	}
}

func (b *panicBox) rethrow() {
	if b.set {
		panic(b.value)
	}
}

// Schedule runs every test: Simultaneously on every device, Concurrent on
// the first device that is free. Each device takes its tests in order.
func (s Scheduler) Schedule(ctx context.Context, devices []DeviceMap, tests []models.UnityTestFunction) {
	queues := s.queues(len(devices), s.starts(tests))
	var wg sync.WaitGroup
	var failure panicBox
	for i, d := range devices {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer failure.capture()
			s.work(ctx, queues[i], d)
		}()
	}
	wg.Wait()
	failure.rethrow()
}

// starts builds the test start of every test once.
func (s Scheduler) starts(tests []models.UnityTestFunction) []action.TestStart {
	starts := make([]action.TestStart, 0, len(tests))
	for _, t := range tests {
		starts = append(starts, action.TestStart{Assembly: t.Assembly, Class: t.Class, Method: t.Method, Env: s.Env})
	}
	return starts
}

// queues returns the queue of each device: its own copy of all starts when
// the tests run simultaneously, else one queue all devices share.
func (s Scheduler) queues(devices int, starts []action.TestStart) []<-chan action.TestStart {
	shared := queueOf(starts)
	queues := make([]<-chan action.TestStart, devices)
	for i := range queues {
		queues[i] = shared
		if s.Mode == models.SimultaneouslyExecutionType {
			queues[i] = queueOf(starts)
		}
	}
	return queues
}

func queueOf(starts []action.TestStart) chan action.TestStart {
	queue := make(chan action.TestStart, len(starts))
	for _, start := range starts {
		queue <- start
	}
	close(queue)
	return queue
}

func (s Scheduler) work(ctx context.Context, queue <-chan action.TestStart, dev DeviceMap) {
	for task := range queue {
		if ctx.Err() != nil {
			return
		}
		s.Run(ctx, dev, task)
	}
}
