package rezi

// slices.go contains functions for encoding and decoding slices of basic types.

import (
	"encoding"
	"fmt"
)

func EncSliceString(sl []string) []byte {
	if sl == nil {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	for i := range sl {
		enc = append(enc, EncString(sl[i])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

func DecSliceString(data []byte) ([]string, int, error) {
	var totalConsumed int

	toConsume, n, err := DecInt(data)
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
		return nil, 0, fmt.Errorf("unexpected EOF")
	}

	sl := []string{}

	var i int
	for i < toConsume {
		s, n, err := DecString(data)
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

func EncSliceBinary[E encoding.BinaryMarshaler](sl []E) []byte {
	if sl == nil {
		return EncInt(-1)
	}

	enc := make([]byte, 0)

	for i := range sl {
		enc = append(enc, EncBinary(sl[i])...)
	}

	enc = append(EncInt(len(enc)), enc...)
	return enc
}

func DecSliceBinary[E encoding.BinaryUnmarshaler](data []byte) ([]E, int, error) {
	var totalConsumed int

	toConsume, n, err := DecInt(data)
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
		return nil, 0, fmt.Errorf("unexpected EOF")
	}

	sl := []E{}

	var i int
	for i < toConsume {
		v := initType[E]()

		n, err := DecBinary(data, v)
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
