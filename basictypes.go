package rezi

// basictypes.go contains functions for encoding and decoding ints, strings,
// bools, and objects that directly implement BinaryUnmarshaler and
// BinaryMarshaler.

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode/utf8"
)

const (
	infoBitsSign  = 0b10000000
	infoBitsExt   = 0b01000000
	infoBitsNil   = 0b00100000
	infoBitsIndir = 0b00010000
	infoBitsLen   = 0b00001111
)

// AnyInt is a union interface that combines all basic Go integer types. It
// allows int, uint, and all of their specifically-sized varieties.
type AnyInt interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

func encIndirect[E any](value interface{}, ti typeInfo, convFn func(reflect.Value) E, encFn func(E) []byte) []byte {
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
}

// encPrim encodes the primitve REZI value as rezi-format bytes. The type of the
// value is examined to determine how to encode it. No type information is
// included in the returned bytes so it is up to the caller to keep track of it.
//
// This function may only be called with a value with type or underlying type of
// int, string, or bool, or a value that implements encoding.BinaryMarshaler.
// For a more generic encoding function that can handle map and slice types, see
// Enc. Generally this function is used internally and users of REZI are better
// off calling the specific type-safe encoding function (EncInt, EncBool,
// EncString, or EncBinary) for the type being encoded.
func encPrim(value interface{}, ti typeInfo) []byte {
	switch ti.Main {
	case tString:
		if ti.Indir > 0 {
			return encIndirect(value, ti, reflect.Value.String, encString)
		} else {
			return encString(value.(string))
		}
	case tBool:
		return encBool(value.(bool))
	case tIntegral:
		if ti.Signed {
			switch ti.Bits {
			case 8:
				return encInt(value.(int8))
			case 16:
				return encInt(value.(int16))
			case 32:
				return encInt(value.(int32))
			case 64:
				return encInt(value.(int64))
			default:
				return encInt(value.(int))
			}
		} else {
			switch ti.Bits {
			case 8:
				return encInt(value.(uint8))
			case 16:
				return encInt(value.(uint16))
			case 32:
				return encInt(value.(uint32))
			case 64:
				return encInt(value.(uint64))
			default:
				return encInt(value.(uint))
			}
		}
	case tBinary:
		return encBinary(value.(encoding.BinaryMarshaler))
	default:
		panic(fmt.Sprintf("%T cannot be encoded as REZI primitive type", value))
	}
}

// decPrim decodes a primitive value from rezi-format bytes into the value
// pointed-to by v. V must point to a REZI primitive value (int, bool, string)
// or implement encoding.BinaryUnmarshaler.
//
// This function may only be called with a value with type or underlying type of
// int, string, or bool, or a value that implements encoding.BinaryUnmarshaler.
// For a more generic encoding function that can handle map and slice types, see
// Enc. Generally this function is used internally and users of REZI are better
// off calling the specific type-safe decoding function (DecInt, DecBool,
// DecString, or DecBinary) for the type being decoded.
func decPrim(data []byte, v interface{}, ti typeInfo) (int, error) {
	// by nature of doing an encoding, v MUST be a pointer to the typeinfo type,
	// or an implementor of BinaryUnmarshaler.

	switch ti.Main {
	case tString:
		s, n, err := decWithIndirectAssignment(data, v, ti, decString)
		if err != nil {
			return n, err
		}
		if ti.Indir == 0 {
			tVal := v.(*string)
			*tVal = s
		}
		return n, nil
	case tBool:
		tVal := v.(*bool)
		b, n, err := decBool(data)
		if err != nil {
			return n, err
		}
		*tVal = b
		return n, nil
	case tIntegral:
		i, n, err := decInt(data)
		if err != nil {
			return n, err
		}
		if ti.Signed {
			switch ti.Bits {
			case 64:
				tVal := v.(*int64)
				*tVal = int64(i)
			case 32:
				tVal := v.(*int32)
				*tVal = int32(i)
			case 16:
				tVal := v.(*int16)
				*tVal = int16(i)
			case 8:
				tVal := v.(*int8)
				*tVal = int8(i)
			default:
				tVal := v.(*int)
				*tVal = int(i)
			}
		} else {
			switch ti.Bits {
			case 64:
				tVal := v.(*uint64)
				*tVal = uint64(i)
			case 32:
				tVal := v.(*uint32)
				*tVal = uint32(i)
			case 16:
				tVal := v.(*uint16)
				*tVal = uint16(i)
			case 8:
				tVal := v.(*uint8)
				*tVal = uint8(i)
			default:
				tVal := v.(*uint)
				*tVal = uint(i)
			}
		}

		return n, nil
	case tBinary:
		// if we just got handed a pointer-to binaryUnmarshaler, we need to undo
		// that

		receiver := v.(encoding.BinaryUnmarshaler)
		return decBinary(data, receiver)
	default:
		panic(fmt.Sprintf("%T cannot receive decoded REZI primitive type", v))
	}
}

