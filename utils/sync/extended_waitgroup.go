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
	Wait() error
	IsCanceled() bool
}

type extendedWaitGroup struct {
	group    sync.WaitGroup
	until    time.Time
	ctx      context.Context
	canceled bool
}

func (wg *extendedWaitGroup) IsCanceled() bool {
	return wg.canceled
}

func (wg *extendedWaitGroup) Done() {
	wg.group.Done()
}

func (wg *extendedWaitGroup) Add(delta int) {
	wg.group.Add(delta)
}

func (wg *extendedWaitGroup) UpdateUntil(waitUntil time.Time) {
	wg.until = waitUntil
}

func (wg *extendedWaitGroup) WaitWithTimeout(duration time.Duration) error {
	timeout := make(chan bool, 1)
	//defer close(timeout)
	go func() {
		time.Sleep(duration)
		timeout <- true
	}()

	wait := make(chan bool, 1)
	//defer close(wait)
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
		wg.canceled = true
		return wg.ctx.Err()
	}
}

func (wg *extendedWaitGroup) WaitUntil(waitUntil time.Time) error {
	wg.until = waitUntil

	timeout := make(chan bool, 1)
	//defer close(timeout)
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
	//defer close(wait)
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
		wg.canceled = true
		return wg.ctx.Err()
	}
}

func (wg *extendedWaitGroup) Wait() error {
	wait := make(chan bool, 1)
	//defer close(wait)

	go func() {
		wg.group.Wait()
		wait <- true
	}()

	select {
	case <-wait:
		return nil
	case <-wg.ctx.Done():
		wg.canceled = true
		return wg.ctx.Err()
	}
}

func NewExtendedWaitGroup(ctx context.Context) *extendedWaitGroup {
	return &extendedWaitGroup{
		ctx: ctx,
	}
}
