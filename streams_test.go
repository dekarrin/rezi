package rezi

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Reader_Dec_sequential(t *testing.T) {
	assert := assert.New(t)
	var input []byte

	input = append(input, 0x02, 0x01, 0x9d) // 413
	var dest1Int int
	expect1Int := 413

	input = append(input, 0x01) // true
	var dest2Bool bool
	expect2Bool := true

	input = append(input, 0xa0) // nil
	var dest3Slice1 []int
	var expect3Slice1 []int

	input = append(input, // slice: {"VRISKA", "NEPETA", "TEREZI"}
		0x01, 0x1b, // len = 27
		0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
		0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
		0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"
	)
	var dest4Slice2 []string
	expect4Slice2 := []string{"VRISKA", "NEPETA", "TEREZI"}

	input = append(input, 0x41, 0x80, 0x01, 0x31) // "1"
	var dest5String string
	expect5String := "1"

	input = append(input, 0x02, 0x02, 0x64) // 612
	var dest6IntPtr *int
	expect6IntPtr := ref(612)

	input = append(input, // map: {413: "JOHN", 612: "VRISKA"}
		0x01, 0x16, // len=22

		0x02, 0x01, 0x9d, // 413:
		0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"

		0x02, 0x02, 0x64, // 612:
		0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
	)
	var dest7Map map[int]string
	expect7Map := map[int]string{413: "JOHN", 612: "VRISKA"}

	input = append(input, 0xa0) // nil
	var dest8BoolPtr *bool
	expect8BoolPtr := nilRef[bool]()

	input = append(input, // testBinary{data: "ABC", number: 8}
		/* byte count = 8  */ 0x01, 0x08,
		/*  data  (string) */ 0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
		/* number (int32)  */ 0x01, 0x08, // 8
	)
	var dest9Bin testBinary
	expect9Bin := testBinary{number: 8, data: "ABC"}

	r, err := NewReader(bytes.NewReader(input), nil)
	if !assert.NoError(err, "creating Reader returned error") {
		return
	}

	err = r.Dec(&dest1Int)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect1Int, dest1Int, "dest1Int mismatch")

	err = r.Dec(&dest2Bool)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect2Bool, dest2Bool, "dest2Bool mismatch")

	err = r.Dec(&dest3Slice1)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect3Slice1, dest3Slice1, "dest3Slice1 mismatch")

	err = r.Dec(&dest4Slice2)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect4Slice2, dest4Slice2, "dest4Slice2 mismatch")

	err = r.Dec(&dest5String)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect5String, dest5String, "dest5String mismatch")

	err = r.Dec(&dest6IntPtr)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect6IntPtr, dest6IntPtr, "dest6IntPtr mismatch")

	err = r.Dec(&dest7Map)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect7Map, dest7Map, "dest7Map mismatch")

	err = r.Dec(&dest8BoolPtr)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect8BoolPtr, dest8BoolPtr, "dest8BoolPtr mismatch")

	err = r.Dec(&dest9Bin)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect9Bin, dest9Bin, "dest9Bin mismatch")

}

