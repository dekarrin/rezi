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

	enc := make([]byte, 0)

	for _, fi := range ti.Fields.ByOrder {
		v := refVal.Field(fi.Index)

		fNameData, err := Enc(fi.Name)
		if err != nil {
			return nil, errorf("field name .%s: %s", fi.Name, err)
		}
		fValData, err := Enc(v.Interface())
		if err != nil {
			return nil, errorf("field .%s: %v", fi.Name, err)
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

	st, n, err := decWithNilCheck(data, v, ti, fn_DecToWrappedReceiver(v, ti,
		func(t reflect.Type) bool {
			// TODO: might need to remove reflect.Pointer
			return t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct
		},
		func(data []byte, v interface{}) (int, error) {
			return decStruct(data, v, ti)
		},
	))
	if err != nil {
		return n, err
	}
	if ti.Indir == 0 {
		refReceiver := reflect.ValueOf(v)
		refReceiver.Elem().Set(reflect.ValueOf(st))
	}
	return n, err
}

func decStruct(data []byte, v interface{}, ti typeInfo) (int, error) {
	var totalConsumed int

	refVal := reflect.ValueOf(v)
	refStructType := refVal.Type().Elem()
	msgTypeName := refStructType.Name()
	if msgTypeName == "" {
		msgTypeName = "(anonymous type)"
	}

	toConsume, n, err := decInt[tLen](data)
	if err != nil {
		return 0, errorDecf(0, "decode %s byte count: %s", msgTypeName, err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		// initialize to an empty struct
		emptyStruct := reflect.New(refStructType)

		// set it to the value
		refVal.Elem().Set(emptyStruct)
		return totalConsumed, nil
	}

	if len(data) < toConsume {
		s := "s"
		verbS := ""
		if len(data) == 1 {
			s = ""
			verbS = "s"
		}
		const errFmt = "decoded %s byte count is %d but only %d byte%s remain%s in data at offset"
		err := errorDecf(totalConsumed, errFmt, msgTypeName, toConsume, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return totalConsumed, err
	}

	// clamp values we are allowed to read so we don't try to read other data
	data = data[:toConsume]

	target := refVal.Elem()
	var i int
	for i < toConsume {
		// get field name
		var fNameVal string
		n, err = Dec(data, &fNameVal)
		if err != nil {
			return totalConsumed, errorDecf(totalConsumed, "decode field name: %s", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		// get field info from name
		fi, ok := ti.Fields.ByName[fNameVal]
		if !ok {
			return totalConsumed, errorDecf(totalConsumed, "decoded field name .%s does not exist in decoded-to struct", fNameVal).wrap(ErrMalformedData, ErrInvalidType)
		}
		fieldPtr := target.Field(fi.Index).Addr()
		n, err = Dec(data, fieldPtr.Interface())
		if err != nil {
			return totalConsumed, errorDecf(totalConsumed, "field .%s: %v", fi.Name, err)
		}
		totalConsumed += n
		i += n
		data = data[n:]
	}

	return totalConsumed, nil
}
