// Package rezi provides the ability to encode and decode data in Rarefied
// Encoding (Compressible) Interchange format. It allows Go types and
// user-defined types to be easily read from and written to byte slices, with
// customization possible by implementing encoding.BinaryUnmarshaler and
// encoding.BinaryMarshaler on a type, or alternatively by implementing
// encoding.TextUnmarshaler and encoding.TextMarshaler. REZI has an interface
// similar to the json package; one function is used to encode all supported
// types, and another function receives bytes and a receiver for decoded data
// and infers how to decode the bytes based on the receiver.
//
// The [Enc] function is used to encode any supported type to REZI bytes:
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
//	var n, offset int
//	var err error
//
//	n, err = rezi.Dec(allData[offset:], &readNumber)
//	if err != nil {
//		panic(err.Error())
//	}
//	offset += n
//
//	n, err = rezi.Dec(allData[offset:], &readName)
//	if err != nil {
//		panic(err.Error())
//	}
//	offset += n
//
// Alternatively, instead of calling Dec and Enc directly on a pre-loaded slice
// of data bytes, the [Reader] and [Writer] types can be used to operate on a
// stream of data. See the section below for more information.
//
// # Compression
//
// Compression can be enabled by passing a [Format] struct with Compression
// options set to any method which accepts a Format. At this time, this is
// possible only with Readers and Writers.
//
// The zlib library is used for compression, with a compression ratio that may
// be specified at write time.
//
// # Readers and Writers
//
// For reading and writing from data streams, [Reader] and [Writer] are
// provided. They each have their own Dec and Enc methods and do not require
// that manual tracking be provided for proper offset error-reporting.
//
// Additionally, the Reader and Writer both support being used for writing
// arbitrary streams of bytes encoded as REZI byte slices. Using the typical
// Write method on Writer will result in writing them as a single REZI-encoded
// slice of bytes. Those bytes can later be read from via calls to Reader.Read,
// which will automatically read across multiple encoded slices if needed. This
// allows both sides to operate without needing to the "full" length of their
// data ahead of time, although it should be noted that this is not a
// particularly efficient use of REZI encoding.
//
// # Error Checking
//
// Errors in REZI have specific types that they can be checked against to
// determine their cause. These errors conform to the [errors] interface and
// must be checked by using [errors.Is].
//
// As mentioned in that library's documentation, errors should not be checked
// with simple equality checks. REZI enforces this fully. Non-nil errors that
// are checked with `==` will never return true.
//
//	if err == rezi.Error
//
// The above expression is not simply the non-preferred way of checking an
// error, but rather is entirely non-functional, as it will always return false.
// Instead, do:
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
// See the individual functions for a list of error types that returned errors
// may be checked against.
//
// # Supported Data Types
//
// REZI supports all built-in basic Go types: int (as well as all of its
// unsigned and specific-size varieties), float32, float64, complex64,
// complex128, string, bool, and any type that implements
// encoding.BinaryMarshaler or encoding.TextMarshaler (for encoding) or whose
// pointer type implements encoding.BinaryUnmarshaler or
// encoding.TextUnmarshaler (for decoding). Implementations of
// encoding.BinaryUnmarshaler should use [Wrapf] when encountering an error from
// a REZI function called from within UnmarshalBinary to supply additional
// offset information, but this is not strictly required.
//
// Slices, arrays, and maps are supported with some stipulations. Slices and
// arrays must contain only other supported types (or pointers to them). Maps
// have the same restrictions on their values, but only maps with a key type of
// string, int (or any of its unsigned or specific-size varieties), float32,
// float64, or bool are supported.
//
// Pointers to any supported type are also accepted, including to other pointer
// types with any number of indirections. The REZI format encodes information on
// how many levels of indirection are valid, though of course note that it does
// not have any concept of two different pointer variables pointing to the same
// data.
//
// All non-struct types whose underlying type is a supported type are themselves
// supported as well. For example, time.Duration has an underlying type of
// int64, and is therefore supported in REZI.
//
// Struct types are supported even if they do not implement text or binary
// marshaling functions, provided all of their exported fields are of a
// supported type. Both decoding and encoding ignore all unexported fields. If a
// field is not present in the given bytes during decoding, its original value
// is left intact, even if it is exported.
//
// # Binary Data Format
//
// REZI uses a binary format for all supported types. Other than bool, which is
// encoded as a single byte, an encoded value will start with one or more "info"
// bytes that contain metadata on the value itself. This is typically the length
// of the full value but may include additional information such as whether the
// encoded value is a nil pointer.
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
// where each letter from left to right stands for a particular bit from most to
// least significant.
//
// The bit labeled "S" is the sign bit; when high (1), it indicates that the
// following integer value is negative.
//
// The "X" bit is the extension flag, and indicates that the next byte is a
// second info byte with additional information, called the info extension byte.
// At this time, only encoded string values use this extension byte.
//
// The "N" bit is the explicit nil flag, and when set it indicates that the
// value is a nil and that there are no following bytes in the encoded value
// other than any indirection amount indicators.
//
// The "I" bit is the indirection bit, and if set, indicates that the following
// bytes encode the number of additional indirections of the pointer beyond the
// initial indirection at which the nil occurs; for instance, a nil *int value
// is encoded as simply the info byte 0b00100000, but a non-nil **int that
// points at a nil *int would be encoded with one level of additional
// indirection and the info byte's I bit would be set.
//
// The "L" bits make up the length of the value. Together, they are a 4-bit
// unsigned integer that indicates how many of the following bytes are part of
// the encoded value. If the I bit is set on the info byte, the L bits give the
// number of bytes that make up the indirection level rather than the actual
// value.
//
//	The EXT Byte
//
//	Layout:
//
//	BXUUVVVV
//	|      |
//	MSB  LSB
//
// The initial INFO byte may be followed by a second byte, the info extension
// byte (EXT for short). This encodes additional metadata about the encoded
// value.
//
// The "B" bit is the byte count flag. If this is set, it explicitly indicates
// that a count in bytes is given immediately after all extension bytes in the
// header have been scanned. This count is given as the data bytes of a
// regularly-encoded int value sans its own header (its header is the one that
// the EXT byte is a part of). Note that the lack of this flag or the extension
// byte as a whole does not necessarily indicate that the count is *not*
// byte-based; an encoded type format that explicitly notes that the count is
// byte-based without an EXT byte in its layout diagram will be assumed to have
// a byte-based length.
//
// The "V" bits make up the version field of the extension byte. This indicates
// the version of encoding of the particular type that is represented, encoded
// as a 4-bit unsigned integer. If not present (all 0's, or the EXT byte itself
// is not present), it is assumed to be 1. This version number is purely
// informative and does not affect decoding in any way.
//
// The "U" bits are unused at this time and are reserved for future use.
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
// no bytes in INT VALUE at all; the 64-bit representation for 0 is all 0x00's,
// and therefore has no significant bytes. Likewise, the 64-bit representation
// for -1 using two's complement representation is all 0xff's. Both of these are
// encoded by an INFO byte that gives a length of zero; distinguishing between
// the two is done via the sign bit in the INFO byte.
//
// All Go integer types are encoded in the same way. This includes int, int8,
// int16, int32, int64, uint, uint8, uint16, uint32, and uint64. The specific
// interpretation into a value is handled at decoding time by infering the type
// from the pointer passed to Enc.
//
//	Float Values
//
//	Full Layout:
//
//	[ INFO ] [ COMP-EXPONENT-HIGHS ] [ MIXED ] [ MANTISSA-LOWS ]
//	 1 byte          1 byte            1 byte      0..6 bytes
//
//	Short-Form Layout:
//
//	[ INFO ]
//	 1 byte
//
// A non-zero float value is encoded by taking the components of its
// representation in IEEE-754 double-precision and encoding them across 1 to 9
// bytes, using compaction where possible. These components are a 1-bit sign, an
// 11-bit exponent, and a 52-bit fraction (also known as the mantissa). Float
// values of 0.0 and -0.0 are instead encoded using an abbreviated "short-form"
// that consists of only a single byte.
//
// All float values begin with an INFO byte. Assuming it does not denote a nil
// value, the 4 L bits of the info byte give the number of bytes following all
// header bytes that are used to encode the value, and the S bit represents
// whether the value is negative, thus encoding the 1-bit sign. If the L bits
// give a non-zero value, the float value uses the full encoding layout; if the
// L bits give a zero value, the float value uses the short-form.
//
// An INFO byte in full-form is followed by the COMP-EXPONENT-HIGHS byte. This
// contains two fields, organized in the byte bits as CEEEEEEE. The first field
// is a 1-bit flag, denoted by "C", that indicates whether compaction of the
// mantissa is performed from the right or the left side. If set, it is from the
// right; if not set, it is from the left. The remaining bits in the byte,
// denoted by "E", are the 7 high-order bits of the exponent component of the
// represented value.
//
// The next byte in full-form is a MIXED byte containing two fields, organized
// in the byte bits as EEEEMMMM. The first field, denoted by "E", contains the 4
// lower-order bits of the exponent. The second field, denoted by "MMMM",
// contains the 4 high-order bits of the mantissa.
//
// After the MIXED byte, the remaining 48 low-order bits of the mantissa are
// encoded with compaction similar to that performed on integer values, but with
// some modifications. First, only 0x00 bytes are removed from the
// representation to compact them; 0xff bytes are never removed, as the mantissa
// is itself is never represented as a two's complement negative value. Second,
// consecutive 0x00 bytes may be removed from either the left or the right side
// of those 48 bits, whatever would make it more compact. The "C" bit being set
// in the COMP-EXPONENT-HIGHS byte indicates that they are removed from the
// right, otherwise they are removed from the left as in compaction of integer
// values. If all 48 low-order bits of the Mantissa are 0x00, they will all be
// compacted and the entire float will take up only the initial three bytes.
//
// Note that the above compaction applies only to the 48 low-order bits of the
// mantissa; the high 4 bits will always be present in the MIXED byte regardless
// of their value.
//
// Zero-valued floats, 0.0 and -0.0, are not encoded using the full layout
// described above, but instead as special cases are encoded in a short-form
// layout as a single INFO byte whose L bits are all set to 0. 0.0 is encoded in
// as a single 0x00 byte, and -0.0 is encoded as a single 0x80 byte. These are
// the only values of float that are encoded in short-form; all others use the
// full form.
//
//	Complex Values
//
//	Layout:
//	[ INFO ] [ EXT ] [ INT VALUE ] [ INFO ] [ FLOAT VALUE ] [ INFO ] [ FLOAT VALUE ]
//	<-----------COUNT------------> <------REAL PART-------> <----IMAGINARY PART---->
//	         2..10 bytes                  3..9 bytes               3..9 bytes
//
//	Short-Form Layout:
//
//	[ INFO ]
//	 1 byte
//
// Complex values are, in general, encoded as a count of bytes in the header
// bytes given as an explicit byte count followed by that many bytes containing
// first the real component and then the imaginary component in sequence,
// encoded as float values.
//
// As special cases, a complex value with a positive 0.0 real part and positive
// 0.0 imaginary part is encoded using the short-form layout as only a single
// info byte with a value of 0x00, and a complex value with a negative 0.0 real
// part and negative 0.0 imaginary part is encoded as only a single info byte
// with a value of 0x80. This only applies to values of (+0.0)+(+0.0)i and
// (-0.0)+(-0.0)i; there is no special case for when both are zero but of
// opposite signs or for when one part is some zero but the other part is not.
//
//	String Values
//
//	Full Layout:
//
//	[ INFO ] [ EXT ] [ INT VALUE ] [ CODEPOINT 1 ] ... [ CODEPOINT N ]
//	<-----------COUNT------------> <------------CODEPOINTS----------->
//	         2..10 bytes                       COUNT bytes
//
//	Short-Form Layout:
//
//	[ INFO ]
//	 1 byte
//
// String values are encoded as a count of bytes in the info header section
// followed by the Unicode codepoints that make up the string encoded using
// UTF-8. Non-empty strings will use the full layout; an empty string will use
// the abbreviated short-form layout.
//
// A non-empty string value's first info byte will have its extension bit set
// and will indicate explicitly that it uses a byte-based count in the
// extension byte that follows. This is to distinguish it from older-style (V0)
// string encodings, which encoded data length as the count of codepoints rather
// than bytes.
//
// Ann empty string is encoded using the short-form layout as a single info
// byte, 0x00.
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
//	encoding.TextMarshaler Values
//
//	Layout:
//
//	[ INFO ] [ INT VALUE ] [ MARSHALED BYTES ]
//	<-------COUNT--------> <-MARSHALED BYTES->
//	      1..9 bytes           COUNT bytes
//
// Any type that implements [encoding.TextMarshaler] is encoded by taking the
// result of calling its MarshalText() method and encoding that value as a
// string.
//
// Note that BinarayMarshaler encoding takes precedence over TextMarshaler
// encoding; if a type implements both, it will be encoded as a BinaryMarshaler,
// not a TextMarshaler.
//
//	Struct Values
//
//	Layout:
//
//	[ INFO ] [ INT VALUE ] [ FIELD 1 ] [ VALUE 1 ] ... [ FIELD N ] [ VALUE N ]
//	<-------COUNT--------> <---------------------VALUES---------------------->
//	      1..9 bytes                           COUNT bytes
//
// Structs that do not implement binary marshaling or text marshaling funcitons
// are encoded as a count of all bytes that make up the entire struct, followed
// by pairs of the names and associated values for each exported field of the
// struct. Each pair consists of the case-sensitive name of the field encoded as
// a string, followed immediately by the encoded value of that field. There is
// no special delimiter between name-value pairs or between the name and value
// in a pair; where one ends, the next one begins.
//
// The encoded names are placed in a consistent order; encoding the same struct
// will result in the same encoding.
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
//	Array Values
//
// Arrays are encoded in an identical fashion to slices. They do not record the
// size of the array type.
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
// of the map. Each pair consists of the encoded key, followed immediately by
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
// *int which is nil would be encoded as a nil with an indirection level of 2.
//
// Encoded nil values are not typed; they will be interpreted as the same type
// as the pointed-to value of the receiver passed to REZI during decoding.
//
//	Pointer Values
//
//	Layout:
//
//	(either encoded value type, or encoded nil)
//
// Pointers do not have their own dedicated encoding format. Instead, the value
// a pointer points to is encoded as though it were not a pointer type, and when
// decoding to a pointer, the value is first decoded, then a pointer to the
// decoded value is created and used as the returned value.
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
// # Backward Compatibility
//
// Older versions of the REZI library use a binary data format that differs from
// the current one. The current version retains compatibility for reading data
// produced by prior versions of this library, regardless of whether they were
// major version releases. The binary format outlined above and the changes
// noted below are all considered a part of "V1" of the binary format itself
// separate from the version of the Go module.
//
// REZI library versions prior to v1.1.0 indicate nil by giving -1 as the byte
// count, and could only encode a nil value for slices and maps. This older
// format is only able to encode a single level of indirection, i.e. a nil
// pointer-to-type, with no additional indirections. Due to this limitation,
// decoding these values will result in either a nil pointer or all levels
// indirected up to the non-nil value; it will never be decoded as, for example,
// a pointer to a pointer which is then nil. When writing a nil value, REZI sets
// the sign bit and keeps the length bytes clear in the first INFO header byte;
// this allows versions prior to v1.1.0 to be able to read it, as long as it has
// only a single level of indirection.
//
// REZI library versions prior to v2.1.0 encode string data length as the number
// of Unicode codepoints rather than the number of bytes and do so in the info
// byte with no info extension byte. These strings can be decoded as normal with
// [Dec] and [Reader.Dec].
package rezi

