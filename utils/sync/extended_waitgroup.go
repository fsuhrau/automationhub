package sync

import (
	"context"
	"fmt"
	"sync"
	"time"
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
	Wait()
}

type extendedWaitGroup struct {
	group sync.WaitGroup
	until time.Time
	ctx   context.Context
}

func (wg *extendedWaitGroup) Done() {
	wg.group.Done()
}

func (wg *extendedWaitGroup) Add(delta int) {
	wg.group.Add(delta)
}

func (wg *extendedWaitGroup) WaitWithTimeout(duration time.Duration) error {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(duration)
		timeout <- true
	}()

	wait := make(chan bool, 1)
	go func() {
		wg.group.Wait()
		wait <- true
	}()

	select {
	case <-timeout:
		return TimeoutError
	case <-wait:
		return nil
	case <-wg.ctx.Done():
		return wg.ctx.Err()
	}
}

func (wg *extendedWaitGroup) WaitUntil(waitUntil time.Time) error {
	wg.until = waitUntil

	timeout := make(chan bool, 1)
	go func() {
		for {
			time.Sleep(50 * time.Millisecond)
			if time.Now().After(wg.until) {
				timeout <- true
				return
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
		return TimeoutError
	case <-wait:
		return nil
	case <-wg.ctx.Done():
		return wg.ctx.Err()
	}
}

func (wg *extendedWaitGroup) UpdateUntil(waitUntil time.Time) {
	wg.until = waitUntil
}

func (wg *extendedWaitGroup) Wait() {
	wg.group.Wait()
}

func NewExtendedWaitGroup(ctx context.Context) *extendedWaitGroup {
	return &extendedWaitGroup{
		ctx: ctx,
	}
}
