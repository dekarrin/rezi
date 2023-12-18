package rezi

// slices.go contains functions for encoding and decoding slices and arrays of
// basic types.

import (
	"io"
	"reflect"
)

// encMap encodes a compatible slice as a REZI map.
func encCheckedSlice(value analyzed[any]) ([]byte, error) {
	if value.info.Main != mtSlice && value.info.Main != mtArray {
		panic("not a slice or array type")
	}

	return encWithNilCheck(value, encSlice, reflect.Value.Interface)
}

func encSlice(value analyzed[any]) ([]byte, error) {
	isArray := value.reflect.Type().Kind() == reflect.Array

	if value.v == nil || (!isArray && value.reflect.IsNil()) {
		return encNilHeader(0), nil
	}

	enc := make([]byte, 0)

	for i := 0; i < value.reflect.Len(); i++ {
		v := value.reflect.Index(i)
		encData, err := encWithTypeInfo(v.Interface(), *value.info.ValType)
		if err != nil {
			if isArray {
				return nil, errorf("array item[%d]: %s", i, err)
			} else {
				return nil, errorf("slice item[%d]: %s", i, err)
			}
		}
		enc = append(enc, encData...)
	}

	enc = append(encCount(len(enc), nil), enc...)
	return enc, nil
}

func decCheckedSlice(data []byte, recv analyzed[any]) (decoded[any], error) {
	if recv.info.Main != mtSlice && recv.info.Main != mtArray {
		panic("not a slice or array type")
	}

	sl, err := decWithNilCheck(data, recv, fn_DecToWrappedReceiver(recv,
		func(t reflect.Type) bool {
			return t.Kind() == reflect.Pointer && ((recv.info.Main == mtSlice && t.Elem().Kind() == reflect.Slice) || (recv.info.Main == mtArray && t.Elem().Kind() == reflect.Array))
		},
		decSlice,
	))
	if err != nil {
		return sl, err
	}
	if recv.info.Indir == 0 {
		refReceiver := recv.reflect
		refReceiver.Elem().Set(sl.reflect)
	}
	return sl, err
}

func decSlice(data []byte, recv analyzed[any]) (decoded[any], error) {
	var dec decoded[any]

	toConsume, err := decInt[tLen](data)
	if err != nil {
		return dec, errorDecf(0, "decode byte count: %s", err)
	}
	data = data[toConsume.n:]
	dec.n += toConsume.n

	refSliceVal := recv.reflect
	refSliceType := refSliceVal.Type().Elem()
	isArray := refSliceType.Kind() == reflect.Array
	sliceOrArrStr := "slice"
	var refArrType reflect.Type
	if isArray {
		refArrType = reflect.ArrayOf(refSliceType.Len(), refSliceType.Elem())
		sliceOrArrStr = "array"
	}

	if toConsume.v == 0 {
		// initialize to the empty slice/array
		var empty reflect.Value
		if isArray {
			empty = reflect.New(refArrType).Elem()
		} else {
			empty = reflect.MakeSlice(refSliceType, 0, 0)
		}
		refSliceVal.Elem().Set(empty)
		dec.v = refSliceVal.Elem().Interface()
		dec.reflect = refSliceVal.Elem()
		return dec, nil
	} else if toConsume.v == -1 {
		var nilVal reflect.Value
		if isArray {
			nilVal = reflect.Zero(refArrType)
		} else {
			nilVal = reflect.Zero(refSliceType)
		}
		refSliceVal.Elem().Set(nilVal)
		dec.v = nilVal.Interface()
		dec.reflect = nilVal
		return dec, nil
	}

	if len(data) < toConsume.v {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded %s byte count is %d but only %d byte%s remain%s in data at offset"
		err := errorDecf(dec.n, errFmt, sliceOrArrStr, toConsume.v, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return dec, err
	}

	// clamp values we are allowed to read so we don't try to read other data
	data = data[:toConsume.v]

	var sl reflect.Value

	if isArray {
		sl = reflect.New(refArrType).Elem()
	} else {
		sl = reflect.MakeSlice(refSliceType, 0, 0)
	}

	var i int
	var itemIdx int
	refVType := refSliceType.Elem()
	for i < toConsume.v {
		refValue := reflect.New(refVType)
		n, err := decWithTypeInfo(data, refValue.Interface(), *recv.info.ValType)
		if err != nil {
			return dec, errorDecf(dec.n, "%s item[%d]: %s", sliceOrArrStr, itemIdx, err)
		}
		dec.n += n
		i += n
		data = data[n:]

		if isArray {
			sl.Index(itemIdx).Set(refValue.Elem())
		} else {
			sl = reflect.Append(sl, refValue.Elem())
		}
		itemIdx++
	}

	refSliceVal.Elem().Set(sl)
	dec.v = sl.Interface()
	dec.reflect = sl
	return dec, nil
}