import (
	"io"
	"reflect"
)

// decInfo holds information gained during decoding to be used in further
// processing of the decoded data.
type decInfo struct {
	// Fields is only included in decInfo when a struct has been decoded and
	// gives a slice of all valid fields that were detected (and read) during
	// the decode. This is used to inform which fields to overwrite in the
	// receiver pointer.
	Fields []fieldInfo

	// Ref is the value that was decoded but as a reflect.Value. This is so that
	// creating it more than once can be avoided. Not always set by every decode
	// function; check with IsValid() before using.
	Ref reflect.Value
}

type (
	tLen      = int
	tNilLevel = int

	// the decInfo in decFunc is any additional info needed for further
	// stages of decode. It can be empty if no further info is required.
	decFunc[E any] func([]byte) (E, decInfo, int, error)
	encFunc[E any] func(analyzed[E]) ([]byte, error)
)

// analyzed is used to pass around a value along with its type info and
// reflect.Value to different subroutines. it's mostly just used for argument
// grouping.
type analyzed[E any] struct {
	native E
	ref    reflect.Value
	ti     typeInfo
}

func preAnalyzed[E any](oldAnalysis analyzed[any], newVal E) analyzed[E] {
	return analyzed[E]{native: newVal, ref: oldAnalysis.ref, ti: oldAnalysis.ti}
}

