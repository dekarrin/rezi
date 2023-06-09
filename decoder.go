package rezi

import "encoding"

// Decoder decodes the primitive types bool, int, and string, as well as a type
// that is specified by its type parameter (usually an interface of some
// XMarshaler type, such as BinaryUnmarshaler).
type Decoder[E any] interface {

	// DecodeBool decodes a bool value at the current position within the buffer
	// of the Decoder and advances the current position past the read bytes.
	DecodeBool() (bool, error)

	// DecodeInt decodes an int value at the current position within the buffer
	// of the Decoder and advances the current position past the read bytes.
	DecodeInt() (int, error)

	// DecodeString decodes a string value at the current position within the
	// buffer of the Decoder and advances the current position past the read
	// bytes.
	DecodeString() (string, error)

	// Decode decodes a value at the current position within the buffer of the
	// Decoder and advances the current position past the read bytes. Unlike the
	// other functions, instead of returning the value this one will set the
	// value of the given item.
	Decode(o E) error
}

// simpleBinaryEncoder encodes values as binary. Create with NewBinaryDecoder,
// don't use directly.
type simpleBinaryDecoder struct {
	b   []byte
	cur int
}

func (sbe *simpleBinaryDecoder) DecodeBool() (bool, error) {
	val, n, err := DecBool(sbe.b[sbe.cur:])
	if err != nil {
		return val, err
	}
	sbe.cur += n
	return val, nil
}

func (sbe *simpleBinaryDecoder) DecodeInt() (int, error) {
	val, n, err := DecInt(sbe.b[sbe.cur:])
	if err != nil {
		return val, err
	}
	sbe.cur += n
	return val, nil
}

func (sbe *simpleBinaryDecoder) DecodeString() (string, error) {
	val, n, err := DecString(sbe.b[sbe.cur:])
	if err != nil {
		return val, err
	}
	sbe.cur += n
	return val, nil
}

func (sbe *simpleBinaryDecoder) Decode(o encoding.BinaryUnmarshaler) error {
	n, err := DecBinary(sbe.b[sbe.cur:], o)
	if err != nil {
		return err
	}
	sbe.cur += n
	return nil
}
