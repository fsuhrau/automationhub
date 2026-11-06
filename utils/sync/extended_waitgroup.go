package sync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var (
	TimeoutError = fmt.Errorf("group timed out")
)

// deadlineCheckInterval is how often WaitUntil re-evaluates a deadline that
// may be moved by UpdateUntil.
const deadlineCheckInterval = time.Second

type ExtendedWaitGroup interface {
	Done()
	Add(delta int)
	WaitWithTimeout(duration time.Duration) error
	WaitUntil(waitUntil time.Time) error
	UpdateUntil(waitUntil time.Time)
	Wait() error
	IsCanceled() bool
}

type extendedWaitGroup struct {
	group    sync.WaitGroup
	mu       sync.Mutex
	until    time.Time
	ctx      context.Context
	canceled atomic.Bool
	id       string
}

func (wg *extendedWaitGroup) logAction(action string) {
	if false { // for debugging purposes
		log.Printf("ExtendedWaitGroup [%s]: %s", wg.id, action)
	}
}

func (wg *extendedWaitGroup) IsCanceled() bool {
	return wg.canceled.Load()
}

func (wg *extendedWaitGroup) Done() {
	wg.logAction("Done called")
	wg.group.Done()
}

func (wg *extendedWaitGroup) Add(delta int) {
	wg.logAction(fmt.Sprintf("Add called with delta: %d", delta))
	wg.group.Add(delta)
}

func (wg *extendedWaitGroup) UpdateUntil(waitUntil time.Time) {
	wg.logAction(fmt.Sprintf("UpdateUntil called with time: %s", waitUntil))
	wg.mu.Lock()
	wg.until = waitUntil
	wg.mu.Unlock()
}

func (wg *extendedWaitGroup) getUntil() time.Time {
	wg.mu.Lock()
	defer wg.mu.Unlock()
	return wg.until
}

// waitDone returns a channel that is closed when the group counter reaches zero.
// The helper goroutine only ends when the group completes.
func (wg *extendedWaitGroup) waitDone() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		wg.group.Wait()
		close(done)
	}()
	return done
}

func (wg *extendedWaitGroup) cancelWith(err error) error {
	wg.canceled.Store(true)
	return err
}

func (wg *extendedWaitGroup) WaitWithTimeout(duration time.Duration) error {
	wg.logAction(fmt.Sprintf("WaitWithTimeout called with duration: %s", duration))

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-wg.waitDone():
		return nil
	case <-wg.ctx.Done():
		return wg.cancelWith(wg.ctx.Err())
	case <-timer.C:
		return wg.cancelWith(TimeoutError)
	}
}

func (wg *extendedWaitGroup) WaitUntil(waitUntil time.Time) error {
	wg.logAction(fmt.Sprintf("WaitUntil called with time: %s", waitUntil))
	wg.UpdateUntil(waitUntil)

	done := wg.waitDone()
	for {
		remaining := time.Until(wg.getUntil())
		if remaining <= 0 {
			return wg.cancelWith(TimeoutError)
		}
		if remaining > deadlineCheckInterval {
			remaining = deadlineCheckInterval
		}
		timer := time.NewTimer(remaining)
		select {
		case <-done:
			timer.Stop()
			return nil
		case <-wg.ctx.Done():
			timer.Stop()
			return wg.cancelWith(wg.ctx.Err())
		case <-timer.C:
		}
	}
}

func (wg *extendedWaitGroup) Wait() error {
	wg.logAction("Wait called")
	select {
	case <-wg.waitDone():
		return nil
	case <-wg.ctx.Done():
		return wg.cancelWith(wg.ctx.Err())
	}
}

func NewExtendedWaitGroup(ctx context.Context) *extendedWaitGroup {
	return &extendedWaitGroup{
		ctx: ctx,
		id:  uuid.New().String(),
	}
}
