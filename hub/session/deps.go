// Package session manages the test sessions of the hub and stores their files.
package session

import (
	"io"
	"os"
	"time"
)

// Clock is the time source of sessions and the manager. testkit.FakeClock
// implements it.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// FS is the part of the file system the session storage needs.
type FS interface {
	MkdirAll(path string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
}

// File is a file the session storage writes to.
type File interface {
	io.Writer
	io.Closer
}

// Deps are the clock and file system of the sessions.
type Deps struct {
	Clock Clock
	FS    FS
}

// SystemDeps are the real clock and file system.
func SystemDeps() Deps { return Deps{Clock: SystemClock{}, FS: OSFS{}} }

// SystemClock is the real time.
type SystemClock struct{}

// Now returns the current time.
func (SystemClock) Now() time.Time { return time.Now() }

// After waits for the duration d to elapse and then sends the time on the returned channel.
func (SystemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// OSFS is the real file system.
type OSFS struct{}

// MkdirAll creates a directory and all its parents.
func (OSFS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }

// OpenFile opens a file like os.OpenFile.
func (OSFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}
