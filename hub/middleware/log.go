package middleware

import (
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger logs every request and its result with the system clock.
func Logger(logger logrus.FieldLogger) gin.HandlerFunc {
	return LoggerWithClock(logger, systemClock{})
}

// LoggerWithClock logs every request when it starts and when it is done, and
// measures the latency with clock.
func LoggerWithClock(logger logrus.FieldLogger, clock Clock) gin.HandlerFunc {
	return func(c *gin.Context) {
		rl := &requestLog{
			logger:   logger,
			c:        c,
			path:     c.Request.URL.Path, // other handler can change c.Path so
			clientIP: c.ClientIP(),
		}
		start := clock.Now()
		rl.start()
		c.Next()
		rl.done(latencyMs(clock.Now().Sub(start)))
	}
}

// requestLog logs one request.
type requestLog struct {
	logger   logrus.FieldLogger
	c        *gin.Context
	path     string
	clientIP string
}

func (r *requestLog) start() {
	entry := r.logger.WithFields(logrus.Fields{"address": r.clientIP})
	if len(r.c.Errors) > 0 {
		entry.Error(r.c.Errors.ByType(gin.ErrorTypePrivate).String())
		return
	}
	entry.Info(fmt.Sprintf("[%s] \"%s\"", r.c.Request.Method, r.path))
}

func (r *requestLog) done(latency int) {
	statusCode := r.c.Writer.Status()
	entry := r.logger.WithFields(logrus.Fields{
		"statusCode": statusCode,
		"latency":    latency, // time to process
		"address":    r.clientIP,
	})
	if len(r.c.Errors) > 0 {
		entry.Error(r.c.Errors.ByType(gin.ErrorTypePrivate).String())
		return
	}
	dataLength := max(r.c.Writer.Size(), 0)
	msg := fmt.Sprintf("[%s] \"%s\" %d (%dms)", r.c.Request.Method, r.path, dataLength, latency)
	logByStatus(entry, statusCode, msg)
}

func logByStatus(entry *logrus.Entry, statusCode int, msg string) {
	switch {
	case statusCode > 499:
		entry.Error(msg)
	case statusCode > 399:
		entry.Warn(msg)
	default:
		entry.Info(msg)
	}
}

func latencyMs(d time.Duration) int {
	return int(math.Ceil(float64(d.Nanoseconds()) / 1000000.0))
}
