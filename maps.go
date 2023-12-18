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
	if value.info.Main != mtMap {
		panic("not a map type")
	}

	return encWithNilCheck(value, encMap, reflect.Value.Interface)
}

// requires keyType type info to be avail under *mapVal.ti.KeyType and ref to be
// set.
func encMap(value analyzed[any]) ([]byte, error) {
	if value.v == nil || value.reflect.IsNil() {
		return encNilHeader(0), nil
	}

	mapKeys := value.reflect.MapKeys()
	keysToSort := sortableMapKeys{
		keys: mapKeys,
		ti:   *value.info.KeyType,
	}
	sort.Sort(keysToSort)
	mapKeys = keysToSort.keys

	enc := make([]byte, 0)

	for i := range mapKeys {
		k := mapKeys[i]
		v := value.reflect.MapIndex(k)

		keyData, err := encWithTypeInfo(k.Interface(), *value.info.KeyType)
		if err != nil {
			return nil, errorf("map key %v: %v", k.Interface(), err)
		}
		valData, err := encWithTypeInfo(v.Interface(), *value.info.ValType)
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
func decCheckedMap(data []byte, recv analyzed[any]) (decoded[any], error) {
	if recv.info.Main != mtMap {
		panic("not a map type")
	}

	m, err := decWithNilCheck(data, recv, fn_DecToWrappedReceiver(recv,
		func(t reflect.Type) bool {
			return t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Map
		},
		decMap,
	))
	if err != nil {
		return m, err
	}
	if recv.info.Indir == 0 {
		refReceiver := recv.reflect
		refReceiver.Elem().Set(m.reflect)
	}
	return m, err
}

func decMap(data []byte, recv analyzed[any]) (decoded[any], error) {
	var dec decoded[any]

	toConsume, err := decInt[tLen](data)
	if err != nil {
		return dec, errorDecf(0, "decode byte count: %s", err)
	}
	data = data[toConsume.n:]
	dec.n += toConsume.n

	refVal := recv.reflect
	refMapType := refVal.Type().Elem()

	if toConsume.v == 0 {
		// initialize to the empty map
		emptyMap := reflect.MakeMap(refMapType)

		// set it to the value
		refVal.Elem().Set(emptyMap)
		dec.v = emptyMap.Interface()
		dec.reflect = emptyMap
		return dec, nil
	} else if toConsume.v == -1 {
		// initialize to the nil map
		nilMap := reflect.Zero(refMapType)

		// set it to the value
		refVal.Elem().Set(nilMap)
		dec.v = nilMap.Interface()
		dec.reflect = nilMap
		return dec, nil
	}

	if len(data) < toConsume.v {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded map byte count is %d but only %d byte%s remain%s in data at offset"
		err := errorDecf(dec.n, errFmt, toConsume.v, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return dec, err
	}

	// clamp values we are allowed to read so we don't try to read other data
	data = data[:toConsume.v]

	// create the map we will be populating
	m := reflect.MakeMap(refMapType)

	var i int
	refKType := refMapType.Key()
	refVType := refMapType.Elem()
	for i < toConsume.v {
		// dynamically create the map key type
		refKey := reflect.New(refKType)
		n, err := decWithTypeInfo(data, refKey.Interface(), *recv.info.KeyType)
		if err != nil {
			return dec, errorDecf(dec.n, "map key: %v", err)
		}
		dec.n += n
		i += n
		data = data[n:]

		refValue := reflect.New(refVType)
		n, err = decWithTypeInfo(data, refValue.Interface(), *recv.info.ValType)
		if err != nil {
			return dec, errorDecf(dec.n, "map value[%v]: %v", refKey.Elem().Interface(), err)
		}
		dec.n += n
		i += n
		data = data[n:]

		m.SetMapIndex(refKey.Elem(), refValue.Elem())
	}

	refVal.Elem().Set(m)
	dec.v = m.Interface()
	dec.reflect = m
	return dec, nil
}
