package rezi

import (
	"encoding"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EncMapStringToInt(t *testing.T) {
	testCases := []struct {
		name   string
		input  map[string]int
		expect []byte
	}{
		{
			name:   "empty",
			input:  map[string]int{},
			expect: []byte{0x00},
		},
		{
			name:   "nil",
			input:  nil,
			expect: []byte{0x80},
		},
		{
			name: "one item",
			input: map[string]int{
				"entry": 1,
			},
			expect: []byte{
				0x01, 0x09,
				0x01, 0x05, 0x65, 0x6e, 0x74, 0x72, 0x79,
				0x01, 0x01,
			},
		},
		{
			name: "multiple items",
			input: map[string]int{
				"first":  1,
				"second": 35,
				"count":  2,
			},
			expect: []byte{
				0x01, 0x1c,

				// is always alphabetical, so count first
				0x01, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
				0x01, 0x02,

				// first
				0x01, 0x05, 0x66, 0x69, 0x72, 0x73, 0x74,
				0x01, 0x01,

				// second
				0x01, 0x06, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64,
				0x01, 0x23,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncMapStringToInt(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_DecMapStringToInt(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectValue map[string]int
		expectRead  int
		expectError bool
	}{
		{
			name:        "empty",
			input:       []byte{0x00},
			expectValue: map[string]int{},
			expectRead:  1,
		},
		{
			name:        "nil",
			input:       []byte{0x80},
			expectValue: nil,
			expectRead:  1,
		},
		{
			name: "one item, and extra bytes",
			input: []byte{
				0x01, 0x09,
				0x01, 0x05, 0x65, 0x6e, 0x74, 0x72, 0x79,
				0x01, 0x01,

				// extra
				0x023, 0x04, 0x0ff, 0xee,
			},
			expectValue: map[string]int{
				"entry": 1,
			},
			expectRead: 11,
		},
		{
			name: "multiple items",
			input: []byte{
				0x01, 0x1c,

				// is always alphabetical, so count first
				0x01, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
				0x01, 0x02,

				// first
				0x01, 0x05, 0x66, 0x69, 0x72, 0x73, 0x74,
				0x01, 0x01,

				// second
				0x01, 0x06, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64,
				0x01, 0x23,
			},
			expectValue: map[string]int{
				"first":  1,
				"second": 35,
				"count":  2,
			},
			expectRead: 30,
		},
		{
			name: "not enough bytes for map",
			input: []byte{
				0x01, 0x15,
				0x00, 0x00, 0x00,
			},
			expectError: true,
		},
		{
			name: "not enough bytes for map key",
			input: []byte{
				0x01, 0x06,
				0x01, 0x05, 0x65, 0x6e, 0x74, 0x72,
			},
			expectError: true,
		},
		{
			name: "not enough bytes for map value",
			input: []byte{
				0x01, 0x08,
				0x01, 0x05, 0x65, 0x6e, 0x74, 0x72, 0x79,
				0x01,
			},
			expectError: true,
		},
		{
			name: "missing map value",
			input: []byte{
				0x01, 0x07,
				0x01, 0x05, 0x65, 0x6e, 0x74, 0x72, 0x79,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualValue, actualRead, actualError := DecMapStringToInt(tc.input)
			if tc.expectError {
				assert.Error(actualError)
				return
			} else if !assert.NoError(actualError) {
				return
			}

			assert.Equal(tc.expectValue, actualValue)
			assert.Equal(tc.expectRead, actualRead)
		})
	}
}

func Test_EncMapStringToBinary(t *testing.T) {
	testCases := []struct {
		name   string
		input  map[string]encoding.BinaryMarshaler
		expect []byte
	}{
		{
			name:   "empty",
			input:  map[string]encoding.BinaryMarshaler{},
			expect: []byte{0x00},
		},
		{
			name:   "nil",
			input:  nil,
			expect: []byte{0x80},
		},
		{
			name: "one item",
			input: map[string]encoding.BinaryMarshaler{
				"item": valueThatMarshalsWith(func() []byte {
					return []byte{0x03, 0xaf, 0xff, 0xde, 0x83}
				}),
			},
			expect: []byte{
				0x01, 0x0d,

				0x01, 0x04, 0x69, 0x74, 0x65, 0x6d,
				0x01, 0x05, 0x03, 0xaf, 0xff, 0xde, 0x83,
			},
		},
		{
			name: "multiple items",
			input: map[string]encoding.BinaryMarshaler{
				"someItem": valueThatMarshalsWith(func() []byte {
					return []byte{0x03, 0xaf, 0xff, 0xde, 0x83}
				}),
				"other": valueThatMarshalsWith(func() []byte {
					return []byte{}
				}),
				"Français": valueThatMarshalsWith(func() []byte {
					return []byte{0x04, 0x04, 0x04, 0xe6}
				}),
			},
			expect: []byte{
				0x01, 0x2a,

				0x01, 0x08, 0x46, 0x72, 0x61, 0x6e, 0xc3, 0xa7, 0x61, 0x69, 0x73,
				0x01, 0x04, 0x04, 0x04, 0x04, 0xe6,

				0x01, 0x05, 0x6f, 0x74, 0x68, 0x65, 0x72,
				0x00,

				0x01, 0x08, 0x73, 0x6f, 0x6d, 0x65, 0x49, 0x74, 0x65, 0x6d,
				0x01, 0x05, 0x03, 0xaf, 0xff, 0xde, 0x83,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncMapStringToBinary(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_DecMapStringToBinary(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectValue map[string]*marshaledBytesReceiver
		expectRead  int
		expectError bool
	}{
		{
			name:        "empty",
			input:       []byte{0x00},
			expectValue: map[string]*marshaledBytesReceiver{},
			expectRead:  1,
		},
		{
			name:        "nil",
			input:       []byte{0x80},
			expectValue: nil,
			expectRead:  1,
		},
		{
			name: "one item, and extra bytes",
			input: []byte{
				0x01, 0x0d,

				0x01, 0x04, 0x69, 0x74, 0x65, 0x6d,
				0x01, 0x05, 0x03, 0xaf, 0xff, 0xde, 0x83,

				0xff, 0xff,
			},
			expectValue: map[string]*marshaledBytesReceiver{
				"item": {[]byte{0x03, 0xaf, 0xff, 0xde, 0x83}},
			},
			expectRead: 15,
		},
		{
			name: "multiple items",
			input: []byte{
				0x01, 0x2a,

				0x01, 0x08, 0x46, 0x72, 0x61, 0x6e, 0xc3, 0xa7, 0x61, 0x69, 0x73,
				0x01, 0x04, 0x04, 0x04, 0x04, 0xe6,

				0x01, 0x05, 0x6f, 0x74, 0x68, 0x65, 0x72,
				0x00,

				0x01, 0x08, 0x73, 0x6f, 0x6d, 0x65, 0x49, 0x74, 0x65, 0x6d,
				0x01, 0x05, 0x03, 0xaf, 0xff, 0xde, 0x83,
			},
			expectValue: map[string]*marshaledBytesReceiver{
				"someItem": {[]byte{0x03, 0xaf, 0xff, 0xde, 0x83}},
				"other":    {[]byte{}},
				"Français": {[]byte{0x04, 0x04, 0x04, 0xe6}},
			},
			expectRead: 44,
		},
		{
			name: "not enough bytes for map",
			input: []byte{
				0x01, 0x15,
				0x00, 0x00, 0x00,
			},
			expectError: true,
		},
		{
			name: "not enough bytes for map key",
			input: []byte{
				0x01, 0x06,
				0x01, 0x05, 0x65, 0x6e, 0x74, 0x72,
			},
			expectError: true,
		},
		{
			name: "not enough bytes for map value",
			input: []byte{
				0x01, 0x09,
				0x01, 0x05, 0x65, 0x6e, 0x74, 0x72, 0x79,
				0x01, 0x01,
			},
			expectError: true,
		},
		{
			name: "missing map value",
			input: []byte{
				0x01, 0x0d,
				0x01, 0x05, 0x65, 0x6e, 0x74, 0x72, 0x79,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualValue, actualRead, err := DecMapStringToBinary[*marshaledBytesReceiver](tc.input)
			if tc.expectError {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expectValue, actualValue)
			assert.Equal(tc.expectRead, actualRead)
		})
	}
}
