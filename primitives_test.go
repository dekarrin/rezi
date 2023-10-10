package rezi

import (
	"encoding"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_encBool(t *testing.T) {
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

			actual := encBool(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_decBool(t *testing.T) {
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

			actualValue, actualRead, err := decBool(tc.input)
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

func Test_encInt(t *testing.T) {
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

			actual := encInt(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_decInt(t *testing.T) {
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
			name:        "skip extension bytes - 0",
			input:       []byte{0x40, 0xff, 0xbf},
			expectValue: 0,
			expectRead:  3,
		},
		{
			name:        "skip extension bytes - 8888",
			input:       []byte{0x42, 0xff, 0xbf, 0x22, 0xb8},
			expectValue: 8888,
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

			actualValue, actualRead, err := decInt[int](tc.input)
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

func Test_encFloat(t *testing.T) {
	// operation on an existing float64 var of 0 mult by -1.0 is only way we
	// could find to reliably get a signed negative zero! glub

	var negZero = float64(0.0)
	negZero *= -1.0

	testCases := []struct {
		name   string
		input  float64
		expect []byte
	}{
		{
			name:   "zero",
			input:  0.0,
			expect: []byte{0x00},
		},
		{
			name:   "signed negative zero",
			input:  negZero,
			expect: []byte{0x80},
		},
		{
			name:   "1",
			input:  1.0,
			expect: []byte{0x02, 0x3f, 0xf0},
			// no LSB tag bc nothing to compact
		},
		{
			name:   "-1",
			input:  -1.0,
			expect: []byte{0x82, 0x3f, 0xf0},
		},
		{
			name:   "pad from right",
			input:  256.01220703125,
			expect: []byte{0x04, 0xc0, 0x70, 0x00, 0x32},
		},
		{
			name:   "pad from left",
			input:  1.00000000000159161572810262442,
			expect: []byte{0x04, 0x3f, 0xf0, 0x1c, 0x00},
		},
		{
			name:   "no padding possible",
			input:  2.02499999999999991118215802999,
			expect: []byte{0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := encFloat(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_encString(t *testing.T) {
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
			expect: []byte{0x41, 0x82, 0x01, 0x31},
		},
		{
			name:   "'Hello, 世界'",
			input:  "Hello, 世界",
			expect: []byte{0x41, 0x82, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c},
		},
		{
			name:   "'hi, world!'",
			input:  "hi, world!",
			expect: []byte{0x41, 0x82, 0x0a, 0x68, 0x69, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := encString(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_decStringV0(t *testing.T) {
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

			actualValue, actualRead, err := decStringV0(tc.input)
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

func Test_decStringV1(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectValue string
		expectRead  int
		expectError bool
	}{

		{
			name:        "empty",
			input:       []byte{0x40, 0x80},
			expectValue: "",
			expectRead:  2,
		},
		{
			name:        "one byte char followed by ff field",
			input:       []byte{0x41, 0x80, 0x01, 0x31, 0xff, 0xff},
			expectValue: "1",
			expectRead:  4,
		},
		{
			name:        "'Hello, 世界', followed by other bytes",
			input:       []byte{0x41, 0x80, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c, 0x01, 0x02, 0x03},
			expectValue: "Hello, 世界",
			expectRead:  16,
		},
		{
			name:        "'Hello, 世界', followed by other bytes, with version",
			input:       []byte{0x41, 0x82, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c, 0x01, 0x02, 0x03},
			expectValue: "Hello, 世界",
			expectRead:  16,
		},
		{
			name:        "'hi, world!'",
			input:       []byte{0x41, 0x80, 0x0a, 0x68, 0x69, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
			expectValue: "hi, world!",
			expectRead:  13,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualValue, actualRead, err := decStringV1(tc.input)
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

func Test_encBinary(t *testing.T) {
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

			actual, _ := encBinary(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_decBinary(t *testing.T) {
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

			actualRead, err := decBinary(tc.input, unmarshalTo)
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

func Test_Enc_String(t *testing.T) {
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
			input:  "V",
			expect: []byte{0x41, 0x82, 0x01, 0x56},
		},
		{
			name:   "several chars",
			input:  "Vriska",
			expect: []byte{0x41, 0x82, 0x06, 0x56, 0x72, 0x69, 0x73, 0x6b, 0x61},
		},
		{
			name:   "pre-composed char seq",
			input:  "homeçtuck",
			expect: []byte{0x41, 0x82, 0x0a, 0x68, 0x6f, 0x6d, 0x65, 0xc3, 0xa7, 0x74, 0x75, 0x63, 0x6b},
		},
		{
			name:   "decomposed char seq",
			input:  "homec\u0327tuck",
			expect: []byte{0x41, 0x82, 0x0b, 0x68, 0x6f, 0x6d, 0x65, 0x63, 0xcc, 0xa7, 0x74, 0x75, 0x63, 0x6b},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			actual, err := Enc(tc.input)
			if !assert.NoError(err) {
				return
			}
			assert.Equal(tc.expect, actual)
		})
	}

	t.Run("*string (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *string
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*string", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = "Vriska"
			input    = &inputVal
			expect   = []byte{0x41, 0x82, 0x06, 0x56, 0x72, 0x69, 0x73, 0x6b, 0x61}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**string", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = "Vriska"
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{0x41, 0x82, 0x06, 0x56, 0x72, 0x69, 0x73, 0x6b, 0x61}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**string, but nil string part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *string
			input  = &ptr
			expect = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

}

func Test_Enc_Int(t *testing.T) {

	testCases := []struct {
		name   string
		input  interface{}
		expect []byte
	}{
		{name: "int zero", input: 0, expect: []byte{0x00}},
		{name: "int large pos mag", input: 5320721484761530367, expect: []byte{0x08, 0x49, 0xd6, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{name: "int large neg mag", input: -5320721484761530367, expect: []byte{0x88, 0xb6, 0x29, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{name: "int 1", input: 1, expect: []byte{0x01, 0x01}},
		{name: "int 256", input: 256, expect: []byte{0x02, 0x01, 0x00}},
		{name: "int -1", input: -1, expect: []byte{0x80}},
		{name: "int -413", input: -413, expect: []byte{0x82, 0xfe, 0x63}},

		{name: "int8 zero", input: int8(0), expect: []byte{0x00}},
		{name: "int8 large pos mag", input: int8(122), expect: []byte{0x01, 0x7a}},
		{name: "int8 large neg mag", input: int8(-124), expect: []byte{0x81, 0x84}},
		{name: "int8 1", input: int8(1), expect: []byte{0x01, 0x01}},
		{name: "int8 -1", input: int8(-1), expect: []byte{0x80}},

		{name: "int16 zero", input: int16(0), expect: []byte{0x00}},
		{name: "int16 large pos mag", input: int16(32760), expect: []byte{0x02, 0x7f, 0xf8}},
		{name: "int16 large neg mag", input: int16(-32000), expect: []byte{0x82, 0x83, 0x00}},
		{name: "int16 1", input: int16(1), expect: []byte{0x01, 0x01}},
		{name: "int16 -1", input: int16(-1), expect: []byte{0x80}},

		{name: "int32 zero", input: int32(0), expect: []byte{0x00}},
		{name: "int32 large pos mag", input: int32(2147400413), expect: []byte{0x04, 0x7f, 0xfe, 0xba, 0xdd}},
		{name: "int32 large neg mag", input: int32(-2147400413), expect: []byte{0x84, 0x80, 0x01, 0x45, 0x23}},
		{name: "int32 1", input: int32(1), expect: []byte{0x01, 0x01}},
		{name: "int32 -1", input: int32(-1), expect: []byte{0x80}},

		{name: "int64 zero", input: int64(0), expect: []byte{0x00}},
		{name: "int64 large pos mag", input: int64(8888413612000000000), expect: []byte{0x08, 0x7b, 0x59, 0xfd, 0x16, 0x58, 0x01, 0xb8, 0x00}},
		{name: "int64 large neg mag", input: int64(-8888413612000000000), expect: []byte{0x88, 0x84, 0xa6, 0x02, 0xe9, 0xa7, 0xfe, 0x48, 0x00}},
		{name: "int64 1", input: int64(1), expect: []byte{0x01, 0x01}},
		{name: "int64 -1", input: int64(-1), expect: []byte{0x80}},

		{name: "uint zero", input: uint(0), expect: []byte{0x00}},
		{name: "uint 1", input: uint(1), expect: []byte{0x01, 0x01}},
		{name: "uint large", input: uint(888888880), expect: []byte{0x04, 0x34, 0xfb, 0x5e, 0x30}},

		{name: "uint8 zero", input: uint8(0), expect: []byte{0x00}},
		{name: "uint8 1", input: uint8(1), expect: []byte{0x01, 0x01}},
		{name: "uint8 large", input: uint8(255), expect: []byte{0x01, 0xff}},

		{name: "uint16 zero", input: uint16(0), expect: []byte{0x00}},
		{name: "uint16 1", input: uint16(1), expect: []byte{0x01, 0x01}},
		{name: "uint16 large", input: uint16(58888), expect: []byte{0x02, 0xe6, 0x08}},

		{name: "uint32 zero", input: uint32(0), expect: []byte{0x00}},
		{name: "uint32 1", input: uint32(1), expect: []byte{0x01, 0x01}},
		{name: "uint32 large", input: uint32(4188888888), expect: []byte{0x04, 0xf9, 0xad, 0x5f, 0x38}},

		{name: "uint64 zero", input: uint64(0), expect: []byte{0x00}},
		{name: "uint64 1", input: uint64(1), expect: []byte{0x01, 0x01}},
		{name: "uint64 large", input: uint64(10004138888888800612), expect: []byte{0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x64}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual, err := Enc(tc.input)
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, actual)
		})
	}

	t.Run("*int (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *int
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = 8
			input    = &inputVal
			expect   = []byte{0x01, 0x08}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = 8
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{0x01, 0x08}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**int, but nil int part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *int
			input  = &ptr
			expect = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*uint (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *uint
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*uint", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = uint(8)
			input    = &inputVal
			expect   = []byte{0x01, 0x08}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**uint", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = uint(8)
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{0x01, 0x08}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**uint, but nil uint part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *uint
			input  = &ptr
			expect = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

}

func Test_Enc_Bool(t *testing.T) {
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

			actual, err := Enc(tc.input)
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, actual)
		})
	}

	t.Run("*bool (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *bool
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*bool", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = true
			input    = &inputVal
			expect   = []byte{0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**bool", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = false
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{0x00}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**bool, but nil bool part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *bool
			input  = &ptr
			expect = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

}

func Test_Enc_Binary(t *testing.T) {
	testCases := []struct {
		name   string
		input  encoding.BinaryMarshaler
		expect []byte
	}{
		{
			name:   "encode to nil bytes",
			input:  valueThatMarshalsWith(func() []byte { return nil }),
			expect: []byte{0x00},
		},
		{
			name:   "encode to empty bytes",
			input:  valueThatMarshalsWith(func() []byte { return []byte{} }),
			expect: []byte{0x00},
		},
		{
			name:   "encode to one byte",
			input:  valueThatMarshalsWith(func() []byte { return []byte{0x03} }),
			expect: []byte{0x01, 0x01, 0x03},
		},
		{
			name:   "encode to several bytes",
			input:  valueThatMarshalsWith(func() []byte { return []byte{0x03, 0x44, 0x15} }),
			expect: []byte{0x01, 0x03, 0x03, 0x44, 0x15},
		},
		{
			name:   "actual object",
			input:  testBinary{number: 12, data: "Hello, John!!!!!!!!"},
			expect: []byte{0x01, 0x18, 0x41, 0x82, 0x13, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x4a, 0x6f, 0x68, 0x6e, 0x21, 0x21, 0x21, 0x21, 0x21, 0x21, 0x21, 0x21, 0x01, 0x0c},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual, err := Enc(tc.input)
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, actual)
		})
	}

	t.Run("*binary (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testBinary
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*binary", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = testBinary{number: 8, data: "VRISKA"}
			input    = &inputVal
			expect   = []byte{
				0x01, 0x0b, // len=11

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**binary", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = testBinary{number: 8, data: "VRISKA"}
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{
				0x01, 0x0b, // len=11

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**binary, but nil binary part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *testBinary
			input  = &ptr
			expect = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})
}

func Test_Dec_String(t *testing.T) {
	type result struct {
		val      string
		consumed int
		err      bool
	}

	testCases := []struct {
		name   string
		input  []byte
		expect result
	}{
		{
			name:   "empty (v1 and v2)",
			input:  []byte{0x00},
			expect: result{val: "", consumed: 1},
		},
		{
			name:   "one char (v1)",
			input:  []byte{0x01, 0x01, 0x56},
			expect: result{val: "V", consumed: 3},
		},
		{
			name:   "several chars (v1)",
			input:  []byte{0x01, 0x06, 0x56, 0x72, 0x69, 0x73, 0x6b, 0x61},
			expect: result{val: "Vriska", consumed: 8},
		},
		{
			name:   "pre-composed char seq (v1)",
			input:  []byte{0x01, 0x09, 0x68, 0x6f, 0x6d, 0x65, 0xc3, 0xa7, 0x74, 0x75, 0x63, 0x6b},
			expect: result{val: "homeçtuck", consumed: 12},
		},
		{
			name:   "decomposed char seq (v1)",
			input:  []byte{0x01, 0x0a, 0x68, 0x6f, 0x6d, 0x65, 0x63, 0xcc, 0xa7, 0x74, 0x75, 0x63, 0x6b},
			expect: result{val: "homec\u0327tuck", consumed: 13},
		},
		{
			name:   "err count too big (v1)",
			input:  []byte{0x01, 0x08, 0x68, 0x6f},
			expect: result{err: true},
		},
		{
			name:   "err invalid sequence (v1)",
			input:  []byte{0x01, 0x01, 0xc3, 0x28},
			expect: result{err: true},
		},
		{
			name:   "one char (v2)",
			input:  []byte{0x41, 0x80, 0x01, 0x56},
			expect: result{val: "V", consumed: 4},
		},
		{
			name:   "several chars (v2)",
			input:  []byte{0x41, 0x80, 0x06, 0x56, 0x72, 0x69, 0x73, 0x6b, 0x61},
			expect: result{val: "Vriska", consumed: 9},
		},
		{
			name:   "pre-composed char seq (v2)",
			input:  []byte{0x41, 0x80, 0x0a, 0x68, 0x6f, 0x6d, 0x65, 0xc3, 0xa7, 0x74, 0x75, 0x63, 0x6b},
			expect: result{val: "homeçtuck", consumed: 13},
		},
		{
			name:   "decomposed char seq (v2)",
			input:  []byte{0x41, 0x80, 0x0b, 0x68, 0x6f, 0x6d, 0x65, 0x63, 0xcc, 0xa7, 0x74, 0x75, 0x63, 0x6b},
			expect: result{val: "homec\u0327tuck", consumed: 14},
		},
		{
			name:   "err count too big (v2)",
			input:  []byte{0x41, 0x80, 0x08, 0x68, 0x6f},
			expect: result{err: true},
		},
		{
			name:   "err invalid sequence (v2)",
			input:  []byte{0x41, 0x80, 0x02, 0xc3, 0x28},
			expect: result{err: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			var actual result
			var err error

			actual.consumed, err = Dec(tc.input, &actual.val)
			if tc.expect.err {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, actual)
		})
	}

	t.Run("nil *string", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xa0}
			expect         *string
			expectConsumed = 1
		)

		var actual *string = ref("somefin")
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*string", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x01, 0x06, 0x56, 0x72, 0x69, 0x73, 0x6b, 0x61}
			expectVal      = "Vriska"
			expect         = &expectVal
			expectConsumed = 8
		)

		var actual *string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**string", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x01, 0x06, 0x56, 0x72, 0x69, 0x73, 0x6b, 0x61}
			expectVal      = "Vriska"
			expectValPtr   = &expectVal
			expect         = &expectValPtr
			expectConsumed = 8
		)

		var actual **string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**string, but nil string part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xb0, 0x01, 0x01}
			expectConsumed = 3
		)

		var actual **string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)

		assert.NotNil(actual) // actual should *itself* not be nil
		assert.Nil(*actual)   // but the pointer it points to should be nil
	})
}

func Test_Dec_Int(t *testing.T) {
	type result struct {
		val      interface{}
		consumed int
		err      bool
	}

	testCases := []struct {
		name   string
		input  []byte
		expect result
	}{
		{name: "int zero", input: []byte{0x00}, expect: result{val: int(0), consumed: 1}},
		{name: "int large pos mag", input: []byte{0x08, 0x49, 0xd6, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, expect: result{val: int(5320721484761530367), consumed: 9}},
		{name: "int large neg mag", input: []byte{0x88, 0xb6, 0x29, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, expect: result{val: int(-5320721484761530367), consumed: 9}},
		{name: "int 1", input: []byte{0x01, 0x01}, expect: result{val: int(1), consumed: 2}},
		{name: "int 256", input: []byte{0x02, 0x01, 0x00}, expect: result{val: int(256), consumed: 3}},
		{name: "int -1", input: []byte{0x80}, expect: result{val: int(-1), consumed: 1}},
		{name: "int -413", input: []byte{0x82, 0xfe, 0x63}, expect: result{val: int(-413), consumed: 3}},

		{name: "int8 zero", input: []byte{0x00}, expect: result{val: int8(0), consumed: 1}},
		{name: "int8 large pos mag", input: []byte{0x01, 0x7a}, expect: result{val: int8(122), consumed: 2}},
		{name: "int8 large neg mag", input: []byte{0x81, 0x84}, expect: result{val: int8(-124), consumed: 2}},
		{name: "int8 1", input: []byte{0x01, 0x01}, expect: result{val: int8(1), consumed: 2}},
		{name: "int8 -1", input: []byte{0x80}, expect: result{val: int8(-1), consumed: 1}},

		{name: "int16 zero", input: []byte{0x00}, expect: result{val: int16(0), consumed: 1}},
		{name: "int16 large pos mag", input: []byte{0x02, 0x7f, 0xf8}, expect: result{val: int16(32760), consumed: 3}},
		{name: "int16 large neg mag", input: []byte{0x82, 0x83, 0x00}, expect: result{val: int16(-32000), consumed: 3}},
		{name: "int16 1", input: []byte{0x01, 0x01}, expect: result{val: int16(1), consumed: 2}},
		{name: "int16 -1", input: []byte{0x80}, expect: result{val: int16(-1), consumed: 1}},

		{name: "int32 zero", input: []byte{0x00}, expect: result{val: int32(0), consumed: 1}},
		{name: "int32 large pos mag", input: []byte{0x04, 0x7f, 0xfe, 0xba, 0xdd}, expect: result{val: int32(2147400413), consumed: 5}},
		{name: "int32 large neg mag", input: []byte{0x84, 0x80, 0x01, 0x45, 0x23}, expect: result{val: int32(-2147400413), consumed: 5}},
		{name: "int32 1", input: []byte{0x01, 0x01}, expect: result{val: int32(1), consumed: 2}},
		{name: "int32 -1", input: []byte{0x80}, expect: result{val: int32(-1), consumed: 1}},

		{name: "int64 zero", input: []byte{0x00}, expect: result{val: int64(0), consumed: 1}},
		{name: "int64 large pos mag", input: []byte{0x08, 0x7b, 0x59, 0xfd, 0x16, 0x58, 0x01, 0xb8, 0x00}, expect: result{val: int64(8888413612000000000), consumed: 9}},
		{name: "int64 large neg mag", input: []byte{0x88, 0x84, 0xa6, 0x02, 0xe9, 0xa7, 0xfe, 0x48, 0x00}, expect: result{val: int64(-8888413612000000000), consumed: 9}},
		{name: "int64 1", input: []byte{0x01, 0x01}, expect: result{val: int64(1), consumed: 2}},
		{name: "int64 -1", input: []byte{0x80}, expect: result{val: int64(-1), consumed: 1}},

		{name: "uint zero", input: []byte{0x00}, expect: result{val: uint(0), consumed: 1}},
		{name: "uint 1", input: []byte{0x01, 0x01}, expect: result{val: uint(1), consumed: 2}},
		{name: "uint large", input: []byte{0x04, 0x34, 0xfb, 0x5e, 0x30}, expect: result{val: uint(888888880), consumed: 5}},

		{name: "uint8 zero", input: []byte{0x00}, expect: result{val: uint8(0), consumed: 1}},
		{name: "uint8 1", input: []byte{0x01, 0x01}, expect: result{val: uint8(1), consumed: 2}},
		{name: "uint8 large", input: []byte{0x01, 0xff}, expect: result{val: uint8(255), consumed: 2}},

		{name: "uint16 zero", input: []byte{0x00}, expect: result{val: uint16(0), consumed: 1}},
		{name: "uint16 1", input: []byte{0x01, 0x01}, expect: result{val: uint16(1), consumed: 2}},
		{name: "uint16 large", input: []byte{0x02, 0xe6, 0x08}, expect: result{val: uint16(58888), consumed: 3}},

		{name: "uint32 zero", input: []byte{0x00}, expect: result{val: uint32(0), consumed: 1}},
		{name: "uint32 1", input: []byte{0x01, 0x01}, expect: result{val: uint32(1), consumed: 2}},
		{name: "uint32 large", input: []byte{0x04, 0xf9, 0xad, 0x5f, 0x38}, expect: result{val: uint32(4188888888), consumed: 5}},

		{name: "uint64 zero", input: []byte{0x00}, expect: result{val: uint64(0), consumed: 1}},
		{name: "uint64 1", input: []byte{0x01, 0x01}, expect: result{val: uint64(1), consumed: 2}},
		{name: "uint64 large", input: []byte{0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x64}, expect: result{val: uint64(10004138888888800612), consumed: 9}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			var actual result
			var err error

			// we are about to have type examination on our actual pointer
			// be run, so we'd betta pass it the correct kind of ptr
			switch tc.expect.val.(type) {
			case int:
				var v int
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case int8:
				var v int8
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case int16:
				var v int16
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case int32:
				var v int32
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case int64:
				var v int64
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case uint:
				var v uint
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case uint8:
				var v uint8
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case uint16:
				var v uint16
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case uint32:
				var v uint32
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case uint64:
				var v uint64
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			default:
				panic("bad test case")
			}

			if tc.expect.err {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, actual)
		})
	}

	t.Run("nil *int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xa0}
			expect         *int
			expectConsumed = 1
		)

		var actual *int = ref(12)
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x01, 0x08}
			expectVal      = 8
			expect         = &expectVal
			expectConsumed = 2
		)

		var actual *int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x01, 0x08}
			expectVal      = 8
			expectValPtr   = &expectVal
			expect         = &expectValPtr
			expectConsumed = 2
		)

		var actual **int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**int, but nil int part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xb0, 0x01, 0x01}
			expectConsumed = 3
		)

		var actual **int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)

		assert.NotNil(actual) // actual should *itself* not be nil
		assert.Nil(*actual)   // but the pointer it points to should be nil
	})

	t.Run("nil *uint", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xa0}
			expect         *uint
			expectConsumed = 1
		)

		var actual *uint = ref(uint(12))
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*uint", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x01, 0x08}
			expectVal      = uint(8)
			expect         = &expectVal
			expectConsumed = 2
		)

		var actual *uint
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**uint", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x01, 0x08}
			expectVal      = uint(8)
			expectValPtr   = &expectVal
			expect         = &expectValPtr
			expectConsumed = 2
		)

		var actual **uint
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**uint, but nil uint part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xb0, 0x01, 0x01}
			expectConsumed = 3
		)

		var actual **uint
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)

		assert.NotNil(actual) // actual should *itself* not be nil
		assert.Nil(*actual)   // but the pointer it points to should be nil
	})
}

