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
	//

	// The V2 format is going to be so much 8etter.

	// bytes to be sent to rezi.Dec.
	var decBuf []byte // TODO: replace this check with function that just returns
	// for flatter structure.
	var decReady bool

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

// loadDecodableBytes loads enough bytes for a complete full data item read from
// the underlying stream, ready to be interpreted by Dec. It does its best to
// interpret as few bytes as possible itself. Due to the nature of the V1 data
// format, many bytes sometimes must be interpreted (or in the worst-case, with
// V0 strings it cannot determine length without actually decoding the string).
//
// it is intended to be the top-level function which calls into other loadX
// methods for the V1 data format.
//
// it does not update r.offset. error may be set to io.EOF if the end of stream
// was reached upon successful full read of bytes; if it was not successful due
// to the end of stream, it will instead be a reziError that matches
// io.UnexpectedEOF.
//
// if error is returned, the returned bytes are all bytes that were successfully
// read, in order to properly update offset.
func (r *Reader) loadDecodeableBytes(info typeInfo) ([]byte, error) {
	var err error
	var hdrBytes []byte
	var totalRead int

	// for io.EOF preservation
	var lastErr error

	// before actually trying that header, let's krill a quick use-case! If the
	// input is a bool, and nil is not set, then it will be exactly one byte so
	// load it in.
	if info.Main == mtBool {
		firstByteBuf, err := r.loadBytes(1)
		lastErr = err
		if err != nil && err != io.EOF {
			return firstByteBuf, errorDecf(totalRead, "%s", err)
		}

		firstByte := firstByteBuf[0]
		if firstByte&0xf0 == 0 {
			// no high order bytes, this is ready to be dec'd as is.
			return firstByteBuf, lastErr
		} else {
			hdrBytes, err = r.loadHeaderBytes(&firstByte)
			lastErr = err
			if err != nil {
				return hdrBytes, errorDecf(totalRead, "%s", err)
			}
		}
	} else {
		// load the entire header in, all other types use an info byte.
		hdrBytes, err = r.loadHeaderBytes(nil)
		lastErr = err
		if err != nil {
			return hdrBytes, errorDecf(totalRead, "%s", err)
		}
	}
	totalRead += len(hdrBytes)

	// we have hdrBytes. know what? just make this easier and decode it
	// immediately.
	var hdr countHeader
	err = hdr.UnmarshalBinary(hdrBytes)
	// don't preserve this one; it is not defined to return io.EOF as a non-error.
	if err != nil {
		// keep this offset at 0 because it is propagating a decode err with its own offsets
		return hdrBytes, errorDecf(0, "pre-decode header: %s", err)
	}

	// indir count retrieval is handled by loadHeaderBytes. did we just pull up one? if so, we are done.
	if hdr.IsNil() {
		return hdrBytes, nil
	}

	decodable := make([]byte, len(hdrBytes))
	copy(decodable, hdrBytes)

	// special case: if it's a non-nil v0 string, we need to immediately decode
	// it as we go.
	//
	// This suuuuuuuucks! V2 when?!
	if info.Main == mtString && !hdr.ByteLength {
		r.loadV0Strin()
	}

	// normal circumstances, override due to ByteLength below
	remByteCount := hdr.Length

	// okay, if the info header says we need to load an int for a byte count, do
	// it now
	// Troll-jegus! Isn't the V1 format awful? ::::(
	if hdr.ByteLength {
		// read count int
		buf, err := r.loadCountIntBytes()
		lastErr = err
		if len(buf) > 0 {
			decodable = append(decodable, buf...)
		}
		if err != nil && err != io.EOF {
			return decodable, errorDecf(totalRead, "%s", err)
		}

		count, n, err := decInt[int](buf)
		// do not preserve this error, it will never be io.EOF.
		if err != nil {
			return decodable, errorDecf(totalRead, "%s", err)
		}
		if n != len(buf) {
			return decodable, errorDecf(totalRead, "header byte count int: actual decoded len < read len").wrap(err)
		}
		remByteCount = count
		totalRead += n
	}

	// well, at least NOW we know the exact remain bytes to grab, glub!
	if remByteCount > 0 {
		remBytes, err := r.loadBytes(remByteCount)
		lastErr = err
		if len(remBytes) > 0 {
			decodable = append(decodable, remBytes...)
		}
		if err != nil && err != io.EOF {
			return decodable, errorDecf(totalRead, "%s", err)
		}
	}

	// all bytes present, return
	return decodable, lastErr
}

// given a countHeader, loadV0StringBytes loads the rest of the bytes that make
// up a V0 string. it does this by manually attempting to decode.
func (r *Reader) loadV0StringBytes(hdr countHeader) ([]byte, error) {
	// okay, we have a header but we need to load in the actual info int.

	hdr.Length
}

