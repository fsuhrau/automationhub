package recording

import (
	"io"
	"os"
	"time"
)

// Clock is the time source of a recording.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// File is a file a recording is written to.
type File interface {
	io.Writer
	io.WriterAt
	io.Seeker
	io.Closer
}

// FileSystem creates and removes the files of a recording.
type FileSystem interface {
	Create(name string) (File, error)
	Remove(name string) error
}

// Deps are the clock and the file system of a recorder; zero fields use the
// real ones.
type Deps struct {
	Clock Clock
	FS    FileSystem
}

func (d Deps) withDefaults() Deps {
	if d.Clock == nil {
		d.Clock = systemClock{}
	}
	if d.FS == nil {
		d.FS = osFS{}
	}
	return d
}

type systemClock struct{}

func (systemClock) Now() time.Time                         { return time.Now() }
func (systemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

type osFS struct{}

func (osFS) Create(name string) (File, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (osFS) Remove(name string) error { return os.Remove(name) }
