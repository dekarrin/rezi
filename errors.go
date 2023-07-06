package rezi

import (
	"errors"
)

var (
	Error              = errors.New("a problem related to the binary REZI format has occurred")
	ErrMarshalBinary   = errors.New("MarshalBinary() returned an error")
	ErrUnmarshalBinary = errors.New("UnmarshalBinary() returned an error")
	ErrInvalidType     = errors.New("data is not the correct type")
	ErrMalformedData   = errors.New("data cannot be interpretered")
)

// reziError is the type of error returned by Enc when there is an issue
// with encoding the provided value, due to the value type being unsupported,
// the value being an implementor of encoding.BinaryMarshaler and its
// MarshalBinary function returning an error, or some other reason.
type reziError struct {
	msg   string
	cause []error
}

// Error returns the message defined for the EncodingError.
func (e reziError) Error() string {
	if e.msg == "" {
		if e.cause != nil {
			return e.cause[0].Error()
		}
		return Error.Error()
	}

	return e.msg
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
	// TODO: from go docs re errors: "An Is method should only shallowly compare
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
