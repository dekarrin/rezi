package rezi

import (
	"encoding"
	"fmt"
	"reflect"
	"sort"
)

var (
	refBinaryMarshalerType   = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
	refBinaryUnmarshalerType = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
	refTextMarshalerType     = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	refTextUnmarshalerType   = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

	refPrimitiveKindTypes = map[reflect.Kind]reflect.Type{
		reflect.Bool:       reflect.TypeOf(true),
		reflect.Complex128: reflect.TypeOf(complex128(0 + 1i)),
		reflect.Complex64:  reflect.TypeOf(complex64(0 + 1i)),
		reflect.Float32:    reflect.TypeOf(float32(0.0)),
		reflect.Float64:    reflect.TypeOf(float64(0.0)),
		reflect.Int:        reflect.TypeOf(int(0.0)),
		reflect.Int8:       reflect.TypeOf(int8(0.0)),
		reflect.Int16:      reflect.TypeOf(int16(0.0)),
		reflect.Int32:      reflect.TypeOf(int32(0.0)),
		reflect.Int64:      reflect.TypeOf(int64(0.0)),
		reflect.String:     reflect.TypeOf(""),
		reflect.Uint:       reflect.TypeOf(uint(0.0)),
		reflect.Uint8:      reflect.TypeOf(uint8(0.0)),
		reflect.Uint16:     reflect.TypeOf(uint16(0.0)),
		reflect.Uint32:     reflect.TypeOf(uint32(0.0)),
		reflect.Uint64:     reflect.TypeOf(uint64(0.0)),
	}
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
	mtFloat
	mtComplex
	mtArray
	mtText
	mtStruct
)

func (mt mainType) String() string {
	switch mt {
	case mtUnknown:
		return "mtUnknown"
	case mtIntegral:
		return "mtIntegral"
	case mtBool:
		return "mtBool"
	case mtString:
		return "mtString"
	case mtBinary:
		return "mtBinary"
	case mtMap:
		return "mtMap"
	case mtSlice:
		return "mtSlice"
	case mtNil:
		return "mtNil"
	case mtFloat:
		return "mtFloat"
	case mtComplex:
		return "mtComplex"
	case mtArray:
		return "mtArray"
	case mtText:
		return "mtText"
	case mtStruct:
		return "mtStruct"
	default:
		return fmt.Sprintf("mainType(%d)", mt)
	}
}

// fieldInfo holds REZI-specific type info on fieldds of a struct
type fieldInfo struct {
	Name      string
	Index     int // position in fields by index
	Type      typeInfo
	Anonymous bool // TODO: check if this actually needed on completion of #61
}

type fields struct {
	ByName  map[string]fieldInfo
	ByOrder []fieldInfo
}

// sortableFields can sort a slice of fieldInfo. select whether by Name or by
// Index with the alpha property.
type sortableFields struct {
	fields []fieldInfo
	alpha  bool // if alpha is false, it's sorted by Index of the fields in fs. else by Name.
}

// Len implements sort.Interface
func (sf *sortableFields) Len() int {
	return len(sf.fields)
}

// Less implements sort.Interface
func (sf *sortableFields) Less(i, j int) bool {
	f1 := sf.fields[i]
	f2 := sf.fields[j]

	if sf.alpha {
		return f1.Name < f2.Name
	}
	return f1.Index < f2.Index
}

// Swap implements sort.Interface
func (sf *sortableFields) Swap(i, j int) {
	f1 := sf.fields[i]
	f2 := sf.fields[j]

	sf.fields[i] = f2
	sf.fields[j] = f1
}

// does not sort in place; makes complete copy
func sortFieldsByName(fields []fieldInfo) []fieldInfo {
	sorting := &sortableFields{fields: make([]fieldInfo, len(fields)), alpha: true}
	copy(sorting.fields, fields)
	sort.Sort(sorting)
	return sorting.fields
}

// does not sort in place; makes complete copy
func sortFieldsByIndex(fields []fieldInfo) []fieldInfo {
	sorting := &sortableFields{fields: make([]fieldInfo, len(fields)), alpha: false}
	copy(sorting.fields, fields)
	sort.Sort(sorting)
	return sorting.fields
}

