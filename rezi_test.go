package rezi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			expect: []byte{0x01, 0x01, 0x56},
		},
		{
			name:   "several chars",
			input:  "Vriska",
			expect: []byte{0x01, 0x06, 0x56, 0x72, 0x69, 0x73, 0x6b, 0x61},
		},
		{
			name:   "pre-composed char seq",
			input:  "homeçtuck",
			expect: []byte{0x01, 0x09, 0x68, 0x6f, 0x6d, 0x65, 0xc3, 0xa7, 0x74, 0x75, 0x63, 0x6b},
		},
		{
			name:   "decomposed char seq",
			input:  "homec\u0327tuck",
			expect: []byte{0x01, 0x0a, 0x68, 0x6f, 0x6d, 0x65, 0x63, 0xcc, 0xa7, 0x74, 0x75, 0x63, 0x6b},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			actual := Enc(tc.input)
			assert.Equal(tc.expect, actual)
		})
	}
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

			actual := Enc(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}

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

			actual := Enc(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}

}

func Test_Enc_Binary(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Enc_Map(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Enc_Slice(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

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
			name:   "empty",
			input:  []byte{0x00},
			expect: result{val: "", consumed: 1},
		},
		{
			name:   "one char",
			input:  []byte{0x01, 0x01, 0x56},
			expect: result{val: "V", consumed: 3},
		},
		{
			name:   "several chars",
			input:  []byte{0x01, 0x06, 0x56, 0x72, 0x69, 0x73, 0x6b, 0x61},
			expect: result{val: "Vriska", consumed: 8},
		},
		{
			name:   "pre-composed char seq",
			input:  []byte{0x01, 0x09, 0x68, 0x6f, 0x6d, 0x65, 0xc3, 0xa7, 0x74, 0x75, 0x63, 0x6b},
			expect: result{val: "homeçtuck", consumed: 12},
		},
		{
			name:   "decomposed char seq",
			input:  []byte{0x01, 0x0a, 0x68, 0x6f, 0x6d, 0x65, 0x63, 0xcc, 0xa7, 0x74, 0x75, 0x63, 0x6b},
			expect: result{val: "homec\u0327tuck", consumed: 13},
		},
		{
			name:   "err count too big",
			input:  []byte{0x01, 0x08, 0x68, 0x6f},
			expect: result{err: true},
		},
		{
			name:   "err invalid sequence",
			input:  []byte{0x01, 0x01, 0xc3, 0x28},
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

}

func Test_Dec_Binary(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Dec_Map(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Dec_Slice(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}