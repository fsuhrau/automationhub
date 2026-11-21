// Package apperr classifies errors by what went wrong, not by how a
// transport reports it. Services return errors of a Kind; the HTTP layer
// (endpoints/httpx) turns the kind into a status code, so handlers never
// choose one.
//
// Kinds are the constants KindNotFound, KindInvalid and so on. The
// constructors carry the plain name: NotFound, Invalid, Conflict, Busy,
// PreconditionRequired, Unsupported, Timeout, Forbidden and Internal.
package apperr

import (
	"errors"
	"fmt"
)

// Kind says what went wrong.
type Kind int

// KindNone is the kind of a nil error. Every non-nil error has a real kind;
// the ones that are not an *Error are KindInternal.
const (
	KindNone Kind = iota
	KindNotFound
	KindInvalid
	KindConflict
	KindBusy
	KindPreconditionRequired
	KindUnsupported
	KindTimeout
	KindForbidden
	KindInternal
)

// Error is an error with a Kind. It keeps the error it wraps for errors.Is
// and errors.As.
type Error struct {
	kind Kind
	err  error
}

func (e *Error) Error() string { return e.err.Error() }

func (e *Error) Unwrap() error { return e.err }

// Kind returns the kind of the error.
func (e *Error) Kind() Kind { return e.kind }

func newf(kind Kind, format string, args []any) error {
	return &Error{kind: kind, err: fmt.Errorf(format, args...)}
}

// NotFound is an entity that doesn't exist or isn't visible to the caller.
func NotFound(format string, args ...any) error { return newf(KindNotFound, format, args) }

// Invalid is a request the caller has to correct.
func Invalid(format string, args ...any) error { return newf(KindInvalid, format, args) }

// Conflict is a request that clashes with the current state.
func Conflict(format string, args ...any) error { return newf(KindConflict, format, args) }

// Busy is a resource that is in use by someone else.
func Busy(format string, args ...any) error { return newf(KindBusy, format, args) }

// PreconditionRequired is a request that needs something first, such as a lease.
func PreconditionRequired(format string, args ...any) error {
	return newf(KindPreconditionRequired, format, args)
}

// Unsupported is a capability the target doesn't have.
func Unsupported(format string, args ...any) error { return newf(KindUnsupported, format, args) }

// Timeout is an operation that took too long.
func Timeout(format string, args ...any) error { return newf(KindTimeout, format, args) }

// Forbidden is a request the caller may not make.
func Forbidden(format string, args ...any) error { return newf(KindForbidden, format, args) }

// Internal is a failure of the hub itself.
func Internal(format string, args ...any) error { return newf(KindInternal, format, args) }

// Wrap gives err the kind and prefixes its message with msg ("msg: err").
// The cause stays reachable through errors.Is and errors.As. Wrap of a nil
// error is nil.
func Wrap(kind Kind, err error, msg string) error {
	if err == nil {
		return nil
	}
	return &Error{kind: kind, err: fmt.Errorf("%s: %w", msg, err)}
}

// KindOf returns the kind of the outermost *Error in the chain of err:
// KindNone for nil and KindInternal for an error that has none.
func KindOf(err error) Kind {
	if err == nil {
		return KindNone
	}
	var e *Error
	if errors.As(err, &e) {
		return e.kind
	}
	return KindInternal
}
