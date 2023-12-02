package rezi

import (
	"io"
	"reflect"
)

// encCheckedStruct encodes a compatible struct as a REZI .
func encCheckedStruct(value analyzed[any]) ([]byte, error) {
	if value.ti.Main != mtStruct {
		panic("not a struct type")
	}

	return encWithNilCheck(value, encStruct, reflect.Value.Interface)
}

func encStruct(val analyzed[any]) ([]byte, error) {
	enc := make([]byte, 0)

	for _, fi := range val.ti.Fields.ByOrder {
		v := val.ref.Field(fi.Index)

		fNameData, err := Enc(fi.Name)
		if err != nil {
			msgTypeName := reflect.ValueOf(v).Type().Name()
			if msgTypeName == "" {
				msgTypeName = "(anonymous type)"
			}
			return nil, errorf("%s.%s field name: %s", msgTypeName, fi.Name, err)
		}
		fValData, err := Enc(v.Interface())
		if err != nil {
			msgTypeName := reflect.ValueOf(v).Type().Name()
			if msgTypeName == "" {
				msgTypeName = "(anonymous type)"
			}
			return nil, errorf("%s.%s: %v", msgTypeName, fi.Name, err)
		}

		enc = append(enc, fNameData...)
		enc = append(enc, fValData...)
	}

	enc = append(encCount(len(enc), nil), enc...)
	return enc, nil
}

// decCheckedStruct decodes a REZI bytes representation of a struct into a
// compatible struct type.
func decCheckedStruct(data []byte, v interface{}, ti typeInfo) (int, error) {
	if ti.Main != mtStruct {
		panic("not a struct type")
	}

	var extraInfo []fieldInfo
	st, n, err := decWithNilCheck(data, v, ti, fn_DecToWrappedReceiver(v, ti,
		func(t reflect.Type) bool {
			return t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct
		},
		func(data []byte, v interface{}) (interface{}, int, error) {
			fi, n, err := decStruct(data, v, ti)
			extraInfo = fi
			return fi, n, err
		},
	))
	if err != nil {
		return n, err
	}
	if ti.Indir == 0 {
		refReceiver := reflect.ValueOf(v)

		// if it's a struct, we must get the original value, if one exists, in order
		// to preserve the original member values
		var origStructVal reflect.Value
		if ti.Main == mtStruct {
			origStructVal = unwrapOriginalStructValue(refReceiver)
		}

		refSt := reflect.ValueOf(st)

		if ti.Main == mtStruct && origStructVal.IsValid() {
			refSt = setStructMembers(origStructVal, refSt, extraInfo)
		}

		refReceiver.Elem().Set(refSt)
	}
	return n, err
}

// the fieldInfo is the successfully decoded fields.
func decStruct(data []byte, v interface{}, ti typeInfo) ([]fieldInfo, int, error) {
	var decFields []fieldInfo
	var totalConsumed int

	refVal := reflect.ValueOf(v)
	refStructType := refVal.Type().Elem()
	msgTypeName := refStructType.Name()
	if msgTypeName == "" {
		msgTypeName = "(anonymous type)"
	}

	toConsume, _, n, err := decInt[tLen](data)
	if err != nil {
		return decFields, 0, errorDecf(0, "decode %s byte count: %s", msgTypeName, err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		// initialize to an empty struct
		emptyStruct := reflect.New(refStructType)

		// set it to the value
		refVal.Elem().Set(emptyStruct.Elem())
		return decFields, totalConsumed, nil
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
		return decFields, totalConsumed, err
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
			return decFields, totalConsumed, errorDecf(totalConsumed, "decode %s field name: %s", msgTypeName, err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		// get field info from name
		fi, ok := ti.Fields.ByName[fNameVal]
		if !ok {
			return decFields, totalConsumed, errorDecf(totalConsumed, "field name .%s does not exist in decoded-to %s", fNameVal, msgTypeName).wrap(ErrMalformedData, ErrInvalidType)
		}
		fieldPtr := target.Field(fi.Index).Addr()
		n, err = Dec(data, fieldPtr.Interface())
		if err != nil {
			return decFields, totalConsumed, errorDecf(totalConsumed, "%s.%s: %v", msgTypeName, fi.Name, err)
		}
		totalConsumed += n
		i += n
		data = data[n:]
		decFields = append(decFields, fi)
	}

	return decFields, totalConsumed, nil
}

func setStructMembers(initial, decoded reflect.Value, decodedFields []fieldInfo) reflect.Value {
	newVal := reflect.New(initial.Type())
	newVal.Elem().Set(initial)

	for _, fi := range decodedFields {
		destPtr := newVal.Elem().Field(fi.Index).Addr()
		fieldVal := decoded.Field(fi.Index)
		destPtr.Elem().Set(fieldVal)
	}

	return newVal.Elem()
}

// this will return nil if v does not end up in a struct value after
// dereferences are made
func unwrapOriginalStructValue(refVal reflect.Value) reflect.Value {
	// TODO: move all this to type analysis

	// the user may have passed in a ptr-ptr-to, make shore we get actual
	// target
	for refVal.Kind() == reflect.Pointer && !refVal.IsNil() {
		refVal = refVal.Elem()
	}

	// only pick up orig value if we ended up at a struct type
	if refVal.Kind() == reflect.Struct {
		return refVal
	}

	return reflect.Value{}
}
