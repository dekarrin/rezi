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
		{
			name:              "decode int: empty bytes",
			input:             []byte{},
			expectTotalOffset: 0,
		},
		{
			name:              "decode int: count header is extended past slice",
			input:             []byte{0x40},
			expectTotalOffset: 0,
		},
		{
			name:              "decode int: length past slice",
			input:             []byte{0x42, 0x00, 0x00},
			expectTotalOffset: 2,
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
			assert.Equal(tc.expectTotalOffset, actual)

			// and finally, ensure the offset is in the error output
			assert.Truef(strings.Contains(rErr.Error(), fmt.Sprintf("%d", tc.expectTotalOffset)), "message does not contain offset: %q", rErr.Error())
		})
	}
}

func Test_reziError_totalOffset_bool(t *testing.T) {
	testCases := []struct {
		name              string
		input             []byte
		expectTotalOffset int
	}{
		{
			name:              "decode bool: empty bytes",
			input:             []byte{},
			expectTotalOffset: 0,
		},
		{
			name:              "decode bool: outside set of values",
			input:             []byte{0x02},
			expectTotalOffset: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// create the error
			var dest bool
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
			assert.Equal(tc.expectTotalOffset, actual)

			// and finally, ensure the offset is in the error output
			assert.Truef(strings.Contains(rErr.Error(), fmt.Sprintf("%d", tc.expectTotalOffset)), "message does not contain offset: %q", rErr.Error())
		})
	}
}

func Test_reziError_totalOffset_string(t *testing.T) {
	testCases := []struct {
		name              string
		input             []byte
		expectTotalOffset int
		expectErrText     string
	}{
		{
			name:              "decode string: empty bytes",
			input:             []byte{},
			expectTotalOffset: 0,
			expectErrText:     "unexpected EOF",
		},
		{
			name:              "decode string v1: decode rune len int: length past slice",
			input:             []byte{0x42, 0x00, 0x00},
			expectTotalOffset: 2,
			expectErrText:     "decode string rune count",
		},
		{
			name:              "decode string v1: decode rune len int: len of -1",
			input:             []byte{0x80},
			expectTotalOffset: 0,
			expectErrText:     "rune count < 0",
		},
		{
			name:              "decode string v1: invalid character byte seq at start",
			input:             []byte{0x01, 0x01, 0xc3, 0x28},
			expectTotalOffset: 2,
			expectErrText:     "invalid UTF-8 encoding",
		},
		{
			name:              "decode string v1: invalid character byte seq at char 2",
			input:             []byte{0x01, 0x02, 0x41, 0xc3, 0x28},
			expectTotalOffset: 3,
			expectErrText:     "invalid UTF-8 encoding",
		},
		{
			name:              "decode string v2: decode byte len int: length past slice",
			input:             []byte{0x41, 0x80},
			expectTotalOffset: 2,
			expectErrText:     "byte count:",
		},
		{
			name:              "decode string v2: decode byte len int: len of -1",
			input:             []byte{0xc0, 0x80},
			expectTotalOffset: 2,
			expectErrText:     "byte count < 0",
		},
		{
			name:              "decode string v2: len < byte len",
			input:             []byte{0x41, 0x80, 0x02, 0x41},
			expectTotalOffset: 3,
			expectErrText:     "only 1 byte remains",
		},
		{
			name:              "decode string v2: invalid character byte seq at start",
			input:             []byte{0x41, 0x80, 0x01, 0xc3, 0x28},
			expectTotalOffset: 3,
			expectErrText:     "invalid UTF-8 encoding",
		},
		{
			name:              "decode string v2: invalid character byte seq at char 2",
			input:             []byte{0x41, 0x80, 0x02, 0x41, 0xc3, 0x28},
			expectTotalOffset: 4,
			expectErrText:     "invalid UTF-8 encoding",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// create the error
			var dest string
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
			assert.Equal(tc.expectTotalOffset, actual)

			// and finally, check the err output
			expectOffsetStr := fmt.Sprintf("%d", tc.expectTotalOffset)
			lowerMsgAct := strings.ToUpper(rErr.Error())
			lowerMsgExp := strings.ToUpper(tc.expectErrText)
			assert.Contains(rErr.Error(), expectOffsetStr, "message does not contain offset: %q", rErr.Error())
			assert.Contains(lowerMsgAct, lowerMsgExp, "message does not contain %q: %q", tc.expectErrText, rErr.Error())
		})
	}
}
