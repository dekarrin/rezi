package rezi

import (
	"encoding"
	"fmt"
)

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
	b = append(b, Enc(tbv.data)...)
	b = append(b, Enc(tbv.number)...)
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
