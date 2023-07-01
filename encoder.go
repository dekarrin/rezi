package rezi

import (
	"bytes"
	"encoding"
)

// Encoder encodes the primitive types bool, int, and string, as well as a type
// that is specified by its type parameter (usually an interface of some
// XMarshaler type, such as BinaryMarshaler).
type Encoder[E any] interface {
	EncodeBool(b bool)
	EncodeInt(i int)
	EncodeString(s string)
	Encode(o E)

	// Bytes returns all encoded values as sequential bytes.
	Bytes() []byte
}

// SimpleBinaryEncoder encodes values as binary.
type simpleBinaryEncoder struct {
	b *bytes.Buffer
}

func (sbe *simpleBinaryEncoder) checkInit() {
	if sbe.b == nil {
		sbe.b = &bytes.Buffer{}
	}
}

func (sbe *simpleBinaryEncoder) EncodeBool(b bool) {
	sbe.checkInit()
	val := encBool(b)
	sbe.b.Write(val)
}

func (sbe *simpleBinaryEncoder) EncodeInt(i int) {
	sbe.checkInit()
	val := encInt(i)
	sbe.b.Write(val)
}

func (sbe *simpleBinaryEncoder) EncodeString(s string) {
	sbe.checkInit()
	val := encString(s)
	sbe.b.Write(val)
}

func (sbe *simpleBinaryEncoder) Encode(o encoding.BinaryMarshaler) {
	sbe.checkInit()
	val := encBinary(o)
	sbe.b.Write(val)
}

func (sbe *simpleBinaryEncoder) Bytes() []byte {
	if sbe.b == nil {
		return nil
	}

	return sbe.b.Bytes()
}
