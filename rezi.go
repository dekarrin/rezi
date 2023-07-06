// Package rezi provides the ability to encode and decode data in Rarefied
// Encoding (Compressible) Interchange format. It allows types that implement
// encoding.BinaryUnmarshaler and encoding.BinaryMarshaler to be easily read
// from and written to encoded bytes.
//
// It is generally used in a fashion similar to the json package. The [Enc]
// function is to encode any supported type to bytes.
//
//	import "github.com/dekarrin/rezi"
//
//	func main() {
//		specialNumber := 413
//		person := "TEREZI"
//
//		var numData []byte
//		var personData []byte
//		var err error
//
//		numData, err = rezi.Enc(specialNumber)
//		if err != nil {
//			panic(err.Error())
//		}
//
//		personData, err = rezi.Enc(person)
//		if err != nil {
//			panic(err.Error())
//		}
//	}
//
// Data from multiple calls to Enc() can be combined into a single block of data
// by appending them together.
//
//	var allData []byte
//	allData = append(allData, numData...)
//	allData = append(allData, personData...)
//
// The [Dec] function is used to decode data from REZI bytes:
//
//	var readNumber int
//	var readPerson string
//
//	var n int
//	var err error
//
//	n, err := rezi.Dec(allData, &readNumber)
//	if err != nil {
//		panic(err.Error())
//	}
//	allData = allData[n:]
//
//	n, err := rezi.Dec(allData, &readPerson)
//	if err != nil {
//		panic(err.Error())
//	}
//	allData = allData[n:]
//
// # Supported Data Types
//
// REZI supports several built-in Go types: int and all of its unsigned and
// specific-bitsize varieties, string, bool, and any type that implements
// encoding.BinaryMarshaler (for encoding) or whose pointer type implements
// encoding.BinaryUnmarshaler (for decoding).
//
// Floating point types and complex types are not supported at this time,
// although they may be added in a future release.
//
// Slices and maps are supported with some simulations. Slices must contain only
// other supported types (or pointers to them). Maps have the same restrictions
// on their values, but only maps with a key type of string, int (or any of its
// unsigned and specific-bitsize varieties), or bool are supported.
//
// Pointers to any supported type are also accepted, including to other pointer
// types with any number of indirections. The REZI format encodes information on
// how many levels of indirection are valid, though of course note that it does
// not have any concept of two different pointer variables pointing to the same
// data.
//
// # Data Format
//
// REZI uses a binary format for all supported types. Other than bool, which is
// encoded as simply one of two byte values, an encoded value will start with
// one or more "info" bytes that gives metadata on the value itself. This is
// typically the length of the full value, but may include additional
// information such as whether the encoded value is in fact a nil pointer.
//
// Note that the info byte does not give information on the type of the encoded
// value, besides whether it is nil (and still, the type of the nil is not
// encoded). Types of the encoded values are inferred by the pointer receiver
// that is passed to Dec(). If a pointer to an int is passed to it, the bytes
// will be interpreted as an encoded int; likewise, if a pointer to a string is
// passed to it, the bytes will be interpreted as an encoded string.
//
//	The INFO Byte
//
//	SXNILLLL
//
// The info byte has information coded into its bits represented as SXNILLLL,
// where each letter stands for a particular bit, from most signficant to the
// least significant.
//
// The bit labeled "S" is the sign bit; when high (1), it indicates that the
// following integer value is negative.
//
// The "X" bit is the extension flag, and indicates that the next byte is a
// second info byte with additional information. At this time that bit is
// unused, but is planned to be used in future releases.
//
// The "N" bit is the explicit nil flag, and when set it indicates that the
// value is a nil and that there are no following bytes which make up its
// encoding, with the exception of any indirection amount indicators.
//
// The "I" bit is the indirection bit, and if set, indicates that the following
// bytes encode the number of additional indirections of the pointer beyond the
// initial indirection at which the nil occurs; for instance, a nil *int value
// is encoded as simply the info byte 0b00100000, but a non nil **int that
// points at a nil *int would be encoded with one level of additional
// indirection and the info byte's I bit would be set.
//
// The "L" bits make up the length of the value. Together, they are a 4-bit
// unsigned integer that indicates how many of the following bytes are part of
// the encoded value. If the I bit is set on the info byte, the L bits give the
// number of bytes that make up the indirection level rather than the actual
// value.
//
//	Bool Values
//
//	[ VALUE ]
//	 1 byte
//
// Boolean values are encoded in REZI as the byte value 0x01 for true, or 0x00
// for false. Bool is the only type whose encoded value does not begin with an
// info byte, although a pointer-to-bool may be encoded with an info byte if it
// is nil.
//
//	Integer Values
//
//	[ INFO ] [ INT VALUE ]
//	 1 byte    0-8 bytes
//
// Integer values begin with the info byte. Assuming that it is not nil, the 4
// L bits of the info byte give the number of bytes that are in the value
// itself, and the S bit represents whether the value is negative.
//
// The INT VALUE portion of the integer includes all bytes necessary to rebuild
// the integer value. It is created by first taking the integer's value expanded
// to 64-bits, and then removing all leading insignificant bytes (those with a
// value of 0x00 for positive integers, or those with a value of 0xff for
// negative integers). These bytes are then used as the INT VALUE.
//
// As a result of the above encoding, certain integer values can be encoded with
// no bytes in INT VALUE at all; the 64-bit representation for 0 all 0x00's, and
// therefore has no significant bytes. Likewise, the 64-bit representation for
// -1 using two's complement representation is all 0xff's.
//
// All Go integer types are encoded in the same way. This includes int, int8,
// int16, int32, int64, uint, uint8, uint16, uint32, and uint64. The specific
// interpretation into a value is handled at decoding time by infering the type
// from the pointer passed to Enc.
//
//	String Values
//
//	{   CODEPOINT COUNT  } [ CODEPOINTS ]
//	[ INFO ] [ INT VALUE ] [ CODEPOINTS ]
//
// String values are encoded as a count of codepoints (which is itself encoded
// as an integer value), followed by the unicode codepoints that make up the
// string in UTF-8.
//
// By starting with an integer, strings begin with an info byte automatically.
//
//	encoding.BinaryMarshaler Values
//
//	{     BYTE COUNT     } [ MARSHALED BYTES ]
//	[ INFO ] [ INT VALUE ] [ MARSHALED BYTES ]
//
// Any type that implements [encoding.BinaryMarshaler] is encoded by taking the
// result of calling its MarshalBinary() method and prepending it with an
// integer value giving the number of bytes in it.
//
//	Slice Values
//
//	{     BYTE COUNT     } [ ITEM 1 ] ... [ ITEM N ]
//	[ INFO ] [ INT VALUE ] [ ITEM 1 ] ... [ ITEM N ]
//
// Slices are encoded. WIP.
//
//	Map Values
//
//	{     BYTE COUNT     } [ KEY 1 ] [ VALUE 1 ] ... [ KEY N ] [ VALUE N ]
//	[ INFO ] [ INT VALUE ] [ KEY 1 ] [ VALUE 1 ] ... [ KEY N ] [ VALUE N ]
//
// Map values are encoded. WIP.
//
//	Nil Values
//
//	[ INFO ] [ INT VALUE ]
//
// Nil values are all encoded as the same way regardless of their pointed-to
// type. WIP.
package rezi

