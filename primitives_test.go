package rezi

import (
	"encoding"
	"fmt"
	"math"
	"math/big"
	"net"
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

func Test_decFloat(t *testing.T) {
	// operation on an existing float64 var of 0 mult by -1.0 is only way we
	// could find to reliably get a signed negative zero! glub

	var negZero = float64(0.0)
	negZero *= -1.0

	testCases := []struct {
		name       string
		input      []byte
		expect     float64
		expectRead int
		expectErr  bool
	}{
		{
			name:       "zero - exact value",
			input:      []byte{0x00},
			expect:     0.0,
			expectRead: 1,
		},
		{
			name:       "signed negative zero - exact value",
			input:      []byte{0x80},
			expect:     negZero,
			expectRead: 1,
		},
		{
			name:       "1 - exact value",
			input:      []byte{0x02, 0x3f, 0xf0},
			expect:     1.0,
			expectRead: 3,
		},
		{
			name:       "-1 - exact value",
			input:      []byte{0x82, 0x3f, 0xf0},
			expect:     -1.0,
			expectRead: 3,
		},
		{
			name:       "pad from right - exact value",
			input:      []byte{0x04, 0xc0, 0x70, 0x00, 0x32},
			expect:     256.01220703125,
			expectRead: 5,
		},
		{
			name:       "pad from right - sequence",
			input:      []byte{0x04, 0xc0, 0x70, 0x00, 0x32, 0x04, 0xc0, 0x70, 0x00, 0x32},
			expect:     256.01220703125,
			expectRead: 5,
		},
		{
			name:       "pad from left - exact value",
			input:      []byte{0x04, 0x3f, 0xf0, 0x1c, 0x00},
			expect:     1.00000000000159161572810262442,
			expectRead: 5,
		},
		{
			name:       "pad from left - sequence",
			input:      []byte{0x04, 0x3f, 0xf0, 0x1c, 0x00, 0x04, 0x3f, 0xf0, 0x1c, 0x00},
			expect:     1.00000000000159161572810262442,
			expectRead: 5,
		},
		{
			name:       "no padding possible - exact value",
			input:      []byte{0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33},
			expect:     2.02499999999999991118215802999,
			expectRead: 9,
		},
		{
			name:       "no padding possible - sequence",
			input:      []byte{0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33},
			expect:     2.02499999999999991118215802999,
			expectRead: 9,
		},
		{
			name:       "skip extension bytes - 0",
			input:      []byte{0x40, 0xff, 0xbf},
			expect:     0.0,
			expectRead: 3,
		},
		{
			name:       "skip extension bytes - negative 0",
			input:      []byte{0xc0, 0xff, 0xbf},
			expect:     negZero,
			expectRead: 3,
		},
		{
			name:       "skip extension bytes - 875.0",
			input:      []byte{0x43, 0xff, 0xbf, 0xc0, 0x8b, 0x58},
			expect:     875.0,
			expectRead: 6,
		},
		{
			name:      "error too short",
			input:     []byte{0x03, 0x00, 0x01},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			expectBits := math.Float64bits(tc.expect)

			actual, actualRead, err := decFloat[float64](tc.input)
			if tc.expectErr {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}
			actualBits := math.Float64bits(actual)

			assert.Equal(tc.expect, actual, "float values differ")
			assert.Equal(expectBits, actualBits, "bit values differ")
			assert.Equal(tc.expectRead, actualRead, "num read bytes does not match expected")
		})
	}
}