// encBool encodes the bool value as a slice of bytes. The value can later
// be decoded with DecBool. No type indicator is included in the output;
// it is up to the caller to add this if they so wish it.
//
// The output will always contain exactly 1 byte.
//
// Deprecated: This function has been replaced by [Enc].
func EncBool(b bool) []byte {
	return encBool(b)
}

func encBool(b bool) []byte {
	enc := make([]byte, 1)

	if b {
		enc[0] = 1
	} else {
		enc[0] = 0
	}

	return enc
}

// DecBool decodes a bool value at the start of the given bytes and
// returns the value and the number of bytes read.
//
// Deprecated: This function has been replaced by [Dec].
func DecBool(data []byte) (bool, int, error) {
	return decBool(data)
}

func decBool(data []byte) (bool, int, error) {
	if len(data) < 1 {
		return false, 0, io.ErrUnexpectedEOF
	}

	if data[0] == 0 {
		return false, 1, nil
	} else if data[0] == 1 {
		return true, 1, nil
	} else {
		return false, 0, ErrInvalidType
	}
}

func encNil(indirLevels int) []byte {
	// nils are encoded as a special negative that is distinct from others,
	// should it be checked.
	//
	// ints always start with SXXXLLLL where S is sign bit, L are byte len, and
	// X are unused bits for the number encoding. 0 is a special case, encoded
	// as simply b00000000, and -1 is a special case, encoded as b10000000.
	//
	// for an explicit nil, we will use the additional bits, XXX. We will label
	// them X, N, and I, respectively, for a total info byte scheme of SXNILLLL.
	// X is reserved for use to indicate info extension, which means the next
	// byte has MORE info bits in it. N indicates that the value is not a number
	// but rather an explicit nil. I indicates whether there is more than one
	// level of indirection; if so, the bytes that follow after the extension
	// byte will be a non-nil int that gives the number of indirections.

	infoByte := byte(0)
	infoByte |= infoBitsNil

	// for compat with older format
	infoByte |= infoBitsSign

	if indirLevels <= 0 {
		return []byte{infoByte}
	}

	infoByte |= infoBitsIndir
	enc := []byte{infoByte}

	enc = append(enc, encInt(indirLevels)...)
	return enc
}

// encInt is similar to EncInt but performs specific behavior based on the
// type of int it is given. This allows, for example, the largest value that can
// be held by a uint64 to be properly represented where casting would have
// converted it to a negative integer.
func encInt[E AnyInt](v E) []byte {
	if v == 0 {
		return []byte{0x00}
	}

	negative := v < 0

	i := int64(v)

	b1 := byte((i >> 56) & 0xff)
	b2 := byte((i >> 48) & 0xff)
	b3 := byte((i >> 40) & 0xff)
	b4 := byte((i >> 32) & 0xff)
	b5 := byte((i >> 24) & 0xff)
	b6 := byte((i >> 16) & 0xff)
	b7 := byte((i >> 8) & 0xff)
	b8 := byte(i & 0xff)

	parts := []byte{b1, b2, b3, b4, b5, b6, b7, b8}

	enc := []byte{}
	var hitMSB bool
	for i := range parts {
		if hitMSB {
			enc = append(enc, parts[i])
		} else if (!negative && parts[i] != 0x00) || (negative && parts[i] != 0xff) {
			enc = append(enc, parts[i])
			hitMSB = true
		}
	}

	byteCount := uint8(len(enc))

	// byteCount will never be more than 8 so we can encode sign info in most
	// significant bit
	if negative {
		byteCount |= infoBitsSign
	}

	enc = append([]byte{byteCount}, enc...)

	return enc
}

