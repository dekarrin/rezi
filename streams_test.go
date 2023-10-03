package rezi

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Reader_Dec_int(t *testing.T) {
	testCases := struct {
		name      string
		data      []byte
		expectVal int
		expectErr bool
	}{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.data), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest int
			err := r.Dec(&dest)
		})
	}
}
