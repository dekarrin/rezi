package rezi

// slices.go contains functions for encoding and decoding slices of basic types.

import (
	"encoding"
	"fmt"
	"log"
	"reflect"
)

// encMap encodes a compatible slice as a REZI map.
func encSlice(v interface{}, ti typeInfo) []byte {
	if ti.Main != tSlice {
		panic("not a slice type")
	}

	refVal := reflect.ValueOf(v)

	if v == nil || refVal.IsNil() {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	for i := 0; i < refVal.Len(); i++ {
		v := refVal.Index(i)
		enc = append(enc, Enc(v.Interface())...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

func decSlice(data []byte, v interface{}, ti typeInfo) (int, error) {
	if ti.Main != tSlice {
		panic("not a slice type")
	}
	var totalConsumed int

	toConsume, n, err := DecInt(data)
	if err != nil {
		return 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	refSliceVal := reflect.ValueOf(v)
	refSliceType := refSliceVal.Type().Elem()

	if toConsume == 0 {
		// initialize to the empty slice
		emptySlice := reflect.MakeSlice(refSliceType, 0, 0)
		refSliceVal.Elem().Set(emptySlice)
		return totalConsumed, nil
	} else if toConsume == -1 {
		nilSlice := reflect.Zero(refSliceType)
		refSliceVal.Elem().Set(nilSlice)
		return totalConsumed, nil
	}

	if len(data) < toConsume {
		return totalConsumed, fmt.Errorf("unexpected EOF")
	}

	sl := reflect.MakeSlice(refSliceType, 0, 0)

	var i int
	for i < toConsume {
		refVType := refSliceType.Elem()
		// if we specifically are instructed to deref, then instead of the
		// normal key, get a ptr-to the type of.
		if ti.ValType.Deref {
			refVType = reflect.PointerTo(refVType)
		}
		refValue := reflect.New(refVType)
		n, err := Dec(data, refValue.Interface())
		if err != nil {
			return totalConsumed, fmt.Errorf("decode item: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		sl = reflect.Append(sl, refValue.Elem())
	}

	log.Printf("%s", sl.Kind())

	refSliceVal.Elem().Set(sl)
	return totalConsumed, nil
}

func EncSliceString(sl []string) []byte {
	if sl == nil {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	for i := range sl {
		enc = append(enc, EncString(sl[i])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

func DecSliceString(data []byte) ([]string, int, error) {
	var totalConsumed int

	toConsume, n, err := DecInt(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		return []string{}, totalConsumed, nil
	} else if toConsume == -1 {
		return nil, totalConsumed, nil
	}

	if len(data) < toConsume {
		return nil, 0, fmt.Errorf("unexpected EOF")
	}

	sl := []string{}

	var i int
	for i < toConsume {
		s, n, err := DecString(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode item: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		sl = append(sl, s)
	}

	return sl, totalConsumed, nil
}

func EncSliceBinary[E encoding.BinaryMarshaler](sl []E) []byte {
	if sl == nil {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	for i := range sl {
		enc = append(enc, EncBinary(sl[i])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

func DecSliceBinary[E encoding.BinaryUnmarshaler](data []byte) ([]E, int, error) {
	var totalConsumed int

	toConsume, n, err := DecInt(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		return []E{}, totalConsumed, nil
	} else if toConsume == -1 {
		return nil, totalConsumed, nil
	}

	if len(data) < toConsume {
		return nil, 0, fmt.Errorf("unexpected EOF")
	}

	sl := []E{}

	var i int
	for i < toConsume {
		v := initType[E]()

		n, err := DecBinary(data, v)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode item: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		sl = append(sl, v)
	}

	return sl, totalConsumed, nil
}
