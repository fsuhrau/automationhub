package device_test

import "time"

func ptrNow() *time.Time {
	now := time.Now()
	return &now
}
