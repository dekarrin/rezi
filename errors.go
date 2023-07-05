package rezi

import (
	"errors"
	"fmt"
)

var (
	ErrMarshalBinary = errors.New("MarshalBinary() returned an error")
)

// EncodingError is the type of error returned by Enc when there is an issue
// with encoding the provided value, due to the value type being unsupported,
// the value being an implementor of encoding.BinaryMarshaler and its
// MarshalBinary function returning an error, or some other reason.
type EncodingError struct {
	msg   string
	cause []error
}

// Error returns the message defined for the EncodingError.
func (e EncodingError) Error() string {
	if e.msg == "" {
		if e.cause != nil {
			return e.cause[0].Error()
		}
		return "encoding failed"
	}

	return e.msg
}

// Unwrap returns the causes of Error. The return value will be nil if no causes
// were defined for it.
//
// This function is for interaction with the errors API. It will only be used in
// Go version 1.20 and later; 1.19 will default to use of Error.Is when calling
// errors.Is on the Error.
func (e EncodingError) Unwrap() []error {
	if len(e.cause) > 0 {
		return e.cause
	}
	return nil
}

// Is returns whether Error either Is itself the given target error, or one of
// its causes is.
//
// This function is for interaction with the errors API.
func (e EncodingError) Is(target error) bool {
	// is the target error itself?
	if errTarget, ok := target.(EncodingError); ok {
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
		if sErr, ok := e.cause[i].(EncodingError); ok {
			if sErr.Is(target) {
				return true
			}
		} else if e.cause[i] == target {
			return true
		}
	}
	return false
}

// wrapped must never be nil. subType may be nil.
func wrapEncErr(wrapped error, subType error) EncodingError {
	if subType != nil {
		actualMsg := ""
		if wrapped.Error() != "" {
			actualMsg = fmt.Sprintf("%s: %s", subType, wrapped)
		} else {
			actualMsg = subType.Error()
		}

		return EncodingError{
			msg:   actualMsg,
			cause: []error{subType, wrapped},
		}
	}

	return EncodingError{
		cause: []error{wrapped},
	}
}
