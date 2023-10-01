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
	mtUnknown mainType = iota
	mtIntegral
	mtBool
	mtString
	mtBinary
	mtMap
	mtSlice
	mtNil
)

// typeInfo holds REZI-specific type info on types that can be encoded and
// decoded.
type typeInfo struct {
	Main    mainType
	Bits    int
	Signed  bool
	Indir   int       // Indir is number of times that the value is deref'd. Used for encoding of ptr-to types.
	KeyType *typeInfo // only valid for maps
	ValType *typeInfo // valid for map and slice
}

func (ti typeInfo) Primitive() bool {
	return ti.Main == mtIntegral || ti.Main == mtBool || ti.Main == mtString || ti.Main == mtBinary
}

func canEncode(v interface{}) (typeInfo, error) {
	return encTypeInfo(reflect.TypeOf(v))
}

func encTypeInfo(t reflect.Type) (info typeInfo, err error) {
	if t == nil {
		return typeInfo{Main: mtNil}, nil
	}

	origType := t

	trying := true
	indirCount := 0

	for trying {
		if t.Implements(refBinaryMarshalerType) {
			// does it actually implement it itself? or did we just get handed a
			// ptr type and the pointed-to type defines a value receiver and Go
			// is performing implicit deref to make it be defined on the ptr
			// as well?

			if t.Kind() == reflect.Pointer {
				// is method actually defined on the *value* receiver with
				// implicit deref?
				_, definedOnValue := t.Elem().MethodByName("MarshalBinary")

				// only consider it to be implementing if it is *not* defined
				// on the value type.
				if !definedOnValue {
					return typeInfo{Indir: indirCount, Main: mtBinary}, nil
				}

				// if it *is* defined on the value type, we are getting implicit
				// deref, and should *not* treat the current checked type as
				// implementing. we'll get it on the next deref pass with the
				// correct Indir number set.
			} else {
				// if it's not a pointer type and it implements, there is no
				// ambiguity.
				return typeInfo{Indir: indirCount, Main: mtBinary}, nil
			}
		}

		switch t.Kind() {
		case reflect.String:
			return typeInfo{Indir: indirCount, Main: mtString}, nil
		case reflect.Bool:
			return typeInfo{Indir: indirCount, Main: mtBool}, nil
		case reflect.Uint8:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 8, Signed: false}, nil
		case reflect.Uint16:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 16, Signed: false}, nil
		case reflect.Uint32:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 32, Signed: false}, nil
		case reflect.Uint64:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 64, Signed: false}, nil
		case reflect.Uint:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 0, Signed: false}, nil
		case reflect.Int8:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 8, Signed: true}, nil
		case reflect.Int16:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 16, Signed: true}, nil
		case reflect.Int32:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 32, Signed: true}, nil
		case reflect.Int64:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 64, Signed: true}, nil
		case reflect.Int:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 0, Signed: true}, nil
		case reflect.Map:
			// could be okay, but key and value types must be encodable.
			mValType := t.Elem()
			mKeyType := t.Key()

			mValInfo, err := encTypeInfo(mValType)
			if err != nil {
				return typeInfo{}, errorf("map value type is not encodable: %s", err)
			}
			mKeyInfo, err := encTypeInfo(mKeyType)
			if err != nil {
				return typeInfo{}, errorf("map key type is not encodable: %s", err)
			}

			// maps in general are not supported; the key type MUST be comparable
			// and with an ordering, which p much means we exclusively support
			// non-binary primitives.
			if !mKeyInfo.Primitive() || mKeyInfo.Main == mtBinary {
				return typeInfo{}, errorf("map key type must be bool, string, or castable to int").wrap(ErrInvalidType)
			}

			return typeInfo{Indir: indirCount, Main: mtMap, KeyType: &mKeyInfo, ValType: &mValInfo}, nil
		case reflect.Slice:
			// could be okay, but val type must be encodable
			slValType := t.Elem()
			slValInfo, err := encTypeInfo(slValType)
			if err != nil {
				return typeInfo{}, errorf("slice value is not encodable: %s", err)
			}
			return typeInfo{Indir: indirCount, Main: mtSlice, ValType: &slValInfo}, nil
		case reflect.Pointer:
			// try removing one level of indrection and checking THAT
			t = t.Elem()
			trying = true
			indirCount++
		default:
			return typeInfo{}, errorf("%q is not a REZI-compatible type for encoding", origType.String()).wrap(ErrInvalidType)
		}
	}

	panic("should not be possible to escape check loop")
}

