package rezi

import (
	"encoding"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EncSliceString(t *testing.T) {
	testCases := []struct {
		name   string
		input  []string
		expect []byte
	}{
		{
			name:   "empty",
			input:  []string{},
			expect: []byte{0x00},
		},
		{
			name:   "nil",
			input:  nil,
			expect: []byte{0x80},
		},
		{
			name:   "one item",
			input:  []string{"one"},
			expect: []byte{0x01, 0x05, 0x01, 0x03, 0x6f, 0x6e, 0x65},
		},
		{
			name:   "two items",
			input:  []string{"one", "two"},
			expect: []byte{0x01, 0x0a, 0x01, 0x03, 0x6f, 0x6e, 0x65, 0x01, 0x03, 0x74, 0x77, 0x6f},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncSliceString(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_DecSliceString(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectValue []string
		expectRead  int
		expectError bool
	}{
		{
			name:        "empty",
			input:       []byte{0x00},
			expectValue: []string{},
			expectRead:  1,
		},
		{
			name:        "nil",
			input:       []byte{0x80},
			expectValue: nil,
			expectRead:  1,
		},
		{
			name:        "one item + some extra bytes",
			input:       []byte{0x01, 0x05, 0x01, 0x03, 0x6f, 0x6e, 0x65, 0x00, 0xfe},
			expectValue: []string{"one"},
			expectRead:  7,
		},
		{
			name:        "two items",
			input:       []byte{0x01, 0x0a, 0x01, 0x03, 0x6f, 0x6e, 0x65, 0x01, 0x03, 0x74, 0x77, 0x6f},
			expectValue: []string{"one", "two"},
			expectRead:  12,
		},
		{
			name:        "not enough bytes for list",
			input:       []byte{0x01, 0x05, 0x03, 0x6f, 0x6e, 0x6f},
			expectError: true,
		},
		{
			name:        "not enough bytes for list item",
			input:       []byte{0x01, 0x04, 0x04, 0x6f, 0x6e, 0x6a},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualValue, actualRead, err := DecSliceString(tc.input)
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

func Test_EncSliceBinary(t *testing.T) {
	testCases := []struct {
		name   string
		input  []encoding.BinaryMarshaler
		expect []byte
	}{
		{
			name:   "empty",
			input:  []encoding.BinaryMarshaler{},
			expect: []byte{0x00},
		},
		{
			name:   "nil",
			input:  nil,
			expect: []byte{0x80},
		},
		{
			name: "one item",
			input: []encoding.BinaryMarshaler{
				valueThatMarshalsWith(func() []byte {
					return []byte{0x03, 0xaf, 0xff, 0xde, 0x83}
				}),
			},
			expect: []byte{0x01, 0x07, 0x01, 0x05, 0x03, 0xaf, 0xff, 0xde, 0x83},
		},
		{
			name: "two items",
			input: []encoding.BinaryMarshaler{
				valueThatMarshalsWith(func() []byte {
					return []byte{}
				}),
				valueThatMarshalsWith(func() []byte {
					return []byte{0x28, 0x03, 0x00, 0x00, 0x12, 0x17}
				}),
			},
			expect: []byte{0x01, 0x09, 0x00, 0x01, 0x06, 0x28, 0x03, 0x00, 0x00, 0x12, 0x17},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncSliceBinary(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_DecSliceBinary(t *testing.T) {
	testCases := []struct {
		name         string
		input        []byte
		expectValue  []*marshaledBytesReceiver
		expectRead   int
		expectError  bool
		consumerFunc func([]byte) error
	}{
		{
			name:        "empty",
			input:       []byte{0x00},
			expectValue: []*marshaledBytesReceiver{},
			expectRead:  1,
		},
		{
			name:        "nil",
			input:       []byte{0x80},
			expectValue: nil,
			expectRead:  1,
		},
		{
			name:  "one item, and extra bytes",
			input: []byte{0x01, 0x07, 0x01, 0x05, 0x03, 0xaf, 0xff, 0xde, 0x83, 0x12, 0x33},
			expectValue: []*marshaledBytesReceiver{
				{[]byte{0x03, 0xaf, 0xff, 0xde, 0x83}},
			},
			expectRead: 9,
		},
		{
			name:  "two items",
			input: []byte{0x01, 0x09, 0x00, 0x01, 0x06, 0x28, 0x03, 0x00, 0x00, 0x12, 0x17},
			expectValue: []*marshaledBytesReceiver{
				{[]byte{}},
				{[]byte{0x28, 0x03, 0x00, 0x00, 0x12, 0x17}},
			},
			expectRead: 11,
		},
		{
			name:        "not enough bytes for list",
			input:       []byte{0x01, 0x05, 0x03, 0x6f, 0x6e, 0x6f},
			expectError: true,
		},
		{
			name:        "not enough bytes for list item",
			input:       []byte{0x01, 0x04, 0x04, 0x6f, 0x6e, 0xff},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualValue, actualRead, err := DecSliceBinary[*marshaledBytesReceiver](tc.input)
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
