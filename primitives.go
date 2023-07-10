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

	// used only in extension byte 1:
	infoBitsByteCount = 0b10000000
	infoBitsVersion   = 0b00001111
	// extension bit not listed because it is the same
)

type anyUint interface {
	uint | uint8 | uint16 | uint32 | uint64
}

type anyInt interface {
	int | int8 | int16 | int32 | int64
}

// integral is a union interface that combines all basic Go integer types. It
// allows int, uint, and all of their specifically-sized varieties.
type integral interface {
	anyInt | anyUint
}

// encCheckedPrim encodes the primitve REZI value as rezi-format bytes. The type
// of the value is examined to determine how to encode it. No type information
// is included in the returned bytes so it is up to the caller to keep track of
// it.
//
// This function takes type info for a primitive and encodes it. The value can
// have any level of pointer indirection and will be correctly encoded as the
// value that the eventual pointed-to element is, or a nil indicating the
// correct level of indirection of pointer that the passed-in pointer was nil
// at, which is retrieved by a call to decCheckedPrim with a pointer to *that*
// type.
func encCheckedPrim(value interface{}, ti typeInfo) ([]byte, error) {
	switch ti.Main {
	case mtString:
		return encWithNilCheck(value, ti, nilErrEncoder(encString), reflect.Value.String)
	case mtBool:
		return encWithNilCheck(value, ti, nilErrEncoder(encBool), reflect.Value.Bool)
	case mtIntegral:
		if ti.Signed {
			switch ti.Bits {
			case 8:
				return encWithNilCheck(value, ti, nilErrEncoder(encInt[int8]), func(r reflect.Value) int8 {
					return int8(r.Int())
				})
			case 16:
				return encWithNilCheck(value, ti, nilErrEncoder(encInt[int16]), func(r reflect.Value) int16 {
					return int16(r.Int())
				})
			case 32:
				return encWithNilCheck(value, ti, nilErrEncoder(encInt[int32]), func(r reflect.Value) int32 {
					return int32(r.Int())
				})
			case 64:
				return encWithNilCheck(value, ti, nilErrEncoder(encInt[int64]), func(r reflect.Value) int64 {
					return int64(r.Int())
				})
			default:
				return encWithNilCheck(value, ti, nilErrEncoder(encInt[int]), func(r reflect.Value) int {
					return int(r.Int())
				})
			}
		} else {
			switch ti.Bits {
			case 8:
				return encWithNilCheck(value, ti, nilErrEncoder(encInt[uint8]), func(r reflect.Value) uint8 {
					return uint8(r.Uint())
				})
			case 16:
				return encWithNilCheck(value, ti, nilErrEncoder(encInt[uint16]), func(r reflect.Value) uint16 {
					return uint16(r.Uint())
				})
			case 32:
				return encWithNilCheck(value, ti, nilErrEncoder(encInt[uint32]), func(r reflect.Value) uint32 {
					return uint32(r.Uint())
				})
			case 64:
				return encWithNilCheck(value, ti, nilErrEncoder(encInt[uint64]), func(r reflect.Value) uint64 {
					return uint64(r.Uint())
				})
			default:
				return encWithNilCheck(value, ti, nilErrEncoder(encInt[uint]), func(r reflect.Value) uint {
					return uint(r.Uint())
				})
			}
		}
	case mtBinary:
		return encWithNilCheck(value, ti, encBinary, func(r reflect.Value) encoding.BinaryMarshaler {
			return r.Interface().(encoding.BinaryMarshaler)
		})
	default:
		panic(fmt.Sprintf("%T cannot be encoded as REZI primitive type", value))
	}
}

