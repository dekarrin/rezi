package rezi

import (
	"bufio"
	"compress/zlib"
	"errors"
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

	// readBuf keeps a buffer of read bytes obtained during a call to Read()
	// for 'normal io.Reader' use of Reader. It holds any loaded decoded bytes
	// that were not used to fill the slice passed in by the caller of Read.
	readBuf []byte
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

// Offset returns the current number of bytes that the Reader has interpreted as
// REZI encoded bytes from the stream. Note that if compression is enabled, this
// refers to the number of uncompressed data bytes interpreted, regardless of
// how many actual bytes are read from the underlying reader provided to r at
// construction.
func (r *Reader) Offset() int {
	return r.offset
}

// Read reads up to len(p) bytes from one or more REZI-encoded byte slices
// present in sequence in the data stream and places them into p. Returns the
// number of valid bytes read into p.
//
// Read requires the underlying data stream at the current position to consist
// only of one or more REZI-encoded byte slices. Attempting to read more bytes
// than the current byte slice has will cause more slices to be read from the
// underlying reader until either p can be filled or the end of the stream is
// reached. If the last slice read in this fashion is not completely used by p,
// i.e. if p does not have enough room to hold the complete slice, then the
// remaining bytes decoded are buffered, and the next call to Read will begin
// filling its p with those bytes before reading another slice from the
// underlying reader.
//
// Note that the number of bytes read into p (returned as n) is almost certainly
// less than the total number of bytes read from the underlying data stream; to
// capture this, call Offset before and after calling Read and check the
// difference between them.
//
// Returns io.EOF only in non-error circumstances. It is possible for n > 0 when
// err is non-nil and even when err is not io.EOF. All errors besides io.EOF
// will be wrapped in a special error type from the rezi package; use errors.Is
// to compare the returned error.
//
// If len(p) is greater than the total number of bytes available, but every byte
// that *is* available is organized as valid REZI-encoded byte slices, err will
// be io.EOF and n will be the number of bytes that could be read.
func (r *Reader) Read(p []byte) (n int, err error) {
	// if p is empty, great, we are done
	if len(p) < 1 {
		return 0, nil
	}

	var cur int

	// before anyfin else, do we have bytes left over from the last call to
	// Read? use those first, glub!
	if len(r.readBuf) > 0 {
		needed := len(p) - cur
		if needed < len(r.readBuf) {
			copy(p[cur:], r.readBuf[:needed])
			cur += needed

			// reset cache for next call
			r.readBuf = r.readBuf[needed:]
		} else {
			copy(p[cur:], r.readBuf)
			cur += len(r.readBuf)

			// invalidate cache, we just used it up
			r.readBuf = nil
		}
	}

	for cur < len(p) {
		var loadedBytes []byte

		// need to capture this to check if any bytes actually read
		oldOffset := r.offset
		err := r.Dec(&loadedBytes)
		if err != nil {
			// okay, if we got UnexpectedEOF due to no bytes at all being
			// present, that is okay, actually. we just hit the end of the
			// stream.
			if errors.Is(err, io.ErrUnexpectedEOF) && r.offset == oldOffset {
				return cur, io.EOF
			}

			return cur, err
		}

		// how many of the loaded bytes do we need?
		needed := len(p) - cur

		if needed < len(loadedBytes) {
			copy(p[cur:], loadedBytes[:needed])
			cur += needed

			// cache the rest for next call to Read:
			r.readBuf = loadedBytes[needed:]
		} else {
			copy(p[cur:], loadedBytes)
			cur += len(loadedBytes)
		}
	}

	return cur, nil
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

	datumBytes, err := r.loadDecodeableBytes(info)
	if err != nil && err != io.EOF {
		r.offset += len(datumBytes)
		return err
	}

	n, err := Dec(datumBytes, v)
	if err != nil {
		err = errorDecf(r.offset, "%s", err)
		r.offset += len(datumBytes)
		return err
	}
	if n != len(datumBytes) {
		err = errorDecf(r.offset, "expected decoded data at offset to consume byte len of %d but actual consumed is %d", len(datumBytes), n)
		r.offset += len(datumBytes)
		return err
	}
	r.offset += len(datumBytes)

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

	// special case: if it's a non-nil v0 string, we need to immediately decode
	// it as we go. We can only tell this once we have the header and can
	// determine that it is not, in fact, a nil.
	//
	// This suuuuuuuucks! V2 when?!
	if info.Main == mtString && !hdr.ByteLength {
		loaded, err := r.loadV0StringBytes(hdr, hdrBytes)
		if err != nil && err != io.EOF {
			return loaded, errorDecf(0, "load v0 string: %s", err)
		}
		return loaded, err
	}

	decodable := make([]byte, len(hdrBytes))
	copy(decodable, hdrBytes)

	var remByteCount int

	// okay, if the info header says we need to load the rest of an int for a
	// byte count, do it now
	// Troll-jegus! Isn't the V1 format awful? ::::(
	if hdr.ByteLength {
		// read count int
		buf, err := r.loadCountIntBytes(hdrBytes)
		lastErr = err
		if len(buf) > len(hdrBytes) {
			decodable = append(decodable, buf[len(hdrBytes):]...)
		}
		if err != nil && err != io.EOF {
			return decodable, errorDecf(totalRead, "%s", err)
		}

		count, n, err := decInt[int](buf)
		// do not preserve this error, it will never be io.EOF.
		if err != nil {
			return decodable, errorDecf(totalRead, "header byte-count int: %s", err)
		}
		if n != len(buf) {
			return decodable, errorDecf(totalRead, "header byte-count int: actual decoded len < read len").wrap(err)
		}
		remByteCount = count
		totalRead += (n - len(hdrBytes))
	} else if info.Main != mtIntegral {
		// for non-ints, we need to load the rest of the integer ourselves, then
		// remByteCount is the value of THAT
		intBytes, err := r.loadBytes(hdr.Length)
		lastErr = err
		if len(intBytes) > 0 {
			decodable = append(decodable, intBytes...)
		}
		if err != nil && err != io.EOF {
			return decodable, errorDecf(totalRead, "%s", err)
		}

		// okay, we have complete header and int bytes, decode to int type
		count, n, err := decInt[int](decodable)
		// do not preserve this error, it will never be io.EOF.
		if err != nil {
			return decodable, errorDecf(0, "count header: %s", err)
		}
		if n != len(decodable) {
			return decodable, errorDecf(0, "count header: actual decoded len < read len").wrap(err)
		}
		remByteCount = count
		totalRead += n
	} else {
		// if it is an int, rem bytes is hdr.Length
		remByteCount = hdr.Length
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
// up a V0 string. it will include hdrBytes in the returned output to make a
// complete v0 string.
func (r *Reader) loadV0StringBytes(hdr countHeader, hdrBytes []byte) ([]byte, error) {
	// okay, we have a header but we need to load in the actual info int.
	var lastErr error // to propagate io.EOFs.
	loaded := make([]byte, len(hdrBytes))
	copy(loaded, hdrBytes)

	// [ IHDR ] [ IBYTES ] [ CHAR BYTES ]
	// ^^^^^^^^ - we have THIS part so far

	// we need to load the rest of the integer (IBYTES) ourselves to get a char
	// count
	intBytes, err := r.loadBytes(hdr.Length)
	lastErr = err
	if len(intBytes) > 0 {
		loaded = append(loaded, intBytes...)
	}
	if err != nil && err != io.EOF {
		return intBytes, err
	}

	// okay, we have complete header and int bytes, decode to int type
	runeCount, n, err := decInt[int](loaded)
	// do not preserve this error, it will never be io.EOF.
	if err != nil {
		return loaded, errorDecf(0, "count header: %s", err)
	}
	if n != len(loaded) {
		return loaded, errorDecf(0, "count header: actual decoded len < read len").wrap(err)
	}

	totalRead := len(loaded)

	// we now have a rune count. begin loading bytes until we have hit it
	for loadedRunes := 0; loadedRunes < runeCount; loadedRunes++ {
		// first, load in byte 1. this will tell us if we need more
		firstByteBuf, err := r.loadBytes(1)
		lastErr = err
		if len(firstByteBuf) > 0 {
			loaded = append(loaded, firstByteBuf...)
		}
		if err != nil && err != io.EOF {
			return loaded, errorDecf(totalRead, "load rune byte 1: %s", err)
		}
		totalRead += len(firstByteBuf)

		firstRuneByte := firstByteBuf[0]

		var additionalBytes int

		// check UTF-8 len bits to see if we have more to load
		if (firstRuneByte&0xc0 == 0xc0) && (firstRuneByte&0x20 == 0) { // matches 0b110xxxxx, two bytes
			additionalBytes = 1
		} else if (firstRuneByte&0xe0 == 0xe0) && (firstRuneByte&0x10 == 0) { // matches 0b1110xxxx, three bytes
			additionalBytes = 2
		} else if (firstRuneByte&0xf0 == 0xf0) && (firstRuneByte&0x08 == 0) { // matches 0b11110xxx, four bytes
			additionalBytes = 3
		}

		if additionalBytes > 0 {
			nextBytes, err := r.loadBytes(additionalBytes)
			lastErr = err
			if len(nextBytes) > 0 {
				loaded = append(loaded, nextBytes...)
			}
			if err != nil && err != io.EOF {
				return loaded, errorDecf(totalRead, "load next rune byte(s): %s", err)
			}
			totalRead++
		}

		// load of next rune complete
	}

	return loaded, lastErr
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
		indirBytes, err := r.loadCountIntBytes(nil)
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

// loadCountIntBytes loads an int that is known to be unsigned and a non-nil
// value.
// preloadedHeader can be nil, in which case this function will load a header
// for it as well.
func (r *Reader) loadCountIntBytes(preloadedHeader []byte) ([]byte, error) {
	var loaded []byte
	var totalRead int

	// for io.EOF preservation
	var lastErr error

	var intHdr []byte
	if preloadedHeader != nil {
		intHdr = preloadedHeader
		loaded = append(loaded, preloadedHeader...)
	} else {
		var err error
		intHdr, err = r.loadHeaderBytes(nil)
		lastErr = err

		if len(intHdr) > 0 {
			loaded = append(loaded, intHdr...)
		}
		if err != nil && err != io.EOF {
			// genuine error
			return loaded, errorDecf(totalRead, "%s", err)
		}
	}

	totalRead += len(intHdr)

	// okay, now peek at the int byte to see if we need to load

	// this had better be a positive int that is not itself nil or indirected.
	if intHdr[0]&infoBitsSign != 0 {
		return loaded, errorDecf(totalRead, "count int header indicates negative").wrap(ErrMalformedData)
	}
	if intHdr[0]&infoBitsNil != 0 {
		return loaded, errorDecf(totalRead, "count int header is nil").wrap(ErrMalformedData)
	}
	if intHdr[0]&infoBitsIndir != 0 {
		return loaded, errorDecf(totalRead, "count int header marks itself as also being an indirected nil").wrap(ErrMalformedData)
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