func nilErrEncoder[E any](fn func(analyzed[E]) []byte) encFunc[E] {
	return func(val analyzed[E]) ([]byte, error) {
		return fn(val), nil
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
// ErrMarshalText if an implementor of encoding.TextMarshal returns an error
// from its MarshalText method.
func Enc(v interface{}) (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errorf("%v", r)
		}
	}()

	info, err := canEncode(v)
	if err != nil {
		return nil, err
	}

	value := analyzed[any]{
		native: v,
		ref:    reflect.ValueOf(v),
		ti:     info,
	}

	if info.Primitive() {
		return encCheckedPrim(value)
	} else if info.Main == mtNil {
		return encNilHeader(0), nil
	} else if info.Main == mtMap {
		return encCheckedMap(value)
	} else if info.Main == mtSlice || info.Main == mtArray {
		return encCheckedSlice(value)
	} else if info.Main == mtStruct {
		return encCheckedStruct(value)
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
// returned from UnmarshalBinary would match). ErrUnmarshalText if an
// implementor of encoding.TextUnmarshaler returns an error from its
// UnmarshalText method. io.ErrUnexpectedEOF if there are fewer bytes than
// necessary to decode the value. ErrMalformedData if there is any problem with
// the data itself (including there being fewer bytes than necessary to decode
// the value).
func Dec(data []byte, v interface{}) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errorf("%v", r)
		}
	}()

	info, err := canDecode(v)
	if err != nil {
		return 0, err
	}

	value := analyzed[any]{
		native: v,
		ref:    reflect.ValueOf(v),
		ti:     info,
	}

	if info.Primitive() {
		return decCheckedPrim(data, value)
	} else if info.Main == mtMap {
		return decCheckedMap(data, value)
	} else if info.Main == mtSlice || info.Main == mtArray {
		return decCheckedSlice(data, value)
	} else if info.Main == mtStruct {
		return decCheckedStruct(data, value)
	} else {
		panic("no possible decoding")
	}
}