// if EOF encountered at before all bytes are loaded, a reziError that matches
// io.UnexpectedEOF is returned as error. If it is encountered AS the header is
// read, io.EOF is returned.
//
// this function will not cause offset to be incremented; caller should do so
// by []byte amount when it is to be advanced.
//
// if withFirst is given and non nil, it will be assumed to be the first byte.
func (r *Reader) loadHeaderBytes(withFirst *byte) ([]byte, error) {
	// read a byte until we have a complete header.
	var hdrBytes []byte
	var totalRead int

	// for io.EOF preservation
	var lastErr error

	var extCheck byte
	if withFirst != nil {
		hdrBytes = append(hdrBytes, *withFirst)
		extCheck = hdrBytes[0]
	} else {
		extCheck = infoBitsExt
	}

	// read in initial info byte first

	for extCheck&infoBitsExt != 0 {
		extBuf, err := r.loadBytes(1)
		lastErr = err
		if len(extBuf) > 0 {
			hdrBytes = append(hdrBytes, extBuf[0])
			extCheck = extBuf[0]
		}
		if err != nil && err != io.EOF {
			return hdrBytes, errorDecf(totalRead, "%s", err)
		}
		totalRead++
	}

	// okay, got our bytes, now check if we have a nil indir level encoding we
	// need to grab
	if hdrBytes[0]&infoBitsIndir != 0 {
		// load info bytes. we should get nothing special here, just a normal int.
		indirBytes, err := r.loadCountIntBytes()
		lastErr = err
		if len(indirBytes) > 0 {
			hdrBytes = append(hdrBytes, indirBytes...)
		}
		if err != nil && err != io.EOF {
			return hdrBytes, errorDecf(totalRead, "%s", err)
		}
		totalRead += len(indirBytes)
	}

	// explicitly do not load the "byte-based count" bit.
	return hdrBytes, lastErr
}

// loadCountIntBytes loads specifically a header and int that is known to be
// unsigned and a non-nil value.
func (r *Reader) loadCountIntBytes() ([]byte, error) {
	var loaded []byte
	var totalRead int

	// for io.EOF preservation
	var lastErr error

	intHdr, err := r.loadHeaderBytes(nil)
	lastErr = err

	if len(intHdr) > 0 {
		totalRead += len(intHdr)
		loaded = append(loaded, intHdr...)
	}
	if err != nil && err != io.EOF {
		// genuine error
		return loaded, errorDecf(totalRead, "%s", err)
	}

	// okay, now peek at the int byte to see if we need to load

	// this had better be a positive int that is not itself nil or indirected.
	if intHdr[0]&infoBitsSign != 0 {
		return loaded, errorDecf(totalRead, "count int header indicates negative", ErrMalformedData)
	}
	if intHdr[0]&infoBitsNil != 0 {
		return loaded, errorDecf(totalRead, "count int header is nil", ErrMalformedData)
	}
	if intHdr[0]&infoBitsIndir != 0 {
		return loaded, errorDecf(totalRead, "count int header marks itself as also being an indirected nil", ErrMalformedData)
	}

	// ext bit doesn't actually matter, grab the LLLL bytes.
	intByteCount := intHdr[0] & infoBitsLen

	if intByteCount > 0 {
		// read in rest of the int
		iBytes, err := r.loadBytes(int(intByteCount))
		lastErr = err
		if len(iBytes) > 0 {
			loaded = append(loaded, iBytes...)
		}
		if err != nil && err != io.EOF {
			return loaded, errorDecf(totalRead, "%s", err)
		}
	}

	return loaded, lastErr
}

// loadBytes calls read on underlying reader until c bytes have been read or an
// error is encountered. if io.EOF is encountered before count bytes are read,
// it is converted to io.UnexpectedEOF. If it is encountered at count bytes, it
// is returned unchanged.
//
// the returned bytes will have as many bytes as WERE read from the stream even
// in the case of a non-nil error.
func (r *Reader) loadBytes(count int) ([]byte, error) {
	read := make([]byte, count)
	var curRead int
	var err error

	for curRead < count {
		buf := make([]byte, count-curRead)

		var n int
		n, err = r.src.Read(buf)
		if n > 0 {
			copy(read[curRead:], buf[:n])
			curRead += n
		}
		if err != nil {
			if err == io.EOF {
				// if we have not loaded enough bytes yet, this is Unexpected.
				if curRead < count {
					return read[:curRead], io.ErrUnexpectedEOF
				}
			} else {
				return read[:curRead], err
			}
		}
	}

	return read, err
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
