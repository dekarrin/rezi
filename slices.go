package rezi

// slices.go contains functions for encoding and decoding slices and arrays of
// basic types.

import (
	"io"
	"reflect"
)

// encMap encodes a compatible slice as a REZI map.
func encCheckedSlice(value analyzed[any]) ([]byte, error) {
	if value.ti.Main != mtSlice && value.ti.Main != mtArray {
		panic("not a slice or array type")
	}

	return encWithNilCheck(value, encSlice, reflect.Value.Interface)
}

func encSlice(value analyzed[any]) ([]byte, error) {
	isArray := value.ref.Type().Kind() == reflect.Array

	if value.native == nil || (!isArray && value.ref.IsNil()) {
		return encNilHeader(0), nil
	}

	enc := make([]byte, 0)

	for i := 0; i < value.ref.Len(); i++ {
		v := value.ref.Index(i)
		encData, err := encWithTypeInfo(v.Interface(), *value.ti.ValType)
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
	if recv.ti.Main != mtSlice && recv.ti.Main != mtArray {
		panic("not a slice or array type")
	}

	sl, err := decWithNilCheck(data, recv, fn_DecToWrappedReceiver(recv,
		func(t reflect.Type) bool {
			return t.Kind() == reflect.Pointer && ((recv.ti.Main == mtSlice && t.Elem().Kind() == reflect.Slice) || (recv.ti.Main == mtArray && t.Elem().Kind() == reflect.Array))
		},
		decSlice,
	))
	if err != nil {
		return sl, err
	}
	if recv.ti.Indir == 0 {
		refReceiver := recv.ref
		refReceiver.Elem().Set(sl.ref)
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

	refSliceVal := recv.ref
	refSliceType := refSliceVal.Type().Elem()
	isArray := refSliceType.Kind() == reflect.Array
	sliceOrArrStr := "slice"
	var refArrType reflect.Type
	if isArray {
		refArrType = reflect.ArrayOf(refSliceType.Len(), refSliceType.Elem())
		sliceOrArrStr = "array"
	}

	if toConsume.native == 0 {
		// initialize to the empty slice/array
		var empty reflect.Value
		if isArray {
			empty = reflect.New(refArrType).Elem()
		} else {
			empty = reflect.MakeSlice(refSliceType, 0, 0)
		}
		refSliceVal.Elem().Set(empty)
		dec.native = refSliceVal.Elem().Interface()
		dec.ref = refSliceVal.Elem()
		return dec, nil
	} else if toConsume.native == -1 {
		var nilVal reflect.Value
		if isArray {
			nilVal = reflect.Zero(refArrType)
		} else {
			nilVal = reflect.Zero(refSliceType)
		}
		refSliceVal.Elem().Set(nilVal)
		dec.native = nilVal.Interface()
		dec.ref = nilVal
		return dec, nil
	}

	if len(data) < toConsume.native {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded %s byte count is %d but only %d byte%s remain%s in data at offset"
		err := errorDecf(dec.n, errFmt, sliceOrArrStr, toConsume.native, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return dec, err
	}

	// clamp values we are allowed to read so we don't try to read other data
	data = data[:toConsume.native]

	var sl reflect.Value

	if isArray {
		sl = reflect.New(refArrType).Elem()
	} else {
		sl = reflect.MakeSlice(refSliceType, 0, 0)
	}

	var i int
	var itemIdx int
	refVType := refSliceType.Elem()
	for i < toConsume.native {
		refValue := reflect.New(refVType)
		n, err := decWithTypeInfo(data, refValue.Interface(), *recv.ti.ValType)
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
	dec.native = sl.Interface()
	dec.ref = sl
	return dec, nil
}
