package device

import (
	"bytes"
	"fmt"
)

// CommandError is the error of a command that started and then failed. It
// keeps the output, which the error text alone would lose: adb, for example,
// says why an installation failed on stderr only.
type CommandError struct {
	Name   string
	Stdout []byte
	Stderr []byte
	Err    error
}

// Error is "run <name>: <err>", followed by the trimmed stderr when there is
// some.
func (e *CommandError) Error() string {
	msg := fmt.Sprintf("run %s: %v", e.Name, e.Err)
	if stderr := bytes.TrimSpace(e.Stderr); len(stderr) > 0 {
		msg += ": " + string(stderr)
	}
	return msg
}

// Unwrap returns the exit error of the command.
func (e *CommandError) Unwrap() error { return e.Err }
