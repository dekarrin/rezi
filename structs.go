package rezi

import (
	"io"
	"reflect"
)

// encCheckedStruct encodes a compatible struct as a REZI .
func encCheckedStruct(v interface{}, ti typeInfo) ([]byte, error) {
	if ti.Main != mtStruct {
		panic("not a struct type")
	}

	return encWithNilCheck(v, ti, func(val interface{}) ([]byte, error) {
		return encStruct(val, ti)
	}, reflect.Value.Interface)
}

func encStruct(v interface{}, ti typeInfo) ([]byte, error) {
	refVal := reflect.ValueOf(v)

	if v == nil || refVal.IsNil() {
		return encNilHeader(0), nil
	}

	enc := make([]byte, 0)

	for _, fi := range ti.Fields.ByOrder {
		v := refVal.Field(fi.Index)
		fValData, err := Enc(v.Interface())
		if err != nil {
			return nil, errorf("field .%s: %v", fi.Name, err)
		}
		fNameData, err := Enc(fi.Name)
		if err != nil {
			return nil, errorf("field name .%s: %s", fi.Name, err)
		}

		enc = append(enc, fNameData...)
		enc = append(enc, fValData...)
	}

	enc = append(encInt(tLen(len(enc))), enc...)
	return enc, nil
}

// decCheckedStruct decodes a REZI bytes representation of a struct into a
// compatible struct type.
func decCheckedStruct(data []byte, v interface{}, ti typeInfo) (int, error) {
	if ti.Main != mtStruct {
		panic("not a struct type")
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

func decStruct(data []byte, v interface{}) (int, error) {
	var totalConsumed int

	toConsume, n, err := decInt[tLen](data)
	if err != nil {
		return 0, errorDecf(0, "decode byte count: %s", err)
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