func Test_encComplex(t *testing.T) {
	// operation on an existing float64 var of 0 mult by -1.0 is only way we
	// could find to reliably get a signed negative zero! glub

	var negZero = float64(0.0)
	negZero *= -1.0

	testCases := []struct {
		name   string
		input  complex128
		expect []byte
	}{
		{
			name:   "zeros",
			input:  complex(0.0, 0.0),
			expect: []byte{0x00},
		},
		{
			name:   "signed negative zeros",
			input:  complex(negZero, negZero),
			expect: []byte{0x80},
		},
		{
			name:  "mixed signed zeros",
			input: complex(negZero, 0.0),
			expect: []byte{
				0x41, 0x80, 0x02, // (explicit byte count) len=2

				0x80, // (-0.0)
				0x00, // 0.0i
			},
		},
		{
			name:  "mixed signed zeros (flipped)",
			input: complex(0.0, negZero),
			expect: []byte{
				0x41, 0x80, 0x02, // (explicit byte count) len=2

				0x00, // 0.0
				0x80, // (-0.0)i
			},
		},
		{
			name:  "1.0",
			input: 1.0,
			expect: []byte{
				0x41, 0x80, 0x04, // (explicit byte count) len=4

				0x02, 0x3f, 0xf0, // 1.0
				0x00, // 0.0i
			},
		},
		{
			name:  "1.0i",
			input: 1.0i,
			expect: []byte{
				0x41, 0x80, 0x04, // (explicit byte count) len=4

				0x00,             // 0.0
				0x02, 0x3f, 0xf0, // 1.0i
			},
		},
		{
			name:  "valued r and i parts",
			input: -1.0 + 8.25i,
			expect: []byte{
				0x41, 0x80, 0x07, // (explicit byte count) len=7

				0x82, 0x3f, 0xf0, // -1.0
				0x03, 0xc0, 0x20, 0x80, // 8.25i
			},
		},
		{
			name:  "largest possible",
			input: 2.02499999999999991118215802999 + 2.02499999999999991118215802999i,
			expect: []byte{
				0x41, 0x80, 0x12, // (explicit byte count) len=18

				0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33,
				0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := encComplex(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_decComplex(t *testing.T) {
	// operation on an existing float64 var of 0 mult by -1.0 is only way we
	// could find to reliably get a signed negative zero! glub

	var negZero = float64(0.0)
	negZero *= -1.0

	testCases := []struct {
		name       string
		input      []byte
		expect     complex128
		expectRead int
		expectErr  bool
	}{
		{
			name:       "zeros",
			input:      []byte{0x00},
			expect:     complex(0.0, 0.0),
			expectRead: 1,
		},
		{
			name:       "signed negative zeros",
			input:      []byte{0x80},
			expect:     complex(negZero, negZero),
			expectRead: 1,
		},
		{
			name: "mixed signed zeros",
			input: []byte{
				0x41, 0x80, 0x02, // (explicit byte count) len=2

				0x80, // (-0.0)
				0x00, // 0.0i
			},
			expect:     complex(negZero, 0.0),
			expectRead: 5,
		},
		{
			name: "mixed signed zeros (flipped)",
			input: []byte{
				0x41, 0x80, 0x02, // (explicit byte count) len=2

				0x00, // 0.0
				0x80, // (-0.0)i
			},
			expect:     complex(0.0, negZero),
			expectRead: 5,
		},
		{
			name: "1.0",
			input: []byte{
				0x41, 0x80, 0x04, // (explicit byte count) len=4

				0x02, 0x3f, 0xf0, // 1.0
				0x00, // 0.0i
			},
			expect:     1.0,
			expectRead: 7,
		},
		{
			name: "1.0i",
			input: []byte{
				0x41, 0x80, 0x04, // (explicit byte count) len=4

				0x00,             // 0.0
				0x02, 0x3f, 0xf0, // 1.0i
			},
			expect:     1.0i,
			expectRead: 7,
		},
		{
			name: "valued r and i parts",
			input: []byte{
				0x41, 0x80, 0x07, // (explicit byte count) len=7

				0x82, 0x3f, 0xf0, // -1.0
				0x03, 0xc0, 0x20, 0x80, // 8.25i
			},
			expect:     -1.0 + 8.25i,
			expectRead: 10,
		},
		{
			name: "largest possible",
			input: []byte{
				0x41, 0x80, 0x12, // (explicit byte count) len=18

				0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33,
				0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33,
			},
			expect:     2.02499999999999991118215802999 + 2.02499999999999991118215802999i,
			expectRead: 21,
		},
		{
			name:      "error too short",
			input:     []byte{0x03, 0x00, 0x01},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			expectRBits := math.Float64bits(real(tc.expect))
			expectIBits := math.Float64bits(imag(tc.expect))

			actual, actualRead, err := decComplex[complex128](tc.input)
			if tc.expectErr {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}
			actualRBits := math.Float64bits(real(actual))
			actualIBits := math.Float64bits(imag(actual))

			assert.Equal(tc.expect, actual, "complex values differ")
			assert.Equal(expectRBits, actualRBits, "real-part bit values differ")
			assert.Equal(expectIBits, actualIBits, "real-part bit values differ")
			assert.Equal(tc.expectRead, actualRead, "num read bytes does not match expected")
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

func Test_encText(t *testing.T) {
	testCases := []struct {
		name   string
		input  encoding.TextMarshaler
		expect []byte
	}{
		{
			name:   "user-defined type",
			input:  testText{name: "VRISKA", value: 8, enabled: true},
			expect: []byte{0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41},
		},
		{
			name:   "IPv4",
			input:  net.ParseIP("128.0.0.1"),
			expect: []byte{0x41, 0x82, 0x09, 0x31, 0x32, 0x38, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31},
		},
		{
			name:   "IPv6",
			input:  net.ParseIP("2001:db8::1"),
			expect: []byte{0x41, 0x82, 0x0b, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x64, 0x62, 0x38, 0x3a, 0x3a, 0x31},
		},
		{
			name:   "big.Int",
			input:  big.NewInt(2023),
			expect: []byte{0x41, 0x82, 0x04, 0x32, 0x30, 0x32, 0x33},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual, _ := encText(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_decText(t *testing.T) {
	t.Run("user-defined type", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41}
			expect         = testText{name: "VRISKA", value: 8, enabled: true}
			expectConsumed = 16
		)

		var actual testText
		actualRead, err := decText(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
		assert.Equal(expectConsumed, actualRead, "num read bytes does not match expected")
	})

	t.Run("IPv4", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x09, 0x31, 0x32, 0x38, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31}
			expect         = net.ParseIP("128.0.0.1")
			expectConsumed = 12
		)

		var actual net.IP
		actualRead, err := decText(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
		assert.Equal(expectConsumed, actualRead, "num read bytes does not match expected")
	})

	t.Run("IPv6", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x0b, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x64, 0x62, 0x38, 0x3a, 0x3a, 0x31}
			expect         = net.ParseIP("2001:db8::1")
			expectConsumed = 14
		)

		var actual net.IP
		actualRead, err := decText(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
		assert.Equal(expectConsumed, actualRead, "num read bytes does not match expected")
	})

	t.Run("big.Int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x04, 0x32, 0x30, 0x32, 0x33}
			expect         = big.NewInt(2023)
			expectConsumed = 7
		)

		actual := big.NewInt(1)
		actualRead, err := decText(input, actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
		assert.Equal(expectConsumed, actualRead, "num read bytes does not match expected")
	})
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

func Test_Enc_Float(t *testing.T) {
	var negZero64 = float64(0.0)
	negZero64 *= -1.0
	var negZero32 = float32(0.0)
	negZero32 *= -1.0

	testCases := []struct {
		name   string
		input  interface{}
		expect []byte
	}{
		{name: "float64 0.0", input: float64(0.0), expect: []byte{0x00}},
		{name: "float64 -0.0", input: negZero64, expect: []byte{0x80}},
		{name: "float64 wide pos", input: float64(2.02499999999999991118215802999), expect: []byte{0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}},
		{name: "float64 wide neg", input: float64(-2.02499999999999991118215802999), expect: []byte{0x88, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}},
		{name: "float64 1.0", input: float64(1.0), expect: []byte{0x02, 0x3f, 0xf0}},
		{name: "float64 256ish", input: float64(256.01220703125), expect: []byte{0x04, 0xc0, 0x70, 0x00, 0x32}},
		{name: "float64 -1.0", input: float64(-1.0), expect: []byte{0x82, 0x3f, 0xf0}},
		{name: "float64 -413.0", input: float64(-413.0), expect: []byte{0x83, 0xc0, 0x79, 0xd0}},
		{name: "float64 +inf", input: float64(math.Inf(0)), expect: []byte{0x02, 0x7f, 0xf0}},
		{name: "float64 -inf", input: float64(math.Inf(-1)), expect: []byte{0x82, 0x7f, 0xf0}},

		{name: "float32 0.0", input: float32(0.0), expect: []byte{0x00}},
		{name: "float32 -0.0", input: negZero32, expect: []byte{0x80}},
		{name: "float32 wide pos", input: float32(8.38218975067138671875), expect: []byte{0x05, 0xc0, 0x20, 0xc3, 0xae, 0x60}},
		{name: "float32 wide neg", input: float32(-8.38218975067138671875), expect: []byte{0x85, 0xc0, 0x20, 0xc3, 0xae, 0x60}},
		{name: "float32 1.0", input: float32(1.0), expect: []byte{0x02, 0x3f, 0xf0}},
		{name: "float32 256ish", input: float32(256.01220703125), expect: []byte{0x04, 0xc0, 0x70, 0x00, 0x32}},
		{name: "float32 -1.0", input: float32(-1.0), expect: []byte{0x82, 0x3f, 0xf0}},
		{name: "float32 -413.0", input: float32(-413.0), expect: []byte{0x83, 0xc0, 0x79, 0xd0}},
		{name: "float32 +inf", input: float32(math.Inf(0)), expect: []byte{0x02, 0x7f, 0xf0}},
		{name: "float32 -inf", input: float32(math.Inf(-1)), expect: []byte{0x82, 0x7f, 0xf0}},
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

	t.Run("*float32 (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *float32
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*float32", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = float32(8.0)
			input    = &inputVal
			expect   = []byte{0x02, 0x40, 0x20}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**float32", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = float32(8)
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{0x02, 0x40, 0x20}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**float32, but nil float32 part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *float32
			input  = &ptr
			expect = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*float64 (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *float64
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*float64", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = float64(8.0)
			input    = &inputVal
			expect   = []byte{0x02, 0x40, 0x20}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**float64", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = float64(8)
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{0x02, 0x40, 0x20}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**float64, but nil float64 part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *float64
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

func Test_Enc_Complex(t *testing.T) {
	var negZero64 = float64(0.0)
	negZero64 *= -1.0
	var negZero32 = float32(0.0)
	negZero32 *= -1.0

	testCases := []struct {
		name   string
		input  interface{}
		expect []byte
	}{
		{name: "complex128 0.0+0.0i", input: complex128(0.0), expect: []byte{0x00}},
		{name: "complex128 (-0.0)+(-0.0)i", input: complex128(complex(negZero64, negZero64)), expect: []byte{0x80}},
		{name: "complex128 (-0.0)+0.0i", input: complex128(complex(negZero64, 0.0)), expect: []byte{0x41, 0x80, 0x02, 0x80, 0x00}},
		{name: "complex128 wide pos real", input: complex128(2.02499999999999991118215802999 + 1.0i), expect: []byte{0x41, 0x80, 0x0c /*=len*/, 0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/}},
		{name: "complex128 wide neg imag", input: complex128(1.0 + -2.02499999999999991118215802999i), expect: []byte{0x41, 0x80, 0x0c /*=len*/, 0x02, 0x3f, 0xf0 /*=real*/, 0x88, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33 /*=imag*/}},
		{name: "complex128 1.0", input: complex128(1.0), expect: []byte{0x41, 0x80, 0x04 /*=len*/, 0x02, 0x3f, 0xf0 /*=real*/, 0x00 /*=imag*/}},
		{name: "complex128 8.25 + 256ish", input: complex128(8.25 + 256.01220703125i), expect: []byte{0x41, 0x80, 0x09 /*=len*/, 0x03, 0xc0, 0x20, 0x80 /*=real*/, 0x04, 0xc0, 0x70, 0x00, 0x32 /*=imag*/}},
		{name: "complex128 -1.0i", input: complex128(-1.0i), expect: []byte{0x41, 0x80, 0x04 /*=len*/, 0x00 /*=real*/, 0x82, 0x3f, 0xf0 /*=imag*/}},
		{name: "complex128 -413.0 + 8.25i", input: complex128(-413.0 + 8.25i), expect: []byte{0x41, 0x80, 0x08 /*=len*/, 0x83, 0xc0, 0x79, 0xd0 /*=real*/, 0x03, 0xc0, 0x20, 0x80 /*=imag*/}},

		{name: "complex64 0.0+0.0i", input: complex64(0.0), expect: []byte{0x00}},
		{name: "complex64 (-0.0)+(-0.0)i", input: complex64(complex(negZero64, negZero64)), expect: []byte{0x80}},
		{name: "complex64 wide pos real", input: complex64(8.38218975067138671875 + 1.0i), expect: []byte{0x41, 0x80, 0x09 /*=len*/, 0x05, 0xc0, 0x20, 0xc3, 0xae, 0x60 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/}},
		{name: "complex64 wide neg imag", input: complex64(1.0 + -8.38218975067138671875i), expect: []byte{0x41, 0x80, 0x09 /*=len*/, 0x02, 0x3f, 0xf0 /*=real*/, 0x85, 0xc0, 0x20, 0xc3, 0xae, 0x60 /*=imag*/}},
		{name: "complex64 1.0", input: complex64(1.0), expect: []byte{0x41, 0x80, 0x04 /*=len*/, 0x02, 0x3f, 0xf0 /*=real*/, 0x00 /*=imag*/}},
		{name: "complex64 8.25 + 256ish", input: complex64(8.25 + 256.01220703125i), expect: []byte{0x41, 0x80, 0x09 /*=len*/, 0x03, 0xc0, 0x20, 0x80 /*=real*/, 0x04, 0xc0, 0x70, 0x00, 0x32 /*=imag*/}},
		{name: "complex64 -1.0i", input: complex64(-1.0i), expect: []byte{0x41, 0x80, 0x04 /*=len*/, 0x00 /*=real*/, 0x82, 0x3f, 0xf0 /*=imag*/}},
		{name: "complex64 -413.0 + 8.25i", input: complex64(-413.0 + 8.25i), expect: []byte{0x41, 0x80, 0x08 /*=len*/, 0x83, 0xc0, 0x79, 0xd0 /*=real*/, 0x03, 0xc0, 0x20, 0x80 /*=imag*/}},
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

	t.Run("*complex64 (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *complex64
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*complex64", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = complex64(8.0 + 8.0i)
			input    = &inputVal
			expect   = []byte{0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**complex64", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = complex64(8.0 + 8.0i)
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**complex64, but nil complex64 part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *complex64
			input  = &ptr
			expect = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*complex128 (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *complex128
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*complex128", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = complex128(8.0 + 8.0i)
			input    = &inputVal
			expect   = []byte{0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**complex128", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = complex128(8.0 + 8.0i)
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**complex128, but nil float64 part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *complex128
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

func Test_Enc_Text(t *testing.T) {
	testCases := []struct {
		name   string
		input  encoding.TextMarshaler
		expect []byte
	}{
		{
			name:   "user-defined type",
			input:  testText{name: "VRISKA", value: 8, enabled: true},
			expect: []byte{0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41},
		},
		{
			name:   "IPv4",
			input:  net.ParseIP("128.0.0.1"),
			expect: []byte{0x41, 0x82, 0x09, 0x31, 0x32, 0x38, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31},
		},
		{
			name:   "IPv6",
			input:  net.ParseIP("2001:db8::1"),
			expect: []byte{0x41, 0x82, 0x0b, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x64, 0x62, 0x38, 0x3a, 0x3a, 0x31},
		},
		{
			name:   "big.Int",
			input:  big.NewInt(2023),
			expect: []byte{0x41, 0x82, 0x04, 0x32, 0x30, 0x32, 0x33},
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

	t.Run("*text (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testText
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*text", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = testText{value: 8, name: "VRISKA", enabled: true}
			input    = &inputVal
			expect   = []byte{
				0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**text", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = testText{value: 8, name: "VRISKA", enabled: true}
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{
				0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**text, but nil binary part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *testText
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

func Test_Dec_Float(t *testing.T) {
	var negZero64 = float64(0.0)
	negZero64 *= -1.0
	var negZero32 = float32(0.0)
	negZero32 *= -1.0

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
		{name: "float64 0.0", input: []byte{0x00}, expect: result{val: float64(0.0), consumed: 1}},
		{name: "float64 -0.0", input: []byte{0x80}, expect: result{val: negZero64, consumed: 1}},
		{name: "float64 wide pos", input: []byte{0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}, expect: result{val: float64(2.02499999999999991118215802999), consumed: 9}},
		{name: "float64 wide neg", input: []byte{0x88, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}, expect: result{val: float64(-2.02499999999999991118215802999), consumed: 9}},
		{name: "float64 1.0", input: []byte{0x02, 0x3f, 0xf0}, expect: result{val: float64(1.0), consumed: 3}},
		{name: "float64 256ish", input: []byte{0x04, 0xc0, 0x70, 0x00, 0x32}, expect: result{val: float64(256.01220703125), consumed: 5}},
		{name: "float64 -1.0", input: []byte{0x82, 0x3f, 0xf0}, expect: result{val: float64(-1.0), consumed: 3}},
		{name: "float64 -413.0", input: []byte{0x83, 0xc0, 0x79, 0xd0}, expect: result{val: float64(-413.0), consumed: 4}},
		{name: "float64 +inf", input: []byte{0x02, 0x7f, 0xf0}, expect: result{val: float64(math.Inf(0)), consumed: 3}},
		{name: "float64 -inf", input: []byte{0x82, 0x7f, 0xf0}, expect: result{val: float64(math.Inf(-1)), consumed: 3}},

		{name: "float32 0.0", input: []byte{0x00}, expect: result{val: float32(0.0), consumed: 1}},
		{name: "float32 -0.0", input: []byte{0x80}, expect: result{val: negZero32, consumed: 1}},
		{name: "float32 wide pos", input: []byte{0x05, 0xc0, 0x20, 0xc3, 0xae, 0x60}, expect: result{val: float32(8.38218975067138671875), consumed: 6}},
		{name: "float32 wide neg", input: []byte{0x85, 0xc0, 0x20, 0xc3, 0xae, 0x60}, expect: result{val: float32(-8.38218975067138671875), consumed: 6}},
		{name: "float32 1.0", input: []byte{0x02, 0x3f, 0xf0}, expect: result{val: float32(1.0), consumed: 3}},
		{name: "float32 256ish", input: []byte{0x04, 0xc0, 0x70, 0x00, 0x32}, expect: result{val: float32(256.01220703125), consumed: 5}},
		{name: "float32 -1.0", input: []byte{0x82, 0x3f, 0xf0}, expect: result{val: float32(-1.0), consumed: 3}},
		{name: "float32 -413.0", input: []byte{0x83, 0xc0, 0x79, 0xd0}, expect: result{val: float32(-413.0), consumed: 4}},
		{name: "float32 +inf", input: []byte{0x02, 0x7f, 0xf0}, expect: result{val: float32(math.Inf(0)), consumed: 3}},
		{name: "float32 -inf", input: []byte{0x82, 0x7f, 0xf0}, expect: result{val: float32(math.Inf(-1)), consumed: 3}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			var actual result
			var err error

			// we are about to have type examination on our actual pointer
			// be run, so we'd betta pass it the correct kind of ptr
			switch tc.expect.val.(type) {
			case float32:
				var v float32
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case float64:
				var v float64
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

	// rly weird case we need a specific test to debug
	t.Run("Sequential 'weird values' do not corrupt data", func(t *testing.T) {
		assert := assert.New(t)

		input := []byte{
			0x04, 0xc0, 0x70, 0x00, 0x32, // 256.01220703125
			0x04, 0xc0, 0x70, 0x00, 0x32, // 256.01220703125
		}

		var dest float64
		n, err := Dec(input, &dest)
		assert.NoError(err)

		assert.Equal(256.01220703125, dest, "read bad")
		assert.Equal([]byte{0x04, 0xc0, 0x70, 0x00, 0x32}, input[n:], "read corrupted data")
	})

	// need multiple checks for NaN in particular - neg, qNaN, sNaN, whatever NaN.
	t.Run("NaN", func(t *testing.T) {
		assert := assert.New(t)

		// NaN is all 1 exponent field and a non-zero mantissa. Try it.

		// 0 11111111111 0000000000000000000000000000000000000000000000000001
		// "sNaN"
		input_sNaN := []byte{
			0x03,
			0x7f, 0xf0, // exponent and COMP
			0x01, // compacted 1
		}

		// 0 11111111111 1000000000000000000000000000000000000000000000000001
		// "qNaN"
		input_qNaN := []byte{
			0x08,
			0x7f, 0xf8, // exponent and COMP
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		}

		// 1 11111111111 1000000000000000000000000000000000000000000000000001
		// signed shouldn't matter
		input_signedNaN := []byte{
			0x88,
			0x7f, 0xf8, // exponent and COMP
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		}

		// 0 11111111111 1111000011110000000000000000000000000000000000000000
		// even an "irregular" NaN should be decoded as NaN
		input_otherNaN := []byte{
			0x03,
			0xff, 0xff, // exponent and COMP
			0x0f,
		}

		var n int
		var err error
		var f float64

		n, err = Dec(input_sNaN, &f)
		assert.NoError(err, "error decoding sNaN")
		assert.Equal(len(input_sNaN), n, "sNaN did not decode expected len")
		assert.True(math.IsNaN(f), "sNaN input did not decode to an NaN")

		n, err = Dec(input_qNaN, &f)
		assert.NoError(err, "error decoding qNaN")
		assert.Equal(len(input_qNaN), n, "qNaN did not decode expected len")
		assert.True(math.IsNaN(f), "qNaN input did not decode to an NaN")

		n, err = Dec(input_signedNaN, &f)
		assert.NoError(err, "error decoding signed NaN")
		assert.Equal(len(input_signedNaN), n, "signed NaN did not decode expected len")
		assert.True(math.IsNaN(f), "signed NaN input did not decode to an NaN")

		n, err = Dec(input_otherNaN, &f)
		assert.NoError(err, "error decoding other NaN")
		assert.Equal(len(input_otherNaN), n, "other NaN did not decode expected len")
		assert.True(math.IsNaN(f), "other NaN input did not decode to an NaN")
	})

	t.Run("nil *float32", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xa0}
			expect         *float32
			expectConsumed = 1
		)

		var actual *float32 = ref(float32(12.0))
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*float32", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x02, 0x40, 0x20}
			expectVal      = float32(8.0)
			expect         = &expectVal
			expectConsumed = 3
		)

		var actual *float32
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**float32", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x02, 0x40, 0x20}
			expectVal      = float32(8.0)
			expectValPtr   = &expectVal
			expect         = &expectValPtr
			expectConsumed = 3
		)

		var actual **float32
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**float32, but nil float32 part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xb0, 0x01, 0x01}
			expectConsumed = 3
		)

		var actual **float32
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)

		assert.NotNil(actual) // actual should *itself* not be nil
		assert.Nil(*actual)   // but the pointer it points to should be nil
	})

	t.Run("nil *float64", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xa0}
			expect         *float64
			expectConsumed = 1
		)

		var actual *float64 = ref(float64(12.0))
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*float64", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x02, 0x40, 0x20}
			expectVal      = float64(8.0)
			expect         = &expectVal
			expectConsumed = 3
		)

		var actual *float64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**float64", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x02, 0x40, 0x20}
			expectVal      = float64(8.0)
			expectValPtr   = &expectVal
			expect         = &expectValPtr
			expectConsumed = 3
		)

		var actual **float64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**float64, but nil float64 part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xb0, 0x01, 0x01}
			expectConsumed = 3
		)

		var actual **float64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)

		assert.NotNil(actual) // actual should *itself* not be nil
		assert.Nil(*actual)   // but the pointer it points to should be nil
	})
}

func Test_Dec_Complex(t *testing.T) {
	var negZero64 = float64(0.0)
	negZero64 *= -1.0
	var negZero32 = float32(0.0)
	negZero32 *= -1.0

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
		{name: "complex128 0.0+0.0i", input: []byte{0x00}, expect: result{val: complex128(0.0 + 0.0i), consumed: 1}},
		{name: "complex128 (-0.0)+(-0.0)i", input: []byte{0x80}, expect: result{val: complex128(complex(negZero64, negZero64)), consumed: 1}},
		{name: "complex128 (-0.0)+0.0i", input: []byte{0x41, 0x80, 0x02, 0x80, 0x00}, expect: result{val: complex128(complex(negZero64, 0.0)), consumed: 5}},
		{name: "complex128 wide pos real", input: []byte{0x41, 0x80, 0x0c /*=len*/, 0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/}, expect: result{val: complex128(2.02499999999999991118215802999 + 1.0i), consumed: 15}},
		{name: "complex128 wide neg imag", input: []byte{0x41, 0x80, 0x0c /*=len*/, 0x02, 0x3f, 0xf0 /*=real*/, 0x88, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33 /*=imag*/}, expect: result{val: complex128(1.0 + -2.02499999999999991118215802999i), consumed: 15}},
		{name: "complex128 1.0", input: []byte{0x41, 0x80, 0x04 /*=len*/, 0x02, 0x3f, 0xf0 /*=real*/, 0x00 /*=imag*/}, expect: result{val: complex128(1.0), consumed: 7}},
		{name: "complex128 8.25 + 256ish", input: []byte{0x41, 0x80, 0x09 /*=len*/, 0x03, 0xc0, 0x20, 0x80 /*=real*/, 0x04, 0xc0, 0x70, 0x00, 0x32 /*=imag*/}, expect: result{val: complex128(8.25 + 256.01220703125i), consumed: 12}},
		{name: "complex128 -1.0i", input: []byte{0x41, 0x80, 0x04 /*=len*/, 0x00 /*=real*/, 0x82, 0x3f, 0xf0 /*=imag*/}, expect: result{val: complex128(-1.0i), consumed: 7}},
		{name: "complex128 -413.0 + 8.25i", input: []byte{0x41, 0x80, 0x08 /*=len*/, 0x83, 0xc0, 0x79, 0xd0 /*=real*/, 0x03, 0xc0, 0x20, 0x80 /*=imag*/}, expect: result{val: complex128(-413.0 + 8.25i), consumed: 11}},

		{name: "complex64 0.0+0.0i", input: []byte{0x00}, expect: result{val: complex64(0.0), consumed: 1}},
		{name: "complex64 (-0.0)+(-0.0)i", input: []byte{0x80}, expect: result{val: complex64(complex(negZero32, negZero32)), consumed: 1}},
		{name: "complex64 (-0.0)+0.0i", input: []byte{0x41, 0x80, 0x02, 0x80, 0x00}, expect: result{val: complex64(complex(negZero32, 0.0)), consumed: 5}},
		{name: "complex64 wide pos real", input: []byte{0x41, 0x80, 0x09 /*=len*/, 0x05, 0xc0, 0x20, 0xc3, 0xae, 0x60 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/}, expect: result{val: complex64(8.38218975067138671875 + 1.0i), consumed: 12}},
		{name: "complex64 wide neg imag", input: []byte{0x41, 0x80, 0x09 /*=len*/, 0x02, 0x3f, 0xf0 /*=real*/, 0x85, 0xc0, 0x20, 0xc3, 0xae, 0x60 /*=imag*/}, expect: result{val: complex64(1.0 + -8.38218975067138671875i), consumed: 12}},
		{name: "complex64 1.0", input: []byte{0x41, 0x80, 0x04 /*=len*/, 0x02, 0x3f, 0xf0 /*=real*/, 0x00 /*=imag*/}, expect: result{val: complex64(1.0), consumed: 7}},
		{name: "complex64 8.25 + 256ish", input: []byte{0x41, 0x80, 0x09 /*=len*/, 0x03, 0xc0, 0x20, 0x80 /*=real*/, 0x04, 0xc0, 0x70, 0x00, 0x32 /*=imag*/}, expect: result{val: complex64(8.25 + 256.01220703125i), consumed: 12}},
		{name: "complex64 -1.0i", input: []byte{0x41, 0x80, 0x04 /*=len*/, 0x00 /*=real*/, 0x82, 0x3f, 0xf0 /*=imag*/}, expect: result{val: complex64(-1.0i), consumed: 7}},
		{name: "complex64 -413.0 + 8.25i", input: []byte{0x41, 0x80, 0x08 /*=len*/, 0x83, 0xc0, 0x79, 0xd0 /*=real*/, 0x03, 0xc0, 0x20, 0x80 /*=imag*/}, expect: result{val: complex64(-413.0 + 8.25i), consumed: 11}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			var actual result
			var err error

			// we are about to have type examination on our actual pointer
			// be run, so we'd betta pass it the correct kind of ptr
			switch tc.expect.val.(type) {
			case complex64:
				var v complex64
				actual.consumed, err = Dec(tc.input, &v)
				actual.val = v
			case complex128:
				var v complex128
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

	t.Run("nil *complex64", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xa0}
			expect         *complex64
			expectConsumed = 1
		)

		var actual *complex64 = ref(complex64(12.0))
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*complex64", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20}
			expectVal      = complex64(8.0 + 8.0i)
			expect         = &expectVal
			expectConsumed = 9
		)

		var actual *complex64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**complex64", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20}
			expectVal      = complex64(8.0 + 8.0i)
			expectValPtr   = &expectVal
			expect         = &expectValPtr
			expectConsumed = 9
		)

		var actual **complex64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**complex64, but nil complex64 part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xb0, 0x01, 0x01}
			expectConsumed = 3
		)

		var actual **complex64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)

		assert.NotNil(actual) // actual should *itself* not be nil
		assert.Nil(*actual)   // but the pointer it points to should be nil
	})

	t.Run("nil *complex128", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xa0}
			expect         *complex128
			expectConsumed = 1
		)

		var actual *complex128 = ref(complex128(12.0))
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*complex128", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20}
			expectVal      = complex128(8.0 + 8.0i)
			expect         = &expectVal
			expectConsumed = 9
		)

		var actual *complex128
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**complex128", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20}
			expectVal      = complex128(8.0 + 8.0i)
			expectValPtr   = &expectVal
			expect         = &expectValPtr
			expectConsumed = 9
		)

		var actual **complex128
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**complex128, but nil complex128 part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xb0, 0x01, 0x01}
			expectConsumed = 3
		)

		var actual **complex128
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

