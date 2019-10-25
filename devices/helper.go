package devices

import (
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type LogWriter struct {
	Tag string
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	logrus.WithField("tag", w.Tag).Errorf("%s", string(p))
	return len(p), nil
}

func NewCommand(executable string, params ...string) *exec.Cmd {
	writer := &LogWriter{
		Tag: executable + " " + strings.Join(params, " "),
	}
	logrus.Debugf("New Command: %s %v", executable, params)
	cmd := exec.Command(executable, params...)
	cmd.Stderr = writer
	return cmd
}
