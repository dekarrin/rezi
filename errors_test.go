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

func Test_reziError_totalOffset_binary(t *testing.T) {
	testCases := []struct {
		name              string
		input             []byte
		expectTotalOffset int
		expectErrText     string
	}{
		{
			name: "decode bin: bad byte count",
			input: []byte{
				/* byte count = 11 */ 0x01, 0x0b,
				/*  data  (string) */ 0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
				/* number (int32)  */ 0x01, 0x01, 0x0a, // 10, len=3
			},
			expectTotalOffset: 2,
			expectErrText:     "decoded binary value byte count",
		},
		{
			name: "decode bin: string error",
			input: []byte{
				/* byte count = 11 */ 0x01, 0x0a,
				/*  data  (string) */ 0x41, 0x80, 0x04, 0x41, 0x42, 0x43, 0xc3, // "ABC" followed by invalid seq, len=8
				/* number (int32)  */ 0x01, 0x01, 0x0a, // 10, len=3
			},
			expectTotalOffset: 8,
			expectErrText:     "data: invalid UTF-8 encoding",
		},
		{
			name: "decode bin: number error",
			input: []byte{
				/* byte count = 11 */ 0x01, 0x07,
				/*  data  (string) */ 0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
				/* number (int32)  */ 0x01, // a length but no actual num
			},
			expectTotalOffset: 9,
			expectErrText:     "number: decoded int byte count",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			var dest testBinary
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
