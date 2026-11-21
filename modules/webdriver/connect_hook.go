package webdriver

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var serverRegex = regexp.MustCompile(`ServerURLHere->(.*)<-ServerURLHere`)

// WDAConnectHook picks the WebDriverAgent server URL out of the go-ios log
// output. go-ios logs the URL without the device id, so startups are
// serialized through Await: only one device waits for a URL at a time.
//
// The hook only sees the log output once it is registered with a logger, e.g.
// logrus.AddHook(hook). The zero value is ready to use.
type WDAConnectHook struct {
	startMu sync.Mutex // serializes Await calls

	mu     sync.Mutex
	waiter chan string
}

// ParseServerURL finds the server URL that WebDriverAgent logs as
// "ServerURLHere->http://…:8100<-ServerURLHere".
func ParseServerURL(line string) (string, bool) {
	if !strings.Contains(line, "ServerURLHere") {
		return "", false
	}
	matches := serverRegex.FindStringSubmatch(line)
	if len(matches) < 2 || len(matches[1]) == 0 {
		return "", false
	}
	return matches[1], true
}

func (hook *WDAConnectHook) Fire(entry *logrus.Entry) error {
	address, ok := ParseServerURL(entry.Message)
	if !ok {
		return nil
	}

	hook.mu.Lock()
	waiter := hook.waiter
	hook.mu.Unlock()
	if waiter != nil {
		// never block the logger; the waiter channel is buffered
		select {
		case waiter <- address:
		default:
		}
	}
	return nil
}

// Await runs start and waits for the server URL it logs. start calls fail
// when the runner stops before it reported the URL, which ends the wait. The
// wait gives up after timeout, measured with after (time.After or a clock's).
func (hook *WDAConnectHook) Await(timeout time.Duration, after func(time.Duration) <-chan time.Time, start func(fail func(error))) (string, error) {
	hook.startMu.Lock()
	defer hook.startMu.Unlock()

	waiter := make(chan string, 1)
	hook.mu.Lock()
	hook.waiter = waiter
	hook.mu.Unlock()
	defer func() {
		hook.mu.Lock()
		hook.waiter = nil
		hook.mu.Unlock()
	}()

	failed := make(chan error, 1)
	start(func(err error) {
		select {
		case failed <- err:
		default:
		}
	})

	select {
	case address := <-waiter:
		return address, nil
	case err := <-failed:
		return "", fmt.Errorf("WebDriverAgent runner failed to start: %w", err)
	case <-after(timeout):
		return "", fmt.Errorf("WebDriverAgent did not report its server URL within %v", timeout)
	}
}

func (hook *WDAConnectHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