func canDecode(v interface{}) (typeInfo, error) {
	if v == nil {
		return typeInfo{}, errorf("receiver is nil").wrap(ErrInvalidType)
	}

	checkVal := reflect.ValueOf(v)
	checkType := checkVal.Type()

	if checkType.Kind() == reflect.Pointer {
		// make shore it's a not a pointer to nil
		if checkVal.Elem().Kind() == reflect.Invalid {
			return typeInfo{}, errorf("receiver is nil").wrap(ErrInvalidType)
		}
		checkType = checkType.Elem()
	}

	info, err := decTypeInfo(checkType)
	if err != nil {
		return info, err
	}

	return info, nil
}

func decTypeInfo(t reflect.Type) (info typeInfo, err error) {
	origType := t

	trying := true
	indirCount := 0

	for trying {
		trying = false

		if reflect.PointerTo(t).Implements(refBinaryUnmarshalerType) {
			return typeInfo{Indir: indirCount, Main: mtBinary}, nil
		}

		switch t.Kind() {
		case reflect.String:
			return typeInfo{Indir: indirCount, Main: mtString}, nil
		case reflect.Bool:
			return typeInfo{Indir: indirCount, Main: mtBool}, nil
		case reflect.Uint8:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 8, Signed: false}, nil
		case reflect.Uint16:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 16, Signed: false}, nil
		case reflect.Uint32:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 32, Signed: false}, nil
		case reflect.Uint64:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 64, Signed: false}, nil
		case reflect.Uint:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 0, Signed: false}, nil
		case reflect.Int8:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 8, Signed: true}, nil
		case reflect.Int16:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 16, Signed: true}, nil
		case reflect.Int32:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 32, Signed: true}, nil
		case reflect.Int64:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 64, Signed: true}, nil
		case reflect.Int:
			return typeInfo{Indir: indirCount, Main: mtIntegral, Bits: 0, Signed: true}, nil
		case reflect.Map:
			// could be okay, but key and value types must be decodable.
			mValType := t.Elem()
			mKeyType := t.Key()

			mValInfo, err := decTypeInfo(mValType)
			if err != nil {
				return typeInfo{}, errorf("map value type is not decodable: %s", err)
			}
			mKeyInfo, err := decTypeInfo(mKeyType)
			if err != nil {
				return typeInfo{}, errorf("map key type is not decodable: %s", err)
			}

			// maps in general are not supported; the key type MUST be comparable
			// and with an ordering, which p much means we exclusively support
			// non-binary primitives.
			if !mKeyInfo.Primitive() || mKeyInfo.Main == mtBinary {
				return typeInfo{}, errorf("map key type must be bool, string, or castable to int").wrap(ErrInvalidType)
			}

			return typeInfo{Indir: indirCount, Main: mtMap, KeyType: &mKeyInfo, ValType: &mValInfo}, nil
		case reflect.Slice:
			// could be okay, but val type must be encodable
			slValType := t.Elem()
			slValInfo, err := decTypeInfo(slValType)
			if err != nil {
				return typeInfo{}, errorf("slice value is not decodable: %s", err)
			}
			return typeInfo{Indir: indirCount, Main: mtSlice, ValType: &slValInfo}, nil
		case reflect.Pointer:
			// try removing one level of indrection and checking THAT
			t = t.Elem()
			trying = true
			indirCount++
		default:
			return typeInfo{}, errorf("%q is not a REZI-compatible type for decoding", origType.String()).wrap(ErrInvalidType)
		}
	}

	panic("should not be possible to escape check loop")
}
