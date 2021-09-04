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

func (wg *ExtendedWaitGroup) Wait() {
	wg.group.Wait()
}