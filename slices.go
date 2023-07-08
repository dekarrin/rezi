package rezi

// slices.go contains functions for encoding and decoding slices of basic types.

import (
	"fmt"
	"io"
	"reflect"
)

// encMap encodes a compatible slice as a REZI map.
func encCheckedSlice(v interface{}, ti typeInfo) ([]byte, error) {
	if ti.Main != mtSlice {
		panic("not a slice type")
	}

	return encWithNilCheck(v, ti, encSlice, reflect.Value.Interface)
}

func encSlice(v interface{}) ([]byte, error) {
	refVal := reflect.ValueOf(v)

	if v == nil || refVal.IsNil() {
		return encNil(0), nil
	}

	enc := make([]byte, 0)

	for i := 0; i < refVal.Len(); i++ {
		v := refVal.Index(i)
		encData, err := Enc(v.Interface())
		if err != nil {
			return nil, reziError{
				msg:   fmt.Sprintf("slice[%d]: %s", i, err.Error()),
				cause: []error{err},
			}
		}
		enc = append(enc, encData...)
	}

	enc = append(encInt(tLen(len(enc))), enc...)
	return enc, nil
}

func decCheckedSlice(data []byte, v interface{}, ti typeInfo) (int, error) {
	if ti.Main != mtSlice {
		panic("not a slice type")
	}

	sl, n, err := decWithNilCheck(data, v, ti, fn_DecToWrappedReceiver(v, ti,
		func(t reflect.Type) bool {
			return t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Slice
		},
		decSlice,
	))
	if err != nil {
		return n, err
	}
	if ti.Indir == 0 {
		refReceiver := reflect.ValueOf(v)
		refReceiver.Elem().Set(reflect.ValueOf(sl))
	}
	return n, err
}

func decSlice(data []byte, v interface{}) (int, error) {
	var totalConsumed int

	toConsume, n, err := decInt[tLen](data)
	if err != nil {
		return 0, reziError{
			msg:   fmt.Sprintf("decode byte count: %s", err.Error()),
			cause: []error{err},
		}
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
		return totalConsumed, reziError{
			cause: []error{io.ErrUnexpectedEOF, ErrMalformedData},
		}
	}

	// clamp values we are allowed to read so we don't try to read other data
	data = data[:toConsume]

	sl := reflect.MakeSlice(refSliceType, 0, 0)

	var i int
	for i < toConsume {
		refVType := refSliceType.Elem()
		refValue := reflect.New(refVType)
		n, err := Dec(data, refValue.Interface())
		if err != nil {
			return totalConsumed, reziError{
				msg:   fmt.Sprintf("slice item: %s", err.Error()),
				cause: []error{err},
			}
		}
		totalConsumed += n
		i += n
		data = data[n:]

		sl = reflect.Append(sl, refValue.Elem())
	}

	refSliceVal.Elem().Set(sl)
	return totalConsumed, nil
}
