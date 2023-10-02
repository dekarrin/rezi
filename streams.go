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
	// 3. if header bytes are NIL with multiple indirs, read an int header byte. call to get rest of bytes, then pass to Dec.
	// 4. elif only one byte was read, and the passed in type is bool, we are done. pass off to Dec.
	// 5. elif header bytes indicate explicit int byte count, read an int. call to get reast of byte, then pass to Dec.
	// 6. else call to get LLLL bytes, pass all bytes to int dec, this is rest of bytes. call to rest, then pass to Dec.
	// 7. special case for v0 strings. we literally need to decode right there, it sucks.
	//

	// bytes to be sent to rezi.Dec.
	var decBuf []byte

	// before actually trying that header, let's krill a quick use-case! If the
	// input is a bool, and nil is not set, then it will be exactly one byte so
	// load it in.
	//

	// 1. load first byte. are we a bool and is that byte zeroed in the high
	// nibbles? if so, it is a raw bool and we can pass to decode.
	// 2. if not, we need to load the rest of the header with that byte starting.
	//
	// then, continue algo
	if info.Main == mtBool {
		// it count be a ptr, if so it will be a count header. if not, it will
		// be
	}

	hdrBytes, err := r.loadHeaderBytes()
	if err != nil {
		return errorDecf(0, "examine bytes for v1 count header: %s", err)
	}

	return nil
}

// never returns io.EOF; if at end of stream, it just returns nil error, as long
// as it loaded everyfin in needs to. If at end before it is loaded,
// io.UnexpectedEOF is returned as error.
//
// this function will not cause offset to be incremented; caller should do so
// by []byte amount when it is to be advanced.
func (r *Reader) loadHeaderBytes() ([]byte, error) {
	// read a byte until we have a complete header.
	var hdrBytes []byte

	var totalRead int

	// read in initial info byte first
	extBuf := make([]byte, 1)

	for {
		n, err := r.src.Read(extBuf)
		if n <= 0 {
			// if error is set
			if err != nil {
				if err == io.EOF {
					// in fact, this is unexpected
					return hdrBytes, io.ErrUnexpectedEOF
				}
				return hdrBytes, err
			}
			return hdrBytes, io.ErrUnexpectedEOF
		}
		totalRead++
		hdrBytes = append(hdrBytes, extBuf[0])
		if extBuf[1]&infoBitsExt == 0 {
			// no more count bytes
			break
		}
	}

	// okay, got our bytes, now check if we have a nil indir level encoding we
	// need to grab
	if hdrBytes[0]&infoBitsIndir != 0 {
		// load info bytes. we should get nothing special here, just a normal int.
		intHdr, err := r.loadHeaderBytes()
		if len(intHdr) > 0 {
			hdrBytes = append(hdrBytes, intHdr...)
		}
		if err != nil && err != io.EOF {
			// genuine error
			return hdrBytes, err
		}

		// okay, now peek at the int byte to see if we need to load

		// this had better be a positive int that is not itself nil or indirected.
		if intHdr[0]&infoBitsSign != 0 {
			return hdrBytes, errorDecf(totalRead, "indirection count in nil-byte header is negative", ErrMalformedData)
		}
		if intHdr[0]&infoBitsNil != 0 {
			return hdrBytes, errorDecf(totalRead, "indirection count in nil-byte header is itself nil", ErrMalformedData)
		}
		if intHdr[0]&infoBitsIndir != 0 {
			return hdrBytes, errorDecf(totalRead, "indirection count in nil-byte header marks itself as also being an indirected nil", ErrMalformedData)
		}

		totalRead += len(intHdr)
		// ext bit doesn't actually matter, grab the LLLL bytes.
		intByteCount := intHdr[0] & infoBitsLen

		if intByteCount > 0 {
			// read in rest of the int
			indirCountBuf := make([]byte, intByteCount)
			n, err := r.src.Read(indirCountBuf)
			if n <= 0 {
				// if error is set
				if err != nil {
					if err == io.EOF {
						// in fact, this is unexpected. there should be at least one.
						return hdrBytes, io.ErrUnexpectedEOF
					}
					return hdrBytes, err
				}
				return hdrBytes, io.ErrUnexpectedEOF
			}
			hdrBytes = append(hdrBytes, indirCountBuf...)
			totalRead += len(intHdr)
		}
	}

	// explicitly do not load the "byte-based count" bit.
	return hdrBytes, nil
}

// readSrc reads (but does not interpret) the given number of bytes from the
// underlying data stream. offset is not incremented, callers must do this. the
// stream is called until that many bytes are present, or an error occurs.
//
// returned slice will be <= count in size. if it ever it is < count, it
// indicates that the underlying stream returned io.EOF.
//
// returns io.EOF when at end of stream.
func (r *Reader) readSrc(count int) ([]byte, error) {
	if count < 1 {
		return nil, nil
	}

	var readBytes []byte

	buf := make([]byte, count)
	n, err := r.src.Read(buf)
	if n > 0 {
		readBytes = buf[:n]
	}

	if err != nil && err == io.EOF {
		err = nil
	}

	return readBytes, err
}
