// Package rezi provides the ability to encode and decode data in Rarefied
// Encoding (Compressible) Interchange format. It allows types that implement
// encoding.BinaryUnmarshaler and encoding.BinaryMarshaler to be easily read
// from and written to byte slices. It has an interface similar to the json
// package, where one function is used to encode all supported types, and
// another function receives bytes and a receiver for decoded data and infers
// how to decode the bytes based on the receiver.
//
// The [Enc] function is to encode any supported type to REZI bytes:
//
//	import "github.com/dekarrin/rezi/v2"
//
//	func main() {
//		specialNumber := 413
//		name := "TEREZI"
//
//		var numData []byte
//		var nameData []byte
//		var err error
//
//		numData, err = rezi.Enc(specialNumber)
//		if err != nil {
//			panic(err.Error())
//		}
//
//		nameData, err = rezi.Enc(name)
//		if err != nil {
//			panic(err.Error())
//		}
//	}
//
// Data from multiple calls to Enc() can be combined into a single block of data
// by appending them together:
//
//	var allData []byte
//	allData = append(allData, numData...)
//	allData = append(allData, nameData...)
//
// The [Dec] function is used to decode data from REZI bytes:
//
//	var readNumber int
//	var readName string
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
//	n, err := rezi.Dec(allData, &readName)
//	if err != nil {
//		panic(err.Error())
//	}
//	allData = allData[n:]
//
// # Error Checking
//
// Errors in REZI have specific types that can be checked in order to determine
// the cause of an error. These errors conform to the [errors] interface and
// must be checked by using [errors.Is].
//
// As mentioned in that library's documentation, errors should not be checked
// with simple equality checks. REZI enforces this fully; non-nil errors that
// are checked with `==` will never return true.
//
// That is
//
//	if err == rezi.Error
//
// is not only the non-preferred way of checking an error, but will always
// return false. Instead, do:
//
//	if errors.Is(err, rezi.Error)
//
// There are several error types defined for checking non-nil errors. [Error] is
// the type that all non-nil errors from REZI will match. It may be caused by
// some other underlying error; again, use errors.Is to check this, even if a
// non-rezi error is being checked. For instance, to check if an error was
// caused due to the supplied bytes being shorter than expected, use
// errors.Is(err, io.ErrUnexpectedEOF).
//
// See the individual functions for a list of error types that non-nil returned
// errors may be checked against.
//
// # Supported Data Types
//
// REZI supports several built-in basic Go types: int (as well as all of its
// unsigned and specific-bitsize varieties), string, bool, and any type that
// implements encoding.BinaryMarshaler (for encoding) or whose pointer type
// implements encoding.BinaryUnmarshaler (for decoding).
//
// Floating point types and complex types are not supported at this time,
// although they may be added in a future release.
//
// Slices and maps are supported with some stipulations. Slices must contain
// only other supported types (or pointers to them). Maps have the same
// restrictions on their values, but only maps with a key type of string, int
// (or any of its unsigned and specific-bitsize varieties), or bool are
// supported.
//
// Pointers to any supported type are also accepted, including to other pointer
// types with any number of indirections. The REZI format encodes information on
// how many levels of indirection are valid, though of course note that it does
// not have any concept of two different pointer variables pointing to the same
// data.
//
// # Binary Data Format
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
//	Layout:
//
//	SXNILLLL
//	|      |
//	MSB  LSB
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
//	Layout:
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
//	Layout:
//
//	[ INFO ] [ INT VALUE ]
//	 1 byte    0..8 bytes
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
//	Layout:
//
//	[ INFO ] [ INT VALUE ] [ CODEPOINT 1 ] ... [ CODEPOINT N ]
//	<---CODEPOINT COUNT--> <------------CODEPOINTS----------->
//	      1..9 bytes               COUNT..COUNT*4 bytes
//
// String values are encoded as a count of codepoints (which is itself encoded
// as an integer value), followed by the Unicode codepoints that make up the
// string encoded with UTF-8. Due to the count being of Unicode codepoints
// rather than bytes, the actual number of bytes in an encoded string will be
// between the minimum and maximum number of bytes needed to encode a codepoint
// in UTF-8, multiplied by the number of codepoints.
//
//	encoding.BinaryMarshaler Values
//
//	Layout:
//
//	[ INFO ] [ INT VALUE ] [ MARSHALED BYTES ]
//	<-------COUNT--------> <-MARSHALED BYTES->
//	      1..9 bytes           COUNT bytes
//
// Any type that implements [encoding.BinaryMarshaler] is encoded by taking the
// result of calling its MarshalBinary() method and prepending it with an
// integer value giving the number of bytes in it.
//
//	Slice Values
//
//	Layout:
//
//	[ INFO ] [ INT VALUE ] [ ITEM 1 ] ... [ ITEM N ]
//	<-------COUNT--------> <--------VALUES--------->
//	      1..9 bytes              COUNT bytes
//
// Slices are encoded as a count of bytes that make up the entire slice,
// followed by the encoded value of each element in the slice. There is no
// special delimiter between the encoded elements; when one ends, the next one
// begins.
//
//	Map Values
//
//	Layout:
//
//	[ INFO ] [ INT VALUE ] [ KEY 1 ] [ VALUE 1 ] ... [ KEY N ] [ VALUE N ]
//	<-------COUNT--------> <-------------------VALUES-------------------->
//	      1..9 bytes                         COUNT bytes
//
// Map values are encoded as a count of all bytes that make up the entire map,
// followed by pairs of the encoded keys and associated values for each element
// of the map. Each pair consistes of the encoded key, followed immediately by
// the encoded value that the key maps to. There is no special delimiter between
// key-value pairs or between the key and value in a pair; where one ends, the
// next one begins.
//
// The encoded keys are placed in a consistent order; encoding the same map will
// result in the same encoding regardless of the order of keys encountered
// during iteration over the keys.
//
//	Nil Values
//
//	Layout:
//
//	[ INFO ] [ INT INFO ] [ INT VALUE ]
//	<-INFO-> <---EXTRA INDIRECTIONS--->
//	 1 byte          0..9 bytes
//
// Nil values are encoded similarly to integers, with one major exception: the
// nil bit in the info byte is set to true. This allows a nil to be stored in
// the same place as a length count, so when interpreting data, a length count
// can be checked for nil and if nil, instead of the normal value being decoded,
// a nil value is decoded.
//
// Nil pointers to a non-pointer type of any kind are encoded as a single info
// byte with the nil bit set and the indirection bit unset.
//
// Pointers that are themselves not nil but point to another pointer which is
// nil are encoded slightly differently. In this case, the info byte will have
// both the nil bit and the indirection bit set, and will then be followed by a
// normal encoded integer with its own info byte. The encoded integer gives the
// number of indirections that are done before a nil pointer is arrived at. For
// instance, a ***int that points to a valid **int that itself points to a valid
// *int which is nil would be encoded as a nil with indirection level of 2.
//
// Encoded nil values are *not* typed; they will be interpreted as the same type
// as the pointed-to value of the receiver passed to REZI during decoding.
//
//	Pointer Values
//
//	Layout:
//
//	(either encoded value type, or encoded nil)
//
// A pointer is not encoded in a special manner. Instead, the value they point
// to is encoded as though it were not pointer, and when decoding to a pointer,
// the value is first decoded, then a pointer to the decoded value is used as
// the value of the pointer.
//
// If a pointer is nil, it is instead encoded as a nil value.
//
// Pointers that have multiple levels of indirection before arriving at the
// pointed-to value are not treated any differently when non-nil; i.e. an **int
// which points to an *int which points to an int with value 413 would be
// encoded as an integer value representing 413. If a pointer with multiple
// levels of indirection has a nil somewhere in the indirection chain, it is
// encoded as a nil value; see the section on nil value encodings for a
// description of how this information is captured.
//
// Compatibility:
//
// Older versions of the REZI encoding indicated nil by giving -1 as the byte
// count. This version of REZI will read this as well and can interpret it
// correctly, however do note that it will only be able to handle a single level
// of indirection, i.e. a nil pointer-to-type, with no additional indirections.
package rezi

