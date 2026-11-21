package wiring

import (
	"io"

	"github.com/sirupsen/logrus"
)

func quietLoggerInternal() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logger
}
