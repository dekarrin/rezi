package rezi

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Wrapf(t *testing.T) {
	type params struct {
		offset  int
		format  string
		err     error
		fmtArgs []interface{}
	}

	testCases := []struct {
		name          string
		input         params
		expectErrText string
		expectPanic   bool
	}{
		{
			name: "add positive offset",
			input: params{
				offset: 6,
				format: "err: %s",
				err:    reziError{msg: "some error", offsetValid: true, offset: 2},
			},
			expectErrText: "(0x08): err: some error",
		},
		{
			name: "add negative offset",
			input: params{
				offset: -1,
				format: "err: %s",
				err:    reziError{msg: "some error", offsetValid: true, offset: 2},
			},
			expectErrText: "(0x01): err: some error",
		},
		{
			name: "add no offset",
			input: params{
				offset: 0,
				format: "err: %s",
				err:    reziError{msg: "some error", offsetValid: true, offset: 2},
			},
			expectErrText: "(0x02): err: some error",
		},
		{
			name: "non-rezi error causes panic",
			input: params{
				offset: 0,
				format: "err: %s",
				err:    errors.New("some error"),
			},
			expectPanic: true,
		},
		{
			name: "%w causes panic",
			input: params{
				offset: 0,
				format: "err: %w",
				err:    reziError{msg: "some error", offsetValid: true, offset: 2},
			},
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			p := tc.input

			if tc.expectPanic {
				assert.Panics(func() {
					Wrapf(p.offset, p.format, p.err, p.fmtArgs...)
				})
				return
			}

			actual := Wrapf(p.offset, p.format, p.err, p.fmtArgs...)

			// and finally, check the err output
			lowerMsgAct := strings.ToUpper(actual.Error())
			lowerMsgExp := strings.ToUpper(tc.expectErrText)
			assert.Contains(lowerMsgAct, lowerMsgExp, "message does not contain %q: %q", tc.expectErrText, actual.Error())
		})
	}
}

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
			expectOffsetStr := fmt.Sprintf("%d", tc.expectTotalOffset)
			assert.Contains(rErr.Error(), expectOffsetStr, "message does not contain offset: %q", rErr.Error())
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
			expectOffsetStr := fmt.Sprintf("%d", tc.expectTotalOffset)
			assert.Contains(rErr.Error(), expectOffsetStr, "message does not contain offset: %q", rErr.Error())
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

func Test_reziError_totalOffset_slice(t *testing.T) {
	stringSlice := func(data []byte) error {
		var sDest []string
		_, err := Dec(data, &sDest)
		return err
	}
	intSlice := func(data []byte) error {
		var iDest []int
		_, err := Dec(data, &iDest)
		return err
	}

	testCases := []struct {
		name              string
		input             []byte
		dest              func([]byte) error
		expectTotalOffset int
		expectErrText     string
	}{
		{
			name: "decode []int: bad byte count",
			dest: intSlice,
			input: []byte{
				0x01, 0x0d, // len=13

				0x01, 0x01, // 1
				0x01, 0x03, // 3
				0x01, 0x04, // 4
				0x01, 0xc8, // 200
				0x03, 0x04, 0x4b, 0x41, // 281409
			},
			expectTotalOffset: 2,
			expectErrText:     "decoded slice byte count",
		},
		{
			name: "decode []string: problem with element 3",
			dest: stringSlice,
			input: []byte{
				0x01, 0x11, // len=17

				0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
				0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
				0x41, 0x80, 0x02, 0xc3, 0x41, // an invalid sequence followed by "A"
			},
			expectTotalOffset: 17,
			expectErrText:     "item[2]: invalid UTF-8",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			err := tc.dest(tc.input)
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

func Test_reziError_totalOffset_map(t *testing.T) {
	stringIntMap := func(data []byte) error {
		var siDest map[string]int
		_, err := Dec(data, &siDest)
		return err
	}
	intStringMap := func(data []byte) error {
		var isDest map[int]string
		_, err := Dec(data, &isDest)
		return err
	}

	testCases := []struct {
		name              string
		input             []byte
		dest              func([]byte) error
		expectTotalOffset int
		expectErrText     string
	}{
		{
			name: "decode map[int]string: bad byte count",
			dest: intStringMap,
			input: []byte{
				0x01, 0x23, // len=35 (one more than actual)

				0x02, 0x01, 0x9d, // 413:
				0x41, 0x80, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"

				0x02, 0x02, 0x64, // 612:
				0x41, 0x80, 0x06, 0x4b, 0x41, 0x52, 0x4b, 0x41, 0x54, // "KARKAT"

				0x02, 0xbe, 0x57, // 48727 ("BEST"):
				0x41, 0x80, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
			},
			expectTotalOffset: 2,
			expectErrText:     "decoded map byte count",
		},
		{
			name: "decode map[int]string: value issue in 2nd entry",
			dest: intStringMap,
			input: []byte{
				0x01, 0x23, // len=35

				0x02, 0x01, 0x9d, // 413:
				0x41, 0x80, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"

				0x02, 0x02, 0x64, // 612:
				0x41, 0x80, 0x07, 0x4b, 0x41, 0x52, 0x4b, 0x41, 0x54, 0xc3, // "KARKAT" followed by bad seq

				0x02, 0xbe, 0x57, // 48727 ("BEST"):
				0x41, 0x80, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
			},
			expectTotalOffset: 24,
			expectErrText:     "map value[612]: invalid",
		},
		{
			name: "decode map[string]int: key issue in 3rd entry",
			dest: stringIntMap,
			input: []byte{
				0x01, 0x23, // len=35

				0x41, 0x80, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN":
				0x02, 0x01, 0x9d, // 413

				0x41, 0x80, 0x06, 0x4b, 0x41, 0x52, 0x4b, 0x41, 0x54, // "KARKAT":
				0x02, 0x02, 0x64, // 612

				0x41, 0x80, 0x07, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, 0xc3, // "NEPETA" (followed by invalid seq):
				0x02, 0xbe, 0x57, // 48727 ("BEST")
			},
			expectTotalOffset: 33,
			expectErrText:     "map key: invalid UTF-8 encoding",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			err := tc.dest(tc.input)
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
