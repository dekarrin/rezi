package rezi

// basictypes.go contains functions for encoding and decoding ints, strings,
// bools, and objects that directly implement BinaryUnmarshaler and
// BinaryMarshaler.

import (
	"encoding"
	"fmt"
	"strings"
	"unicode/utf8"
)

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
		return EncString(value.(string))
	case tBool:
		return EncBool(value.(bool))
	case tIntegral:
		if ti.Signed {
			switch ti.Bits {
			case 8:
				return EncInt(int(value.(int8)))
			case 16:
				return EncInt(int(value.(int16)))
			case 32:
				return EncInt(int(value.(int32)))
			case 64:
				return EncInt(int(value.(int64)))
			default:
				return EncInt(value.(int))
			}
		} else {
			switch ti.Bits {
			case 8:
				return EncInt(int(value.(uint8)))
			case 16:
				return EncInt(int(value.(uint16)))
			case 32:
				return EncInt(int(value.(uint32)))
			case 64:
				return EncInt(int(value.(uint64)))
			default:
				return EncInt(int(value.(uint)))
			}
		}
	case tBinary:
		return EncBinary(value.(encoding.BinaryMarshaler))
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
		tVal := v.(*string)
		s, n, err := DecString(data)
		if err != nil {
			return n, err
		}
		*tVal = s
		return n, nil
	case tBool:
		tVal := v.(*bool)
		b, n, err := DecBool(data)
		if err != nil {
			return n, err
		}
		*tVal = b
		return n, nil
	case tIntegral:
		i, n, err := DecInt(data)
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
		return DecBinary(data, receiver)
	default:
		panic(fmt.Sprintf("%T cannot receive decoded REZI primitive type", v))
	}
}

// EncBool encodes the bool value as a slice of bytes. The value can later
// be decoded with DecBool. No type indicator is included in the output;
// it is up to the caller to add this if they so wish it.
//
// The output will always contain exactly 1 byte.
func EncBool(b bool) []byte {
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
func DecBool(data []byte) (bool, int, error) {
	if len(data) < 1 {
		return false, 0, fmt.Errorf("unexpected EOF")
	}

	if data[0] == 0 {
		return false, 1, nil
	} else if data[0] == 1 {
		return true, 1, nil
	} else {
		return false, 0, fmt.Errorf("unknown non-bool value")
	}
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
func EncInt(i int) []byte {
	if i == 0 {
		return []byte{0x00}
	}

	negative := i < 0

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
		byteCount |= 0x80
	}

	enc = append([]byte{byteCount}, enc...)

	return enc
}

// DecInt decodes an integer value at the start of the given bytes and
// returns the value and the number of bytes read.
func DecInt(data []byte) (int, int, error) {
	if len(data) < 1 {
		return 0, 0, fmt.Errorf("data does not contain at least 1 byte")
	}

	byteCount := data[0]

	if byteCount == 0 {
		return 0, 1, nil
	}
	data = data[1:]

	// pull count and sign out of byteCount
	negative := byteCount&0x80 != 0
	byteCount &= 0x0f

	// do not examine the 2nd, 3rd, and 4th left-most bits; they are reserved
	// for future use

	if len(data) < int(byteCount) {
		return 0, 0, fmt.Errorf("unexpected EOF")
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

// EncString encodes a string value as a slice of bytes. The value can
// later be decoded with DecString. Encoded string output starts with an
// integer (as encoded by EncInt) indicating the number of bytes following
// that make up the string, followed by that many bytes containing the string
// encoded as UTF-8.
//
// The output will be variable length; it will contain 8 bytes followed by the
// bytes that make up X characters, where X is the int value contained in the
// first 8 bytes. Due to the specifics of how UTF-8 strings are encoded, this
// may or may not be the actual number of bytes used.
func EncString(s string) []byte {
	enc := make([]byte, 0)

	chCount := 0
	for _, ch := range s {
		chBuf := make([]byte, utf8.UTFMax)
		byteLen := utf8.EncodeRune(chBuf, ch)
		enc = append(enc, chBuf[:byteLen]...)
		chCount++
	}

	countBytes := EncInt(chCount)
	enc = append(countBytes, enc...)

	return enc
}

// DecString decodes a string value at the start of the given bytes and
// returns the value and the number of bytes read.
func DecString(data []byte) (string, int, error) {
	if len(data) < 1 {
		return "", 0, fmt.Errorf("unexpected EOF")
	}
	runeCount, n, err := DecInt(data)
	if err != nil {
		return "", 0, fmt.Errorf("decoding string rune count: %w", err)
	}
	data = data[n:]

	if runeCount < 0 {
		return "", 0, fmt.Errorf("string rune count < 0")
	}

	readBytes := n

	var sb strings.Builder

	for i := 0; i < runeCount; i++ {
		ch, charBytesRead := utf8.DecodeRune(data)
		if ch == utf8.RuneError {
			if charBytesRead == 0 {
				return "", 0, fmt.Errorf("unexpected EOF")
			} else if charBytesRead == 1 {
				return "", 0, fmt.Errorf("invalid UTF-8 encoding in string")
			} else {
				return "", 0, fmt.Errorf("invalid unicode replacement character in rune")
			}
		}

		sb.WriteRune(ch)
		readBytes += charBytesRead
		data = data[charBytesRead:]
	}

	return sb.String(), readBytes, nil
}

// EncBinary encodes a BinaryMarshaler as a slice of bytes. The value can later
// be decoded with DecBinary. Encoded output starts with an integer (as encoded
// by EncBinaryInt) indicating the number of bytes following that make up the
// object, followed by that many bytes containing the encoded value.
//
// The output will be variable length; it will contain 8 bytes followed by the
// number of bytes encoded in those 8 bytes.
func EncBinary(b encoding.BinaryMarshaler) []byte {
	if b == nil {
		return EncInt(-1)
	}

	enc, _ := b.MarshalBinary()

	enc = append(EncInt(len(enc)), enc...)

	return enc
}

// DecBinary decodes a value at the start of the given bytes and calls
// UnmarshalBinary on the provided object with those bytes. If a nil value was
// encoded, then a nil byte slice is passed to the UnmarshalBinary func.
//
// It returns the total number of bytes read from the data bytes.
func DecBinary(data []byte, b encoding.BinaryUnmarshaler) (int, error) {
	var readBytes int
	var byteLen int
	var err error

	byteLen, readBytes, err = DecInt(data)
	if err != nil {
		return 0, err
	}

	data = data[readBytes:]

	if len(data) < byteLen {
		return readBytes, fmt.Errorf("unexpected end of data")
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
