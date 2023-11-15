package rezi

import (
	"bytes"
	"compress/zlib"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_WriteRead_Cycle_Compressed(t *testing.T) {
	assert := assert.New(t)

	startValue := []byte{
		0x01, 0x02, 0xff, 0xff, 0x2f, 0xf1, 0x57, 0x1c, 0x0f, 0xf5,
	}

	// write data to bytes across multiple calls to Write:
	buf := &bytes.Buffer{}
	w, err := NewWriter(buf, &Format{Version: 1, Compression: true, CompressionLevel: zlib.BestCompression})
	if !assert.NoError(err, "error creating writer") {
		return
	}
	_, err = w.Write(startValue[:2])
	if !assert.NoError(err, "error writing value") {
		return
	}
	_, err = w.Write(startValue[2:])
	if !assert.NoError(err, "error writing value") {
		return
	}
	w.Flush()

	// read data from bytes across a single call to Read:
	rBuf := bytes.NewReader(buf.Bytes())
	r, err := NewReader(rBuf, &Format{Version: 1, Compression: true})
	if !assert.NoError(err, "error creating reader") {
		return
	}
	actual := make([]byte, len(startValue))
	_, err = r.Read(actual)
	if !assert.NoError(err, "error reading value") {
		return
	}

	assert.Equal(startValue, actual)
}

func Test_WriteRead_Cycle(t *testing.T) {
	assert := assert.New(t)

	startValue := []byte{
		0x01, 0x02, 0xff, 0xff, 0x2f, 0xf1, 0x57, 0x1c, 0x0f, 0xf5,
	}

	// write data to bytes across multiple calls to Write:
	buf := &bytes.Buffer{}
	w, err := NewWriter(buf, nil)
	if !assert.NoError(err, "error creating writer") {
		return
	}
	_, err = w.Write(startValue[:2])
	if !assert.NoError(err, "error writing value") {
		return
	}
	_, err = w.Write(startValue[2:])
	if !assert.NoError(err, "error writing value") {
		return
	}
	w.Flush()

	// read data from bytes across a single call to Read:
	rBuf := bytes.NewReader(buf.Bytes())
	r, err := NewReader(rBuf, nil)
	if !assert.NoError(err, "error creating reader") {
		return
	}
	actual := make([]byte, len(startValue))
	_, err = r.Read(actual)
	if !assert.NoError(err, "error reading value") {
		return
	}

	assert.Equal(startValue, actual)
}

func Test_EncDec_Cycle_Compressed(t *testing.T) {
	assert := assert.New(t)

	startValue := testBinary{data: "NEPETA", number: 413}

	// write data to bytes:
	buf := &bytes.Buffer{}
	w, err := NewWriter(buf, &Format{Version: 1, Compression: true})
	if !assert.NoError(err, "error creating writer") {
		return
	}
	err = w.Enc(startValue)
	if !assert.NoError(err, "error writing value") {
		return
	}
	w.Flush()

	// read data from bytes:
	rBuf := bytes.NewReader(buf.Bytes())
	r, err := NewReader(rBuf, &Format{Version: 1, Compression: true})
	if !assert.NoError(err, "error creating reader") {
		return
	}
	var actual testBinary
	err = r.Dec(&actual)
	if !assert.NoError(err, "error reading value") {
		return
	}

	assert.Equal(startValue, actual)
}

func Test_Writer_Enc(t *testing.T) {
	assert := assert.New(t)

	var expect []byte

	buf := &bytes.Buffer{}
	w, err := NewWriter(buf, nil)
	if !assert.NoError(err, "error creating writer") {
		return
	}

	strData := "NEPETA"
	floatData := 256.01220703125
	complexData := 1.0 + 8.25i
	intData := 413
	expect = []byte{
		0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41,
		0x04, 0xc0, 0x70, 0x00, 0x32,
		0x41, 0x80, 0x07, 0x02, 0x3f, 0xf0, 0x03, 0xc0, 0x20, 0x80,
		0x02, 0x01, 0x9d,
	}

	err = w.Enc(strData)
	if !assert.NoError(err, "error writing first time") {
		return
	}
	err = w.Enc(floatData)
	if !assert.NoError(err, "error writing second time") {
		return
	}
	err = w.Enc(complexData)
	if !assert.NoError(err, "error writing third time") {
		return
	}
	err = w.Enc(intData)
	if !assert.NoError(err, "error writing fourth time") {
		return
	}
	w.Flush()

	actual := buf.Bytes()
	assert.Equal(expect, actual)
}

func Test_Reader_Read_oneCall(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		p         []byte
		expect    []byte
		expectN   int
		expectErr bool
	}{
		{
			name:    "nil source bytes",
			input:   nil,
			p:       make([]byte, 4),
			expect:  []byte{0x00, 0x00, 0x00, 0x00},
			expectN: 0,
		},
		{
			name:    "empty source bytes",
			input:   []byte{},
			p:       make([]byte, 4),
			expect:  []byte{0x00, 0x00, 0x00, 0x00},
			expectN: 0,
		},
		{
			name:    "one single encoded empty slice",
			input:   []byte{0x00},
			p:       make([]byte, 4),
			expect:  []byte{0x00, 0x00, 0x00, 0x00},
			expectN: 0,
		},
		{
			name:    "one several encoded empty slices",
			input:   []byte{0x00, 0x00, 0x00},
			p:       make([]byte, 4),
			expect:  []byte{0x00, 0x00, 0x00, 0x00},
			expectN: 0,
		},
		{
			name:    "one non-empty encoded slice, len(slice) < len(p)",
			input:   []byte{0x01, 0x04, 0x01, 0x88, 0x01, 0x20},
			p:       make([]byte, 4),
			expect:  []byte{0x88, 0x20, 0x00, 0x00},
			expectN: 2,
		},
		{
			name:    "one non-empty encoded slice, len(slice) = len(p)",
			input:   []byte{0x01, 0x08, 0x01, 0x88, 0x01, 0x20, 0x01, 0xff, 0x01, 0x7f},
			p:       make([]byte, 4),
			expect:  []byte{0x88, 0x20, 0xff, 0x7f},
			expectN: 4,
		},
		{
			name:    "one non-empty encoded slice, len(slice) > len(p)",
			input:   []byte{0x01, 0x0a, 0x01, 0x88, 0x01, 0x20, 0x01, 0xff, 0x01, 0x7f, 0x01, 0x12},
			p:       make([]byte, 4),
			expect:  []byte{0x88, 0x20, 0xff, 0x7f},
			expectN: 4,
		},
		{
			name: "multiple non-empty encoded slices, len(slices) < len(p)",
			input: []byte{
				0x01, 0x04, 0x01, 0x88, 0x01, 0x20, // []byte{0x88, 0x20}
				0x01, 0x02, 0x01, 0x01, // []byte{0x01}
			},
			p:       make([]byte, 4),
			expect:  []byte{0x88, 0x20, 0x01, 0x00},
			expectN: 3,
		},
		{
			name: "multiple non-empty encoded slices, len(slices) = len(p)",
			input: []byte{
				0x01, 0x04, 0x01, 0x88, 0x01, 0x20, // []byte{0x88, 0x20}
				0x01, 0x03, 0x00, 0x01, 0x01, // []byte{0x00, 0x01}
			},
			p:       make([]byte, 4),
			expect:  []byte{0x88, 0x20, 0x00, 0x01},
			expectN: 4,
		},
		{
			name: "multiple non-empty encoded slices, len(slices) > len(p)",
			input: []byte{
				0x01, 0x04, 0x01, 0x88, 0x01, 0x20, // []byte{0x88, 0x20}
				0x01, 0x05, 0x00, 0x01, 0x01, 0x01, 0xae, // []byte{0x00, 0x01, 0xae}
			},
			p:       make([]byte, 4),
			expect:  []byte{0x88, 0x20, 0x00, 0x01},
			expectN: 4,
		},
		{
			name: "multiple non-empty encoded slices, one empty between them",
			input: []byte{
				0x01, 0x04, 0x01, 0x88, 0x01, 0x20, // []byte{0x88, 0x20}
				0x00,                                     // []byte{}
				0x01, 0x05, 0x00, 0x01, 0x01, 0x01, 0xae, // []byte{0x00, 0x01, 0xae}
			},
			p:       make([]byte, 4),
			expect:  []byte{0x88, 0x20, 0x00, 0x01},
			expectN: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			n, err := r.Read(tc.p)
			if tc.expectErr {
				assert.Error(err)
				return
			}
			if err != io.EOF {
				if !assert.NoError(err) {
					return
				}
			}

			assert.Equal(tc.expect, tc.p, "read bytes differ from expected")
			assert.Equal(tc.expectN, n, "read count n differs from expected")
		})
	}

}

