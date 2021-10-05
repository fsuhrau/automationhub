package sync

import (
	"fmt"
	"sync"
	"time"
)

var (
	TimeoutError = fmt.Errorf("group timed out")
)

type ExtendedWaitGroup struct {
	group sync.WaitGroup
	until time.Time
}

func (wg *ExtendedWaitGroup) Done() {
	wg.group.Done()
}

func (wg *ExtendedWaitGroup) Add(delta int) {
	wg.group.Add(delta)
}

func (wg *ExtendedWaitGroup) WaitWithTimeout(duration time.Duration) error {
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
	case _ = <-timeout:
		return TimeoutError
	case _ = <-wait:
		return nil
	}
}

func (wg *ExtendedWaitGroup) WaitUntil(waitUntil time.Time) error {
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
	case _ = <-timeout:
		return TimeoutError
	case _ = <-wait:
		return nil
	}
}

func (wg *ExtendedWaitGroup) UpdateUntil(waitUntil time.Time) {
	wg.until = waitUntil
}

func (wg *ExtendedWaitGroup) Wait() {
	wg.group.Wait()
}