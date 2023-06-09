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
	"reflect"
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
