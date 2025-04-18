package sync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	TimeoutError = fmt.Errorf("group timed out")
)

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
	until    time.Time
	ctx      context.Context
	canceled bool
	id       string
}

func (wg *extendedWaitGroup) logAction(action string) {
	if false { // for debugging purposes
		log.Printf("ExtendedWaitGroup [%s]: %s", wg.id, action)
	}
}

func (wg *extendedWaitGroup) IsCanceled() bool {
	wg.logAction("Checked if canceled")
	return wg.canceled
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
	wg.until = waitUntil
}

func (wg *extendedWaitGroup) WaitWithTimeout(duration time.Duration) error {
	wg.logAction(fmt.Sprintf("WaitWithTimeout called with duration: %s", duration))

	ctx, cancel := context.WithTimeout(wg.ctx, duration)
	defer cancel() // Ensure the context is canceled to release resources

	wait := make(chan bool, 1)
	defer close(wait)

	go func() {
		wg.group.Wait()
		if wg.canceled {
			return
		}
		if wait != nil {
			wait <- true
		}
	}()

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			wg.canceled = true
			wg.logAction("WaitWithTimeout timed out")
			return TimeoutError
		}
		wg.logAction("WaitWithTimeout canceled by context")
		return ctx.Err()
	case <-wait:
		wg.logAction("WaitWithTimeout completed")
		return nil
	}
}

func (wg *extendedWaitGroup) WaitUntil(waitUntil time.Time) error {
	wg.logAction(fmt.Sprintf("WaitUntil called with time: %s", waitUntil))
	wg.until = waitUntil

	timeout := make(chan bool, 1)
	done := make(chan struct{})
	defer close(timeout)
	defer close(done)

	ticker := time.NewTicker(50 * time.Minute)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case _, ok := <-ticker.C:
				if !ok {
					return
				}
				if time.Now().After(wg.until) {
					timeout <- true
					return
				}
			}
		}
	}()

	wait := make(chan bool, 1)
	go func() {
		wg.group.Wait()
		wait <- true
	}()

	select {
	case <-timeout:
		wg.canceled = true
		wg.logAction("WaitUntil timed out")
		return TimeoutError
	case <-wait:
		wg.logAction("WaitUntil completed")
		return nil
	case <-wg.ctx.Done():
		wg.canceled = true
		wg.logAction("WaitUntil canceled by context")
		return wg.ctx.Err()
	}
}

func (wg *extendedWaitGroup) Wait() error {
	wg.logAction("Wait called")
	wait := make(chan bool, 1)
	defer close(wait)
	go func() {
		wg.group.Wait()
		wait <- true
	}()

	select {
	case <-wait:
		wg.logAction("Wait completed")
		return nil
	case <-wg.ctx.Done():
		wg.canceled = true
		wg.logAction("Wait canceled by context")
		return wg.ctx.Err()
	}
}

func NewExtendedWaitGroup(ctx context.Context) *extendedWaitGroup {
	id := uuid.New().String()
	log.Printf("ExtendedWaitGroup [%s]: Created", id)
	return &extendedWaitGroup{
		ctx: ctx,
		id:  id,
	}
}
