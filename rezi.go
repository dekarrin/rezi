// Package rezi, the Rarefied Encoding (Compressible) Interpreter, contains
// functions for marshaling data as binary bytes using a simple encoding scheme.
// Encoding output length is variable based on data size. For bools, this is
// accomplished by having constant length of encoded output. Encoded ints are
// represented as one or more bytes depending on their value, with values closer
// to 0 taking up fewer bytes. Encoded strings are represented by an encoded int
// that gives the unicode codepoint count followed by the contents of the
// string encoded as UTF-8. For other types, an encoded integer is placed at the
// head of the bytes which indicates how many bytes that follow are part of the
// value being decoded.
package rezi

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
)

type (
	tLen      = int
	tNilLevel = int

	decFunc[E any] func([]byte) (E, int, error)
	encFunc[E any] func(E) []byte
)

var (
	ErrInvalidType   = errors.New("data is not the correct type")
	ErrMalformedData = errors.New("data cannot be interpretered")
)

// NewBinaryEncoder creates an Encoder that can encode to bytes and uses an
// object's MarshalBinary method to encode non-trivial types.
func NewBinaryEncoder() Encoder[encoding.BinaryMarshaler] {
	enc := &simpleBinaryEncoder{}
	return enc
}

// NewBinaryDecoder creates a Decoder that can decode bytes and uses an object's
// UnmarshalBinary method to decode non-trivial types.
func NewBinaryDecoder() Decoder[encoding.BinaryUnmarshaler] {
	dec := &simpleBinaryDecoder{}
	return dec
}

// get zero value for a type if not a pointer, or a pointer to a valid 0 value
// if a pointer type.
func initType[E any]() E {
	var v E

	vType := reflect.TypeOf(v)

	if vType == nil {
		panic("cannot initialize an interface value; must decode to a concrete type")
	}

	if vType.Kind() == reflect.Pointer {
		pointedTo := vType.Elem()
		pointedVal := reflect.New(pointedTo)
		pointedIFace := pointedVal.Interface()
		var ok bool
		v, ok = pointedIFace.(E)
		if !ok {
			// should never happen
			panic("could not convert returned type")
		}
	}

	return v
}

// Enc encodes the value as rezi-format bytes. The type of the value is
// examined to determine how to encode it. No type information is included in
// the returned bytes so it is up to the caller to keep track of it.
//
// The value must be one of the supported REZI types. The supported types are:
// string, bool, uint and its sized variants, int and its sized variants, and
// any implementor of encoding.BinaryMarshaler. Map and slice types are also
// supported, as long as their contents are REZI-supported.
//
// This function will panic if the type is not rezi supported.
func Enc(v interface{}) []byte {
	info, err := canEncode(v)
	if err != nil {
		panic(err.Error())
	}

	if info.Primitive() {
		return encCheckedPrim(v, info)
	} else if info.Main == tNil {
		return encNil(0)
	} else if info.Main == tMap {
		return encCheckedMap(v, info)
	} else if info.Main == tSlice {
		return encCheckedSlice(v, info)
	} else {
		panic("no possible encoding")
	}
}

// Dec decodes a value as rezi-format bytes. The argument v must be a pointer to
// a supported type (or directly implement binary.BinaryMarshaler).
func Dec(data []byte, v interface{}) (int, error) {
	info, err := canDecode(v)
	if err != nil {
		panic(err.Error())
	}

	if info.Primitive() {
		return decCheckedPrim(data, v, info)
	} else if info.Main == tMap {
		return decMap(data, v, info)
	} else if info.Main == tSlice {
		return decCheckedSlice(data, v, info)
	} else {
		panic("no possible decoding")
	}
}

func encWithNilCheck[E any](value interface{}, ti typeInfo, encFn encFunc[E], convFn func(reflect.Value) E) []byte {
	if ti.Indir > 0 {
		// we cannot directly encode, we must get at the reel value.
		encodeTarget := reflect.ValueOf(value)
		// encodeTarget is a *THING but we want a THING

		nilLevel := -1
		for i := 0; i < ti.Indir; i++ {
			if encodeTarget.IsNil() {
				// so if it were a *string we deal w, nil level can only be 0.
				// if it were a **string we deal w, nil level can be 0 or 1.
				// *string -> indir is 1 - nl 0
				// **string -> indir is 2 - nl 0-1
				nilLevel = i
				break
			}
			encodeTarget = encodeTarget.Elem()
		}
		if nilLevel > -1 {
			return encNil(nilLevel)
		}
		return encFn(convFn(encodeTarget))
	} else {
		return encFn(value.(E))
	}
}

