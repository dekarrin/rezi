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
		return encPrim(v, info)
	} else if info.Main == tNil {
		return encNil(0)
	} else if info.Main == tMap {
		return encMap(v, info)
	} else if info.Main == tSlice {
		return encSlice(v, info)
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
		return decPrim(data, v, info)
	} else if info.Main == tMap {
		return decMap(data, v, info)
	} else if info.Main == tSlice {
		return decSlice(data, v, info)
	} else {
		panic("no possible decoding")
	}
}