func Test_Dec_Bool(t *testing.T) {
	type result struct {
		val      bool
		consumed int
		err      bool
	}

	testCases := []struct {
		name   string
		input  []byte
		expect result
	}{
		{
			name:   "true",
			input:  []byte{0x01},
			expect: result{val: true, consumed: 1},
		},
		{
			name:   "false",
			input:  []byte{0x00},
			expect: result{val: false, consumed: 1},
		},
		{
			name:   "err not enough bytes",
			input:  []byte{},
			expect: result{err: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			var actual result
			var err error

			actual.consumed, err = Dec(tc.input, &actual.val)
			if tc.expect.err {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, actual)
		})
	}

	t.Run("nil *bool (no sign bit)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x20}
			expect         *bool
			expectConsumed = 1
		)

		var actual *bool = ref(true)
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("nil *bool", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xa0}
			expect         *bool
			expectConsumed = 1
		)

		var actual *bool = ref(true)
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*bool", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x01}
			expectVal      = true
			expect         = &expectVal
			expectConsumed = 1
		)

		var actual *bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**bool", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x00}
			expectVal      = false
			expectValPtr   = &expectVal
			expect         = &expectValPtr
			expectConsumed = 1
		)

		var actual **bool

		// give it original value "true" so we can check it was set to false.
		actualOrig := true
		actualPtr := &actualOrig
		actual = &actualPtr

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**bool, but nil bool part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xb0, 0x01, 0x01}
			expectConsumed = 3
		)

		var actual **bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)

		assert.NotNil(actual) // actual should *itself* not be nil
		assert.Nil(*actual)   // but the pointer it points to should be nil
	})

}

