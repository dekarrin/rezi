package rezi

// slices.go contains functions for encoding and decoding slices of basic types.

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
)

// encMap encodes a compatible slice as a REZI map.
func encCheckedSlice(v interface{}, ti typeInfo) ([]byte, error) {
	if ti.Main != tSlice {
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
	if ti.Main != tSlice {
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

// EncSliceString encodes a slice of strings from raw REZI bytes.
//
// Deprecated: This function has been replaced by [Enc].
func EncSliceString(sl []string) []byte {
	if sl == nil {
		return encNil(0)
	}

	enc := make([]byte, 0)

	for i := range sl {
		enc = append(enc, encString(sl[i])...)
	}

	enc = append(encInt(tLen(len(enc))), enc...)
	return enc
}

// DecSliceString decodes a slice of strings from raw REZI bytes.
//
// Deprecated: This function has been replaced by [Dec].
func DecSliceString(data []byte) ([]string, int, error) {
	var totalConsumed int

	toConsume, n, err := decInt[tLen](data)
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
		return nil, 0, io.ErrUnexpectedEOF
	}

	sl := []string{}

	var i int
	for i < toConsume {
		s, n, err := decString(data)
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

// EncSliceBinary encodes a slice of implementors of encoding.BinaryMarshaler
// from the data bytes.
//
// Deprecated: This function has been replaced by [Enc].
func EncSliceBinary[E encoding.BinaryMarshaler](sl []E) []byte {
	if sl == nil {
		return encNil(0)
	}

	enc := make([]byte, 0)

	for i := range sl {
		enc = append(enc, EncBinary(sl[i])...)
	}

	enc = append(encInt(len(enc)), enc...)
	return enc
}

// DecSliceBinary decodes a slice of implementors of encoding.BinaryUnmarshaler
// from the data bytes.
//
// Deprecated: This function requires the slice value type to directly implement
// encoding.BinaryUnmarshaler. Use [Dec] instead, which allows any type as a
// slice value provided that a *pointer* to it implements
// encoding.BinaryUnmarshaler.
func DecSliceBinary[E encoding.BinaryUnmarshaler](data []byte) ([]E, int, error) {
	var totalConsumed int

	toConsume, n, err := decInt[tLen](data)
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
		return nil, 0, io.ErrUnexpectedEOF
	}

	sl := []E{}

	var i int
	for i < toConsume {
		v := initType[E]()

		n, err := decBinary(data, v)
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