func Test_Dec_Text(t *testing.T) {
	// cannot be table-driven easily due to trying different types

	t.Run("normal TextUnmarshaler result", func(t *testing.T) {
		// setup
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41}
			expect         = testText{name: "VRISKA", value: 8, enabled: true}
			expectConsumed = 16
		)

		// exeucte
		actual := testText{}
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("nil *TextUnmarshaler", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xa0}
			expect         *testText
			expectConsumed = 1
		)

		var actual *testText = &testText{enabled: true}
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*TextUnmarshaler", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41}
			expectVal      = testText{name: "VRISKA", value: 8, enabled: true}
			expect         = &expectVal
			expectConsumed = 16
		)

		var actual *testText
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**TextUnmarshaler", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41}
			expectVal      = testText{name: "VRISKA", value: 8, enabled: true}
			expectValPtr   = &expectVal
			expect         = &expectValPtr
			expectConsumed = 16
		)

		var actual **testText
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**TextUnmarshaler, but nil TextUnmarshaler part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0xb0, 0x01, 0x01}
			expectConsumed = 3
		)

		var actual **testText
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)

		assert.NotNil(actual) // actual should *itself* not be nil
		assert.Nil(*actual)   // but the pointer it points to should be nil
	})

	t.Run("IPv4", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x09, 0x31, 0x32, 0x38, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31}
			expect         = net.ParseIP("128.0.0.1")
			expectConsumed = 12
		)

		var actual net.IP
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*IPv4", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x09, 0x31, 0x32, 0x38, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31}
			expectVal      = net.ParseIP("128.0.0.1")
			expect         = &expectVal
			expectConsumed = 12
		)

		var actual *net.IP
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("IPv6", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x0b, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x64, 0x62, 0x38, 0x3a, 0x3a, 0x31}
			expect         = net.ParseIP("2001:db8::1")
			expectConsumed = 14
		)

		var actual net.IP
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*IPv6", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input          = []byte{0x41, 0x82, 0x0b, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x64, 0x62, 0x38, 0x3a, 0x3a, 0x31}
			expectVal      = net.ParseIP("2001:db8::1")
			expect         = &expectVal
			expectConsumed = 14
		)

		var actual *net.IP
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("big.Int", func(t *testing.T) {
		assert := assert.New(t)

		// big.Int is an always-ptr type, so we will make expect be the deref'd
		// value

		var (
			input          = []byte{0x41, 0x82, 0x04, 0x32, 0x30, 0x32, 0x33}
			val            = big.NewInt(2023)
			expect         = *val
			expectConsumed = 7
		)

		var actual big.Int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*big.Int", func(t *testing.T) {
		assert := assert.New(t)

		// big.Int is an always-ptr type, so we need not adjust expect, here.

		var (
			input          = []byte{0x41, 0x82, 0x04, 0x32, 0x30, 0x32, 0x33}
			expect         = big.NewInt(2023)
			expectConsumed = 7
		)

		var actual *big.Int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**big.Int", func(t *testing.T) {
		assert := assert.New(t)

		// big.Int is an always-ptr type, so even though its a double * check,
		// expect only needs one level of indir from the expected val.

		var (
			input          = []byte{0x41, 0x82, 0x04, 0x32, 0x30, 0x32, 0x33}
			expectVal      = big.NewInt(2023)
			expect         = &expectVal
			expectConsumed = 7
		)

		var actual **big.Int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

}