// EncInt encodes the int value as a slice of bytes. The value can later
// be decoded with DecInt. No type indicator is included in the output;
// it is up to the caller to add this if they so wish it. Integers up to 64 bits
// are supported with this encoding scheme.
//
// The returned slice will be 1 to 9 bytes long. Integers larger in magnitude
// will result in longer slices; only 0 is encoded as a single byte.
//
// Encoded integers start with an info byte that packs the sign and the number
// of following bytes needed to represent the value together. The sign is
// encoded as the most significant bit (the first/leftmost bit) of the byte,
// with 0 being positive and 1 being negative. The next significant 3 bits are
// unused. The least significant 4 bits contain the number of bytes that are
// used to encode the integer value. The bits in the info byte can be
// represented as `SXXXLLLL`, where S is the sign bit, X are unused bits, and L
// are the bits that encode the remaining length.
//
// The remaining bytes give the value being encoded as a 2's complement 64-bit
// big-endian integer, omitting any leading bytes that would be encoded as 0x00
// if the integer is positive, or 0xff if the integer is negative. The value 0
// is special and is encoded as with infobyte 0x00 with no additional bytes.
// Because two's complement is used and as a result of the rules, -1 also
// requires no bytes besides the info byte (because it would simply be a series
// of eight 0xff bytes), and is therefore encoded as 0x80.
//
// Additional examples: 1 would be encoded as [0x01 0x01], 2 as [0x01 0x02],
// 500 as [0x02 0x01 0xf4], etc. -2 would be encoded as [0x81 0xfe], -500 as
// [0x82 0xfe 0x0c], etc.
//
// Deprecated: This function has been replaced by [Enc].
func EncInt(i int) []byte {
	return encInt(i)
}

// DecInt decodes an integer value at the start of the given bytes and
// returns the value and the number of bytes read.
//
// Deprecated: this function has been replaced by [Dec].
func DecInt(data []byte) (int, int, error) {
	return decInt(data)
}

// decNilableInt decodes an integer that could also represent a nil value. It's
// rly only used in places where nil is allowed to be directly encoded, such as
// when decoding a byte/element count.
//
// This function DOES respect the info extension bit, unless it is interpreted
// as an int.
func decNilableInt(data []byte) (isNil bool, iVal int, indir int, consumed int, err error) {
	if len(data) < 1 {
		return false, 0, 0, 0, io.ErrUnexpectedEOF
	}

	infoByte := data[0]
	if infoByte&infoBitsNil != infoBitsNil {
		// not a nil, regular number, do no more manipulation of data and
		// interpret as a regular int.
		iVal, consumed, err = decInt(data)
		return false, iVal, 0, consumed, err
	}

	// it is a nil. do other checks.

	// skip over any extension bytes in the info header
	for data[0]&infoBitsExt == infoBitsExt {
		data = data[1:]
		consumed++
	}

	// data now starts with the last info byte, skip it.
	data = data[1:]
	consumed++

	if infoByte&infoBitsIndir == infoBitsIndir {
		// the level of indirection is encoded in following bytes
		var n int
		indir, n, err = decInt(data)
		consumed += n
		if err != nil {
			return true, 0, indir, consumed, fmt.Errorf("decode ptr indirection level: %w", err)
		}
	}

	return true, 0, indir, consumed, nil
}

// decInt decodes an integer value at the start of the given bytes and
// returns the value and the number of bytes read.
func decInt(data []byte) (int, int, error) {
	if len(data) < 1 {
		return 0, 0, io.ErrUnexpectedEOF
	}

	byteCount := data[0]

	if byteCount == 0 {
		return 0, 1, nil
	}
	data = data[1:]

	// pull count and sign out of byteCount
	negative := byteCount&infoBitsSign != 0
	byteCount &= infoBitsLen

	// do not examine the 2nd, 3rd, and 4th left-most bits; they are reserved
	// for future use

	if len(data) < int(byteCount) {
		return 0, 0, io.ErrUnexpectedEOF
	}

	intData := data[:byteCount]

	// put missing other bytes back in

	padByte := byte(0x00)
	if negative {
		padByte = 0xff
	}
	for len(intData) < 8 {
		// if we're negative, we need to pad with 0xff bytes, otherwise 0x00
		intData = append([]byte{padByte}, intData...)
	}

	// keep value as uint until we return so we avoid logical shift semantics
	var iVal uint
	iVal |= (uint(intData[0]) << 56)
	iVal |= (uint(intData[1]) << 48)
	iVal |= (uint(intData[2]) << 40)
	iVal |= (uint(intData[3]) << 32)
	iVal |= (uint(intData[4]) << 24)
	iVal |= (uint(intData[5]) << 16)
	iVal |= (uint(intData[6]) << 8)
	iVal |= (uint(intData[7]))

	return int(iVal), int(byteCount + 1), nil
}

