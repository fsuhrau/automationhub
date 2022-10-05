package webdriver

import (
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

var (
	serverRegex = regexp.MustCompile(`ServerURLHere->(.*)<-ServerURLHere`)
	WDAHook     *WDAConnectHook
)

func init() {
	WDAHook = &WDAConnectHook{
		Connected: make(chan string),
	}
	logrus.AddHook(WDAHook)
}

type WDAConnectHook struct {
	Connected chan string
}

func (hook *WDAConnectHook) Fire(entry *logrus.Entry) error {

	if strings.Contains(entry.Message, "ServerURLHere") {
		matches := serverRegex.FindAllStringSubmatch(entry.Message, 1)
		address := matches[0][1]
		if len(address) > 0 {
			hook.Connected <- address
		}
	}
	return nil
}

func (hook *WDAConnectHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}
