package rezi

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// Error is a general error returned from encoding and decoding functions.
	// All non-nil errors returned from this package will return true for the
	// expression errors.Is(err, Error).
	Error = errors.New("a problem related to the binary REZI format has occurred")

	// ErrMarshalBinary indicates that calling a MarshalBinary method on a type
	// that was being encoded returned a non-nil error. Any error returned from
	// this package that was caused by this will return true for the expression
	// errors.Is(err, ErrMarshalBinary).
	ErrMarshalBinary = errors.New("MarshalBinary() returned an error")

	// ErrUnmarshalBinary indicates that calling an UnmarshalBinary method on a
	// type that was being decoded returned a non-nil error. Any error returned
	// from this package that was caused by this will return true for the
	// expression errors.Is(err, ErrUnmarshalBinary).
	ErrUnmarshalBinary = errors.New("UnmarshalBinary() returned an error")

	// ErrInvalidType indicates that the value to be encoded or decoded to is
	// not of a valid type. Any error returned from this package that was caused
	// by this will return true for the expression
	// errors.Is(err, ErrInvalidType).
	ErrInvalidType = errors.New("data is not the correct type")

	// ErrMalformedData indicates that there is a problem with the data being
	// decoded. Any error returned from this package that was caused by this
	// will return true for the expression errors.Is(err, ErrMalformedData).
	ErrMalformedData = errors.New("data cannot be interpretered")
)

// reziError is the concrete type of errors returned by all exported functions.
// It is intended to be used and compared against error types with the errors.Is
// API.
//
// Generally should not be created by hand. create one by calling errorf(), add
// wrapped errors with reziError.wrap().
type reziError struct {
	msg         string
	cause       []error
	offsetValid bool
	offset      int
}

// errorf works like fmt.Errorf, except a %w is not needed to wrap an err; any
// error type in the a list will be in the wrapped errors, and %s or %v can be
// used to get their error value.
func errorf(msgFmt string, a ...interface{}) reziError {
	if strings.Contains(strings.ReplaceAll(msgFmt, "%%", "--"), "%w") {
		panic("don't use %w in errorf; use %s/%v and give the error in the args")
	}

	e := reziError{
		msg: fmt.Sprintf(msgFmt, a...),
	}

	for i := range a {
		if wrappedErr, ok := a[i].(error); ok {
			e = e.wrap(wrappedErr)
		}
	}

	return e
}

// decErrorf is same as errorf but requires an offset
func decErrorf(offset int, msgFmt string, a ...interface{}) reziError {
	e := errorf(msgFmt, a...)
	e.offsetValid = true
	e.offset = offset
	return e
}

func (e reziError) totalOffset() (offset int, ok bool) {
	if !e.offsetValid {
		return 0, false
	}

	// are we wrapping a rezi error? find one and use it if so
	for _, wrapped := range e.cause {
		if reziErr, ok := wrapped.(reziError); ok {
			if wrappedOff, wrappedOk := reziErr.totalOffset(); wrappedOk {
				return e.offset + wrappedOff, true
			}
		}
	}

	// if we got here, we aren't wrapping a rezi error that has an offset for us
	// to add. just return own
	return e.offset, true
}

func (e reziError) wrap(wrapped ...error) reziError {
	e.cause = append(e.cause, wrapped...)
	for _, w := range wrapped {
		if reziErr, ok := w.(reziError); ok {
			if reziErr.offsetValid {
				e.offsetValid = true
			}
		}
	}
	return e
}

// Error returns the message defined for the EncodingError.
func (e reziError) Error() string {
	// lead with offset if provided
	prefix := ""
	if offset, ok := e.totalOffset(); ok {
		asHex := fmt.Sprintf("%x", offset)
		if len(asHex)%2 != 0 {
			asHex = "0" + asHex
		}
		prefix = fmt.Sprintf("byte offset %d (0x%s): ", offset, asHex)
	}

	if e.msg == "" {
		if e.cause != nil {
			return prefix + e.cause[0].Error()
		}
		return prefix + Error.Error()
	}

	return prefix + e.msg
}

// Unwrap returns the causes of Error. The return value will be nil if no causes
// were defined for it.
//
// This function is for interaction with the errors API. It will only be used in
// Go version 1.20 and later; 1.19 will default to use of Error.Is when calling
// errors.Is on the Error.
func (e reziError) Unwrap() []error {
	wrapped := []error{Error}

	if len(e.cause) > 0 {
		wrapped = append(wrapped, e.cause...)
	}

	return wrapped
}

// Is returns whether Error either Is itself the given target error, or one of
// its causes is.
//
// This function is for interaction with the errors API.
func (e reziError) Is(target error) bool {
	// a reziError will always return true for the Error type.
	if errors.Is(target, Error) {
		return true
	}

	// is the target error itself?
	if errTarget, ok := target.(reziError); ok {
		if e.msg == errTarget.msg {
			if len(e.cause) == len(errTarget.cause) {
				allCausesEqual := true
				for i := range e.cause {
					if e.cause[i] != errTarget.cause[i] {
						allCausesEqual = false
						break
					}
				}
				if allCausesEqual {
					return true
				}
			}
		}
	}

	// otherwise, check if any cause equals target
	// from go docs re errors: "An Is method should only shallowly compare
	// err and the target and not call Unwrap on either.". Okay. But the thing
	// is, Go 1.19 does not support wrapping multiple errors so we have opted to
	// do things this way. In future, let's use build tags and separate files to
	// split based on go version and ensure that we have unit tests for each.
	for i := range e.cause {

		// we must check if any are of type Error, because if they are, we need
		// to run the normal Is.
		if sErr, ok := e.cause[i].(reziError); ok {
			if sErr.Is(target) {
				return true
			}
		} else if e.cause[i] == target {
			return true
		}
	}
	return false
}