import (
	"encoding"
	"fmt"
	"reflect"
)

type (
	tLen      = int
	tNilLevel = int

	decFunc[E any] func([]byte) (E, int, error)
	encFunc[E any] func(E) ([]byte, error)
)

func nilErrEncoder[E any](fn func(E) []byte) encFunc[E] {
	return func(e E) ([]byte, error) {
		return fn(e), nil
	}
}

// NewBinaryEncoder creates an Encoder that can encode to bytes and uses an
// object's MarshalBinary method to encode non-trivial types.
//
// Deprecated: Use [NewWriter] instead.
func NewBinaryEncoder() Encoder[encoding.BinaryMarshaler] {
	enc := &simpleBinaryEncoder{}
	return enc
}

// NewBinaryDecoder creates a Decoder that can decode bytes and uses an object's
// UnmarshalBinary method to decode non-trivial types.
//
// Deprecated: Use [NewReader] instead.
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

// MustEnc is identitical to Enc, but panics if an error would be returned.
func MustEnc(v interface{}) []byte {
	enc, err := Enc(v)
	if err != nil {
		panic(err.Error())
	}
	return enc
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
// TODO: during docs, add note on returned error types and using errors.Is.
func Enc(v interface{}) ([]byte, error) {
	info, err := canEncode(v)
	if err != nil {
		panic(err.Error())
	}

	if info.Primitive() {
		return encCheckedPrim(v, info)
	} else if info.Main == tNil {
		return encNil(0), nil
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
		return decCheckedMap(data, v, info)
	} else if info.Main == tSlice {
		return decCheckedSlice(data, v, info)
	} else {
		panic("no possible decoding")
	}
}

func encWithNilCheck[E any](value interface{}, ti typeInfo, encFn encFunc[E], convFn func(reflect.Value) E) ([]byte, error) {
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
			return encNil(nilLevel), nil
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
		} else {
			zeroVal := reflect.Zero(assignTarget.Elem().Type())
			assignTarget.Elem().Set(zeroVal)
		}
	}

	return decoded, n, nil
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