func Test_Reader_Read_twoCalls(t *testing.T) {
	type afterCall struct {
		bytes                   []byte
		n                       int
		cacheLen                int
		totalReadFromUnderlying int
	}
	twoSeqTestCases := []struct {
		name  string
		input []byte
		pLen  int

		expect1 afterCall
		expect2 afterCall
	}{
		{
			name: "both 1 full slice - no cache",
			input: []byte{
				0x01, 0x08, 0x01, 0x01, 0x01, 0x02, 0x01, 0x03, 0x01, 0x04, // []byte{0x01, 0x02, 0x03, 0x04}
				0x01, 0x08, 0x01, 0x05, 0x01, 0x06, 0x01, 0x07, 0x01, 0x08, // []byte{0x05, 0x06, 0x07, 0x08}
			},
			pLen:    4,
			expect1: afterCall{bytes: []byte{0x01, 0x02, 0x03, 0x04}, n: 4, cacheLen: 0, totalReadFromUnderlying: 10},
			expect2: afterCall{bytes: []byte{0x05, 0x06, 0x07, 0x08}, n: 4, cacheLen: 0, totalReadFromUnderlying: 10},
		},
		{
			name: "first is 2 full, 2nd is 1 full - no cache",
			input: []byte{
				0x01, 0x02, 0x01, 0x01, // []byte{0x01}
				0x01, 0x06, 0x01, 0x02, 0x01, 0x03, 0x01, 0x04, // []byte{0x02, 0x03, 0x04}
				0x01, 0x08, 0x01, 0x05, 0x01, 0x06, 0x01, 0x07, 0x01, 0x08, // []byte{0x05, 0x06, 0x07, 0x08}
			},
			pLen:    4,
			expect1: afterCall{bytes: []byte{0x01, 0x02, 0x03, 0x04}, n: 4, cacheLen: 0, totalReadFromUnderlying: 12},
			expect2: afterCall{bytes: []byte{0x05, 0x06, 0x07, 0x08}, n: 4, cacheLen: 0, totalReadFromUnderlying: 10},
		},
		{
			name: "1st call leaves cache after 1, 2nd call reads entire cache and none from underlying",
			input: []byte{
				// []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
				0x01, 0x10, 0x01, 0x01, 0x01, 0x02, 0x01, 0x03, 0x01, 0x04, 0x01, 0x05, 0x01, 0x06, 0x01, 0x07, 0x01, 0x08,
			},
			pLen:    4,
			expect1: afterCall{bytes: []byte{0x01, 0x02, 0x03, 0x04}, n: 4, cacheLen: 4, totalReadFromUnderlying: 18},
			expect2: afterCall{bytes: []byte{0x05, 0x06, 0x07, 0x08}, n: 4, cacheLen: 0, totalReadFromUnderlying: 0},
		},
		{
			name: "1st call leaves cache after 1, 2nd call reads entire cache and 1 slice from underlying",
			input: []byte{
				// []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
				0x01, 0x0e, 0x01, 0x01, 0x01, 0x02, 0x01, 0x03, 0x01, 0x04, 0x01, 0x05, 0x01, 0x06, 0x01, 0x07,

				// []byte{0x08}
				0x01, 0x02, 0x01, 0x08,
			},
			pLen:    4,
			expect1: afterCall{bytes: []byte{0x01, 0x02, 0x03, 0x04}, n: 4, cacheLen: 3, totalReadFromUnderlying: 16},
			expect2: afterCall{bytes: []byte{0x05, 0x06, 0x07, 0x08}, n: 4, cacheLen: 0, totalReadFromUnderlying: 4},
		},
		{
			name: "1st call leaves cache after 1, 2nd call reads entire cache and 2 slices from remaining",
			input: []byte{
				// []byte{0x01, 0x02, 0x03, 0x04, 0x05}
				0x01, 0x0a, 0x01, 0x01, 0x01, 0x02, 0x01, 0x03, 0x01, 0x04, 0x01, 0x05,

				// []byte{0x06, 0x07}
				0x01, 0x04, 0x01, 0x06, 0x01, 0x07,

				// []byte{0x08}
				0x01, 0x02, 0x01, 0x08,
			},
			pLen:    4,
			expect1: afterCall{bytes: []byte{0x01, 0x02, 0x03, 0x04}, n: 4, cacheLen: 1, totalReadFromUnderlying: 12},
			expect2: afterCall{bytes: []byte{0x05, 0x06, 0x07, 0x08}, n: 4, cacheLen: 0, totalReadFromUnderlying: 10},
		},
		{
			name: "1st call leaves cache after 1, 2nd call reads entire cache and cannot be filled further",
			input: []byte{
				// []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
				0x01, 0x0e, 0x01, 0x01, 0x01, 0x02, 0x01, 0x03, 0x01, 0x04, 0x01, 0x05, 0x01, 0x06, 0x01, 0x07,
			},
			pLen:    4,
			expect1: afterCall{bytes: []byte{0x01, 0x02, 0x03, 0x04}, n: 4, cacheLen: 3, totalReadFromUnderlying: 16},
			expect2: afterCall{bytes: []byte{0x05, 0x06, 0x07, 0x00}, n: 3, cacheLen: 0, totalReadFromUnderlying: 0},
		},
		{
			name: "1st call leaves cache after 1, 2nd call reads some cache",
			input: []byte{
				// []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}
				0x01, 0x12, 0x01, 0x01, 0x01, 0x02, 0x01, 0x03, 0x01, 0x04, 0x01, 0x05, 0x01, 0x06, 0x01, 0x07, 0x01, 0x08, 0x01, 0x09,
			},
			pLen:    4,
			expect1: afterCall{bytes: []byte{0x01, 0x02, 0x03, 0x04}, n: 4, cacheLen: 5, totalReadFromUnderlying: 20},
			expect2: afterCall{bytes: []byte{0x05, 0x06, 0x07, 0x08}, n: 4, cacheLen: 1, totalReadFromUnderlying: 0},
		},
		{
			name: "1st call leaves cache after 2, 2nd call reads some cache",
			input: []byte{
				// []byte{0x01, 0x02}
				0x01, 0x04, 0x01, 0x01, 0x01, 0x02,

				// []byte{0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}
				0x01, 0x0e, 0x01, 0x03, 0x01, 0x04, 0x01, 0x05, 0x01, 0x06, 0x01, 0x07, 0x01, 0x08, 0x01, 0x09,
			},
			pLen:    4,
			expect1: afterCall{bytes: []byte{0x01, 0x02, 0x03, 0x04}, n: 4, cacheLen: 5, totalReadFromUnderlying: 22},
			expect2: afterCall{bytes: []byte{0x05, 0x06, 0x07, 0x08}, n: 4, cacheLen: 1, totalReadFromUnderlying: 0},
		},
	}
	for _, tc := range twoSeqTestCases {
		t.Run("two sequential calls - "+tc.name, func(t *testing.T) {
			assert := assert.New(t)

			var p []byte
			var beforeOff int
			var actualRead int

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			// first call:
			p = make([]byte, tc.pLen)
			beforeOff = r.offset
			actualN, err := r.Read(p)
			if err != io.EOF && err != nil {
				t.Fatalf("first r.Read: returned error: %s", err)
				return
			}
			actualRead = r.offset - beforeOff
			assert.Equal(tc.expect1.n, actualN, "first r.Read: n doesn't match expected")
			assert.Equal(tc.expect1.bytes, p, "first r.Read: output doesn't match expected")
			assert.Len(r.readBuf, tc.expect1.cacheLen, "first r.Read: read buffer after call doesn't match expected")
			assert.Equal(tc.expect1.totalReadFromUnderlying, actualRead, "first r.Read: total read bytes from stream doesn't match expected")

			// second call:
			p = make([]byte, tc.pLen)
			beforeOff = r.offset
			actualN, err = r.Read(p)
			if err != io.EOF && err != nil {
				t.Fatalf("second r.Read: returned error: %s", err)
				return
			}
			actualRead = r.offset - beforeOff
			assert.Equal(tc.expect2.n, actualN, "second r.Read: n doesn't match expected")
			assert.Equal(tc.expect2.bytes, p, "second r.Read: output doesn't match expected")
			assert.Len(r.readBuf, tc.expect2.cacheLen, "second r.Read: read buffer after call doesn't match expected")
			assert.Equal(tc.expect2.totalReadFromUnderlying, actualRead, "second r.Read: total read bytes from stream doesn't match expected")
		})
	}
}