// encString encodes a string value as a slice of bytes. The value can
// later be decoded with DecString. Encoded string output starts with an
// integer (as encoded by EncInt) indicating the number of bytes following
// that make up the string, followed by that many bytes containing the string
// encoded as UTF-8.
//
// The output will be variable length; it will contain 8 bytes followed by the
// bytes that make up X characters, where X is the int value contained in the
// first 8 bytes. Due to the specifics of how UTF-8 strings are encoded, this
// may or may not be the actual number of bytes used.
//
// Deprecated: This function has been replaced by [Enc].
func EncString(s string) []byte {
	return encString(s)
}

func encString(s string) []byte {
	enc := make([]byte, 0)

	chCount := 0
	for _, ch := range s {
		chBuf := make([]byte, utf8.UTFMax)
		byteLen := utf8.EncodeRune(chBuf, ch)
		enc = append(enc, chBuf[:byteLen]...)
		chCount++
	}

	countBytes := encInt(chCount)
	enc = append(countBytes, enc...)

	return enc
}

// DecString decodes a string value at the start of the given bytes and
// returns the value and the number of bytes read.
//
// Deprecated: This function has been replaced by [Dec].
func DecString(data []byte) (string, int, error) {
	return decString(data)
}

func decString(data []byte) (string, int, error) {
	if len(data) < 1 {
		return "", 0, io.ErrUnexpectedEOF
	}
	runeCount, n, err := decInt(data)
	if err != nil {
		return "", 0, fmt.Errorf("decoding string rune count: %w", err)
	}
	data = data[n:]

	if runeCount < 0 {
		return "", 0, fmt.Errorf("string rune count < 0: %w", ErrMalformedData)
	}

	readBytes := n

	var sb strings.Builder

	for i := 0; i < runeCount; i++ {
		ch, charBytesRead := utf8.DecodeRune(data)
		if ch == utf8.RuneError {
			if charBytesRead == 0 {
				return "", 0, io.ErrUnexpectedEOF
			} else if charBytesRead == 1 {
				return "", 0, fmt.Errorf("invalid UTF-8 encoding in string: %w", ErrMalformedData)
			} else {
				return "", 0, fmt.Errorf("invalid unicode replacement character in rune: %w", ErrMalformedData)
			}
		}

		sb.WriteRune(ch)
		readBytes += charBytesRead
		data = data[charBytesRead:]
	}

	return sb.String(), readBytes, nil
}

// encBinary encodes a BinaryMarshaler as a slice of bytes. The value can later
// be decoded with DecBinary. Encoded output starts with an integer (as encoded
// by EncBinaryInt) indicating the number of bytes following that make up the
// object, followed by that many bytes containing the encoded value.
//
// The output will be variable length; it will contain 8 bytes followed by the
// number of bytes encoded in those 8 bytes.
//
// Deprecated: This function has been replaced by [Enc].
func EncBinary(b encoding.BinaryMarshaler) []byte {
	return encBinary(b)
}

func encBinary(b encoding.BinaryMarshaler) []byte {
	if b == nil {
		return encNil(0)
	}

	enc, _ := b.MarshalBinary()

	enc = append(encInt(len(enc)), enc...)

	return enc
}

// decBinary decodes a value at the start of the given bytes and calls
// UnmarshalBinary on the provided object with those bytes. If a nil value was
// encoded, then a nil byte slice is passed to the UnmarshalBinary func.
//
// It returns the total number of bytes read from the data bytes.
//
// Deprecated: this function has been replaced by [Dec].
func DecBinary(data []byte, b encoding.BinaryUnmarshaler) (int, error) {
	return decBinary(data, b)
}

func decBinary(data []byte, b encoding.BinaryUnmarshaler) (int, error) {
	var readBytes int
	var byteLen int
	var err error

	byteLen, readBytes, err = decInt(data)
	if err != nil {
		return 0, err
	}

	data = data[readBytes:]

	if len(data) < byteLen {
		return readBytes, io.ErrUnexpectedEOF
	}
	var binData []byte

	if byteLen >= 0 {
		binData = data[:byteLen]
	}

	err = b.UnmarshalBinary(binData)
	if err != nil {
		return readBytes, err
	}

	return byteLen + readBytes, nil
}

// if ti.Indir > 0, this will assign to the interface at the appropriate
// indirection level. If ti.Indir == 0, this will not assign. Callers should use
// that check to determine if it is safe to do their own assignment of the
// decoded value this function returns.
func decWithIndirectAssignment[E any](data []byte, v interface{}, ti typeInfo, decFn func([]byte) (E, int, error)) (decoded E, n int, err error) {
	var isNil bool
	var nilLevel int

	if ti.Indir > 0 {
		isNil, _, nilLevel, n, err = decNilableInt(data)
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
