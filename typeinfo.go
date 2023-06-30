package rezi

import (
	"encoding"
	"fmt"
	"reflect"
)

var (
	refBinaryMarshalerType   = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
	refBinaryUnmarshalerType = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
)

type mainType int64

const (
	tUnknown mainType = iota
	tIntegral
	tBool
	tString
	tBinary
	tMap
	tSlice
	tNil
)

// typeInfo holds REZI-specific type info on types that can be encoded and
// decoded.
type typeInfo struct {
	Main      mainType
	Bits      int
	Signed    bool
	ViaNonPtr bool
	Indir     int       // Indir is number of times that the value is deref'd. Used for encoding of ptr-to types.
	KeyType   *typeInfo // only valid for maps
	ValType   *typeInfo // valid for map and slice
}

func (ti typeInfo) Primitive() bool {
	return ti.Main == tIntegral || ti.Main == tBool || ti.Main == tString || ti.Main == tBinary
}

func canEncode(v interface{}) (typeInfo, error) {
	return encTypeInfo(reflect.TypeOf(v))
}

func encTypeInfo(t reflect.Type) (info typeInfo, err error) {
	if t == nil {
		return typeInfo{Main: tNil}, nil
	}

	if t.Implements(refBinaryMarshalerType) {
		return typeInfo{Main: tBinary}, nil
	}

	switch t.Kind() {
	case reflect.String:
		return typeInfo{Main: tString}, nil
	case reflect.Bool:
		return typeInfo{Main: tBool}, nil
	case reflect.Uint8:
		return typeInfo{Main: tIntegral, Bits: 8, Signed: false}, nil
	case reflect.Uint16:
		return typeInfo{Main: tIntegral, Bits: 16, Signed: false}, nil
	case reflect.Uint32:
		return typeInfo{Main: tIntegral, Bits: 32, Signed: false}, nil
	case reflect.Uint64:
		return typeInfo{Main: tIntegral, Bits: 64, Signed: false}, nil
	case reflect.Uint:
		return typeInfo{Main: tIntegral, Bits: 0, Signed: false}, nil
	case reflect.Int8:
		return typeInfo{Main: tIntegral, Bits: 8, Signed: true}, nil
	case reflect.Int16:
		return typeInfo{Main: tIntegral, Bits: 16, Signed: true}, nil
	case reflect.Int32:
		return typeInfo{Main: tIntegral, Bits: 32, Signed: true}, nil
	case reflect.Int64:
		return typeInfo{Main: tIntegral, Bits: 64, Signed: true}, nil
	case reflect.Int:
		return typeInfo{Main: tIntegral, Bits: 0, Signed: true}, nil
	default:
		// is it a map type?
		if t.Kind() == reflect.Map {
			// could be okay, but key and value types must be encodable.
			mValType := t.Elem()
			mKeyType := t.Key()

			mValInfo, err := encTypeInfo(mValType)
			if err != nil {
				return typeInfo{}, fmt.Errorf("map value type is not encodable: %w", err)
			}
			mKeyInfo, err := encTypeInfo(mKeyType)
			if err != nil {
				return typeInfo{}, fmt.Errorf("map key type is not encodable: %w", err)
			}

			// maps in general are not supported; the key type MUST be comparable
			// and with an ordering, which p much means we exclusively support
			// non-binary primitives.
			if !mKeyInfo.Primitive() || mKeyInfo.Main == tBinary {
				return typeInfo{}, fmt.Errorf("map key type must be bool, string, or castable to int")
			}

			return typeInfo{Main: tMap, KeyType: &mKeyInfo, ValType: &mValInfo}, nil
		}

		// is it a slice type?
		if t.Kind() == reflect.Slice {
			// could be okay, but val type must be encodable
			slValType := t.Elem()
			slValInfo, err := encTypeInfo(slValType)
			if err != nil {
				return typeInfo{}, fmt.Errorf("slice value is not encodable: %w", err)
			}
			return typeInfo{Main: tSlice, ValType: &slValInfo}, nil
		}

		return typeInfo{}, fmt.Errorf("%q is not a REZI-compatible type for encoding", t.String())
	}
}