import (
	"fmt"
	"io"
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

// MustEnc is identical to Enc, but panics if an error would be returned.
func MustEnc(v interface{}) []byte {
	enc, err := Enc(v)
	if err != nil {
		panic(err.Error())
	}
	return enc
}

// Enc encodes a value to REZI-format bytes. The type of the value is examined
// to determine how to encode it. No type information is included in the
// returned bytes, so it is up to the caller to keep track of it and use a
// receiver of a compatible type when decoding.
//
// If a problem occurs while encoding, the returned error will be non-nil and
// will return true for errors.Is(err, rezi.Error). Additionally, the same
// expression will return true for other error types, depending on the cause of
// the error. Do not check error types with the equality operator ==; this will
// always return false.
//
// Non-nil errors from this function can match the following error types: Error
// in all cases. ErrInvalidType if the type of v is not supported.
// ErrMarshalBinary if an implementor of encoding.BinaryMarshaler returns an
// error from its MarshalBinary method (additionally, the returned error will
// match the same types that the error returned from MarshalBinary would match).
func Enc(v interface{}) (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = reziError{
				msg: fmt.Sprintf("%v", r),
			}
		}
	}()

	info, err := canEncode(v)
	if err != nil {
		return nil, err
	}

	if info.Primitive() {
		return encCheckedPrim(v, info)
	} else if info.Main == mtNil {
		return encNilHeader(0), nil
	} else if info.Main == mtMap {
		return encCheckedMap(v, info)
	} else if info.Main == mtSlice {
		return encCheckedSlice(v, info)
	} else {
		panic("no possible encoding")
	}
}

