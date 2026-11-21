package middleware

import "time"

// Clock is the time source of the log middleware.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }
