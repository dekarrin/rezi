package rezi

import (
	"encoding"
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
)

// typeInfo holds REZI-specific type info on types that can be encoded and
// decoded.
type typeInfo struct {
	Main    mainType
	Pointer bool
	KeyType *typeInfo // only valid for maps
	ValType *typeInfo // valid for map and slice
}

func (ti typeInfo) Primitive() bool {
	return ti.Main == tIntegral || ti.Main == tBool || ti.Main == tString || ti.Main == tBinary
}

func canEncode(v interface{}) (typeInfo, bool) {
	var ti typeInfo
}

func encTypeInfo(t reflect.Type) (info typeInfo, err error) {
	switch t.Kind() {
	case reflect.String:
		return typeInfo{Main: tString}, nil
	case reflect.Bool:
		return typeInfo{Main: tBool}, nil
	case reflect.Uint8:
		return typeInfo{Main: tIntegral}, nil
	case reflect.Uint16:
		return typeInfo{Main: tIntegral}, true
	case reflect.Uint32:
		return typeInfo{Main: tIntegral}, true
	case reflect.Uint64:
		return typeInfo{Main: tIntegral}, true
	case reflect.Uint:
		return typeInfo{Main: tIntegral}, true
	case reflect.Int8:
		return typeInfo{Main: tIntegral}, true
	case reflect.Int16:
		return typeInfo{Main: tIntegral}, true
	case reflect.Int32:
		return typeInfo{Main: tIntegral}, true
	case reflect.Int64:
		return typeInfo{Main: tIntegral}, true
	case reflect.Int:
		return typeInfo{Main: tIntegral}, true
	default:
		// does it implement binary encoder?
		if t.Implements(refBinaryMarshalerType) {
			return typeInfo{Main: tBinary}, true
		}

		// is it a map type?
		if t.Kind() == reflect.Map {
			// could be okay, but key and value types must be encodable.
			mValType := t.Elem()
			mKeyType := t.Key()

			mValInfo, ok := encTypeInfo(mValType)
			if !ok {
				return typeInfo{}, false
			}
			mKeyInfo, ok := encTypeInfo(mKeyType)
			if !ok {
				return typeInfo{}, false
			}
			return typeInfo{Main: tMap, KeyType: &mKeyInfo, ValType: &mValInfo}, true
		}

		// is it a slice type?
		if t.Kind() == reflect.Slice {
			// could be okay, but val type must be encodable
			slValType := t.Elem()
			slValInfo, ok := encTypeInfo(slValType)
			if !ok {
				return typeInfo{}, false
			}
			return typeInfo{Main: tSlice, ValType: &slValInfo}, true
		}

		return typeInfo{}, false
	}
}