// MustDec is identical to Dec, but panics if an error would be returned.
func MustDec(data []byte, v interface{}) int {
	n, err := Dec(data, v)
	if err != nil {
		panic(err.Error())
	}
	return n
}

// Dec decodes a value from REZI-format bytes in data, starting with the first
// byte in it. Returns the number of bytes consumed in order to read the
// complete value. If the data slice was constructed by appending encoded values
// together, then skipping over n bytes after a successful call to Dec will
// result in the next call to Dec reading the next subsequent value.
//
// V must be a non-nil pointer to a type supported by REZI. The type of v is
// examined to determine how to decode the value. The data itself is not
// examined for type inference, therefore v must be a pointer to a compatible
// type. V is only assigned to on successful decoding; if this function returns
// a non-nil error, v will not have been assigned to.
//
// If a problem occurs while decoding, the returned error will be non-nil and
// will return true for errors.Is(err, rezi.Error). Additionally, the same
// expression will return true for other error types, depending on the cause of
// the error. Do not check error types with the equality operator ==; this will
// always return false.
//
// Non-nil errors from this function can match the following error types: Error
// in all cases. ErrInvalidType if the type pointed to by v is not supported or
// if v is a nil pointer. ErrUnmarshalBinary if an implementor of
// encoding.BinaryUnmarshaler returns an error from its UnmarshalBinary method
// (additionally, the returned error will match the same types that the error
// returned from UnmarshalBinary would match). io.ErrUnexpectedEOF if there are
// fewer bytes than necessary to decode the value. ErrMalformedData if there is
// any problem with the data itself (including there being fewer bytes than
// necessary to decode the value).
func Dec(data []byte, v interface{}) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = reziError{
				msg: fmt.Sprintf("%v", r),
			}
		}
	}()

	info, err := canDecode(v)
	if err != nil {
		return 0, err
	}

	if info.Primitive() {
		return decCheckedPrim(data, v, info)
	} else if info.Main == mtMap {
		return decCheckedMap(data, v, info)
	} else if info.Main == mtSlice {
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
			return encNilHeader(nilLevel), nil
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
	var hdr countHeader

	if ti.Indir > 0 {
		hdr, n, err = decCountHeader(data)
		if err != nil {
			return decoded, n, reziError{
				msg:   fmt.Sprintf("check count header: %s", err.Error()),
				cause: []error{err},
			}
		}
	}

	effectiveExtraIndirs := hdr.ExtraNilIndirections()

	if !hdr.IsNil() {
		decoded, n, err = decFn(data)
		if err != nil {
			return decoded, n, err
		}
		effectiveExtraIndirs = ti.Indir
	}

	if ti.Indir > 0 {
		// the user has passed in a ptr-ptr-to. We cannot directly assign.
		assignTarget := reflect.ValueOf(v)
		// assignTarget is a **string but we want a *string

		for i := 0; i < ti.Indir && i < effectiveExtraIndirs; i++ {
			// *double indirection ALL THE WAY~*
			// *acrosssss the sky*
			// *what does it mean*

			// **string     // *string  // string
			newTarget := reflect.New(assignTarget.Type().Elem().Elem())
			assignTarget.Elem().Set(newTarget)
			assignTarget = newTarget
		}

		if !hdr.IsNil() {
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
			if receiverType.Elem().Kind() == reflect.Func {
				// if we have been given a *function* pointer, reject it, we
				// cannot do this.
				return nil, 0, reziError{
					msg:   "function pointer type receiver is not supported",
					cause: []error{ErrInvalidType},
				}
			}
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

// countHeader represents the info available in the initial "info byte" of an
// encoded int.
type countHeader struct {
	Negative bool

	// NilAt indicates that it is nil at that level of indirection. 1 is nil at
	// first, 2 is nil at the second (1 level of additional indirection), etc.
	// 0 is not nil at all. A value greater than 0 means there are no further
	// bytes to read for the value being checked (it's nil), a value greater
	// than 1 additionally means that an integer following the info byte(s) was
	// decoded/will be encoded to give that number (NilAt - 1).
	NilAt tNilLevel

	// num following bytes that make up an integer, must be representable as a
	// 4-bit unsigned int (so must be between 0 and 15). will be 0 for nil.
	Length int

	// Whether count is explicitly byte count (used to distinguish new-style
	// string format from old style, which used a count of *runes*). If true,
	// automatically implies ExtensionLevel >= 1.
	ByteLength bool

	// Must be representable as a 4-bit unsigned int. If not 0, automatically
	// implies ExtensionLevel >= 1
	Version int

	// ExtensionLevel is number of extension bytes that are in the
	// representation. Caveat - this can be "wrong". When encoding, regardless
	// of this value as many extension bytes as are needed to encode non-default
	// values are included; if this number is *higher*, extension bytes up to
	// the ExtensionLevel (up to the maximum extension bytes possible) are
	// included.
	//
	// during decoding this explicitly notes how many extension bytes were
	// present even if not needed.
	//
	// at this time, 1 is the maximum level supported for encoding. If higher,
	// encoding will not succeed. It can be higher than 1 after decoding; this
	// indicates that that many Extension bytes were read (but only the first N
	// will be interpreted).
	ExtensionLevel int

	// DecodedCount is the total number of bytes that were consumed during
	// decode. it is completely ignored during encoding. Includes bytes that
	// make up integer count of extra indirections of a nil; does NOT include
	// bytes that make up the "content" of an int that is started by the count
	// header, as that data is not included in a countHeader and is not parsed.
	DecodedCount int
}

func (hdr countHeader) IsNil() bool {
	return hdr.NilAt > 0
}

func (hdr countHeader) ExtraNilIndirections() int {
	return hdr.NilAt - 1
}

// encode the header info as valid bytes.
func (hdr countHeader) MarshalBinary() ([]byte, error) {

	// infobyte bit layout for ref:
	// SXNILLLL
	//
	// S = Sign
	// X = eXtesnion
	// N = Nil
	// I = has Indirection past 1
	// L = Length (4-bit unsigned int)
	//
	//
	// extension byte 1 layout for ref:
	// BXUUVVVV
	//
	// B = length is Byte count. not included if not needed.
	// X = eXtension
	// U = Unused
	// V = binary format explicit Version

	if hdr.Length > 15 || hdr.Length < 0 {
		return nil, reziError{msg: "countHeader.Length cannot fit into nibble", cause: []error{ErrMalformedData}}
	}

	if hdr.Version > 15 || hdr.Version < 0 {
		return nil, reziError{msg: "countHeader.Version cannot fit into nibble", cause: []error{ErrMalformedData}}
	}

	var encoded []byte

	// L bits
	infoByte := uint8(hdr.Length)

	// S bit
	if hdr.Negative {
		infoByte |= infoBitsSign
	}

	// N bit
	if hdr.NilAt > 0 {
		infoByte |= infoBitsNil
	}

	// I bit
	if hdr.NilAt > 1 {
		infoByte |= infoBitsIndir
	}

	encoded = append(encoded, infoByte)

	// if later things require more info bytes, continue to the next
	if hdr.ByteLength || hdr.Version > 0 || hdr.ExtensionLevel >= 1 {
		// do the extension byte

		extByte := uint8(hdr.Version)

		if hdr.ByteLength {
			extByte |= infoBitsByteCount
		}

		encoded = append(encoded, extByte)
	}

	// okay, if nilAt is > 1 then we need to additionally encode an int of that
	// value
	if hdr.NilAt > 1 {
		encoded = append(encoded, encInt(hdr.NilAt-1)...)
	}

	return encoded, nil
}

// decode the header info from valid bytes. will consume following int if
// needed to fill the NilAt value. Will *not* consume regular int value bytes.
func (hdr *countHeader) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return reziError{msg: "no bytes to decode", cause: []error{io.ErrUnexpectedEOF, ErrMalformedData}}
	}

	decoded := countHeader{}

	infoByte := data[0]

	decoded.Length = int(infoByte & infoBitsLen)
	decoded.Negative = infoByte&infoBitsSign != 0
	decoded.DecodedCount = 1 // for the initial info byte

	if infoByte&infoBitsNil != 0 {
		decoded.NilAt++
		// need to hold off on indir check until after we've processed all bytes
		// so we will hold on to info byte and check back later
	}

	// scan all extension bytes
	extByte := infoByte
	for extByte&infoBitsExt != 0 {
		decoded.ExtensionLevel++
		if len(data) < decoded.ExtensionLevel+1 {
			return reziError{cause: []error{io.ErrUnexpectedEOF, ErrMalformedData}}
		}
		extByte = data[decoded.ExtensionLevel]

		// we have consumed an additional byte, add it to total decoded
		decoded.DecodedCount++

		// interpret the extension byte based on which one it is
		if decoded.ExtensionLevel == 1 {
			// first extension byte, layout: BXUUVVVV.
			decoded.Version = int(extByte & infoBitsVersion)
			decoded.ByteLength = extByte&infoBitsByteCount != 0
		}

		// future: more extension bytes, if needed. for now, just run through
		// and process them.
	}

	// all extension bytes processed, now decode any indirection level int if
	// present
	if infoByte&infoBitsIndir != 0 {
		extraIndirs, n, err := decInt[tNilLevel](data[decoded.DecodedCount:])
		if err != nil {
			return err
		}
		decoded.DecodedCount += n
		decoded.NilAt += extraIndirs
	}

	*hdr = decoded

	return nil
}
