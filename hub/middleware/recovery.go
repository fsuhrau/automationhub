// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

const (
	logPrefix = "\n\n\x1b[31m"
	reset     = "\033[0m"
)

// recoverer turns panics of the handlers into a 500 and logs them.
type recoverer struct {
	logger *log.Logger // nil: nothing is logged
	now    func() time.Time
}

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() gin.HandlerFunc {
	return RecoveryWithWriter(gin.DefaultErrorWriter)
}

// RecoveryWithWriter returns a middleware for a given writer that recovers from any panics and writes a 500 if there was one.
func RecoveryWithWriter(out io.Writer) gin.HandlerFunc {
	r := &recoverer{now: time.Now}
	if out != nil {
		r.logger = log.New(out, logPrefix, log.LstdFlags)
	}
	return r.middleware
}

func (r *recoverer) middleware(c *gin.Context) {
	defer func() {
		if v := recover(); v != nil {
			r.handle(c, v)
		}
	}()

	if !c.IsAborted() {
		c.Next()
	}
}

func (r *recoverer) handle(c *gin.Context, v any) {
	err := asError(v)
	brokenPipe := isBrokenConnection(v)
	r.logPanic(c, err, brokenPipe)
	abort(c, err, brokenPipe)
	sentry.CaptureException(err)
}

// abort answers with a 500. If the connection is dead, we can't write a
// status to it, so the error is only attached.
func abort(c *gin.Context, err error, brokenPipe bool) {
	if brokenPipe {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.AbortWithStatus(http.StatusInternalServerError)
}

func asError(v any) error {
	if err, ok := v.(error); ok {
		return err
	}
	return fmt.Errorf("%v", v)
}

// isBrokenConnection checks for a broken connection, as it is not really a
// condition that warrants a panic stack trace.
func isBrokenConnection(v any) bool {
	ne, ok := v.(*net.OpError)
	if !ok {
		return false
	}
	se, ok := ne.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	msg := strings.ToLower(se.Error())
	return strings.Contains(msg, "broken pipe") || strings.Contains(msg, "connection reset by peer")
}

func (r *recoverer) logPanic(c *gin.Context, err error, brokenPipe bool) {
	if r.logger == nil {
		return
	}
	httpRequest, _ := httputil.DumpRequest(c.Request, false)
	switch {
	case brokenPipe:
		r.logger.Printf("%s\n%s%s", err, string(httpRequest), reset)
	case gin.IsDebugging():
		r.logger.Printf("[Recovery] %s panic recovered:\n%s\n%s\n%s%s",
			timeFormat(r.now()), maskAuthorization(httpRequest), err, stack(5), reset)
	default:
		r.logger.Printf("[Recovery] %s panic recovered:\n%s\n%s%s",
			timeFormat(r.now()), err, stack(5), reset)
	}
}

// maskAuthorization hides the value of the Authorization header of a dumped request.
func maskAuthorization(httpRequest []byte) string {
	headers := strings.Split(string(httpRequest), "\r\n")
	for idx, header := range headers {
		current := strings.Split(header, ":")
		if current[0] == "Authorization" {
			headers[idx] = current[0] + ": *"
		}
	}
	return strings.Join(headers, "\r\n")
}

func timeFormat(t time.Time) string {
	return t.Format("2006/01/02 - 15:04:05")
}
