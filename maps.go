package rezi

import (
	"fmt"
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

// encMap encodes a compatible map as a REZI map.
func encMap(v interface{}, ti typeInfo) []byte {
	if ti.Main != tMap {
		panic("not a map type")
	}

	refVal := reflect.ValueOf(v)

	if v == nil || refVal.IsNil() {
		return EncInt(-1)
	}

	mapKeys := refVal.MapKeys()
	keysToSort := sortableMapKeys{
		keys: mapKeys,
		ti:   *ti.KeyType,
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

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

// decMap decodes a REZI map as a compatible map type.
func decMap(data []byte, v interface{}, ti typeInfo) (int, error) {
	if ti.Main != tMap {
		panic("not a map type")
	}

	var totalConsumed int

	toConsume, n, err := DecInt(data)
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
		return totalConsumed, fmt.Errorf("unexpected EOF")
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
		// if we specifically are instructed to deref, then instead of the
		// normal key, get a ptr-to the type of.
		if ti.ValType.ViaNonPtr {
			refVType = reflect.PointerTo(refVType)
		}
		refValue := reflect.New(refVType)
		n, err = Dec(data, refValue.Interface())
		if err != nil {
			return totalConsumed, fmt.Errorf("decode value: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		m.SetMapIndex(refKey, refValue)
	}

	refVal.Elem().Set(m)

	return totalConsumed, nil
}