func Test_Dec_Binary(t *testing.T) {
	// cannot be table-driven easily due to trying different types

	t.Run("normal BinaryUnmarshaler result", func(t *testing.T) {
		// setup
		assert := assert.New(t)

		input := []byte{
			0x01, 0x17, 0x01, 0x13, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20,
			0x4a, 0x6f, 0x68, 0x6e, 0x21, 0x21, 0x21, 0x21, 0x21, 0x21, 0x21,
			0x21, 0x01, 0x0c,
		}
		expect := testBinary{number: 12, data: "Hello, John!!!!!!!!"}
		expectConsumed := 25

		// exeucte
		actual := testBinary{}
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("nil *BinaryUnmarshaler", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xa0}
			expect         *testBinary
			expectConsumed = 1
		)

		var actual *testBinary = &testBinary{data: "Test", number: 1}
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*BinaryUnmarshaler", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0x01, 0x0a, // len=10

				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
			expectVal      = testBinary{number: 8, data: "VRISKA"}
			expect         = &expectVal
			expectConsumed = 12
		)

		var actual *testBinary
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**BinaryUnmarshaler", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0x01, 0x0a, // len=10

				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
			expectVal      = testBinary{number: 8, data: "VRISKA"}
			expectValPtr   = &expectVal
			expect         = &expectValPtr
			expectConsumed = 12
		)

		var actual **testBinary
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**BinaryUnmarshaler, but nil BinaryUnmarshaler part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xb0, 0x01, 0x01}
			expectConsumed = 3
		)

		var actual **testBinary
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)

		assert.NotNil(actual) // actual should *itself* not be nil
		assert.Nil(*actual)   // but the pointer it points to should be nil
	})

}
