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
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
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
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}
func Test_Dec_Bool(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
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
