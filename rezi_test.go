package rezi

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testUnsupportedOOB struct {
	Name        string
	unsupported func()
}

func (oob testUnsupportedOOB) MarshalBinary() ([]byte, error) {
	return Enc(oob.Name)
}

func (oob *testUnsupportedOOB) UnmarshalBinary(data []byte) error {
	_, err := Dec(data, &(oob.Name))
	return err
}

type testUnsupportedOOBLevel2 testUnsupportedOOB

func Test_Underlying_Supported_Via_Marshal(t *testing.T) {
	assert := assert.New(t)

	myData := testUnsupportedOOBLevel2{Name: "Jack", unsupported: nil}

	data, err := Enc(myData)
	if !assert.NoError(err, "encode error") {
		return
	}

	var decodeTo testUnsupportedOOBLevel2
	_, err = Dec(data, &decodeTo)
	if !assert.NoError(err, "decode error") {
		return
	}

	assert.Equal(myData, decodeTo)
}

func Test_Enc_Errors(t *testing.T) {
	dummyTyped := make(chan struct{})

	ErrFakeMarshal := errors.New("fake marshal error")

	testCases := []struct {
		name      string
		input     interface{}
		expectErr error
	}{
		{
			name:      "unknown type - Error",
			input:     dummyTyped,
			expectErr: Error,
		},
		{
			name:      "unknown type - ErrInvalidType",
			input:     dummyTyped,
			expectErr: ErrInvalidType,
		},
		{
			name:      "marshal failure - Error",
			input:     testBinary{encErr: ErrFakeMarshal},
			expectErr: Error,
		},
		{
			name:      "marshal failure - ErrMarshalBinary",
			input:     testBinary{encErr: ErrFakeMarshal},
			expectErr: ErrMarshalBinary,
		},
		{
			name:      "marshal failure - wrapped error",
			input:     testBinary{encErr: ErrFakeMarshal},
			expectErr: ErrFakeMarshal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			_, actual := Enc(tc.input)

			assert.ErrorIs(actual, tc.expectErr)
		})
	}
}

func Test_Dec_Errors(t *testing.T) {
	dummyTyped := make(chan struct{})

	ErrTestError := errors.New("test error")

	testCases := []struct {
		name      string
		data      []byte
		recv      interface{}
		expectErr error
	}{
		{
			name:      "receiver is nil - Error",
			recv:      nil,
			expectErr: Error,
		},
		{
			name:      "receiver is nil - ErrInvalidType",
			recv:      nil,
			expectErr: ErrInvalidType,
		},
		{
			name:      "receiver is typed nil - Error",
			recv:      nilRef[int](),
			expectErr: Error,
		},
		{
			name:      "receiver is typed nil - ErrInvalidType",
			recv:      nilRef[int](),
			expectErr: ErrInvalidType,
		},
		{
			name:      "receiver is unsupported type - Error",
			recv:      ref(dummyTyped),
			expectErr: Error,
		},
		{
			name:      "receiver is unsupported type - ErrInvalidType",
			recv:      ref(dummyTyped),
			expectErr: ErrInvalidType,
		},
		{
			name:      "unmarshal failure - Error",
			data:      []byte{0x01, 0x01, 0x00},
			recv:      &testBinary{decErr: ErrTestError},
			expectErr: Error,
		},
		{
			name:      "unmarshal failure - ErrUnmarshalBinary",
			data:      []byte{0x01, 0x01, 0x00},
			recv:      &testBinary{decErr: ErrTestError},
			expectErr: ErrUnmarshalBinary,
		},
		{
			name:      "unmarshal failure - wrapped error",
			data:      []byte{0x01, 0x01, 0x00},
			recv:      &testBinary{decErr: ErrTestError},
			expectErr: ErrTestError,
		},
		{
			name:      "not enough bytes - Error",
			data:      []byte{0x01},
			recv:      ref(0),
			expectErr: Error,
		},
		{
			name:      "not enough bytes - io.ErrUnexpectedEOF",
			data:      []byte{0x01},
			recv:      ref(0),
			expectErr: io.ErrUnexpectedEOF,
		},
		{
			name:      "not enough bytes - ErrMalformedData",
			data:      []byte{0x01},
			recv:      ref(0),
			expectErr: ErrMalformedData,
		},
		{
			name:      "incorrect bytes - Error",
			data:      []byte{0x02},
			recv:      ref(true),
			expectErr: Error,
		},
		{
			name:      "incorrect bytes - ErrMalformedData",
			data:      []byte{0x02},
			recv:      ref(true),
			expectErr: ErrMalformedData,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			_, actual := Dec(tc.data, tc.recv)

			assert.ErrorIs(actual, tc.expectErr)
		})
	}
}

