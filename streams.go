package rezi

import (
	"bufio"
	"compress/zlib"
	"io"
)

// Format is a specification of a binary data format used by REZI. It specifies
// how data should be laid out and contains any options needed to do so.
//
// A nil or empty Format can be passed to functions which use it, and will be
// interpreted as a version 1 Format with no compression.
type Format struct {
	// Version is the version of the Format used. At this time only data format
	// V1 exists.
	//
	// As a special case, a Version value of 0 is interpreted as data format V1;
	// all other values are interpreted as that exact data format version.
	Version int

	// Compression is whether compression is enabled.
	Compression bool
}

// Reader is an io.ReadCloser that reads from REZI data streams. A Reader may be
// opened in compression mode or normal mode; compression mode can only read
// streams written by a [Writer] in compression mode.
//
// The zero-value is a Reader ready to read REZI data streams in the default
// V1 data format.
type Reader struct {
	f         Format
	src       io.Reader
	srcCloser func() error // does any closing of src, if needed

	// offset of decoded bytes into the stream we are. used for error reporting.
	offset int
}

// Format returns the Format that r interprets data as.
func (r *Reader) Format() Format {
	return r.f
}

// NewReader creates a new Reader ready to read data from r. If Compression is
// enabled in the supplied Format, it will interpret compressed data returned
// from r.
//
// If f is nil or points to the zero-value of Format, the default format of V1
// with compression disabled is selected, compatible for reading all written
// data that did not specify a Format (including in older releases of REZI).
// This function will make a copy of the Format pointed to; changes to it from
// outside this function will not be reflected in the returned Reader.
//
// This function returns a non-nil error only in cases where compression is
// selected via the format and an error occurs when opening a zlib reader on r.
func NewReader(r io.Reader, f *Format) (*Reader, error) {
	// prep format, check args
	if f == nil {
		f = &Format{}
	}
	usedFormat := *f
	if usedFormat.Version == 0 {
		usedFormat.Version = 1
	}

	if r == nil {
		panic("NewReader called on nil io.Reader")
	}

	streamReader := &Reader{
		f: usedFormat,
	}

	if f.Compression {
		// if it is compressed, open a zlib reader on the stream.
		zReader, err := zlib.NewReader(r)
		if err != nil {
			return nil, err
		}

		streamReader.src = bufio.NewReader(zReader)
		streamReader.srcCloser = zReader.Close
	} else {
		streamReader.src = r
		streamReader.srcCloser = func() error { return nil }
	}

	return streamReader, nil
}

// Close frees any resources needed from opening the Reader.
func (r *Reader) Close() error {
	return r.srcCloser()
}

// TODO: need Dec for reader, and Read, which should read REZI-encoded bytes but
// in an on-going basis in case more is given. Also, need to create a Writer.

// Dec decodes REZI-encoded bytes in r at the current position into the supplied
// value v, then advances the data stream past those bytes.
//
// Parameter v must be a pointer to a type supported by REZI.
func (r *Reader) Dec(v interface{}) (err error) {
	// job is to, based on what we are given, read the number of bytes we need
	// to read.
	defer func() {
		if r := recover(); r != nil {
			err = errorf("%v", r)
		}
	}()

	info, err := canDecode(v)
	if err != nil {
		return err
	}

	// data len possibilities:
	// - it is an encoded nil. the amount of data read will be 1+ext bytes+len of
	// next int (only if indir bit set)
	//   NEED TO READ: 1 byte for detect of INFO props, any more ext bytes to get their count, then reg INT INFO after that+ext bytes.
	//
	// - it is a bool, no indirection. the amount of data read will be 1 byte.
	//   NEED TO READ: 1 byte for detection of INFO props + any more ext bytes needed.
	//
	// - it is an int, no indirection. the amount of data read will be given in LLLL nibble of INFO.
	//   NEED TO READ: 1 byte for detection of INFO props&len + any more ext bytes needed.
	//
	// - it is a string, no indirection. the amount of data read will be all int bytes OR that many utf-8 chars.
	//   NEED TO READ: 1 byte for detect of INFO props, any ext bytes, then, if extended mode, 0-9 bytes to read int w byte len.
	//		if not extended mode: 1 byte for detect of INFO props, any ext bytes, then, all bytes until UTF-8 chars read.
	//
	// - it is a binary, no indirection. the amount of data read will be all int bytes
	//	 NEED TO READ: 1 byte for INFO, any ext, then, LLLL bytes to read int len.
	//
	// - it is a map, no indirection. the amount of data read will be all int bytes
	//	 NEED TO READ: 1 byte for INFO, any ext, then, LLLL bytes to read int len.
	//
	// - it is a slice, no indirection. the amount of data read will be all int bytes
	//	 NEED TO READ: 1 byte for INFO, any ext, then, LLLL bytes to read int len.
	//
	// ALGO:
	// 1. send any bytes read to buf for later re-read of core DEC routine.
	// 2. read byte. while ext bit is set on last byte read, read byte. if it specifies
	// 5. if header bytes are NIL with multiple indirs, read an int header byte. call to get rest of bytes, then pass to Dec.
	// 3. elif only one byte was read, and the passed in type is bool, we are done. pass off to Dec.
	// 4. elif header bytes indicate explicit int byte count, read an int. call to get reast of byte, then pass to Dec.
	// 6. else call to get LLLL bytes, pass all bytes to int dec, this is rest of bytes. call to rest, then pass to Dec.
	// 7. special case for v0 strings. we literally need to decode right there, it sucks.
	//

	// bytes to be sent to rezi.Dec.
	//var decBuf []byte

	// read a byte until we have a complete reader.

	if info.Main == mtBool {

	}

	return nil
}
