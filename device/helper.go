package device

import (
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

type CommandLogWriter struct {
	Tag string
}

func (w *CommandLogWriter) Write(p []byte) (n int, err error) {
	logrus.WithField("tag", w.Tag).Errorf("%s", string(p))
	return len(p), nil
}

func NewCommand(executable string, params ...string) *exec.Cmd {
	writer := &CommandLogWriter{
		Tag: executable + " " + strings.Join(params, " "),
	}
	logrus.Debugf("New Command: %s %v", executable, params)
	cmd := exec.Command(executable, params...)
	cmd.Stderr = writer
	return cmd
}