func Test_EncAndDec_NontrivialStructure(t *testing.T) {
	assert := assert.New(t)

	original := testNontrivial{
		ptr: ref(208),
		goodNums: map[int]bool{
			5: true,
			6: false,
		},
		actions: []**uint{
			ref(ref(uint(22))),
			ref(ref(uint(208))),
		},
		arr:  [3]int{1, 7, 20},
		frac: 0.8,
		jobs: nil,
		friend: &testNontrivial{
			ptr: nil,
			goodNums: map[int]bool{
				600: true,
				612: false,
				420: true,
				15:  true,
			},
			actions: nil,
			jobs:    []string{},
			comp:    8.25 + 1.0i,
			friend: &testNontrivial{
				ptr:      ref(413),
				goodNums: nil,
				actions: []**uint{
					ref(ref(uint(8))),
					ref(ref(uint(88))),
					ref(ref(uint(8888))),
					ref(ref(uint(88888888))),
				},
				jobs:   []string{"SECURE", "CONTAIN", "PROTECT"},
				friend: nil,
			},
		},
	}

	// we should be able to *encode* it
	data, err := Enc(original)
	if !assert.NoError(err) {
		return
	}

	// and then, we should be able to get the original back without error
	var rebuilt testNontrivial
	_, err = Dec(data, &rebuilt)
	if !assert.NoError(err) {
		return
	}

	// first check that there are at least as many friends as the original, 3.
	// (the first one is a given, we need to check n - 1 levels above that)
	if !assert.NotNil(rebuilt.friend) {
		return
	}
	if !assert.NotNil(rebuilt.friend.friend) {
		return
	}

	// okay, check each nontrivial from deepest level to highest so that error
	// messages can be well defined
	if !assert.Equal(original.friend.friend, rebuilt.friend.friend, "mismatch of rebuilt struct at level 3") {
		return
	}
	if !assert.Equal(original.friend, rebuilt.friend, "mismatch of rebuilt struct at level 2") {
		return
	}
	assert.Equal(original, rebuilt, "mismatch of rebuilt struct at level 1")
}

func nilRef[E any]() *E {
	var ref *E
	return ref
}

func ref[E any](v E) *E {
	return &v
}

// testText is a small struct that implements TextMarshaler and TextUnmarshaler.
// It has three fields, that it lays out as such in encoding: "value", a uint16,
// followed by "enabled", a bool, followed by "name", a string. Each field is
// separated with delimiter character ',', and the string field comes last to
// avoid needing to escape it within.
type testText struct {
	enabled bool
	name    string
	value   uint16
}

func (ttv testText) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d,%t,%s", ttv.value, ttv.enabled, ttv.name)), nil
}

func (ttv *testText) UnmarshalText(data []byte) error {
	str := string(data)

	var value int64
	var enabled bool
	var name string
	var err error

	parts := strings.SplitN(str, ",", 3)
	if len(parts) != 3 {
		return fmt.Errorf("text data does not contain 3 comma-separated fields")
	}

	value, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("value: %w", err)
	}

	enabled, err = strconv.ParseBool(parts[1])
	if err != nil {
		return fmt.Errorf("enabled: %w", err)
	}

	name = parts[2]

	ttv.value = uint16(value)
	ttv.enabled = enabled
	ttv.name = name

	return nil
}

