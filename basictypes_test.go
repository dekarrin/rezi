package rezi

import (
	"encoding"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EncBool(t *testing.T) {
	testCases := []struct {
		name   string
		input  bool
		expect []byte
	}{
		{
			name:   "true",
			input:  true,
			expect: []byte{0x01},
		},
		{
			name:   "false",
			input:  false,
			expect: []byte{0x00},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncBool(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_DecBool(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectValue bool
		expectRead  int
		expectError bool
	}{
		{
			name:        "true from exact value",
			input:       []byte{0x01},
			expectValue: true,
			expectRead:  1,
		},
		{
			name:        "true from sequence",
			input:       []byte{0x01, 0x00},
			expectValue: true,
			expectRead:  1,
		},
		{
			name:        "false from exact value",
			input:       []byte{0x00},
			expectValue: false,
			expectRead:  1,
		},
		{
			name:        "false from sequence",
			input:       []byte{0x00, 0x01},
			expectValue: false,
			expectRead:  1,
		},
		{
			name:        "error from exact value - 0x02",
			input:       []byte{0x02},
			expectError: true,
		},
		{
			name:        "error from exact value - 0xff",
			input:       []byte{0xff},
			expectError: true,
		},
		{
			name:        "error from sequence",
			input:       []byte{0x25, 0xab, 0xcc},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualValue, actualRead, err := DecBool(tc.input)
			if tc.expectError {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expectValue, actualValue)
			assert.Equal(tc.expectRead, actualRead, "num read bytes does not match expected")
		})
	}
}

func Test_EncInt(t *testing.T) {
	testCases := []struct {
		name   string
		input  int
		expect []byte
	}{
		{
			name:   "zero",
			input:  0,
			expect: []byte{0x00},
		},
		{
			name:   "1",
			input:  1,
			expect: []byte{0x01, 0x01},
		},
		{
			name:   "256",
			input:  256,
			expect: []byte{0x02, 0x01, 0x00},
		},
		{
			name:   "328493",
			input:  328493,
			expect: []byte{0x03, 0x05, 0x03, 0x2d},
		},
		{
			name:   "413",
			input:  413,
			expect: []byte{0x02, 0x01, 0x9d},
		},
		{
			name:   "-413",
			input:  -413,
			expect: []byte{0x82, 0xfe, 0x63},
		},
		{
			name:   "-5,320,721,484,761,530,367",
			input:  -5320721484761530367,
			expect: []byte{0x88, 0xb6, 0x29, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncInt(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_DecInt(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectValue int
		expectRead  int
		expectError bool
	}{
		{
			name:        "0 from exact value",
			input:       []byte{0x00},
			expectValue: 0,
			expectRead:  1,
		},
		{
			name:        "1 from exact value",
			input:       []byte{0x01, 0x01},
			expectValue: 1,
			expectRead:  2,
		},
		{
			name:        "-1 from exact value",
			input:       []byte{0x80},
			expectValue: -1,
			expectRead:  1,
		},

		{
			name:        "413 from exact value",
			input:       []byte{0x02, 0x01, 0x9d},
			expectValue: 413,
			expectRead:  3,
		},
		{
			name:        "-413413413 from sequence",
			input:       []byte{0x84, 0xe7, 0x5b, 0xcf, 0xdb, 0x00},
			expectValue: -413413413,
			expectRead:  5,
		},
		{
			name:        "error too short",
			input:       []byte{0x03, 0x00, 0x01},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualValue, actualRead, err := DecInt(tc.input)
			if tc.expectError {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expectValue, actualValue)
			assert.Equal(tc.expectRead, actualRead, "num read bytes does not match expected")
		})
	}
}

func Test_EncString(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect []byte
	}{
		{
			name:   "empty",
			input:  "",
			expect: []byte{0x00},
		},
		{
			name:   "one char",
			input:  "1",
			expect: []byte{0x01, 0x01, 0x31},
		},
		{
			name:   "'Hello, 世界'",
			input:  "Hello, 世界",
			expect: []byte{0x01, 0x09, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c},
		},
		{
			name:   "'hi, world!'",
			input:  "hi, world!",
			expect: []byte{0x01, 0x0a, 0x68, 0x69, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncString(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_DecString(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectValue string
		expectRead  int
		expectError bool
	}{

		{
			name:        "empty",
			input:       []byte{0x00},
			expectValue: "",
			expectRead:  1,
		},
		{
			name:        "one char followed by ff field",
			input:       []byte{0x01, 0x01, 0x31, 0xff, 0xff},
			expectValue: "1",
			expectRead:  3,
		},
		{
			name:        "'Hello, 世界', followed by other bytes",
			input:       []byte{0x01, 0x09, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c, 0x01, 0x02, 0x03},
			expectValue: "Hello, 世界",
			expectRead:  15,
		},
		{
			name:        "'hi, world!'",
			input:       []byte{0x01, 0x0a, 0x68, 0x69, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
			expectValue: "hi, world!",
			expectRead:  12,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualValue, actualRead, err := DecString(tc.input)
			if tc.expectError {
				if !assert.Error(err) {
					return
				}
			} else if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expectValue, actualValue)
			assert.Equal(tc.expectRead, actualRead, "num read bytes does not match expected")
		})
	}
}

func Test_EncBinary(t *testing.T) {
	testCases := []struct {
		name   string
		input  encoding.BinaryMarshaler
		expect []byte
	}{
		{
			name: "nil bytes",
			input: valueThatMarshalsWith(func() []byte {
				return nil
			}),
			expect: []byte{0x00},
		},
		{
			name: "empty bytes",
			input: valueThatMarshalsWith(func() []byte {
				return []byte{}
			}),
			expect: []byte{0x00},
		},
		{
			name: "1 byte",
			input: valueThatMarshalsWith(func() []byte {
				return []byte{0xff}
			}),
			expect: []byte{0x01, 0x01, 0xff},
		},
		{
			name: "several bytes",
			input: valueThatMarshalsWith(func() []byte {
				return []byte{0xff, 0x0a, 0x0b, 0x0c, 0x0e}
			}),
			expect: []byte{0x01, 0x05, 0xff, 0x0a, 0x0b, 0x0c, 0x0e},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncBinary(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_DecBinary(t *testing.T) {
	var received []byte

	sendToReceived := func(b []byte) error {
		received = make([]byte, len(b))
		copy(received, b)
		return nil
	}

	testCases := []struct {
		name          string
		input         []byte
		expectReceive []byte
		expectRead    int
		expectError   bool
		consumerFunc  func([]byte) error
	}{
		{
			name:          "empty",
			input:         []byte{0x00},
			expectReceive: []byte{},
			expectRead:    1,
			consumerFunc:  sendToReceived,
		},
		{
			name:          "nil",
			input:         []byte{0x00},
			expectReceive: []byte{},
			expectRead:    1,
			consumerFunc:  sendToReceived,
		},
		{
			name:          "1 byte",
			input:         []byte{0x01, 0x01, 0xff},
			expectReceive: []byte{0xff},
			expectRead:    3,
			consumerFunc:  sendToReceived,
		},
		{
			name:          "several bytes, followed by unrelated",
			input:         []byte{0x01, 0x05, 0xff, 0x0a, 0x0b, 0x0c, 0x0e, 0xff},
			expectReceive: []byte{0xff, 0x0a, 0x0b, 0x0c, 0x0e},
			expectRead:    7,
			consumerFunc:  sendToReceived,
		},
		{
			name:  "several bytes, but it will error",
			input: []byte{0x01, 0x05, 0xff, 0x0a, 0x0b, 0x0c, 0x0e},
			consumerFunc: func(b []byte) error {
				return fmt.Errorf("error")
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			unmarshalTo := valueThatUnmarshalsWith(tc.consumerFunc)

			actualRead, err := DecBinary(tc.input, unmarshalTo)
			if tc.expectError {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expectReceive, received)
			assert.Equal(tc.expectRead, actualRead, "num read bytes does not match expected")
		})
	}
}