func canDecode(v interface{}) (typeInfo, error) {
	if v == nil {
		return typeInfo{}, fmt.Errorf("receiver is nil")
	}

	checkType := reflect.TypeOf(v)
	origType := checkType

	var checkPtr bool

	if checkType.Kind() == reflect.Pointer {
		checkType = checkType.Elem()
		checkPtr = true
	}

	info, err := decTypeInfo(checkType)
	if err != nil {
		return info, err
	}

	// we do not allow a ref-to binaryUnmarshaler here
	if info.Main == tBinary && checkPtr && info.ViaNonPtr {
		// no, you pass in an implementor of encoding.BinaryUnmarshaler... not
		// a ptr to *that*
		return typeInfo{}, fmt.Errorf("%q is not a REZI-compatible type for decoding", origType.String())
	}
	return info, nil
}

func decTypeInfo(t reflect.Type) (info typeInfo, err error) {
	origType := t

	trying := true
	indirCount := 0

	for trying {
		trying = false

		if t.Implements(refBinaryUnmarshalerType) {
			// the 'via an embedded struct' way of getting a binary value
			return typeInfo{Indir: indirCount, ViaNonPtr: true, Main: tBinary}, nil
		} else if reflect.PointerTo(t).Implements(refBinaryUnmarshalerType) {
			// the 'normal' way of getting a binary value
			return typeInfo{Indir: indirCount, ViaNonPtr: false, Main: tBinary}, nil
		}

		switch t.Kind() {
		case reflect.String:
			return typeInfo{Indir: indirCount, Main: tString}, nil
		case reflect.Bool:
			return typeInfo{Indir: indirCount, Main: tBool}, nil
		case reflect.Uint8:
			return typeInfo{Indir: indirCount, Main: tIntegral, Bits: 8, Signed: false}, nil
		case reflect.Uint16:
			return typeInfo{Indir: indirCount, Main: tIntegral, Bits: 16, Signed: false}, nil
		case reflect.Uint32:
			return typeInfo{Indir: indirCount, Main: tIntegral, Bits: 32, Signed: false}, nil
		case reflect.Uint64:
			return typeInfo{Indir: indirCount, Main: tIntegral, Bits: 64, Signed: false}, nil
		case reflect.Uint:
			return typeInfo{Indir: indirCount, Main: tIntegral, Bits: 0, Signed: false}, nil
		case reflect.Int8:
			return typeInfo{Indir: indirCount, Main: tIntegral, Bits: 8, Signed: true}, nil
		case reflect.Int16:
			return typeInfo{Indir: indirCount, Main: tIntegral, Bits: 16, Signed: true}, nil
		case reflect.Int32:
			return typeInfo{Indir: indirCount, Main: tIntegral, Bits: 32, Signed: true}, nil
		case reflect.Int64:
			return typeInfo{Indir: indirCount, Main: tIntegral, Bits: 64, Signed: true}, nil
		case reflect.Int:
			return typeInfo{Indir: indirCount, Main: tIntegral, Bits: 0, Signed: true}, nil
		case reflect.Map:
			// could be okay, but key and value types must be decodable.
			mValType := t.Elem()
			mKeyType := t.Key()

			mValInfo, err := decTypeInfo(mValType)
			if err != nil {
				return typeInfo{}, fmt.Errorf("map value type is not decodable: %w", err)
			}
			mKeyInfo, err := decTypeInfo(mKeyType)
			if err != nil {
				return typeInfo{}, fmt.Errorf("map key type is not decodable: %w", err)
			}

			// maps in general are not supported; the key type MUST be comparable
			// and with an ordering, which p much means we exclusively support
			// non-binary primitives.
			if !mKeyInfo.Primitive() || mKeyInfo.Main == tBinary {
				return typeInfo{}, fmt.Errorf("map key type must be bool, string, or castable to int")
			}

			return typeInfo{Indir: indirCount, Main: tMap, KeyType: &mKeyInfo, ValType: &mValInfo}, nil
		case reflect.Slice:
			// could be okay, but val type must be encodable
			slValType := t.Elem()
			slValInfo, err := decTypeInfo(slValType)
			if err != nil {
				return typeInfo{}, fmt.Errorf("slice value is not decodable: %w", err)
			}
			return typeInfo{Indir: indirCount, Main: tSlice, ValType: &slValInfo}, nil
		case reflect.Pointer:
			// try removing one level of indrection and checking THAT
			t = t.Elem()
			trying = true
			indirCount++
		default:
			return typeInfo{}, fmt.Errorf("%q is not a REZI-compatible type for decoding", origType.String())
		}
	}
	panic("should not be possible to escape check loop")
}
