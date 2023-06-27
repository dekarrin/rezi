package rezi

// stringmaps.go contains functions for encoding and decoding maps of string to
// the basic types.

import (
	"encoding"
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

	if v == nil {
		return EncInt(-1)
	}

	refVal := reflect.ValueOf(v)
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

// EncMapStringToInt encodes a map of string-to-int as bytes. The order of keys
// in the output is gauranteed to be consistent.
func EncMapStringToInt(m map[string]int) []byte {
	if m == nil {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := range keys {
		enc = append(enc, EncString(keys[i])...)
		enc = append(enc, EncInt(m[keys[i]])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

// DecMapStringToBinary decodes a map of string-to-int from bytes.
func DecMapStringToInt(data []byte) (map[string]int, int, error) {
	var totalConsumed int

	toConsume, n, err := DecInt(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		return map[string]int{}, totalConsumed, nil
	} else if toConsume == -1 {
		return nil, totalConsumed, nil
	}

	if len(data) < toConsume {
		return nil, 0, fmt.Errorf("unexpected EOF")
	}

	m := map[string]int{}

	var i int
	for i < toConsume {
		k, n, err := DecString(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		v, n, err := DecInt(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		m[k] = v
	}

	return m, totalConsumed, nil
}

// EncMapStringToBinary encodes a map of string to an implementer of
// encoding.BinaryMarshaler as bytes. The order of keys in output is gauranteed
// to be consistent.
func EncMapStringToBinary[E encoding.BinaryMarshaler](m map[string]E) []byte {
	if m == nil {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := range keys {
		enc = append(enc, EncString(keys[i])...)
		enc = append(enc, EncBinary(m[keys[i]])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

// DecMapStringToBinary decodes a map of string to an implementer of
// encoding.BinaryMarshaler from bytes.
func DecMapStringToBinary[E encoding.BinaryUnmarshaler](data []byte) (map[string]E, int, error) {
	var totalConsumed int

	toConsume, n, err := DecInt(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		return map[string]E{}, totalConsumed, nil
	} else if toConsume == -1 {
		return nil, totalConsumed, nil
	}

	if len(data) < toConsume {
		return nil, 0, fmt.Errorf("unexpected EOF")
	}

	m := map[string]E{}

	var i int
	for i < toConsume {
		k, n, err := DecString(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		v := initType[E]()
		n, err = DecBinary(data, v)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		m[k] = v
	}

	return m, totalConsumed, nil
}
