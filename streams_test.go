package rezi

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			expectOff: 4,
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
			expectOff: 4,
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

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x80}
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

	// nil with multiple indirs not possible with V0
}
