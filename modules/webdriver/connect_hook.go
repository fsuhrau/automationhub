package webdriver

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	serverRegex = regexp.MustCompile(`ServerURLHere->(.*)<-ServerURLHere`)
	WDAHook     *WDAConnectHook
)

func init() {
	WDAHook = &WDAConnectHook{}
	logrus.AddHook(WDAHook)
}

// WDAConnectHook picks the WebDriverAgent server URL out of the go-ios log
// output. go-ios logs the URL without the device id, so startups are
// serialized through Await: only one device waits for a URL at a time.
type WDAConnectHook struct {
	startMu sync.Mutex // serializes Await calls

	mu     sync.Mutex
	waiter chan string
}

func (hook *WDAConnectHook) Fire(entry *logrus.Entry) error {
	if !strings.Contains(entry.Message, "ServerURLHere") {
		return nil
	}
	matches := serverRegex.FindStringSubmatch(entry.Message)
	if len(matches) < 2 || len(matches[1]) == 0 {
		return nil
	}

	hook.mu.Lock()
	waiter := hook.waiter
	hook.mu.Unlock()
	if waiter != nil {
		// never block the logger; the waiter channel is buffered
		select {
		case waiter <- matches[1]:
		default:
		}
	}
	return nil
}

// Await runs start and waits for the server URL it logs.
func (hook *WDAConnectHook) Await(timeout time.Duration, start func()) (string, error) {
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

	start()

	select {
	case address := <-waiter:
		return address, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("WebDriverAgent did not report its server URL within %v", timeout)
	}
}

func (hook *WDAConnectHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
