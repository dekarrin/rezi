package rezi

// stringmaps.go contains functions for encoding and decoding maps of string to
// the basic types.

import (
	"encoding"
	"fmt"
	"io"
	"sort"
)

// EncMapStringToInt encodes a map of string-to-int as bytes. The order of keys
// in the output is gauranteed to be consistent.
//
// Deprecated: This function has been replaced by [Enc].
func EncMapStringToInt(m map[string]int) []byte {
	if m == nil {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := range keys {
		enc = append(enc, encString(keys[i])...)
		enc = append(enc, EncInt(m[keys[i]])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

// DecMapStringToBinary decodes a map of string-to-int from bytes.
//
// Deprecated: This function has been replaced by [Dec].
func DecMapStringToInt(data []byte) (map[string]int, int, error) {
	var totalConsumed int

	toConsume, n, err := decInt(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		return map[string]int{}, totalConsumed, nil
	} else if toConsume == -1 {
		return nil, totalConsumed, nil
	}

	if len(data) < toConsume {
		return nil, 0, io.ErrUnexpectedEOF
	}

	m := map[string]int{}

	var i int
	for i < toConsume {
		k, n, err := decString(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		v, n, err := decInt(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode value: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		m[k] = v
	}

	return m, totalConsumed, nil
}

// EncMapStringToBinary encodes a map of string to an implementer of
// encoding.BinaryMarshaler as bytes. The order of keys in output is gauranteed
// to be consistent.
//
// Deprecated: This function has been replaced by [Enc].
func EncMapStringToBinary[E encoding.BinaryMarshaler](m map[string]E) []byte {
	if m == nil {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := range keys {
		enc = append(enc, encString(keys[i])...)
		enc = append(enc, encBinary(m[keys[i]])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

// DecMapStringToBinary decodes a map of string to an implementer of
// encoding.BinaryMarshaler from bytes.
//
// Deprecated: This function has been replaced by [Dec].
func DecMapStringToBinary[E encoding.BinaryUnmarshaler](data []byte) (map[string]E, int, error) {
	var totalConsumed int

	toConsume, n, err := decInt(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode byte count: %w", err)
	}
	data = data[n:]
	totalConsumed += n

	if toConsume == 0 {
		return map[string]E{}, totalConsumed, nil
	} else if toConsume == -1 {
		return nil, totalConsumed, nil
	}

	if len(data) < toConsume {
		return nil, 0, io.ErrUnexpectedEOF
	}

	m := map[string]E{}

	var i int
	for i < toConsume {
		k, n, err := decString(data)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		v := initType[E]()
		n, err = decBinary(data, v)
		if err != nil {
			return nil, totalConsumed, fmt.Errorf("decode key: %w", err)
		}
		totalConsumed += n
		i += n
		data = data[n:]

		m[k] = v
	}

	return m, totalConsumed, nil
}