// testBinary is a small struct that implements BinaryMarshaler and
// BinaryUnmarshaler. It has two fields that it lays out as such in encoding:
// "data", a string, followed by "number", an int32. decErr is an error string
// to return from UnmarshalBinary instead of decoding, if left nil no error is
// returned; same with encErr but for encoding. Neither err field is encoded. If
// encOverride is set, the byte slice pointed to will be returned from
// MarshalBinary instead of normal encoding. encOverride is therefore encoded,
// in a sense, but not directly. If both encOverride and encErr are set, encErr
// takes precedence.
//
// encPanic and decPanic are the same thing but for panics. They take precedence
// over all others.
type testBinary struct {
	// encoded fields:
	number int32
	data   string

	// non-encoded fields:
	decErr      error
	encErr      error
	encOverride *[]byte
}

func (tbv testBinary) MarshalBinary() ([]byte, error) {
	if tbv.encErr != nil {
		return nil, tbv.encErr
	}
	if tbv.encOverride != nil {
		return *tbv.encOverride, nil
	}
	var b []byte
	b = append(b, MustEnc(tbv.data)...)
	b = append(b, MustEnc(tbv.number)...)
	return b, nil
}

func (tbv *testBinary) UnmarshalBinary(data []byte) error {
	if tbv.decErr != nil {
		return tbv.decErr
	}
	var n int
	var err error

	offset := 0
	n, err = Dec(data[offset:], &tbv.data)
	if err != nil {
		return errorDecf(offset, "data: %v", err)
	}
	offset += n

	_, err = Dec(data[offset:], &tbv.number)
	if err != nil {
		return errorDecf(offset, "number: %v", err)
	}

	return nil
}

type testNontrivial struct {
	ptr      *int
	friend   *testNontrivial
	goodNums map[int]bool
	actions  []**uint
	jobs     []string
	arr      [3]int
	frac     float64
	comp     complex128
}

func (tn testNontrivial) MarshalBinary() ([]byte, error) {
	var enc []byte

	enc = append(enc, MustEnc(tn.ptr)...)
	enc = append(enc, MustEnc(tn.goodNums)...)
	enc = append(enc, MustEnc(tn.actions)...)
	enc = append(enc, MustEnc(tn.jobs)...)
	enc = append(enc, MustEnc(tn.comp)...)
	enc = append(enc, MustEnc(tn.frac)...)
	enc = append(enc, MustEnc(tn.arr)...)
	enc = append(enc, MustEnc(tn.friend)...)

	return enc, nil
}

func (tn *testNontrivial) UnmarshalBinary(data []byte) error {
	var err error
	var n int
	var offset int

	newNontriv := testNontrivial{}

	n, err = Dec(data[offset:], &newNontriv.ptr)
	if err != nil {
		return Wrapf(offset, "ptr: %s", err)
	}
	offset += n

	n, err = Dec(data[offset:], &newNontriv.goodNums)
	if err != nil {
		return Wrapf(offset, "goodNums: %s", err)
	}
	offset += n

	n, err = Dec(data[offset:], &newNontriv.actions)
	if err != nil {
		return Wrapf(offset, "actions: %s", err)
	}
	offset += n

	n, err = Dec(data[offset:], &newNontriv.jobs)
	if err != nil {
		return Wrapf(offset, "jobs: %s", err)
	}
	offset += n

	n, err = Dec(data[offset:], &newNontriv.comp)
	if err != nil {
		return Wrapf(offset, "comp: %s", err)
	}
	offset += n

	n, err = Dec(data[offset:], &newNontriv.frac)
	if err != nil {
		return Wrapf(offset, "frac: %s", err)
	}
	offset += n

	n, err = Dec(data[offset:], &newNontriv.arr)
	if err != nil {
		return Wrapf(offset, "arr: %s", err)
	}
	offset += n

	_, err = Dec(data[offset:], &newNontriv.friend)
	if err != nil {
		return Wrapf(offset, "friend: %s", err)
	}

	*tn = newNontriv
	return nil
}