// decCheckedPrim decodes a primitive value from rezi-format bytes into the
// value pointed-to by v. V must point to a REZI primitive value (int, bool,
// string), or implement encoding.BinaryUnmarshaler, or be a pointer to one of
// those types with any level of indirection.
func decCheckedPrim(data []byte, v interface{}, ti typeInfo) (int, error) {
	// by nature of doing an encoding, v MUST be a pointer to the typeinfo type,
	// or an implementor of BinaryUnmarshaler.

	switch ti.Main {
	case mtString:
		s, n, err := decWithNilCheck(data, v, ti, decString)
		if err != nil {
			return n, err
		}
		if ti.Indir == 0 {
			tVal := v.(*string)
			*tVal = s
		}
		return n, nil
	case mtBool:
		b, n, err := decWithNilCheck(data, v, ti, decBool)
		if err != nil {
			return n, err
		}
		if ti.Indir == 0 {
			tVal := v.(*bool)
			*tVal = b
		}
		return n, nil
	case mtIntegral:
		var n int
		var err error

		if ti.Signed {
			switch ti.Bits {
			case 64:
				var i int64
				i, n, err = decWithNilCheck(data, v, ti, decInt[int64])
				if err != nil {
					return n, err
				}
				if ti.Indir == 0 {
					tVal := v.(*int64)
					*tVal = i
				}
			case 32:
				var i int32
				i, n, err = decWithNilCheck(data, v, ti, decInt[int32])
				if err != nil {
					return n, err
				}
				if ti.Indir == 0 {
					tVal := v.(*int32)
					*tVal = i
				}
			case 16:
				var i int16
				i, n, err = decWithNilCheck(data, v, ti, decInt[int16])
				if err != nil {
					return n, err
				}
				if ti.Indir == 0 {
					tVal := v.(*int16)
					*tVal = i
				}
			case 8:
				var i int8
				i, n, err = decWithNilCheck(data, v, ti, decInt[int8])
				if err != nil {
					return n, err
				}
				if ti.Indir == 0 {
					tVal := v.(*int8)
					*tVal = i
				}
			default:
				var i int
				i, n, err = decWithNilCheck(data, v, ti, decInt[int])
				if err != nil {
					return n, err
				}
				if ti.Indir == 0 {
					tVal := v.(*int)
					*tVal = i
				}
			}
		} else {
			switch ti.Bits {
			case 64:
				var i uint64
				i, n, err = decWithNilCheck(data, v, ti, decInt[uint64])
				if err != nil {
					return n, err
				}
				if ti.Indir == 0 {
					tVal := v.(*uint64)
					*tVal = i
				}
			case 32:
				var i uint32
				i, n, err = decWithNilCheck(data, v, ti, decInt[uint32])
				if err != nil {
					return n, err
				}
				if ti.Indir == 0 {
					tVal := v.(*uint32)
					*tVal = i
				}
			case 16:
				var i uint16
				i, n, err = decWithNilCheck(data, v, ti, decInt[uint16])
				if err != nil {
					return n, err
				}
				if ti.Indir == 0 {
					tVal := v.(*uint16)
					*tVal = i
				}
			case 8:
				var i uint8
				i, n, err = decWithNilCheck(data, v, ti, decInt[uint8])
				if err != nil {
					return n, err
				}
				if ti.Indir == 0 {
					tVal := v.(*uint8)
					*tVal = i
				}
			default:
				var i uint
				i, n, err = decWithNilCheck(data, v, ti, decInt[uint])
				if err != nil {
					return n, err
				}
				if ti.Indir == 0 {
					tVal := v.(*uint)
					*tVal = i
				}
			}
		}

		return n, nil
	case mtBinary:
		// if we just got handed a pointer-to binaryUnmarshaler, we need to undo
		// that
		bu, n, err := decWithNilCheck(data, v, ti, fn_DecToWrappedReceiver(v, ti,
			func(t reflect.Type) bool {
				return t.Implements(refBinaryUnmarshalerType)
			},
			func(b []byte, unwrapped interface{}) (int, error) {
				recv := unwrapped.(encoding.BinaryUnmarshaler)
				return decBinary(b, recv)
			},
		))
		if err != nil {
			return n, err
		}
		if ti.Indir == 0 {
			// assume v is a *T, no future-proofing here.

			// due to complicated forcing of decBinary into the decFunc API,
			// we do now have a T (as an interface{}). We must use reflection to
			// assign it.

			refReceiver := reflect.ValueOf(v)
			refReceiver.Elem().Set(reflect.ValueOf(bu))
		}
		return n, nil
	default:
		panic(fmt.Sprintf("%T cannot receive decoded REZI primitive type", v))
	}
}