func Test_Reader_Dec_int(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    int
		expectOff int
		expectErr bool
	}{
		{
			name:      "normal value",
			input:     []byte{0x02, 0x01, 0x9d},
			expect:    413,
			expectOff: 3,
		},
		{
			// due to header sniffing for V1, this will fail if second byte has 0x80 bit set
			name:      "skip extension bytes",
			input:     []byte{0x42, 0x7f, 0xbf, 0x22, 0xb8},
			expect:    8888,
			expectOff: 5,
		},
		{
			name:      "normal value - multi value",
			input:     []byte{0x02, 0x01, 0x9d, 0x02, 0x01, 0x9d},
			expect:    413,
			expectOff: 3,
		},
		{
			// due to header sniffing for V1, this will fail if second byte has 0x80 bit set
			name:      "skip extension bytes - multi value",
			input:     []byte{0x42, 0x7f, 0xbf, 0x22, 0xb8, 0x42, 0x7f, 0xbf, 0x22, 0xb8},
			expect:    8888,
			expectOff: 5,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest int
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x20}
		expect := nilRef[int]()
		expectOff := 1

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *int
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		expectPtr := nilRef[int]()
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest **int
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_bool(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    bool
		expectErr bool
		expectOff int
	}{
		{
			name:      "normal value - false",
			input:     []byte{0x00},
			expect:    false,
			expectOff: 1,
		},
		{
			name:      "normal value - true",
			input:     []byte{0x01},
			expect:    true,
			expectOff: 1,
		},
		{
			name:      "normal value - false - multi value",
			input:     []byte{0x00, 0x00},
			expect:    false,
			expectOff: 1,
		},
		{
			name:      "normal value - true - multi value",
			input:     []byte{0x01, 0x01},
			expect:    true,
			expectOff: 1,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest bool
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x20}
		expect := nilRef[bool]()
		expectOff := 1

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *bool
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		expectPtr := nilRef[bool]()
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest **bool
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_stringV0(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    string
		expectErr bool
		expectOff int
	}{
		{
			name:      "empty",
			input:     []byte{0x00},
			expect:    "",
			expectOff: 1,
		},
		{
			name:      "string with no multibyte chars",
			input:     []byte{0x01, 0x01, 0x31},
			expect:    "1",
			expectOff: 3,
		},
		{
			name:      "string with multibyte chars",
			input:     []byte{0x01, 0x09, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c},
			expect:    "Hello, 世界",
			expectOff: 15,
		},
		{
			name:      "empty - multiple values",
			input:     []byte{0x00, 0x00},
			expect:    "",
			expectOff: 1,
		},
		{
			name:      "string with no multibyte chars - multiple values",
			input:     []byte{0x01, 0x01, 0x31, 0x01, 0x01, 0x31},
			expect:    "1",
			expectOff: 3,
		},
		{
			name:      "string with multibyte chars - multiple values",
			input:     []byte{0x01, 0x09, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c, 0x01, 0x09, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c},
			expect:    "Hello, 世界",
			expectOff: 15,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest string
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	// cannot encode a v0 string that is nil so we also cannot decode one
}

func Test_Reader_Dec_stringV1(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    string
		expectErr bool
		expectOff int
	}{
		{
			name:      "empty",
			input:     []byte{0x00},
			expect:    "",
			expectOff: 1,
		},
		{
			name:      "string with no multibyte chars",
			input:     []byte{0x41, 0x80, 0x01, 0x31},
			expect:    "1",
			expectOff: 4,
		},
		{
			name:      "string with multibyte chars",
			input:     []byte{0x41, 0x80, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c},
			expect:    "Hello, 世界",
			expectOff: 16,
		},
		{
			name:      "empty - multiple values",
			input:     []byte{0x00, 0x00},
			expect:    "",
			expectOff: 1,
		},
		{
			name:      "string with no multibyte chars - multiple values",
			input:     []byte{0x41, 0x80, 0x01, 0x31, 0x40, 0x80, 0x01, 0x01, 0x31},
			expect:    "1",
			expectOff: 4,
		},
		{
			name:      "string with multibyte chars - multiple values",
			input:     []byte{0x41, 0x80, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c, 0x40, 0x80, 0x01, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c},
			expect:    "Hello, 世界",
			expectOff: 16,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest string
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x20}
		expect := nilRef[string]()
		expectOff := 1

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *string
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		expectPtr := nilRef[string]()
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest **string
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_binary(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    testBinary
		expectErr bool
		expectOff int
	}{
		{
			name: "empty values",
			input: []byte{
				/* byte count = 2  */ 0x01, 0x02,
				/*  data  (string) */ 0x00, // ""
				/* number (int32)  */ 0x00, // 0
			},
			expect:    testBinary{},
			expectOff: 4,
		},
		{
			name: "filled values",
			input: []byte{
				/* byte count = 8  */ 0x01, 0x08,
				/*  data  (string) */ 0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
				/* number (int32)  */ 0x01, 0x08, // 8
			},
			expect:    testBinary{data: "ABC", number: 8},
			expectOff: 10,
		},
		{
			name: "empty values x2",
			input: []byte{
				/* byte count = 2  */ 0x01, 0x02,
				/*  data  (string) */ 0x00, // ""
				/* number (int32)  */ 0x00, // 0

				/* byte count = 2  */ 0x01, 0x02,
				/*  data  (string) */ 0x00, // ""
				/* number (int32)  */ 0x00, // 0
			},
			expect:    testBinary{},
			expectOff: 4,
		},
		{
			name: "filled values x2",
			input: []byte{
				/* byte count = 8  */ 0x01, 0x08,
				/*  data  (string) */ 0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
				/* number (int32)  */ 0x01, 0x08, // 8

				/* byte count = 8  */ 0x01, 0x08,
				/*  data  (string) */ 0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
				/* number (int32)  */ 0x01, 0x08, // 8
			},
			expect:    testBinary{data: "ABC", number: 8},
			expectOff: 10,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest testBinary
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x20}
		expect := nilRef[testBinary]()
		expectOff := 1

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *testBinary
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		expectPtr := nilRef[testBinary]()
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest **testBinary
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_slice(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    []string
		expectErr bool
		expectOff int
	}{
		{
			name: "nil",
			input: []byte{
				0xa0, // nil=true
			},
			expect:    nil,
			expectOff: 1,
		},
		{
			name: "empty",
			input: []byte{
				0x00, // len=0
			},
			expect:    []string{},
			expectOff: 1,
		},
		{
			name: "3 value",
			input: []byte{
				0x01, 0x1b, // len=27

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
				0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"
			},
			expect:    []string{"VRISKA", "NEPETA", "TEREZI"},
			expectOff: 29,
		},
		{
			name: "nil x2",
			input: []byte{
				0xa0, // nil=true
				0xa0, // nil=true
			},
			expect:    nil,
			expectOff: 1,
		},
		{
			name: "empty x2",
			input: []byte{
				0x00, // len=0
				0x00, // len=0
			},
			expect:    []string{},
			expectOff: 1,
		},
		{
			name: "3 value x2",
			input: []byte{
				0x01, 0x1b, // len=27

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
				0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"

				0x01, 0x1b, // len=27

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
				0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"
			},
			expect:    []string{"VRISKA", "NEPETA", "TEREZI"},
			expectOff: 29,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest []string
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		var expectPtr []string
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *[]string
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_map(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    map[int]string
		expectErr bool
		expectOff int
	}{
		{
			name: "nil",
			input: []byte{
				0xa0, // nil=true
			},
			expect:    nil,
			expectOff: 1,
		},
		{
			name: "empty",
			input: []byte{
				0x00, // len=0
			},
			expect:    map[int]string{},
			expectOff: 1,
		},
		{
			name: "2 values",
			input: []byte{
				0x01, 0x16, // len=22

				0x02, 0x01, 0x9d, // 413:
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"

				0x02, 0x02, 0x64, // 612:
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
			},
			expect:    map[int]string{413: "JOHN", 612: "VRISKA"},
			expectOff: 24,
		},
		{
			name: "nil x2",
			input: []byte{
				0xa0, // nil=true
				0xa0, // nil=true
			},
			expect:    nil,
			expectOff: 1,
		},
		{
			name: "empty x2",
			input: []byte{
				0x00, // len=0
				0x00, // len=0
			},
			expect:    map[int]string{},
			expectOff: 1,
		},
		{
			name: "2 values x2",
			input: []byte{
				0x01, 0x16, // len=22

				0x02, 0x01, 0x9d, // 413:
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"

				0x02, 0x02, 0x64, // 612:
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"

				0x01, 0x16, // len=22

				0x02, 0x01, 0x9d, // 413:
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"

				0x02, 0x02, 0x64, // 612:
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
			},
			expect:    map[int]string{413: "JOHN", 612: "VRISKA"},
			expectOff: 24,
		},

		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest map[int]string
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		var expectPtr map[int]string
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *map[int]string
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}