// typeInfo holds REZI-specific type info on types that can be encoded and
// decoded.
type typeInfo struct {
	Main       mainType
	Bits       int
	Signed     bool
	Underlying bool      // can be valid for any type that has the Main one as an underlying. will never be valid for mtText or mtBinary.
	Indir      int       // Indir is number of times that the value is deref'd. Used for encoding of ptr-to types.
	KeyType    *typeInfo // only valid for maps
	ValType    *typeInfo // valid for map, slice, and array
	Len        int       // only valid for array
	Dec        bool      // whether the info is for a decoded value. if false, it's for an encoded one.
	Fields     fields    // valid for struct only
}

// MainReflectType returns the reflect.Type that represents the main type of the
// value that this typeInfo is created from. It will be nil when Main is a
// mainType which does not have a strictly associated kind (mtUnknown, mtNil),
// or if the typeInfo represents an invalid type (such as mtIntegral with bit
// size of 3).
//
// The returned type will be entirely based off of the typeInfo, As a result,
// typeInfos created from implementors of marshaler classes will return the
// interface implemented in the result as opposed to the actual implementor, as
// that info is not available in the typeInfo. In addition, those with
// underlying types that are basic Go types will return the underlying type as
// opposed to the actual type.
//
// If indirected is set to false, the returned type strictly refers to the main
// type with no indirection; a typeInfo created from a *uint8 will return the
// Type of uint8, even though it was made from a pointer. This does not affect
// the value and/or key types when the returned Type is a container type (map,
// array, slice); these will always be properly indirected.
func (ti typeInfo) MainReflectType(indirected bool) reflect.Type {
	var t reflect.Type
	switch ti.Main {
	case mtArray:
		vrt := ti.ValType.MainReflectType(true)
		t = reflect.ArrayOf(ti.Len, vrt)
	case mtBool:
		t = refPrimitiveKindTypes[reflect.Bool]
	case mtComplex:
		switch ti.Bits {
		case 128:
			t = refPrimitiveKindTypes[reflect.Complex128]
		case 64:
			t = refPrimitiveKindTypes[reflect.Complex64]
		}
	case mtFloat:
		switch ti.Bits {
		case 64:
			t = refPrimitiveKindTypes[reflect.Float64]
		case 32:
			t = refPrimitiveKindTypes[reflect.Float32]
		}
	case mtIntegral:
		if ti.Signed {
			switch ti.Bits {
			case 0:
				t = refPrimitiveKindTypes[reflect.Int]
			case 8:
				t = refPrimitiveKindTypes[reflect.Int8]
			case 16:
				t = refPrimitiveKindTypes[reflect.Int16]
			case 32:
				t = refPrimitiveKindTypes[reflect.Int32]
			case 64:
				t = refPrimitiveKindTypes[reflect.Int64]
			}
		} else {
			switch ti.Bits {
			case 0:
				t = refPrimitiveKindTypes[reflect.Uint]
			case 8:
				t = refPrimitiveKindTypes[reflect.Uint8]
			case 16:
				t = refPrimitiveKindTypes[reflect.Uint16]
			case 32:
				t = refPrimitiveKindTypes[reflect.Uint32]
			case 64:
				t = refPrimitiveKindTypes[reflect.Uint64]
			}
		}
	case mtBinary:
		if ti.Dec {
			t = refBinaryUnmarshalerType
		} else {
			t = refBinaryMarshalerType
		}
	case mtText:
		if ti.Dec {
			t = refTextUnmarshalerType
		} else {
			t = refTextMarshalerType
		}
	case mtMap:
		krt := ti.KeyType.MainReflectType(true)
		vrt := ti.ValType.MainReflectType(true)
		t = reflect.MapOf(krt, vrt)
	case mtSlice:
		vrt := ti.ValType.MainReflectType(true)
		t = reflect.SliceOf(vrt)
	case mtString:
		t = refPrimitiveKindTypes[reflect.String]
	case mtStruct:
		sorted := sortFieldsByIndex(ti.Fields.ByOrder)
		refFields := []reflect.StructField{}
		for _, fi := range sorted {
			structRefType := fi.Type.MainReflectType(true)
			sf := reflect.StructField{
				Name:      fi.Name,
				Anonymous: fi.Anonymous,
				Type:      structRefType,
			}
			refFields = append(refFields, sf)
		}
		t = reflect.StructOf(refFields)
	}

	if t != nil && indirected {
		for i := 0; i < ti.Indir; i++ {
			t = reflect.PointerTo(t)
		}
	}

	return t
}

