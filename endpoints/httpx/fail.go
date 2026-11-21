package httpx

import (
	"net/http"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/gin-gonic/gin"
)

const loggerKey = "httpx.logger"

// ErrorResponse is the body of every failed request.
type ErrorResponse struct {
	Message string `json:"message"`
}

// Logger is what Fail needs to report server errors; *logrus.Logger and
// *logrus.Entry fit.
type Logger interface {
	Errorf(format string, args ...any)
}

// WithLogger is the middleware that gives Fail its logger, so there is no
// global one. Without it Fail doesn't log.
func WithLogger(l Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(loggerKey, l)
		c.Next()
	}
}

// Fail aborts the request with the status of the error and the body
// {"message": "..."}. Server errors (5xx) are logged; a nil error is
// reported as an internal one.
func Fail(c *gin.Context, err error) {
	if err == nil {
		err = apperr.Internal("unknown error")
	}
	status := Status(err)
	if status >= http.StatusInternalServerError {
		logServerError(c, err)
	}
	c.AbortWithStatusJSON(status, ErrorResponse{Message: err.Error()})
}

func logServerError(c *gin.Context, err error) {
	v, _ := c.Get(loggerKey)
	if l, ok := v.(Logger); ok {
		l.Errorf("%s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	}
}
