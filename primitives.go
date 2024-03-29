package rezi

// basictypes.go contains functions for encoding and decoding ints, strings,
// bools, and objects that directly implement BinaryUnmarshaler and
// BinaryMarshaler.

import (
	"encoding"
	"fmt"
	"io"
	"math"
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

// some constants for IEEE-754 float representation
const (
	ieee754NegativeBits = 0x8000000000000000
	ieee754ExponentBits = 0x7ff0000000000000
	ieee754MantissaBits = 0x000fffffffffffff
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

// anyFloat is a union interface that combines anyFloat-types. It allows float32
// and float64.
type anyFloat interface {
	float32 | float64
}

// anyComplex is a union interface that combines anyComplex-types. It allows complex64
// and complex128.
type anyComplex interface {
	complex64 | complex128
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
func encCheckedPrim(value analyzed[any]) ([]byte, error) {
	switch value.info.Main {
	case mtString:
		return encWithNilCheck(value, nilErrEncoder(encString), reflect.Value.String)
	case mtBool:
		return encWithNilCheck(value, nilErrEncoder(encBool), reflect.Value.Bool)
	case mtIntegral:
		if value.info.Signed {
			switch value.info.Bits {
			case 8:
				return encWithNilCheck(value, nilErrEncoder(encInt[int8]), func(r reflect.Value) int8 {
					return int8(r.Int())
				})
			case 16:
				return encWithNilCheck(value, nilErrEncoder(encInt[int16]), func(r reflect.Value) int16 {
					return int16(r.Int())
				})
			case 32:
				return encWithNilCheck(value, nilErrEncoder(encInt[int32]), func(r reflect.Value) int32 {
					return int32(r.Int())
				})
			case 64:
				return encWithNilCheck(value, nilErrEncoder(encInt[int64]), reflect.Value.Int)
			default:
				return encWithNilCheck(value, nilErrEncoder(encInt[int]), func(r reflect.Value) int {
					return int(r.Int())
				})
			}
		} else {
			switch value.info.Bits {
			case 8:
				return encWithNilCheck(value, nilErrEncoder(encInt[uint8]), func(r reflect.Value) uint8 {
					return uint8(r.Uint())
				})
			case 16:
				return encWithNilCheck(value, nilErrEncoder(encInt[uint16]), func(r reflect.Value) uint16 {
					return uint16(r.Uint())
				})
			case 32:
				return encWithNilCheck(value, nilErrEncoder(encInt[uint32]), func(r reflect.Value) uint32 {
					return uint32(r.Uint())
				})
			case 64:
				return encWithNilCheck(value, nilErrEncoder(encInt[uint64]), reflect.Value.Uint)
			default:
				return encWithNilCheck(value, nilErrEncoder(encInt[uint]), func(r reflect.Value) uint {
					return uint(r.Uint())
				})
			}
		}
	case mtFloat:
		switch value.info.Bits {
		case 32:
			return encWithNilCheck(value, nilErrEncoder(encFloat[float32]), func(r reflect.Value) float32 {
				return float32(r.Float())
			})
		default:
			fallthrough
		case 64:
			return encWithNilCheck(value, nilErrEncoder(encFloat[float64]), reflect.Value.Float)
		}
	case mtComplex:
		switch value.info.Bits {
		case 64:
			return encWithNilCheck(value, nilErrEncoder(encComplex[complex64]), func(r reflect.Value) complex64 {
				return complex64(r.Complex())
			})
		default:
			fallthrough
		case 128:
			return encWithNilCheck(value, nilErrEncoder(encComplex[complex128]), reflect.Value.Complex)
		}
	case mtBinary:
		return encWithNilCheck(value, encBinary, func(r reflect.Value) encoding.BinaryMarshaler {
			return r.Interface().(encoding.BinaryMarshaler)
		})
	case mtText:
		return encWithNilCheck(value, encText, func(r reflect.Value) encoding.TextMarshaler {
			return r.Interface().(encoding.TextMarshaler)
		})
	default:
		panic(fmt.Sprintf("%T cannot be encoded as REZI primitive type", value))
	}
}

// zeroIndirAssign performs the assignment of decoded to v, performing a type
// conversion if needed.
func zeroIndirAssign[E any](dec decoded[E], recv analyzed[any]) {
	if recv.info.Underlying {
		// need to get fancier
		refVal := recv.reflect
		refVal.Elem().Set(dec.reflect.Convert(refVal.Type().Elem()))
	} else {
		tVal := recv.v.(*E)
		*tVal = dec.v
	}
}

// decCheckedPrim decodes a primitive value from rezi-format bytes into the
// value pointed-to by v. V must point to a REZI primitive value (int, bool,
// string, float, complex), or implement encoding.BinaryUnmarshaler, or
// implement encoding.TextUnmarshaler, or be a pointer to one of those types
// with any level of indirection.
func decCheckedPrim(data []byte, recv analyzed[any]) (decoded[any], error) {
	// by nature of doing an encoding, v must be a pointer to the typeinfo type,
	// or an implementor of BinaryUnmarshaler.
	var dec decoded[any]

	switch recv.info.Main {
	case mtString:
		s, err := decWithNilCheck(data, recv, decString)
		dec.n += s.n
		dec.v = s.v
		if err != nil {
			return dec, err
		}
		if recv.info.Indir == 0 {
			zeroIndirAssign(s, recv)
		}
		return dec, nil
	case mtBool:
		b, err := decWithNilCheck(data, recv, decBool)
		dec.n += b.n
		dec.v = b.v
		if err != nil {
			return dec, err
		}
		if recv.info.Indir == 0 {
			zeroIndirAssign(b, recv)
		}
		return dec, nil
	case mtIntegral:
		if recv.info.Signed {
			switch recv.info.Bits {
			case 64:
				i, err := decWithNilCheck(data, recv, decInt[int64])
				dec.n += i.n
				dec.v = i.v
				if err != nil {
					return dec, err
				}
				if recv.info.Indir == 0 {
					zeroIndirAssign(i, recv)
				}
			case 32:
				i, err := decWithNilCheck(data, recv, decInt[int32])
				dec.n += i.n
				dec.v = i.v
				if err != nil {
					return dec, err
				}
				if recv.info.Indir == 0 {
					zeroIndirAssign(i, recv)
				}
			case 16:
				i, err := decWithNilCheck(data, recv, decInt[int16])
				dec.n += i.n
				dec.v = i.v
				if err != nil {
					return dec, err
				}
				if recv.info.Indir == 0 {
					zeroIndirAssign(i, recv)
				}
			case 8:
				i, err := decWithNilCheck(data, recv, decInt[int8])
				dec.n += i.n
				dec.v = i.v
				if err != nil {
					return dec, err
				}
				if recv.info.Indir == 0 {
					zeroIndirAssign(i, recv)
				}
			default:
				i, err := decWithNilCheck(data, recv, decInt[int])
				dec.n += i.n
				dec.v = i.v
				if err != nil {
					return dec, err
				}
				if recv.info.Indir == 0 {
					zeroIndirAssign(i, recv)
				}
			}
		} else {
			switch recv.info.Bits {
			case 64:
				i, err := decWithNilCheck(data, recv, decInt[uint64])
				dec.n += i.n
				dec.v = i.v
				if err != nil {
					return dec, err
				}
				if recv.info.Indir == 0 {
					zeroIndirAssign(i, recv)
				}
			case 32:
				i, err := decWithNilCheck(data, recv, decInt[uint32])
				dec.n += i.n
				dec.v = i.v
				if err != nil {
					return dec, err
				}
				if recv.info.Indir == 0 {
					zeroIndirAssign(i, recv)
				}
			case 16:
				i, err := decWithNilCheck(data, recv, decInt[uint16])
				dec.n += i.n
				dec.v = i.v
				if err != nil {
					return dec, err
				}
				if recv.info.Indir == 0 {
					zeroIndirAssign(i, recv)
				}
			case 8:
				i, err := decWithNilCheck(data, recv, decInt[uint8])
				dec.n += i.n
				dec.v = i.v
				if err != nil {
					return dec, err
				}
				if recv.info.Indir == 0 {
					zeroIndirAssign(i, recv)
				}
			default:
				i, err := decWithNilCheck(data, recv, decInt[uint])
				dec.n += i.n
				dec.v = i.v
				if err != nil {
					return dec, err
				}
				if recv.info.Indir == 0 {
					zeroIndirAssign(i, recv)
				}
			}
		}

		return dec, nil
	case mtFloat:
		switch recv.info.Bits {
		case 32:
			f, err := decWithNilCheck(data, recv, decFloat[float32])
			dec.n += f.n
			dec.v = f.v
			if err != nil {
				return dec, err
			}
			if recv.info.Indir == 0 {
				zeroIndirAssign(f, recv)
			}
		default:
			fallthrough
		case 64:
			f, err := decWithNilCheck(data, recv, decFloat[float64])
			dec.n += f.n
			dec.v = f.v
			if err != nil {
				return dec, err
			}
			if recv.info.Indir == 0 {
				zeroIndirAssign(f, recv)
			}
		}

		return dec, nil
	case mtComplex:
		switch recv.info.Bits {
		case 64:
			c, err := decWithNilCheck(data, recv, decComplex[complex64])
			dec.n += c.n
			dec.v = c.v
			if err != nil {
				return dec, err
			}
			if recv.info.Indir == 0 {
				zeroIndirAssign(c, recv)
			}
		default:
			fallthrough
		case 128:
			c, err := decWithNilCheck(data, recv, decComplex[complex128])
			dec.n += c.n
			dec.v = c.v
			if err != nil {
				return dec, err
			}
			if recv.info.Indir == 0 {
				zeroIndirAssign(c, recv)
			}
		}

		return dec, nil
	case mtBinary:
		b, err := decWithNilCheck(data, recv, fn_DecToWrappedReceiver(recv,
			func(t reflect.Type) bool {
				return t.Implements(refBinaryUnmarshalerType)
			},
			decBinary,
		))
		dec.n += b.n
		dec.v = b.v
		dec.reflect = b.reflect
		if err != nil {
			return dec, err
		}
		if recv.info.Indir == 0 {
			// assume v is a *T, no future-proofing here.

			// due to complicated forcing of decBinary into the decFunc API,
			// we do now have a T (as an interface{}). We must use reflection to
			// assign it.

			// do NOT use zeroIndirAssign; that is only for underlying type
			// detection which we do not need if operating on an mtBinary

			refReceiver := recv.reflect
			refReceiver.Elem().Set(dec.reflect)
		}
		return dec, nil
	case mtText:
		t, err := decWithNilCheck(data, recv, fn_DecToWrappedReceiver(recv,
			func(t reflect.Type) bool {
				return t.Implements(refTextUnmarshalerType)
			},
			decText,
		))
		dec.n += t.n
		dec.v = t.v
		dec.reflect = t.reflect
		if err != nil {
			return dec, err
		}
		if recv.info.Indir == 0 {
			// assume v is a *T, no future-proofing here.

			// due to complicated forcing of decText into the decFunc API,
			// we do now have a T (as an interface{}). We must use reflection to
			// assign it.

			// do NOT use zeroIndirAssign; that is only for underlying type
			// detection which we do not need if operating on an metText

			refReceiver := recv.reflect
			refReceiver.Elem().Set(dec.reflect)
		}
		return dec, nil
	default:
		panic(fmt.Sprintf("%T cannot receive decoded REZI primitive type", recv.v))
	}
}

// Negative, NilAt, and Length from extra are all ignored.
func encCount(count tLen, extra *countHeader) []byte {
	intBytes := encInt(analyzed[tLen]{v: count})

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
//
// ref in the returned decValue will not be set.
func decCountHeader(data []byte) (decoded[countHeader], error) {
	var hdr decoded[countHeader]

	if len(data) < 1 {
		return hdr, errorDecf(0, "%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}

	err := hdr.v.UnmarshalBinary(data)
	hdr.n = hdr.v.DecodedCount
	return hdr, err
}

// does not actually use analysis data, only native value. accepts
// analyzed[bool] only to conform to encFunc.
func encBool(value analyzed[bool]) []byte {
	b := value.v
	enc := make([]byte, 1)

	if b {
		enc[0] = 1
	} else {
		enc[0] = 0
	}

	return enc
}

// returned decValue does not set ref automatically.
func decBool(data []byte) (decoded[bool], error) {
	if len(data) < 1 {
		return d(false, 0), errorDecf(0, "%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}

	if data[0] == 0 {
		return d(false, 1), nil
	} else if data[0] == 1 {
		return d(true, 1), nil
	} else {
		return d(false, 0), errorDecf(0, "not a bool value 0x00 or 0x01: %#02x", data[0]).wrap(ErrMalformedData)
	}
}

// does not actually use analysis data, only native value. accepts
// analyzed[anyComplex] only to conform to encFunc.
func encComplex[E anyComplex](value analyzed[E]) []byte {
	v := value.v

	// go 1.18 compat, real() and imag() cannot be done to our E type
	v128 := complex128(v)

	rv := real(v128)
	iv := imag(v128)

	// first off, if both real and imaginary parts are +/-0.0, we can encode as
	// single-byte values
	if rv == 0.0 && iv == 0.0 {
		if math.Signbit(rv) && math.Signbit(iv) {
			return []byte{0x80}
		} else if !math.Signbit(rv) && !math.Signbit(iv) {
			return []byte{0x00}
		}
	}

	// encode the parts
	realEnc := encFloat(analyzed[float64]{v: rv})
	imagEnc := encFloat(analyzed[float64]{v: iv})

	hdrEnc := encCount(len(realEnc)+len(imagEnc), &countHeader{ByteLength: true})

	enc := hdrEnc
	enc = append(enc, realEnc...)
	enc = append(enc, imagEnc...)

	return enc
}

// returned decValue does not set ref automatically.
func decComplex[E anyComplex](data []byte) (decoded[E], error) {
	if len(data) < 1 {
		return d(E(0.0), 0), errorDecf(0, "%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}

	// special case single-byte 0's check
	if data[0] == 0x00 {
		return d(E(0.0+0.0i), 1), nil
	} else if data[0] == 0x80 {
		// only way to reliably get a -0.0 value is by direct calculation on var
		// (cannot be result of consts, I tried, at least as of Go 1.19.4)
		var val float64
		val *= -1.0
		return d(E(complex(val, val)), 1), nil
	}

	// do normal decoding of full-form
	var err error
	var offset int

	// get the byte count as an int
	byteCount, err := decInt[tLen](data[offset:])
	if err != nil {
		return d(E(0.0+0.0i), 0), err
	}
	offset += byteCount.n

	// count check
	if len(data[offset:]) < byteCount.v {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded complex value byte count is %d but only %d byte%s remain%s at offset"
		err := errorDecf(offset, errFmt, byteCount.v, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return d(E(0.0+0.0i), 0), err
	}

	// clamp data to len
	data = data[:offset+byteCount.v]

	// real part
	rPart, err := decFloat[float64](data[offset:])
	if err != nil {
		return d(E(0.0+0.0i), offset), errorDecf(offset, "%s", err)
	}
	offset += rPart.n

	// imaginary part
	iPart, err := decFloat[float64](data[offset:])
	if err != nil {
		return d(E(0.0+0.0i), offset), errorDecf(offset, "%s", err)
	}
	offset += iPart.n

	var v128 complex128 = complex(rPart.v, iPart.v)

	return d(E(v128), offset), nil
}

// does not actually use analysis data, only native value. accepts
// analyzed[anyFloat] only to conform to encFunc.
func encFloat[E anyFloat](value analyzed[E]) []byte {
	v := value.v

	// first off, if it is 0, than we can return special 0-value
	if v == 0.0 {
		if math.Signbit(float64(v)) {
			return []byte{0x80}
		} else {
			return []byte{0x00}
		}
	}

	i := math.Float64bits(float64(v))

	// get its parts
	signPart := i & ieee754NegativeBits
	expoPart := i & ieee754ExponentBits
	mantPart := i & ieee754MantissaBits

	// sign is encoded into the count.
	//
	//	[ INFO ] [ COMP-EXPONENT-HIGHS ] [ MIXED ] [ MANTISSA-LOWS ]
	//  SXNILLLL     CEEEEEEE            EEEEMMMM  MMMMMMMM MMMMMMMM MMMMMMMM MMMMMMMM MMMMMMMM MMMMMMMM

	// first, split the exponent part into 7-bits and 4-bits
	expoHigh7 := byte((expoPart >> 56) & 0x7f)
	expoLow4 := byte((expoPart >> 52) & 0x0f)

	// next, split out the mantissa into 4-bits and 48 bits.
	mantHigh4 := byte((mantPart >> 48) & 0x0f)
	mantLow48b1 := byte((mantPart >> 40) & 0xff)
	mantLow48b2 := byte((mantPart >> 32) & 0xff)
	mantLow48b3 := byte((mantPart >> 24) & 0xff)
	mantLow48b4 := byte((mantPart >> 16) & 0xff)
	mantLow48b5 := byte((mantPart >> 8) & 0xff)
	mantLow48b6 := byte(mantPart & 0xff)

	// great, we now have all of our parts.

	// analyze the mantissa
	mantLow48 := []byte{mantLow48b1, mantLow48b2, mantLow48b3, mantLow48b4, mantLow48b5, mantLow48b6}
	var hitMSBAfter int
	for i := range mantLow48 {
		if mantLow48[i] != 0x00 {
			break
		} else {
			hitMSBAfter++
		}
	}
	var hitLSBAfter int
	for i := range mantLow48 {
		if mantLow48[len(mantLow48)-(i+1)] != 0x00 {
			break
		} else {
			hitLSBAfter++
		}
	}
	useLSBCompaction := hitLSBAfter > hitMSBAfter

	// okay, now ready to start building encoded bytes

	// build COMP-EXPONENT-HIGHS byte CEEEEEEE
	var compactionStyle byte = 0x00
	if useLSBCompaction {
		compactionStyle = 0x80
	}

	var compExpoHighs byte = compactionStyle | expoHigh7

	// build MIXED byte EEEEMMMM
	var mixed byte = (expoLow4 << 4) | mantHigh4

	var encMantLows []byte
	var hitSigByte bool
	for i := range mantLow48 {
		mantPart := mantLow48[i]
		if useLSBCompaction {
			mantPart = mantLow48[len(mantLow48)-(i+1)]
		}

		if hitSigByte {
			if useLSBCompaction {
				encMantLows = append([]byte{mantPart}, encMantLows...)
			} else {
				encMantLows = append(encMantLows, mantPart)
			}
		} else if mantPart != 0x00 {
			hitSigByte = true
			if useLSBCompaction {
				encMantLows = append([]byte{mantPart}, encMantLows...)
			} else {
				encMantLows = append(encMantLows, mantPart)
			}
		}
	}

	// put it all into enc
	enc := []byte{compExpoHighs, mixed}
	enc = append(enc, encMantLows...)

	byteCount := uint8(len(enc))

	// byteCount will never be more than 8 so we can encode sign info in most
	// significant bit
	if signPart&ieee754NegativeBits == ieee754NegativeBits {
		byteCount |= infoBitsSign
	}

	enc = append([]byte{byteCount}, enc...)

	return enc
}

// returned decValue does not set ref automatically.
func decFloat[E anyFloat](data []byte) (decoded[E], error) {
	if len(data) < 1 {
		return d(E(0.0), 0), errorDecf(0, "%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}

	byteCount := data[0]

	// special case single-byte 0's check
	if byteCount == 0 {
		return d(E(0.0), 1), nil
	} else if byteCount == 0x80 {
		var val float64
		val *= -1.0
		return d(E(val), 1), nil
	}

	// pull count and sign out of byteCount
	negative := byteCount&infoBitsSign != 0
	byteCount &= infoBitsLen

	// interpretation of other parts of the count header is handled in different
	// functions. skip over all extension bytes
	numHeaderBytes := 0
	for data[0]&infoBitsExt != 0 {
		if len(data[1:]) < 1 {
			const errFmt = "count header indicates extension byte follows, but at end of data"
			err := errorDecf(numHeaderBytes, errFmt).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
			return d(E(0.0), 0), err
		}
		data = data[1:]
		numHeaderBytes++
	}

	// done reading count header info; move past the last byte of it and
	// interpret data bytes
	data = data[1:]
	numHeaderBytes++

	// it could still have been a zero or neg zero. check now
	if int(byteCount) == 0 {
		var val float64
		if negative {
			val *= -1.0
		}
		return d(E(val), numHeaderBytes), nil
	}

	if int(byteCount) < 2 {
		// the absolute minimum is 2 if not 0
		const errFmt = "min data len for non-zero float is 2, but count from header specifies len of %d starting at offset"
		err := errorDecf(numHeaderBytes, errFmt, int(byteCount)).wrap(ErrMalformedData)
		return d(E(0.0), 0), err
	}

	if len(data) < int(byteCount) {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded float byte count is %d but only %d byte%s remain%s at offset"
		err := errorDecf(numHeaderBytes, errFmt, byteCount, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return d(E(0.0), 0), err
	}

	floatData := data[:byteCount]
	compExpoHighs := floatData[0]
	mixed := floatData[1]
	useLSBCompaction := compExpoHighs&0x80 == 0x80

	// we are about to modify mantissaLows, possibly with append operations. we
	// must therefore enshore we don't modify the underlying data storage of
	// data. we will do this by copying into a new slice if we are about to do
	// an append.
	var mantissaLows []byte
	if useLSBCompaction {
		mantissaLows = make([]byte, len(floatData[2:]))
		copy(mantissaLows, floatData[2:])
	} else {
		// otherwise, perfectly safe to start this as a slice-child of
		// floatData.
		mantissaLows = floatData[2:]
	}

	// put compacted other bytes back in
	for len(mantissaLows) < 6 {
		if useLSBCompaction {
			mantissaLows = append(mantissaLows, 0x00)
		} else {
			mantissaLows = append([]byte{0x00}, mantissaLows...)
		}
	}

	// now reconstruct original byte layout of the float
	var signBit byte
	if negative {
		signBit = 0x80
	}

	compExpoHighs &= 0x7f
	compExpoHighs |= signBit

	// place complete result into a uint64 so we can send it to bit-based
	// interpretation and to avoid logical shift semantics

	var iVal uint64
	iVal |= (uint64(compExpoHighs) << 56)
	iVal |= (uint64(mixed) << 48)
	iVal |= (uint64(mantissaLows[0]) << 40)
	iVal |= (uint64(mantissaLows[1]) << 32)
	iVal |= (uint64(mantissaLows[2]) << 24)
	iVal |= (uint64(mantissaLows[3]) << 16)
	iVal |= (uint64(mantissaLows[4]) << 8)
	iVal |= (uint64(mantissaLows[5]))

	fVal := math.Float64frombits(iVal)

	return d(E(fVal), int(byteCount)+numHeaderBytes), nil
}

// does not actually use analysis data, only native value. accepts
// analyzed[integral] only to conform to encFunc.
func encInt[E integral](value analyzed[E]) []byte {
	v := value.v
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
//
// returned decValue does not set ref automatically.
func decInt[E integral](data []byte) (decoded[E], error) {
	if len(data) < 1 {
		return d(E(0), 0), errorDecf(0, "%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}

	byteCount := data[0]

	if byteCount == 0 {
		return d(E(0), 1), nil
	}

	// pull count and sign out of byteCount
	negative := byteCount&infoBitsSign != 0
	byteCount &= infoBitsLen

	// interpretation of other parts of the count header is handled in different
	// functions. skip over all extension bytes
	numHeaderBytes := 0
	for data[0]&infoBitsExt != 0 {
		if len(data[1:]) < 1 {
			const errFmt = "count header indicates extension byte follows, but at end of data"
			err := errorDecf(numHeaderBytes, errFmt).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
			return d(E(0), 0), err
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
		const errFmt = "decoded int byte count is %d but only %d byte%s remain%s at offset"
		err := errorDecf(numHeaderBytes, errFmt, byteCount, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return d(E(0), 0), err
	}

	intData := data[:byteCount]

	// put missing other bytes back in

	padByte := byte(0x00)
	if negative {
		padByte = 0xff
	}
	for len(intData) < 8 {
		// if we're negative, we need to pad with 0xff bytes, otherwise 0x00.
		intData = append([]byte{padByte}, intData...)

		// NOTE: this has no chance of modifying the original data slice bc it
		// is appending to a brand new slice. if we were appending to the END,
		// this could modify the underlying storage.
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

	return d(E(iVal), int(byteCount)+numHeaderBytes), nil
}

// does not actually use analysis data, only native value. accepts
// analyzed[string] only to conform to encFunc.
func encString(value analyzed[string]) []byte {
	s := value.v

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
//
// returned decValue does not set ref automatically.
func decString(data []byte) (decoded[string], error) {
	if len(data) < 1 {
		return d("", 0), errorDecf(0, "%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}

	// special case; 0x00 is the empty string in all variants
	if data[0] == 0 {
		return d("", 1), nil
	}

	hdr, err := decCountHeader(data)
	if err != nil {
		return d("", 0), err
	}

	// compatibility with older format
	if !hdr.v.ByteLength {
		return decStringV0(data)
	}

	return decStringV1(data)
}

// returned decValue does not set ref automatically.
func decStringV1(data []byte) (decoded[string], error) {
	if len(data) < 1 {
		return d("", 0), errorDecf(0, "%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}
	strLengthDec, err := decInt[tLen](data)
	if err != nil {
		return d("", 0), errorDecf(0, "decode string byte count: %s", err)
	}
	strLength := strLengthDec.v
	countLen := strLengthDec.n
	data = data[countLen:]

	if strLength < 0 {
		return d("", 0), errorDecf(countLen, "string byte count < 0").wrap(ErrMalformedData)
	}

	if len(data) < strLength {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded string byte count is %d but only %d byte%s remain%s at offset"
		err := errorDecf(countLen, errFmt, strLength, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return d("", 0), err
	}
	// clamp it
	data = data[:strLength]

	readBytes := countLen

	var sb strings.Builder
	for readBytes-countLen < strLength {
		ch, charBytesRead, err := decUTF8Codepoint(data)
		if err != nil {
			return d("", 0), errorDecf(readBytes, "%s", err)
		}

		sb.WriteRune(ch)
		readBytes += charBytesRead
		data = data[charBytesRead:]
	}

	return d(sb.String(), readBytes), nil
}

// returned decValue does not set ref automatically.
func decStringV0(data []byte) (decoded[string], error) {
	if len(data) < 1 {
		return d("", 0), errorDecf(0, "%s", io.ErrUnexpectedEOF).wrap(ErrMalformedData)
	}
	runeCountDec, err := decInt[int](data)
	if err != nil {
		return d("", 0), errorDecf(0, "decode string rune count: %s", err)
	}
	n := runeCountDec.n
	runeCount := runeCountDec.v
	data = data[n:]

	if runeCount < 0 {
		return d("", 0), errorDecf(0, "string rune count < 0").wrap(ErrMalformedData)
	}

	readBytes := n

	var sb strings.Builder

	for i := 0; i < runeCount; i++ {
		ch, charBytesRead, err := decUTF8Codepoint(data)
		if err != nil {
			return d("", 0), errorDecf(readBytes, "%s", err)
		}

		sb.WriteRune(ch)
		readBytes += charBytesRead
		data = data[charBytesRead:]
	}

	return d(sb.String(), readBytes), nil
}

func decUTF8Codepoint(data []byte) (rune, int, error) {
	ch, charBytesRead := utf8.DecodeRune(data)
	if ch == utf8.RuneError {
		if charBytesRead == 0 {
			return ch, 0, errorDecf(0, "bytes could not be read as UTF-8 codepoint data").wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		} else if charBytesRead == 1 {
			return ch, 0, errorDecf(0, "invalid UTF-8 encoding in string").wrap(ErrMalformedData)
		} else {
			return ch, 0, errorDecf(0, "invalid unicode replacement character in rune").wrap(ErrMalformedData)
		}
	}
	return ch, charBytesRead, nil
}

// does not actually use analysis data, only native value. accepts
// analyzed[encoding.TextMarshaler] only to conform to encFunc.
func encText(value analyzed[encoding.TextMarshaler]) ([]byte, error) {
	t := value.v

	if t == nil {
		return encNilHeader(0), nil
	}

	tTextSlice, marshalErr := t.MarshalText()
	if marshalErr != nil {
		return nil, errorf("%s: %s", ErrMarshalText, marshalErr)
	}
	tText := string(tTextSlice)

	return encString(analyzed[string]{v: tText}), nil
}

func decText(data []byte, recv analyzed[any]) (decoded[any], error) {
	t := recv.v.(encoding.TextUnmarshaler)

	var textData decoded[string]
	var dec decoded[any]
	var err error

	textData, err = decString(data)
	if err != nil {
		return dec, errorDecf(0, "decode text: %s", err).wrap(ErrMalformedData)
	}
	dec.n = textData.n

	err = t.UnmarshalText([]byte(textData.v))
	if err != nil {
		return dec, errorDecf(0, "%s: %s", ErrUnmarshalText, err).wrap(ErrMalformedData)
	}
	dec.v = t

	return dec, nil
}

// does not actually use analysis data, only native value. accepts
// analyzed[encoding.BinaryMarshaler] only to conform to encFunc.
func encBinary(value analyzed[encoding.BinaryMarshaler]) ([]byte, error) {
	b := value.v

	if b == nil {
		return encNilHeader(0), nil
	}

	enc, marshalErr := b.MarshalBinary()
	if marshalErr != nil {
		return nil, errorf("%s: %s", ErrMarshalBinary, marshalErr)
	}

	enc = append(encCount(len(enc), nil), enc...)

	return enc, nil
}

func decBinary(data []byte, recv analyzed[any]) (decoded[any], error) {
	b := recv.v.(encoding.BinaryUnmarshaler)

	var dec decoded[any]
	var err error

	byteLen, err := decInt[tLen](data)
	if err != nil {
		return decoded[any]{n: byteLen.n}, errorDecf(0, "decode byte count: %s", err)
	}
	dec.n = byteLen.n
	data = data[dec.n:]

	if len(data) < byteLen.v {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded binary value byte count is %d but only %d byte%s remain%s at offset"
		err := errorDecf(dec.n, errFmt, byteLen.v, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return dec, err
	}
	var binData []byte

	if byteLen.v >= 0 {
		binData = data[:byteLen.v]
	}

	err = b.UnmarshalBinary(binData)
	if err != nil {
		return dec, errorDecf(dec.n, "%s: %s", ErrUnmarshalBinary, err)
	}
	dec.v = b
	dec.n += byteLen.v

	return dec, nil
}
