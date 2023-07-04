package rezi

// slices.go contains functions for encoding and decoding slices of basic types.

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
)

// encMap encodes a compatible slice as a REZI map.
func encCheckedSlice(v interface{}, ti typeInfo) []byte {
	if ti.Main != tSlice {
		panic("not a slice type")
	}

	return encWithNilCheck(v, ti, encSlice, reflect.Value.Interface)
}

func encSlice(v interface{}) []byte {
	refVal := reflect.ValueOf(v)

	if v == nil || refVal.IsNil() {
		return encNil(0)
	}

	enc := make([]byte, 0)

	for i := 0; i < refVal.Len(); i++ {
		v := refVal.Index(i)
		enc = append(enc, Enc(v.Interface())...)
	}

	enc = append(encInt(tLen(len(enc))), enc...)
	return enc
}

func decCheckedSlice(data []byte, v interface{}, ti typeInfo) (int, error) {
	if ti.Main != tSlice {
		panic("not a slice type")
	}

	// TODO: adapting this code to slice.

	// sl, n, err := decWithNilCheck(data, v, ti, func(b []byte) (interface{}, int, error) {
	// 	// v is *(...*)T, ret-val of decFn (this lambda) is T.
	// 	// v is *(...*)[]T, ret-val of decFn is []T.

	// 	receiverType := reflect.TypeOf(v)
	// 	// if v is a *T, we are done. but it could be a **T. check now.
	// 	// if v is a *[]T, we are done, but it could be a **[]T, check now.

	// 	if receiverType.Kind() == reflect.Pointer { // future-proofing - binary unmarshaler might come in as a T
	// 		// for every * in the (...*) part of *(...*)T up until the
	// 		// implementor/slice-ptr, do a deref.
	// 		for i := 0; i < ti.Indir; i++ {
	// 			receiverType = receiverType.Elem()
	// 		}
	// 	}

	// 	/* CHOICE START { */
	// 	// receiverType should now be the exact type which implements
	// 	// encoding.BinaryUnmarshaler. Assert this for now.
	// 	if !receiverType.Implements(refBinaryUnmarshalerType) {
	// 		// should never happen, assuming ti.Indir is valid.
	// 		panic("unwrapped binary type receiver does not implement encoding.BinaryUnmarshaler")
	// 	}
	// 	// receiverType should now be the exact ptr-to-slice type. Assert this
	// 	// for now.
	// 	if !(receiverType.Kind() == reflect.Pointer && receiverType.Elem().Kind() == reflect.Slice) {
	// 		// should never happen, assuming ti.Indir is valid.
	// 		panic("unwrapped receiver is not compatible with encoded value")
	// 	}
	// 	/* } CHOICE END */

	// 	var receiverValue reflect.Value
	// 	if receiverType.Kind() == reflect.Pointer {
	// 		// receiverType is *T
	// 		// receiverType is *[]T
	// 		receiverValue = reflect.New(receiverType.Elem())
	// 	} else {
	// 		// receiverType is itself T (future-proofing)
	// 		receiverValue = reflect.Zero(receiverType)
	// 	}

	// 	var decoded interface{}

	// 	receiver := receiverValue.Interface()

	// 	/* CHOICE START (ONLY BIN) { */
	// 	binReceiver := receiver.(encoding.BinaryUnmarshaler)
	// 	/* } */

	// 	var decConsumed int
	// 	var decErr error

	// 	/* CHOICE START { */
	// 	decConsumed, decErr = decBinary(data, binReceiver)
	// 	decConsumed, decErr = decSlice(data, receiver, ti)
	// 	/* CHOICE END } */

	// 	if decErr != nil {
	// 		return nil, decConsumed, decErr
	// 	}

	// 	if receiverType.Kind() == reflect.Pointer {
	// 		decoded = reflect.ValueOf(receiver).Elem().Interface()
	// 	} else {
	// 		decoded = receiver
	// 	}

	// 	return decoded, decConsumed, decErr
	// })
	// if ti.Indir == 0 {
	// 	// assume v is a *T, no future-proofing here.

	// 	// due to complicated forcing of decBinary into the decFunc API,
	// 	// we do now have a T (as an interface{}). We must use reflection to
	// 	// assign it.

	// 	refReceiver := reflect.ValueOf(v)
	// 	refReceiver.Elem().Set(reflect.ValueOf(sl))
	// }
	// if err != nil {
	// 	return n, err
	// }
	return decSlice(data, v, ti)
}

func decSlice(data []byte, v interface{}, ti typeInfo) (int, error) {
	var totalConsumed int

	toConsume, n, err := decInt[tLen](data)
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
		return totalConsumed, io.ErrUnexpectedEOF
	}

	sl := reflect.MakeSlice(refSliceType, 0, 0)

	var i int
	for i < toConsume {
		refVType := refSliceType.Elem()
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
		enc = append(enc, encBinary(sl[i])...)
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