// Negative, NilAt, and Length from extra are all ignored.
func encCount(count tLen, extra *countHeader) []byte {
	intBytes := encInt(count)

	if extra == nil {
		// normal int enc
		return intBytes
	}

	hdr := countHeader{
		Negative:       false,
		NilAt:          0,
		Length:         int(intBytes[0] & infoBitsLen),
		ExtensionLevel: extra.ExtensionLevel,
		Version:        extra.Version,
		ByteLength:     extra.ByteLength,
	}

	hdrBytes, err := hdr.MarshalBinary()
	if err != nil {
		// should never happen
		panic(err.Error())
	}

	var enc []byte

	enc = append(enc, hdrBytes...)
	enc = append(enc, intBytes[1:]...)

	return enc
}

func encNilHeader(indirLevels int) []byte {
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

	// encode it with a count header
	if indirLevels < 0 {
		indirLevels = 0
	}
	hdr := countHeader{
		NilAt: indirLevels + 1,

		// for compat with older format
		Negative: true,
	}

	enc, err := hdr.MarshalBinary()
	if err != nil {
		// should never happen
		panic(fmt.Sprintf("encoding nil-indicating countHeader failed: %s", err.Error()))
	}

	return enc
}

// decCountHeader decodes a count header. It could represent a nil value. It
// will *not* decode the actual count, if in fact the count is present.
func decCountHeader(data []byte) (countHeader, int, error) {
	var hdr countHeader

	if len(data) < 1 {
		return hdr, 0, reziError{
			cause: []error{io.ErrUnexpectedEOF, ErrMalformedData},
		}
	}

	err := hdr.UnmarshalBinary(data)
	return hdr, hdr.DecodedCount, err
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

func decBool(data []byte) (bool, int, error) {
	if len(data) < 1 {
		return false, 0, reziError{cause: []error{io.ErrUnexpectedEOF, ErrMalformedData}}
	}

	if data[0] == 0 {
		return false, 1, nil
	} else if data[0] == 1 {
		return true, 1, nil
	} else {
		return false, 0, errorf("not a bool value 0x00 or 0x01").wrap(ErrMalformedData)
	}
}

func encInt[E integral](v E) []byte {
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

// decInt decodes an integer value at the start of the given bytes and
// returns the value and the number of bytes read.
//
// assumes that first byte specifies a non-nil integer whose L field gives
// number of bytes to decode after all count header bytes and interprets it as
// such. does not do further checks on count header.
func decInt[E integral](data []byte) (E, int, error) {
	if len(data) < 1 {
		return 0, 0, errorf("%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}

	byteCount := data[0]

	if byteCount == 0 {
		return 0, 1, nil
	}

	// pull count and sign out of byteCount
	negative := byteCount&infoBitsSign != 0
	byteCount &= infoBitsLen

	// interpretation of other parts of the count header is handled in different
	// functions. skip over all extension bytes
	numHeaderBytes := 0
	for data[0]&infoBitsExt != 0 {
		if len(data) < 1 {
			s := "s"
			verbS := ""
			if len(data) == 1 {
				s = ""
				verbS = "s"
			}
			const errFmt = "count header length is at least %d but only %d byte%s remain%s in data"
			err := errorf(errFmt, numHeaderBytes+1, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
			return 0, 0, err
		}
		data = data[1:]
		numHeaderBytes++
	}

	// done reading count header info; move past the last byte of it and
	// interpret data bytes
	data = data[1:]
	numHeaderBytes++

	if len(data) < int(byteCount) {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded int byte count is %d but only %d byte%s remain%s in data"
		err := errorf(errFmt, byteCount, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return 0, 0, err
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
	var iVal uint64
	iVal |= (uint64(intData[0]) << 56)
	iVal |= (uint64(intData[1]) << 48)
	iVal |= (uint64(intData[2]) << 40)
	iVal |= (uint64(intData[3]) << 32)
	iVal |= (uint64(intData[4]) << 24)
	iVal |= (uint64(intData[5]) << 16)
	iVal |= (uint64(intData[6]) << 8)
	iVal |= (uint64(intData[7]))

	return E(iVal), int(byteCount) + numHeaderBytes, nil
}

func encString(s string) []byte {
	if s == "" {
		return []byte{0x00}
	}

	strBytes := make([]byte, 0)

	for _, ch := range s {
		chBuf := make([]byte, utf8.UTFMax)
		byteLen := utf8.EncodeRune(chBuf, ch)
		strBytes = append(strBytes, chBuf[:byteLen]...)
	}

	var enc []byte

	enc = append(enc, encCount(len(strBytes), &countHeader{ByteLength: true, Version: 2})...)
	enc = append(enc, strBytes...)

	return enc
}

// decString decodes a string of any version. Assumes header is not nil.
func decString(data []byte) (string, int, error) {
	if len(data) < 1 {
		return "", 0, errorf("%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}

	// special case; 0x00 is the empty string in all variants
	if data[0] == 0 {
		return "", 1, nil
	}

	hdr, _, err := decCountHeader(data)
	if err != nil {
		return "", 0, err
	}

	// compatibility with older format
	if !hdr.ByteLength {
		return decStringV1(data)
	}

	return decStringV2(data)
}

func decStringV2(data []byte) (string, int, error) {
	if len(data) < 1 {
		return "", 0, errorf("%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}
	strLength, countLen, err := decInt[int](data)
	if err != nil {
		return "", 0, errorf("decode string byte count: %s", err)
	}
	data = data[countLen:]

	if strLength < 0 {
		return "", 0, errorf("string byte count < 0").wrap(ErrMalformedData)
	}

	if len(data) < strLength {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded string byte count is %d but only %d byte%s remain%s in data"
		err := errorf(errFmt, strLength, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return "", 0, err
	}
	// clamp it
	data = data[:strLength]

	readBytes := countLen

	var sb strings.Builder
	for readBytes-countLen < strLength {
		ch, charBytesRead, err := decUTF8Codepoint(data)
		if err != nil {
			return "", 0, err
		}

		sb.WriteRune(ch)
		readBytes += charBytesRead
		data = data[charBytesRead:]
	}

	return sb.String(), readBytes, nil
}

func decStringV1(data []byte) (string, int, error) {
	if len(data) < 1 {
		return "", 0, errorf("%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}
	runeCount, n, err := decInt[int](data)
	if err != nil {
		return "", 0, errorf("decoding string rune count: %s", err)
	}
	data = data[n:]

	if runeCount < 0 {
		return "", 0, errorf("string rune count < 0").wrap(ErrMalformedData)
	}

	readBytes := n

	var sb strings.Builder

	for i := 0; i < runeCount; i++ {
		ch, charBytesRead, err := decUTF8Codepoint(data)
		if err != nil {
			return "", 0, err
		}

		sb.WriteRune(ch)
		readBytes += charBytesRead
		data = data[charBytesRead:]
	}

	return sb.String(), readBytes, nil
}

func decUTF8Codepoint(data []byte) (rune, int, error) {
	ch, charBytesRead := utf8.DecodeRune(data)
	if ch == utf8.RuneError {
		if charBytesRead == 0 {
			return ch, 0, errorf("bytes could be read as UTF-8 data for character").wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		} else if charBytesRead == 1 {
			return ch, 0, errorf("invalid UTF-8 encoding in string").wrap(ErrMalformedData)
		} else {
			return ch, 0, errorf("invalid unicode replacement character in rune").wrap(ErrMalformedData)
		}
	}
	return ch, charBytesRead, nil
}

func encBinary(b encoding.BinaryMarshaler) ([]byte, error) {
	if b == nil {
		return encNilHeader(0), nil
	}

	enc, marshalErr := b.MarshalBinary()
	if marshalErr != nil {
		return nil, errorf("%s: %s", ErrMarshalBinary, marshalErr)
	}

	enc = append(encInt(len(enc)), enc...)

	return enc, nil
}

func decBinary(data []byte, b encoding.BinaryUnmarshaler) (int, error) {
	var readBytes int
	var byteLen int
	var err error

	byteLen, readBytes, err = decInt[tLen](data)
	if err != nil {
		return 0, err
	}

	data = data[readBytes:]

	if len(data) < byteLen {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded binary value byte count is %d but only %d byte%s remain%s in data"
		err := errorf(errFmt, byteLen, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return readBytes, err
	}
	var binData []byte

	if byteLen >= 0 {
		binData = data[:byteLen]
	}

	err = b.UnmarshalBinary(binData)
	if err != nil {
		return readBytes, errorf("%s: %s", ErrUnmarshalBinary, err)
	}

	return byteLen + readBytes, nil
}
