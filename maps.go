package rezi

import (
	"fmt"
	"io"
	"reflect"
	"sort"
)

// ti must containt a main type of mtIntegral, mtBool, mtFloat, or mtString
type sortableMapKeys struct {
	keys []reflect.Value
	ti   typeInfo
}

func (smk sortableMapKeys) Len() int {
	return len(smk.keys)
}

func (smk sortableMapKeys) Swap(i, j int) {
	smk.keys[i], smk.keys[j] = smk.keys[j], smk.keys[i]
}

func (smk sortableMapKeys) Less(i, j int) bool {
	if smk.ti.Main == mtBool {
		b1 := smk.keys[i].Bool()
		b2 := smk.keys[j].Bool()
		return !b1 && b2
	} else if smk.ti.Main == mtIntegral {
		if smk.ti.Signed {
			i64v1 := smk.keys[i].Int()
			i64v2 := smk.keys[j].Int()
			switch smk.ti.Bits {
			case 64:
				return i64v1 < i64v2
			case 32:
				return int32(i64v1) < int32(i64v2)
			case 16:
				return int16(i64v1) < int16(i64v2)
			case 8:
				return int8(i64v1) < int8(i64v2)
			default:
				return int(i64v1) < int(i64v2)
			}
		} else {
			u64v1 := smk.keys[i].Uint()
			u64v2 := smk.keys[j].Uint()
			switch smk.ti.Bits {
			case 64:
				return u64v1 < u64v2
			case 32:
				return uint32(u64v1) < uint32(u64v2)
			case 16:
				return uint16(u64v1) < uint16(u64v2)
			case 8:
				return uint8(u64v1) < uint8(u64v2)
			default:
				return uint(u64v1) < uint(u64v2)
			}
		}
	} else if smk.ti.Main == mtFloat {
		f1 := smk.keys[i].Float()
		f2 := smk.keys[j].Float()
		return f1 < f2
	} else if smk.ti.Main == mtString {
		s1 := smk.keys[i].String()
		s2 := smk.keys[j].String()
		return s1 < s2
	} else {
		panic(fmt.Sprintf("invalid map key type: %v", smk.ti.Main))
	}
}

// encCheckedMap encodes a compatible map as a REZI map.
func encCheckedMap(value analyzed[any]) ([]byte, error) {
	if value.ti.Main != mtMap {
		panic("not a map type")
	}

	return encWithNilCheck(value, encMap, reflect.Value.Interface)
}

// requires keyType type info to be avail under *mapVal.ti.KeyType and ref to be
// set.
func encMap(val analyzed[any]) ([]byte, error) {
	if val.native == nil || val.ref.IsNil() {
		return encNilHeader(0), nil
	}

	mapKeys := val.ref.MapKeys()
	keysToSort := sortableMapKeys{
		keys: mapKeys,
		ti:   *val.ti.KeyType,
	}
	sort.Sort(keysToSort)
	mapKeys = keysToSort.keys

	enc := make([]byte, 0)

	for i := range mapKeys {
		k := mapKeys[i]
		v := val.ref.MapIndex(k)

		keyData, err := Enc(k.Interface())
		if err != nil {
			return nil, errorf("map key %v: %v", k.Interface(), err)
		}
		valData, err := Enc(v.Interface())
		if err != nil {
			return nil, errorf("map value[%v]: %v", k.Interface(), err)
		}

		enc = append(enc, keyData...)
		enc = append(enc, valData...)
	}

	enc = append(encCount(len(enc), nil), enc...)
	return enc, nil
}

// decCheckedMap decodes a REZI map as a compatible map type.
func decCheckedMap(data []byte, v analyzed[any]) (int, error) {
	if v.ti.Main != mtMap {
		panic("not a map type")
	}

	m, n, err := decWithNilCheck(data, v, fn_DecToWrappedReceiver(v,
		func(t reflect.Type) bool {
			return t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Map
		},
		func(b []byte, v analyzed[any]) (interface{}, int, error) {
			decN, err := decMap(b, v)
			return nil, decN, err
		},
	))
	if err != nil {
		return n, err
	}
	if v.ti.Indir == 0 {
		refReceiver := v.ref
		refReceiver.Elem().Set(reflect.ValueOf(m))
	}
	return n, err
}

func decMap(data []byte, v analyzed[any]) (int, error) {
	var totalConsumed int

	toConsume, _, n, err := decInt[tLen](data)
	if err != nil {
		return 0, errorDecf(0, "decode byte count: %s", err)
	}
	data = data[n:]
	totalConsumed += n

	refVal := v.ref
	refMapType := refVal.Type().Elem()

	if toConsume == 0 {
		// initialize to the empty map
		emptyMap := reflect.MakeMap(refMapType)

		// set it to the value
		refVal.Elem().Set(emptyMap)
		return totalConsumed, nil
	} else if toConsume == -1 {
		// initialize to the nil map
		nilMap := reflect.Zero(refMapType)

		// set it to the value
		refVal.Elem().Set(nilMap)
		return totalConsumed, nil
	}

	if len(data) < toConsume {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded map byte count is %d but only %d byte%s remain%s in data at offset"
		err := errorDecf(totalConsumed, errFmt, toConsume, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return totalConsumed, err
	}

	// clamp values we are allowed to read so we don't try to read other data
	data = data[:toConsume]

	// create the map we will be populating
	m := reflect.MakeMap(refMapType)

	var i int
	refKType := refMapType.Key()
	refVType := refMapType.Elem()
	for i < toConsume {
		// dynamically create the map key type
		refKey := reflect.New(refKType)
		n, err := Dec(data, refKey.Interface())
		if err != nil {
			return totalConsumed, errorDecf(totalConsumed, "map key: %v", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		refValue := reflect.New(refVType)
		n, err = Dec(data, refValue.Interface())
		if err != nil {
			return totalConsumed, errorDecf(totalConsumed, "map value[%v]: %v", refKey.Elem().Interface(), err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		m.SetMapIndex(refKey.Elem(), refValue.Elem())
	}

	refVal.Elem().Set(m)

	return totalConsumed, nil
}
