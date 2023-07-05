package rezi

import (
	"fmt"
	"io"
	"reflect"
	"sort"
)

// ti must containt a main type of tIntegral, tBool, or tString
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
	if smk.ti.Main == tBool {
		b1 := smk.keys[i].Bool()
		b2 := smk.keys[j].Bool()
		return !b1 && b2
	} else if smk.ti.Main == tIntegral {
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
	} else if smk.ti.Main == tString {
		s1 := smk.keys[i].String()
		s2 := smk.keys[j].String()
		return s1 < s2
	} else {
		panic(fmt.Sprintf("invalid map type: %v", smk.ti.Main))
	}
}

// encCheckedMap encodes a compatible map as a REZI map.
func encCheckedMap(v interface{}, ti typeInfo) []byte {
	if ti.Main != tMap {
		panic("not a map type")
	}

	return encWithNilCheck(v, ti, func(val interface{}) []byte {
		return encMap(val, *ti.KeyType)
	}, reflect.Value.Interface)
}

func encMap(v interface{}, keyType typeInfo) []byte {
	refVal := reflect.ValueOf(v)

	if v == nil || refVal.IsNil() {
		return encNil(0)
	}

	mapKeys := refVal.MapKeys()
	keysToSort := sortableMapKeys{
		keys: mapKeys,
		ti:   keyType,
	}
	sort.Sort(keysToSort)
	mapKeys = keysToSort.keys

	enc := make([]byte, 0)

	for i := range mapKeys {
		k := mapKeys[i]
		v := refVal.MapIndex(k)

		enc = append(enc, Enc(k.Interface())...)
		enc = append(enc, Enc(v.Interface())...)
	}

	enc = append(encInt(tLen(len(enc))), enc...)
	return enc
}

// decCheckedMap decodes a REZI map as a compatible map type.
func decCheckedMap(data []byte, v interface{}, ti typeInfo) (int, error) {
	if ti.Main != tMap {
		panic("not a map type")
	}

	m, n, err := decWithNilCheck(data, v, ti, fn_DecToWrappedReceiver(v, ti,
		func(t reflect.Type) bool {
			return t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Map
		},
		decMap,
	))
	if err != nil {
		return n, err
	}
	if ti.Indir == 0 {
		refReceiver := reflect.ValueOf(v)
		refReceiver.Elem().Set(reflect.ValueOf(m))
	}
	return n, err
}

func decMap(data []byte, v interface{}) (int, error) {
	var totalConsumed int

	toConsume, n, err := decInt[tLen](data)
	if err != nil {
		return 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	refVal := reflect.ValueOf(v)
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
		return totalConsumed, io.ErrUnexpectedEOF
	}

	// create the map we will be populating
	m := reflect.MakeMap(refMapType)

	var i int
	for i < toConsume {
		// dynamically create the map key type
		refKey := reflect.New(refMapType.Key())
		n, err := Dec(data, refKey.Interface())
		if err != nil {
			return totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		refVType := refMapType.Elem()
		refValue := reflect.New(refVType)
		n, err = Dec(data, refValue.Interface())
		if err != nil {
			return totalConsumed, fmt.Errorf("decode value: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		m.SetMapIndex(refKey.Elem(), refValue.Elem())
	}

	refVal.Elem().Set(m)

	return totalConsumed, nil
}
