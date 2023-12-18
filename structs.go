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

func encStruct(value analyzed[any]) ([]byte, error) {
	enc := make([]byte, 0)

	for _, fi := range value.ti.Fields.ByOrder {
		v := value.ref.Field(fi.Index)

		fNameData, err := encWithTypeInfo(fi.Name, typeInfo{Indir: 0, Underlying: false, Main: mtString})
		if err != nil {
			msgTypeName := value.ref.Type().Name()
			if msgTypeName == "" {
				msgTypeName = "(anonymous type)"
			}
			return nil, errorf("%s.%s field name: %s", msgTypeName, fi.Name, err)
		}
		fValData, err := encWithTypeInfo(v.Interface(), fi.Type)
		if err != nil {
			msgTypeName := value.ref.Type().Name()
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
func decCheckedStruct(data []byte, recv analyzed[any]) (decoded[any], error) {
	if recv.ti.Main != mtStruct {
		panic("not a struct type")
	}

	st, err := decWithNilCheck(data, recv, fn_DecToWrappedReceiver(recv,
		func(t reflect.Type) bool {
			return t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct
		},
		decStruct,
	))
	if err != nil {
		return st, err
	}
	if recv.ti.Indir == 0 {
		refReceiver := recv.ref

		// if it's a struct, we must get the original value, if one exists, in order
		// to preserve the original member values
		var origStructVal reflect.Value
		if recv.ti.Main == mtStruct {
			origStructVal = unwrapOriginalStructValue(refReceiver)
		}

		refSt := st.ref

		if recv.ti.Main == mtStruct && origStructVal.IsValid() {
			refSt = makeStructWithFieldValues(origStructVal, refSt, st.fields)
		}

		refReceiver.Elem().Set(refSt)
	}
	return st, err
}

// decInfo will have Fields set to the successfully decoded fields.
func decStruct(data []byte, recv analyzed[any]) (decoded[any], error) {
	var dec decoded[any]

	refVal := recv.ref
	refStructType := refVal.Type().Elem()
	msgTypeName := refStructType.Name()
	if msgTypeName == "" {
		msgTypeName = "(anonymous type)"
	}

	toConsume, err := decInt[tLen](data)
	if err != nil {
		return dec, errorDecf(0, "decode %s byte count: %s", msgTypeName, err)
	}
	data = data[toConsume.n:]
	dec.n += toConsume.n

	if toConsume.native == 0 {
		// initialize to an empty struct
		emptyStruct := reflect.New(refStructType)

		// set it to the value
		refVal.Elem().Set(emptyStruct.Elem())
		dec.native = emptyStruct.Elem().Interface()
		dec.ref = emptyStruct.Elem()
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
		err := errorDecf(dec.n, errFmt, msgTypeName, toConsume.native, len(data), s, verbS).wrap(io.ErrUnexpectedEOF, ErrMalformedData)
		return dec, err
	}

	// clamp values we are allowed to read so we don't try to read other data
	data = data[:toConsume.native]

	target := refVal.Elem()
	var i int
	for i < toConsume.native {
		// get field name
		var fNameVal string
		n, err := decWithTypeInfo(data, &fNameVal, typeInfo{Indir: 0, Underlying: false, Main: mtString, Dec: true})
		if err != nil {
			return dec, errorDecf(dec.n, "decode %s field name: %s", msgTypeName, err)
		}
		dec.n += n
		i += n
		data = data[n:]

		// get field info from name
		fi, ok := recv.ti.Fields.ByName[fNameVal]
		if !ok {
			return dec, errorDecf(dec.n, "field name .%s does not exist in decoded-to %s", fNameVal, msgTypeName).wrap(ErrMalformedData, ErrInvalidType)
		}
		fieldPtr := target.Field(fi.Index).Addr()
		n, err = decWithTypeInfo(data, fieldPtr.Interface(), fi.Type)
		if err != nil {
			return dec, errorDecf(dec.n, "%s.%s: %v", msgTypeName, fi.Name, err)
		}
		dec.n += n
		i += n
		data = data[n:]
		dec.fields = append(dec.fields, fi)
	}

	dec.native = target.Interface()
	dec.ref = target
	return dec, nil
}

func makeStructWithFieldValues(initial, setValues reflect.Value, affectedFields []fieldInfo) reflect.Value {
	newVal := reflect.New(initial.Type())
	newVal.Elem().Set(initial)

	for _, fi := range affectedFields {
		destPtr := newVal.Elem().Field(fi.Index).Addr()
		fieldVal := setValues.Field(fi.Index)
		destPtr.Elem().Set(fieldVal)
	}

	return newVal.Elem()
}

// this will return nil if v does not end up in a struct value after
// dereferences are made
func unwrapOriginalStructValue(refVal reflect.Value) reflect.Value {
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
