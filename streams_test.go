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
		expectErr bool
	}{
		{
			name:   "normal value",
			input:  []byte{0x02, 0x01, 0x9d},
			expect: 413,
		},
		{
			name:   "skip extension bytes",
			input:  []byte{0x42, 0xff, 0xbf, 0x22, 0xb8},
			expect: 8888,
		},
		{
			name:   "normal value - multi value",
			input:  []byte{0x02, 0x01, 0x9d, 0x02, 0x01, 0x9d},
			expect: 413,
		},
		{
			name:   "skip extension bytes - multi value",
			input:  []byte{0x42, 0xff, 0xbf, 0x22, 0xb8, 0x42, 0xff, 0xbf, 0x22, 0xb8},
			expect: 8888,
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
				assert.Error(err)
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest)
		})
	}
}