func Test_Reader_Dec_sequential(t *testing.T) {
	assert := assert.New(t)
	var input []byte

	input = append(input, 0x02, 0x01, 0x9d) // 413
	var dest1Int int
	expect1Int := 413

	input = append(input, 0x01) // true
	var dest2Bool bool
	expect2Bool := true

	input = append(input, 0xa0) // nil
	var dest3Slice1 []int
	var expect3Slice1 []int

	input = append(input, // slice: {"VRISKA", "NEPETA", "TEREZI"}
		0x01, 0x1b, // len = 27
		0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
		0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
		0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"
	)
	var dest4Slice2 []string
	expect4Slice2 := []string{"VRISKA", "NEPETA", "TEREZI"}

	input = append(input, 0x41, 0x80, 0x01, 0x31) // "1"
	var dest5String string
	expect5String := "1"

	input = append(input, 0x02, 0x02, 0x64) // 612
	var dest6IntPtr *int
	expect6IntPtr := ref(612)

	input = append(input, // map: {413: "JOHN", 612: "VRISKA"}
		0x01, 0x16, // len=22

		0x02, 0x01, 0x9d, // 413:
		0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"

		0x02, 0x02, 0x64, // 612:
		0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
	)
	var dest7Map map[int]string
	expect7Map := map[int]string{413: "JOHN", 612: "VRISKA"}

	input = append(input, 0xa0) // nil
	var dest8BoolPtr *bool
	expect8BoolPtr := nilRef[bool]()

	input = append(input, 0x04, 0xc0, 0x70, 0x00, 0x32) // 256.01220703125
	var dest9Float float64
	expect9Float := 256.01220703125

	input = append(input, // testBinary{data: "ABC", number: 8}
		/* byte count = 8  */ 0x01, 0x08,
		/*  data  (string) */ 0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
		/* number (int32)  */ 0x01, 0x08, // 8
	)
	var dest10Bin testBinary
	expect10Bin := testBinary{number: 8, data: "ABC"}

	input = append(input, 0x41, 0x80, 0x07, 0x02, 0x3f, 0xf0, 0x03, 0xc0, 0x20, 0x80) // 1.0+8.25i
	var dest11Complex complex128
	expect11Complex := 1.0 + 8.25i

	r, err := NewReader(bytes.NewReader(input), nil)
	if !assert.NoError(err, "creating Reader returned error") {
		return
	}

	err = r.Dec(&dest1Int)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect1Int, dest1Int, "dest1Int mismatch")

	err = r.Dec(&dest2Bool)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect2Bool, dest2Bool, "dest2Bool mismatch")

	err = r.Dec(&dest3Slice1)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect3Slice1, dest3Slice1, "dest3Slice1 mismatch")

	err = r.Dec(&dest4Slice2)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect4Slice2, dest4Slice2, "dest4Slice2 mismatch")

	err = r.Dec(&dest5String)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect5String, dest5String, "dest5String mismatch")

	err = r.Dec(&dest6IntPtr)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect6IntPtr, dest6IntPtr, "dest6IntPtr mismatch")

	err = r.Dec(&dest7Map)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect7Map, dest7Map, "dest7Map mismatch")

	err = r.Dec(&dest8BoolPtr)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect8BoolPtr, dest8BoolPtr, "dest8BoolPtr mismatch")

	err = r.Dec(&dest9Float)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect9Float, dest9Float, "dest9Float mismatch")

	err = r.Dec(&dest10Bin)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect10Bin, dest10Bin, "dest10Bin mismatch")

	err = r.Dec(&dest11Complex)
	if !assert.NoError(err) {
		return
	}
	assert.Equal(expect11Complex, dest11Complex, "dest11Complex mismatch")
}