func (ti typeInfo) Primitive() bool {
	return ti.Main == mtIntegral || ti.Main == mtBool || ti.Main == mtString || ti.Main == mtBinary || ti.Main == mtFloat || ti.Main == mtComplex || ti.Main == mtText
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
		} else if t.Implements(refTextMarshalerType) {
			// same checks as above but for text
			if t.Kind() == reflect.Pointer {
				_, definedOnValue := t.Elem().MethodByName("MarshalText")

				// only consider it to be implementing if it is *not* defined
				// on the value type.
				if !definedOnValue {
					return typeInfo{Indir: indirCount, Main: mtText}, nil
				}

				// implicit deref, wait for next pass
			} else {
				// if it's not a pointer type and it implements, there is no
				// ambiguity.
				return typeInfo{Indir: indirCount, Main: mtText}, nil
			}
		}

		var under bool
		if pkt, ok := refPrimitiveKindTypes[t.Kind()]; ok {
			// TODO: second check probs not needed; if t.Kind() == X, X(t(val)) should always be valid i think.
			// Verify. Check in Laws of Reflection?
			under = t != pkt && t.ConvertibleTo(pkt)
		}

		switch t.Kind() {
		case reflect.String:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtString}, nil
		case reflect.Bool:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtBool}, nil
		case reflect.Uint8:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 8, Signed: false}, nil
		case reflect.Uint16:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 16, Signed: false}, nil
		case reflect.Uint32:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 32, Signed: false}, nil
		case reflect.Uint64:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 64, Signed: false}, nil
		case reflect.Uint:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 0, Signed: false}, nil
		case reflect.Int8:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 8, Signed: true}, nil
		case reflect.Int16:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 16, Signed: true}, nil
		case reflect.Int32:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 32, Signed: true}, nil
		case reflect.Int64:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 64, Signed: true}, nil
		case reflect.Int:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 0, Signed: true}, nil
		case reflect.Float32:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtFloat, Bits: 32, Signed: true}, nil
		case reflect.Float64:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtFloat, Bits: 64, Signed: true}, nil
		case reflect.Complex64:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtComplex, Bits: 64, Signed: true}, nil
		case reflect.Complex128:
			return typeInfo{Indir: indirCount, Underlying: under, Main: mtComplex, Bits: 128, Signed: true}, nil
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
				return typeInfo{}, errorf("map key type must be bool, string, float, int, or text-encodable type").wrap(ErrInvalidType)
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
		case reflect.Array:
			// could be okay, but val type must be encodable.
			arrValType := t.Elem()
			arrValInfo, err := encTypeInfo(arrValType)
			if err != nil {
				return typeInfo{}, errorf("array value is not encodable: %s", err)
			}
			arrLen := t.Len()
			return typeInfo{Indir: indirCount, Main: mtArray, ValType: &arrValInfo, Len: arrLen}, nil
		case reflect.Struct:
			// could be okay, but all exported fields must be encodable.
			// check while building lists of fields
			fieldsData := fields{ByName: map[string]fieldInfo{}}

			for i := 0; i < t.NumField(); i++ {
				sf := t.Field(i)
				if !sf.IsExported() {
					continue
				}
				fieldValInfo, err := encTypeInfo(sf.Type)
				if err != nil {
					return typeInfo{}, errorf("field %s is not encodeable: %s", sf.Name, err)
				}
				fi := fieldInfo{Index: i, Name: sf.Name, Anonymous: sf.Anonymous, Type: fieldValInfo}
				fieldsData.ByName[fi.Name] = fi
				fieldsData.ByOrder = append(fieldsData.ByOrder, fi)
			}
			fieldsData.ByOrder = sortFieldsByName(fieldsData.ByOrder)
			return typeInfo{Indir: indirCount, Main: mtStruct, Fields: fieldsData}, nil
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
			return typeInfo{Dec: true, Indir: indirCount, Main: mtBinary}, nil
		} else if reflect.PointerTo(t).Implements(refTextUnmarshalerType) {
			return typeInfo{Dec: true, Indir: indirCount, Main: mtText}, nil
		}

		var under bool
		if pkt, ok := refPrimitiveKindTypes[t.Kind()]; ok {
			// TODO: second check probs not needed; if t.Kind() == X, X(t(val)) should always be valid i think.
			// Verify. Check in Laws of Reflection?
			under = t != pkt && t.ConvertibleTo(pkt)
		}

		switch t.Kind() {
		case reflect.String:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtString}, nil
		case reflect.Bool:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtBool}, nil
		case reflect.Uint8:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 8, Signed: false}, nil
		case reflect.Uint16:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 16, Signed: false}, nil
		case reflect.Uint32:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 32, Signed: false}, nil
		case reflect.Uint64:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 64, Signed: false}, nil
		case reflect.Uint:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 0, Signed: false}, nil
		case reflect.Int8:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 8, Signed: true}, nil
		case reflect.Int16:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 16, Signed: true}, nil
		case reflect.Int32:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 32, Signed: true}, nil
		case reflect.Int64:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 64, Signed: true}, nil
		case reflect.Int:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtIntegral, Bits: 0, Signed: true}, nil
		case reflect.Float32:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtFloat, Bits: 32, Signed: true}, nil
		case reflect.Float64:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtFloat, Bits: 64, Signed: true}, nil
		case reflect.Complex64:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtComplex, Bits: 64, Signed: true}, nil
		case reflect.Complex128:
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtComplex, Bits: 128, Signed: true}, nil
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
				return typeInfo{}, errorf("map key type must be bool, string, float, int, or text-encodable type").wrap(ErrInvalidType)
			}

			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtMap, KeyType: &mKeyInfo, ValType: &mValInfo}, nil
		case reflect.Slice:
			// could be okay, but val type must be encodable
			slValType := t.Elem()
			slValInfo, err := decTypeInfo(slValType)
			if err != nil {
				return typeInfo{}, errorf("slice value is not decodable: %s", err)
			}
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtSlice, ValType: &slValInfo}, nil
		case reflect.Array:
			// could be okay, but val type must be encodable
			arrValType := t.Elem()
			arrValInfo, err := decTypeInfo(arrValType)
			if err != nil {
				return typeInfo{}, errorf("array value is not decodable: %s", err)
			}
			arrLen := t.Len()
			return typeInfo{Dec: true, Indir: indirCount, Underlying: under, Main: mtArray, ValType: &arrValInfo, Len: arrLen}, nil
		case reflect.Struct:
			// could be okay, but all exported fields must be encodable.
			// check while building lists of fields
			fieldsData := fields{ByName: map[string]fieldInfo{}}

			for i := 0; i < t.NumField(); i++ {
				sf := t.Field(i)
				if !sf.IsExported() {
					continue
				}
				fieldValInfo, err := decTypeInfo(sf.Type)
				if err != nil {
					return typeInfo{}, errorf("struct field %s is not encodeable: %s", sf.Name, err)
				}
				fi := fieldInfo{Index: i, Name: sf.Name, Anonymous: sf.Anonymous, Type: fieldValInfo}
				fieldsData.ByName[fi.Name] = fi
				fieldsData.ByOrder = append(fieldsData.ByOrder, fi)
			}
			fieldsData.ByOrder = sortFieldsByName(fieldsData.ByOrder)
			// doesn't make sense to set Underlying for a struct; it will ALWAYS be the 'underlying' type.
			return typeInfo{Dec: true, Indir: indirCount, Main: mtStruct, Fields: fieldsData}, nil
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
