package rezi

import (
	"compress/zlib"
	"encoding"
	"fmt"
	"io"
)

// Reader reads data from REZI-encoded bytes.
//
// The zero value of Reader is not valid. Call NewReader to get an instance
// ready for use.
type Reader struct {
	r io.Reader

	// If compression is enabled, z will be the same as r but stored as a
	// ReadCloser().
	z io.ReadCloser
}

// Writer writes data as REZI-encoded bytes.
//
// The zero value of Writer is not valid. Call NewWriter to get an instance
// ready for use.
type Writer struct {
	w io.Writer

	// If compression is enabled, z will be the same as w but stored as a
	// WriteCloser().
	z io.WriteCloser
}

// NewReader creates a new Reader that reads compressed, REZI-encoded bytes from
// r. The passed-in reader must return compressed REZI bytes when Read() is
// called on it; to get a rezi.Reader that reads uncompressed bytes, call
// NewRawReader instead.
//
// The returned error will be non-nil if there is a problem reading the
// compressed data headers from the reader.
//
// It is the caller's responsibility to call Close() on the returned Reader when
// reading is complete.
func NewReader(r io.Reader) (*Reader, error) {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("data format error: %w", err)
	}

	return &Reader{
		r: zr,
		z: zr,
	}, nil
}

// NewRawReader creates a new Reader that reads REZI-encoded bytes from r. The
// passed-in reader r must return uncompressed REZI-encoded bytes when Read() is
// called on it.
//
// It is the caller's responsibility to call Close() on the returned Reader when
// reading is complete.
func NewRawReader(r io.Reader) *Reader {
	return &Reader{
		r: r,
	}
}

// Close closes the Reader. It does not close the underlying reader passed to it
// during construction.
func (r *Reader) Close() error {
	if r.z != nil {
		return r.z.Close()
	}

	// nothing else to do
	return nil
}

// Read reads a series of REZI-encoded bytes from the stream and places them
// into b. It is assumed that the bytes are encoded the REZI representation of a
// slice of bytes, as is returned by a call to Enc with a []byte or []uint8.
//
// If the REZI data at the current position in the reader cannot be interpreted
// as a []byte or []uint8, the returned error will be ErrInvalidType. If the
// REZI data can be interpreted as a []byte or []uint8 of the same size as b,
// the returned error will be io.ShortBuffer or io.UnexpectedEOF with the
// available bytes written to b and the count of read bytes returned as n.
func (r *Reader) Read(b []byte) (n int, err error) {
	return 0, nil
}

// Dec reads REZI-encoded bytes from the stream and interprets it as the type of
// v, assigning the final result to v. The argument v must be a pointer to the
// data to set, or be an implementor of encoding.BinaryUnmarshaler.
//
// If the REZI data at the current position cannot be interpreted as the type of
// v, the returned error will be ErrInvalidType. If the REZI data runs out
// before it is expected to, io.UnexpectedEOF is returned.
func (r *Reader) Dec(v interface{}) error {
	return nil
}

// NewWriter creates a new Writer that writes compressed REZI-encoded bytes to
// w.
//
// It is the caller's responsibility to call Close() on the returned Writer when
// writing is complete.
func NewWriter(w io.Writer) *Writer {
	zw := zlib.NewWriter(w)
	return &Writer{
		w: zw,
		z: zw,
	}
}

// NewRawWriter creates a new Writer that writes uncompressed REZI-encoded bytes
// to w.
//
// It is the caller's responsibility to call Close() on the returned Writer when
// writing is complete.
func NewRawWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

func (w *Writer) Close() error {
	if w.z != nil {
		return w.z.Close()
	}

	// nothing else to do
	return nil
}

// Decoder decodes the primitive types bool, int, and string, as well as a type
// that is specified by its type parameter (usually an interface of some
// XMarshaler type, such as BinaryUnmarshaler).
//
// Deprecated: The Decoder didn't really ever work.
type Decoder[E any] interface {

	// DecodeBool decodes a bool value at the current position within the buffer
	// of the Decoder and advances the current position past the read bytes.
	DecodeBool() (bool, error)

	// DecodeInt decodes an int value at the current position within the buffer
	// of the Decoder and advances the current position past the read bytes.
	DecodeInt() (int, error)

	// DecodeString decodes a string value at the current position within the
	// buffer of the Decoder and advances the current position past the read
	// bytes.
	DecodeString() (string, error)

	// Decode decodes a value at the current position within the buffer of the
	// Decoder and advances the current position past the read bytes. Unlike the
	// other functions, instead of returning the value this one will set the
	// value of the given item.
	Decode(o E) error
}

// simpleBinaryEncoder encodes values as binary. Create with NewBinaryDecoder,
// don't use directly.
type simpleBinaryDecoder struct {
	b   []byte
	cur int
}

func (sbe *simpleBinaryDecoder) DecodeBool() (bool, error) {
	val, n, err := decBool(sbe.b[sbe.cur:])
	if err != nil {
		return val, err
	}
	sbe.cur += n
	return val, nil
}

func (sbe *simpleBinaryDecoder) DecodeInt() (int, error) {
	val, n, err := decInt[int](sbe.b[sbe.cur:])
	if err != nil {
		return val, err
	}
	sbe.cur += n
	return val, nil
}

func (sbe *simpleBinaryDecoder) DecodeString() (string, error) {
	val, n, err := decString(sbe.b[sbe.cur:])
	if err != nil {
		return val, err
	}
	sbe.cur += n
	return val, nil
}

func (sbe *simpleBinaryDecoder) Decode(o encoding.BinaryUnmarshaler) error {
	n, err := decBinary(sbe.b[sbe.cur:], o)
	if err != nil {
		return err
	}
	sbe.cur += n
	return nil
}
