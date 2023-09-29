package rezi

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_reziError_totalOffset_int(t *testing.T) {
	testCases := []struct {
		name              string
		input             []byte
		expectTotalOffset int
	}{
		/*{
			name:              "decode int: empty bytes",
			input:             []byte{},
			expectTotalOffset: 0,
		},*/
		{
			name:              "decode int: count header is extended past slice",
			input:             []byte{0x40},
			expectTotalOffset: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// create the error
			var dest int
			_, err := Dec(tc.input, &dest)
			if !assert.Error(err) {
				return
			}

			// convert it to known type reziError
			rErr, ok := err.(reziError)
			if !assert.Truef(ok, "Dec returned non-reziErr: %v", err) {
				return
			}

			// check if the offset is valid
			actual, ok := rErr.totalOffset()
			if !assert.Truef(ok, "Dec returned reziErr with no offset: %v", err) {
				return
			}

			// assert the offset is the expected
			if !assert.Equal(tc.expectTotalOffset, actual) {
				return
			}

			// and finally, ensure the offset is in the error output
			assert.Falsef(strings.Contains(rErr.Error(), fmt.Sprintf("%d", tc.expectTotalOffset)), "message does not contain offset: %q", rErr.Error())
		})
	}
}
