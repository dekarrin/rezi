package rezi

import (
	"encoding"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		friend: &testNontrivial{
			ptr: nil,
			goodNums: map[int]bool{
				600: true,
				612: false,
				420: true,
				15:  true,
			},
			actions: nil,
			friend: &testNontrivial{
				ptr:      ref(413),
				goodNums: nil,
				actions: []**uint{
					ref(ref(uint(8))),
					ref(ref(uint(88))),
					ref(ref(uint(8888))),
					ref(ref(uint(88888888))),
				},
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

func ref[E any](v E) *E {
	return &v
}

func valueThatUnmarshalsWith(byteConsumer func([]byte) error) encoding.BinaryUnmarshaler {
	return marshaledBytesConsumer{fn: byteConsumer}
}

func valueThatMarshalsWith(byteProducer func() []byte) encoding.BinaryMarshaler {
	return marshaledBytesProducer{fn: byteProducer}
}

type testBinary struct {
	number int32
	data   string
}

func (tbv testBinary) MarshalBinary() ([]byte, error) {
	var b []byte
	b = append(b, MustEnc(tbv.data)...)
	b = append(b, MustEnc(tbv.number)...)
	return b, nil
}

func (tbv *testBinary) UnmarshalBinary(data []byte) error {
	var n int
	var err error

	n, err = Dec(data, &tbv.data)
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	data = data[n:]

	_, err = Dec(data, &tbv.number)
	if err != nil {
		return fmt.Errorf("number: %w", err)
	}

	return nil
}

type marshaledBytesConsumer struct {
	fn func([]byte) error
}

func (mv marshaledBytesConsumer) UnmarshalBinary(b []byte) error {
	return mv.fn(b)
}

type marshaledBytesReceiver struct {
	received []byte
}

func (recv *marshaledBytesReceiver) UnmarshalBinary(b []byte) error {
	recv.received = make([]byte, len(b))
	copy(recv.received, b)
	return nil
}

type marshaledBytesProducer struct {
	fn func() []byte
}

// MarshalBinary converts mv into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (mv marshaledBytesProducer) MarshalBinary() ([]byte, error) {
	return mv.fn(), nil
}

type testNontrivial struct {
	ptr      *int
	friend   *testNontrivial
	goodNums map[int]bool
	actions  []**uint
}

func (tn testNontrivial) MarshalBinary() ([]byte, error) {
	var enc []byte

	enc = append(enc, MustEnc(tn.ptr)...)
	enc = append(enc, MustEnc(tn.goodNums)...)
	enc = append(enc, MustEnc(tn.actions)...)
	enc = append(enc, MustEnc(tn.friend)...)

	return enc, nil
}

func (tn *testNontrivial) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	newNontriv := testNontrivial{}

	n, err = Dec(data, &newNontriv.ptr)
	if err != nil {
		return fmt.Errorf("ptr: %w", err)
	}
	data = data[n:]

	n, err = Dec(data, &newNontriv.goodNums)
	if err != nil {
		return fmt.Errorf("goodNums: %w", err)
	}
	data = data[n:]

	n, err = Dec(data, &newNontriv.actions)
	if err != nil {
		return fmt.Errorf("actions: %w", err)
	}
	data = data[n:]

	_, err = Dec(data, &newNontriv.friend)
	if err != nil {
		return fmt.Errorf("friend: %w", err)
	}

	*tn = newNontriv
	return nil
}
