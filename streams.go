package rezi

import (
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

		streamReader.src = zReader
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