// if ti.Indir > 0, this will assign to the interface at the appropriate
// indirection level. If ti.Indir == 0, this will not assign. Callers should use
// that check to determine if it is safe to do their own assignment of the
// decoded value this function returns.
func decWithNilCheck[E any](data []byte, v interface{}, ti typeInfo, decFn decFunc[E]) (decoded E, n int, err error) {
	var isNil bool
	var nilLevel tNilLevel

	if ti.Indir > 0 {
		isNil, nilLevel, _, n, err = decNilable[E](nil, data)
		if err != nil {
			return decoded, n, fmt.Errorf("check nil value: %w", err)
		}
	}

	if !isNil {
		decoded, n, err = decFn(data)
		if err != nil {
			return decoded, n, err
		}
		nilLevel = ti.Indir
	}

	if ti.Indir > 0 {
		// the user has passed in a ptr-ptr-to. We cannot directly assign.
		assignTarget := reflect.ValueOf(v)
		// assignTarget is a **string but we want a *string

		for i := 0; i < ti.Indir && i < nilLevel; i++ {
			// *double indirection ALL THE WAY~*
			// *acrosssss the sky*
			// *what does it mean*

			// **string     // *string  // string
			newTarget := reflect.New(assignTarget.Type().Elem().Elem())
			assignTarget.Elem().Set(newTarget)
			assignTarget = newTarget
		}

		if !isNil {
			assignTarget.Elem().Set(reflect.ValueOf(decoded))
		}
	}

	return decoded, n, nil
}
func makeFuncDecToWrappedReceiver(wrapped interface{}, ti typeInfo, assertFn func(reflect.Type) bool, decToUnwrappedFn func([]byte, interface{}) (int, error)) decFunc[interface{}] {
	return func(data []byte) (interface{}, int, error) {
		// v is *(...*)T, ret-val of decFn (this lambda) is T.
		receiverType := reflect.TypeOf(wrapped)

		if receiverType.Kind() == reflect.Pointer { // future-proofing - binary unmarshaler might come in as a T
			// for every * in the (...*) part of *(...*)T up until the
			// implementor/slice-ptr, do a deref.
			for i := 0; i < ti.Indir; i++ {
				receiverType = receiverType.Elem()
			}
		}

		// receiverType should now be the exact type which needs to be sent to
		// the reel decode func.
		if !assertFn(receiverType) {
			// should never happen, assuming ti.Indir is valid.
			panic("unwrapped receiver is not compatible with encoded value")
		}

		var receiverValue reflect.Value
		if receiverType.Kind() == reflect.Pointer {
			// receiverType is *T
			receiverValue = reflect.New(receiverType.Elem())
		} else {
			// receiverType is itself T (future-proofing)
			receiverValue = reflect.Zero(receiverType)
		}

		var decoded interface{}

		receiver := receiverValue.Interface()
		decConsumed, decErr := decToUnwrappedFn(data, receiver)

		if decErr != nil {
			return nil, decConsumed, decErr
		}

		if receiverType.Kind() == reflect.Pointer {
			decoded = reflect.ValueOf(receiver).Elem().Interface()
		} else {
			decoded = receiver
		}

		return decoded, decConsumed, decErr
	}
}

func fn_DecToWrappedReceiver(wrapped interface{}, ti typeInfo, assertFn func(reflect.Type) bool, decToUnwrappedFn func([]byte, interface{}) (int, error)) decFunc[interface{}] {
	return func(data []byte) (interface{}, int, error) {
		// v is *(...*)T, ret-val of decFn (this lambda) is T.
		receiverType := reflect.TypeOf(wrapped)

		if receiverType.Kind() == reflect.Pointer { // future-proofing - binary unmarshaler might come in as a T
			// for every * in the (...*) part of *(...*)T up until the
			// implementor/slice-ptr, do a deref.
			for i := 0; i < ti.Indir; i++ {
				receiverType = receiverType.Elem()
			}
		}

		// receiverType should now be the exact type which needs to be sent to
		// the reel decode func.
		if !assertFn(receiverType) {
			// should never happen, assuming ti.Indir is valid.
			panic("unwrapped receiver is not compatible with encoded value")
		}

		var receiverValue reflect.Value
		if receiverType.Kind() == reflect.Pointer {
			// receiverType is *T
			receiverValue = reflect.New(receiverType.Elem())
		} else {
			// receiverType is itself T (future-proofing)
			receiverValue = reflect.Zero(receiverType)
		}

		var decoded interface{}

		receiver := receiverValue.Interface()
		decConsumed, decErr := decToUnwrappedFn(data, receiver)

		if decErr != nil {
			return nil, decConsumed, decErr
		}

		if receiverType.Kind() == reflect.Pointer {
			decoded = reflect.ValueOf(receiver).Elem().Interface()
		} else {
			decoded = receiver
		}

		return decoded, decConsumed, decErr
	}
}