func Test_Reader_Dec_int(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    int
		expectOff int
		expectErr bool
	}{
		{
			name:      "normal value",
			input:     []byte{0x02, 0x01, 0x9d},
			expect:    413,
			expectOff: 3,
		},
		{
			// due to header sniffing for V1, this will fail if second byte has 0x80 bit set
			name:      "skip extension bytes",
			input:     []byte{0x42, 0x7f, 0xbf, 0x22, 0xb8},
			expect:    8888,
			expectOff: 5,
		},
		{
			name:      "normal value - multi value",
			input:     []byte{0x02, 0x01, 0x9d, 0x02, 0x01, 0x9d},
			expect:    413,
			expectOff: 3,
		},
		{
			// due to header sniffing for V1, this will fail if second byte has 0x80 bit set
			name:      "skip extension bytes - multi value",
			input:     []byte{0x42, 0x7f, 0xbf, 0x22, 0xb8, 0x42, 0x7f, 0xbf, 0x22, 0xb8},
			expect:    8888,
			expectOff: 5,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest int
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x20}
		expect := nilRef[int]()
		expectOff := 1

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *int
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		expectPtr := nilRef[int]()
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest **int
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_complex(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    complex128
		expectOff int
		expectErr bool
	}{
		{
			name:      "normal value - no compaction",
			input:     []byte{0x41, 0x80, 0x0c /*=len*/, 0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/},
			expect:    2.02499999999999991118215802999 + 1.0i,
			expectOff: 15,
		},
		{
			name:      "normal value - LSB compaction",
			input:     []byte{0x41, 0x80, 0x08 /*=len*/, 0x02, 0x3f, 0xf0 /*=real*/, 0x04, 0xc0, 0x70, 0x00, 0x32 /*=imag*/},
			expect:    1.0 + 256.01220703125i,
			expectOff: 11,
		},
		{
			name:      "normal value - MSB compaction",
			input:     []byte{0x41, 0x80, 0x08 /*=len*/, 0x04, 0x3f, 0xf0, 0x1c, 0x00 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/},
			expect:    1.00000000000159161572810262442 + 1.0i,
			expectOff: 11,
		},
		{
			name:      "skip extension bytes",
			input:     []byte{0x41, 0xc0, 0xbf, 0x06 /*=len*/, 0x82, 0x3f, 0xf0 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/},
			expect:    -1.0 + 1.0i,
			expectOff: 10,
		},

		{
			name:      "normal value - multi value",
			input:     []byte{0x41, 0x80, 0x06 /*=len*/, 0x82, 0x3f, 0xf0 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/, 0x41, 0x80, 0x06 /*=len*/, 0x82, 0x3f, 0xf0 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/},
			expect:    -1.0 + 1.0i,
			expectOff: 9,
		},
		{
			// due to header sniffing for V1, this will fail if second byte has 0x80 bit set
			name:      "skip extension bytes - multi value",
			input:     []byte{0x41, 0xc0, 0xbf, 0x06 /*=len*/, 0x82, 0x3f, 0xf0 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/, 0x41, 0xc0, 0xbf, 0x06 /*=len*/, 0x82, 0x3f, 0xf0 /*=real*/, 0x02, 0x3f, 0xf0 /*=imag*/},
			expect:    -1.0 + 1.0i,
			expectOff: 10,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest complex128
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x20}
		expect := nilRef[complex128]()
		expectOff := 1

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *complex128
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		expectPtr := nilRef[complex128]()
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest **complex128
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_float(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    float64
		expectOff int
		expectErr bool
	}{
		{
			name:      "normal value - no compaction",
			input:     []byte{0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33},
			expect:    2.02499999999999991118215802999,
			expectOff: 9,
		},
		{
			name:      "normal value - LSB compaction",
			input:     []byte{0x04, 0xc0, 0x70, 0x00, 0x32},
			expect:    256.01220703125,
			expectOff: 5,
		},
		{
			name:      "normal value - MSB compaction",
			input:     []byte{0x04, 0x3f, 0xf0, 0x1c, 0x00},
			expect:    1.00000000000159161572810262442,
			expectOff: 5,
		},
		{
			// due to header sniffing for V1, this will fail if second byte has 0x80 bit set
			name:      "skip extension bytes",
			input:     []byte{0xc2, 0x7f, 0xbf, 0x3f, 0xf0},
			expect:    -1.0,
			expectOff: 5,
		},
		{
			name:      "normal value - multi value",
			input:     []byte{0x82, 0x3f, 0xf0, 0x82, 0x3f, 0xf0},
			expect:    -1.0,
			expectOff: 3,
		},
		{
			// due to header sniffing for V1, this will fail if second byte has 0x80 bit set
			name:      "skip extension bytes - multi value",
			input:     []byte{0xc2, 0x7f, 0xbf, 0x3f, 0xf0, 0xc2, 0x7f, 0xbf, 0x3f, 0xf0},
			expect:    -1.0,
			expectOff: 5,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest float64
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x20}
		expect := nilRef[float64]()
		expectOff := 1

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *float64
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		expectPtr := nilRef[float64]()
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest **float64
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_bool(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    bool
		expectErr bool
		expectOff int
	}{
		{
			name:      "normal value - false",
			input:     []byte{0x00},
			expect:    false,
			expectOff: 1,
		},
		{
			name:      "normal value - true",
			input:     []byte{0x01},
			expect:    true,
			expectOff: 1,
		},
		{
			name:      "normal value - false - multi value",
			input:     []byte{0x00, 0x00},
			expect:    false,
			expectOff: 1,
		},
		{
			name:      "normal value - true - multi value",
			input:     []byte{0x01, 0x01},
			expect:    true,
			expectOff: 1,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest bool
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x20}
		expect := nilRef[bool]()
		expectOff := 1

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *bool
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		expectPtr := nilRef[bool]()
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest **bool
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_stringV0(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    string
		expectErr bool
		expectOff int
	}{
		{
			name:      "empty",
			input:     []byte{0x00},
			expect:    "",
			expectOff: 1,
		},
		{
			name:      "string with no multibyte chars",
			input:     []byte{0x01, 0x01, 0x31},
			expect:    "1",
			expectOff: 3,
		},
		{
			name:      "string with multibyte chars",
			input:     []byte{0x01, 0x09, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c},
			expect:    "Hello, 世界",
			expectOff: 15,
		},
		{
			name:      "empty - multiple values",
			input:     []byte{0x00, 0x00},
			expect:    "",
			expectOff: 1,
		},
		{
			name:      "string with no multibyte chars - multiple values",
			input:     []byte{0x01, 0x01, 0x31, 0x01, 0x01, 0x31},
			expect:    "1",
			expectOff: 3,
		},
		{
			name:      "string with multibyte chars - multiple values",
			input:     []byte{0x01, 0x09, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c, 0x01, 0x09, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c},
			expect:    "Hello, 世界",
			expectOff: 15,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest string
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	// cannot encode a v0 string that is nil so we also cannot decode one
}

func Test_Reader_Dec_stringV1(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    string
		expectErr bool
		expectOff int
	}{
		{
			name:      "empty",
			input:     []byte{0x00},
			expect:    "",
			expectOff: 1,
		},
		{
			name:      "string with no multibyte chars",
			input:     []byte{0x41, 0x80, 0x01, 0x31},
			expect:    "1",
			expectOff: 4,
		},
		{
			name:      "string with multibyte chars",
			input:     []byte{0x41, 0x80, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c},
			expect:    "Hello, 世界",
			expectOff: 16,
		},
		{
			name:      "empty - multiple values",
			input:     []byte{0x00, 0x00},
			expect:    "",
			expectOff: 1,
		},
		{
			name:      "string with no multibyte chars - multiple values",
			input:     []byte{0x41, 0x80, 0x01, 0x31, 0x40, 0x80, 0x01, 0x01, 0x31},
			expect:    "1",
			expectOff: 4,
		},
		{
			name:      "string with multibyte chars - multiple values",
			input:     []byte{0x41, 0x80, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c, 0x40, 0x80, 0x01, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c},
			expect:    "Hello, 世界",
			expectOff: 16,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest string
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x20}
		expect := nilRef[string]()
		expectOff := 1

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *string
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		expectPtr := nilRef[string]()
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest **string
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_binary(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    testBinary
		expectErr bool
		expectOff int
	}{
		{
			name: "empty values",
			input: []byte{
				/* byte count = 2  */ 0x01, 0x02,
				/*  data  (string) */ 0x00, // ""
				/* number (int32)  */ 0x00, // 0
			},
			expect:    testBinary{},
			expectOff: 4,
		},
		{
			name: "filled values",
			input: []byte{
				/* byte count = 8  */ 0x01, 0x08,
				/*  data  (string) */ 0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
				/* number (int32)  */ 0x01, 0x08, // 8
			},
			expect:    testBinary{data: "ABC", number: 8},
			expectOff: 10,
		},
		{
			name: "empty values x2",
			input: []byte{
				/* byte count = 2  */ 0x01, 0x02,
				/*  data  (string) */ 0x00, // ""
				/* number (int32)  */ 0x00, // 0

				/* byte count = 2  */ 0x01, 0x02,
				/*  data  (string) */ 0x00, // ""
				/* number (int32)  */ 0x00, // 0
			},
			expect:    testBinary{},
			expectOff: 4,
		},
		{
			name: "filled values x2",
			input: []byte{
				/* byte count = 8  */ 0x01, 0x08,
				/*  data  (string) */ 0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
				/* number (int32)  */ 0x01, 0x08, // 8

				/* byte count = 8  */ 0x01, 0x08,
				/*  data  (string) */ 0x41, 0x80, 0x03, 0x41, 0x42, 0x43, // "ABC"
				/* number (int32)  */ 0x01, 0x08, // 8
			},
			expect:    testBinary{data: "ABC", number: 8},
			expectOff: 10,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest testBinary
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - single indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x20}
		expect := nilRef[testBinary]()
		expectOff := 1

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *testBinary
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		expectPtr := nilRef[testBinary]()
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest **testBinary
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_slice(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    []string
		expectErr bool
		expectOff int
	}{
		{
			name: "nil",
			input: []byte{
				0xa0, // nil=true
			},
			expect:    nil,
			expectOff: 1,
		},
		{
			name: "empty",
			input: []byte{
				0x00, // len=0
			},
			expect:    []string{},
			expectOff: 1,
		},
		{
			name: "3 value",
			input: []byte{
				0x01, 0x1b, // len=27

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
				0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"
			},
			expect:    []string{"VRISKA", "NEPETA", "TEREZI"},
			expectOff: 29,
		},
		{
			name: "nil x2",
			input: []byte{
				0xa0, // nil=true
				0xa0, // nil=true
			},
			expect:    nil,
			expectOff: 1,
		},
		{
			name: "empty x2",
			input: []byte{
				0x00, // len=0
				0x00, // len=0
			},
			expect:    []string{},
			expectOff: 1,
		},
		{
			name: "3 value x2",
			input: []byte{
				0x01, 0x1b, // len=27

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
				0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"

				0x01, 0x1b, // len=27

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
				0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"
			},
			expect:    []string{"VRISKA", "NEPETA", "TEREZI"},
			expectOff: 29,
		},
		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest []string
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		var expectPtr []string
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *[]string
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}

func Test_Reader_Dec_map(t *testing.T) {
	testCases := []struct {
		name      string
		input     []byte
		expect    map[int]string
		expectErr bool
		expectOff int
	}{
		{
			name: "nil",
			input: []byte{
				0xa0, // nil=true
			},
			expect:    nil,
			expectOff: 1,
		},
		{
			name: "empty",
			input: []byte{
				0x00, // len=0
			},
			expect:    map[int]string{},
			expectOff: 1,
		},
		{
			name: "2 values",
			input: []byte{
				0x01, 0x16, // len=22

				0x02, 0x01, 0x9d, // 413:
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"

				0x02, 0x02, 0x64, // 612:
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
			},
			expect:    map[int]string{413: "JOHN", 612: "VRISKA"},
			expectOff: 24,
		},
		{
			name: "nil x2",
			input: []byte{
				0xa0, // nil=true
				0xa0, // nil=true
			},
			expect:    nil,
			expectOff: 1,
		},
		{
			name: "empty x2",
			input: []byte{
				0x00, // len=0
				0x00, // len=0
			},
			expect:    map[int]string{},
			expectOff: 1,
		},
		{
			name: "2 values x2",
			input: []byte{
				0x01, 0x16, // len=22

				0x02, 0x01, 0x9d, // 413:
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"

				0x02, 0x02, 0x64, // 612:
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"

				0x01, 0x16, // len=22

				0x02, 0x01, 0x9d, // 413:
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"

				0x02, 0x02, 0x64, // 612:
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
			},
			expect:    map[int]string{413: "JOHN", 612: "VRISKA"},
			expectOff: 24,
		},

		{
			// error - invalid (nil) count
			name:      "error - invalid indir count int",
			input:     []byte{0x70, 0x00, 0x20},
			expectErr: true,
			expectOff: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			r, err := NewReader(bytes.NewReader(tc.input), nil)
			if !assert.NoError(err, "creating Reader returned error") {
				return
			}

			var dest map[int]string
			err = r.Dec(&dest)
			if tc.expectErr {
				assert.Error(err, "error not returned")
				assert.Equal(tc.expectOff, r.offset, "offset mismatch")
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, dest, "dest not expected value")
			assert.Equal(tc.expectOff, r.offset, "offset mismatch")
		})
	}

	t.Run("nil value - multi indir", func(t *testing.T) {
		assert := assert.New(t)
		input := []byte{0x30, 0x01, 0x01}
		var expectPtr map[int]string
		expect := &expectPtr
		expectOff := 3

		r, err := NewReader(bytes.NewReader(input), nil)
		if !assert.NoError(err, "creating Reader returned error") {
			return
		}

		var dest *map[int]string
		err = r.Dec(&dest)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, dest, "dest not expected value")
		assert.Equal(expectOff, r.offset, "offset mismatch")
	})
}