func encWithNilCheck[E any](v analyzed[any], encFn encFunc[E], convFn func(reflect.Value) E) ([]byte, error) {
	if v.ti.Indir > 0 {
		// we cannot directly encode, we must get at the reel value.
		encodeTarget := v.ref
		// encodeTarget is a *THING but we want a THING

		nilLevel := -1
		for i := 0; i < v.ti.Indir; i++ {
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
		convTarget := convFn(encodeTarget)
		reAnalyzed := analyzed[E]{
			native: convTarget,
			ref:    reflect.ValueOf(convTarget),
			ti:     v.ti,
		}
		return encFn(reAnalyzed)
	} else {
		// if the type we have is actually a new UDT with some underlying basic
		// Go type, then in fact we want to encode it as the actual kind type.
		if v.ti.Underlying {
			v.native = convFn(v.ref)
		}
		return encFn(preAnalyzed(v, v.native.(E)))
	}
}

// if ti.Indir > 0, this will assign to the interface at the appropriate
// indirection level. If ti.Indir == 0, this will not assign. Callers should use
// that check to determine if it is safe to do their own assignment of the
// decoded value this function returns.
func decWithNilCheck[E any](data []byte, v analyzed[any], decFn decFunc[E]) (decoded E, di decInfo, n int, err error) {
	var hdr countHeader

	if v.ti.Indir > 0 {
		hdr, n, err = decCountHeader(data)
		if err != nil {
			return decoded, di, n, errorDecf(0, "check count header: %s", err)
		}
	}

	countHeaderBytes := n
	effectiveExtraIndirs := hdr.ExtraNilIndirections()

	if !hdr.IsNil() {
		effectiveExtraIndirs = v.ti.Indir
		decoded, di, n, err = decFn(data)
		if err != nil {
			return decoded, di, n, errorDecf(countHeaderBytes, "%s", err)
		}
	}

	if v.ti.Indir > 0 {
		// the user has passed in a ptr-ptr-to. We cannot directly assign.
		assignTarget := v.ref
		// assignTarget is a **string but we want a *string

		// if it's a struct, we must get the original value, if one exists, in order
		// to preserve the original member values
		var origStructVal reflect.Value
		if v.ti.Main == mtStruct {
			origStructVal = unwrapOriginalStructValue(assignTarget)
		}

		for i := 0; i < v.ti.Indir && i < effectiveExtraIndirs; i++ {
			// *double indirection ALL THE WAY~*
			// *acrosssss the sky*
			// *what does it mean*

			// **string     // *string  // string
			newTarget := reflect.New(assignTarget.Type().Elem().Elem())
			assignTarget.Elem().Set(newTarget)
			assignTarget = newTarget
		}

		if !hdr.IsNil() {
			refDecoded := reflect.ValueOf(decoded)
			if v.ti.Underlying {
				refDecoded = refDecoded.Convert(assignTarget.Type().Elem())
			}

			if v.ti.Main == mtStruct && origStructVal.IsValid() {
				refDecoded = setStructMembers(origStructVal, refDecoded, di)
			}

			assignTarget.Elem().Set(refDecoded)
		} else {
			zeroVal := reflect.Zero(assignTarget.Elem().Type())
			assignTarget.Elem().Set(zeroVal)
		}
	}

	return decoded, di, n, nil
}

// decToUnwrappedFn takes the encoded bytes and an interface to decode to and
// returns any extra data (may be nil), bytes consumed, and error status.
func fn_DecToWrappedReceiver(wrapped analyzed[any], assertFn func(reflect.Type) bool, decToUnwrappedFn func([]byte, analyzed[any]) (decInfo, int, error)) decFunc[interface{}] {
	return func(data []byte) (interface{}, decInfo, int, error) {
		// v is *(...*)T, ret-val of decFn (this lambda) is T.
		refWrapped := wrapped.ref
		receiverType := refWrapped.Type()
		refUnwrapped := refWrapped

		if receiverType.Kind() == reflect.Pointer { // future-proofing - binary unmarshaler might come in as a T
			// for every * in the (...*) part of *(...*)T up until the
			// implementor/slice-ptr, do a deref.
			for i := 0; i < wrapped.ti.Indir; i++ {
				receiverType = receiverType.Elem()
				if (wrapped.ti.Main == mtText || wrapped.ti.Main == mtBinary) && refUnwrapped.IsValid() {
					if !refUnwrapped.IsNil() {
						refUnwrapped = refUnwrapped.Elem()
					} else {
						refUnwrapped = reflect.Value{}
					}
				}
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
				return nil, decInfo{}, 0, errorDecf(0, "function pointer type receiver is not supported").wrap(ErrInvalidType)
			}
			// receiverType is *T

			// Make shore we are actually using the populated struct for
			// interface implementors, in case unmarshaling is dependant on any
			// prior-set properties. No need to do this for mtStruct bc that is
			// automatic and we have full control; we don't (and can't) rely on
			// already set things and use reflect after actual decoding to copy
			// the decoded values into the passed-in pointer.
			if (wrapped.ti.Main == mtText || wrapped.ti.Main == mtBinary) && refUnwrapped.IsValid() && !refUnwrapped.IsNil() {
				receiverValue = refUnwrapped
			} else {
				receiverValue = reflect.New(receiverType.Elem())
			}
		} else {
			// receiverType is itself T (future-proofing)
			if (wrapped.ti.Main == mtText || wrapped.ti.Main == mtBinary) && refUnwrapped.IsValid() {
				receiverValue = refUnwrapped
			} else {
				receiverValue = reflect.Zero(receiverType)
			}
		}

		var decoded interface{}

		receiver := receiverValue.Interface()
		recvAnalyzed := analyzed[any]{native: receiver, ref: receiverValue, ti: wrapped.ti}
		extraInfo, decConsumed, decErr := decToUnwrappedFn(data, recvAnalyzed)

		if decErr != nil {
			return nil, extraInfo, decConsumed, decErr
		}

		if receiverType.Kind() == reflect.Pointer {
			decoded = reflect.ValueOf(receiver).Elem().Interface()
		} else {
			decoded = receiver
		}

		return decoded, extraInfo, decConsumed, decErr
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
		return nil, errorf("countHeader.Length cannot fit into nibble").wrap(ErrMalformedData)
	}

	if hdr.Version > 15 || hdr.Version < 0 {
		return nil, errorf("countHeader.Version cannot fit into nibble").wrap(ErrMalformedData)
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
		encoded[0] |= infoBitsExt

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
		encoded = append(encoded, encInt(analyzed[int]{native: hdr.NilAt - 1})...)
	}

	return encoded, nil
}

// decode the header info from valid bytes. will consume following int if
// needed to fill the NilAt value. Will *not* consume regular int value bytes.
func (hdr *countHeader) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errorDecf(0, "no bytes to decode").wrap(io.ErrUnexpectedEOF, ErrMalformedData)
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
			s := "s"
			verbS := ""
			if len(data) == 1 {
				s = ""
				verbS = "s"
			}
			const errFmt = "count header length is at least %d but only %d byte%s remain%s in data"
			err := errorDecf(decoded.DecodedCount, errFmt, decoded.ExtensionLevel+1, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
			return err
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
		extraIndirs, _, n, err := decInt[tNilLevel](data[decoded.DecodedCount:])
		if err != nil {
			return errorDecf(decoded.DecodedCount, "%s", err)
		}
		decoded.DecodedCount += n
		decoded.NilAt += extraIndirs
	}

	*hdr = decoded

	return nil
}
