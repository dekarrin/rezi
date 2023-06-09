package rezi

import "encoding"

func valueThatUnmarshalsWith(byteConsumer func([]byte) error) encoding.BinaryUnmarshaler {
	return marshaledBytesConsumer{fn: byteConsumer}
}

func valueThatMarshalsWith(byteProducer func() []byte) encoding.BinaryMarshaler {
	return marshaledBytesProducer{fn: byteProducer}
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
