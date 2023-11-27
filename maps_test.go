package rezi

import (
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Enc_Map_NoIndirection(t *testing.T) {
	t.Run("nil map[string]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  map[string]int
			expect = []byte{
				0xa0,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[float64]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[float64]int{0.25: 1, 8.5: 8, -1.0: 1}
			expect = []byte{
				0x01, 0x0f, // len=15

				0x82, 0x3f, 0xf0, // -1.0
				0x01, 0x01, // 1

				0x02, 0x3f, 0xd0, // 0.25
				0x01, 0x01, // 1

				0x02, 0x40, 0x21, // 8.5
				0x01, 0x08, // 8
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[string]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[string]int{"ONE": 1, "EIGHT": 8}
			expect = []byte{
				0x01, 0x12, // len=18

				0x41, 0x82, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54, // "EIGHT"
				0x01, 0x08, // 8

				0x41, 0x82, 0x03, 0x4f, 0x4e, 0x45, // "ONE"
				0x01, 0x01, // 1
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]uint64", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[int]uint64{8: 10004138888888800612, 1: 88888888}
			expect = []byte{
				0x01, 0x12,

				0x01, 0x01,
				0x04, 0x05, 0x4c, 0x56, 0x38,

				0x01, 0x08,
				0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x64,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]string", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[bool]string{true: "COOL VRISKA", false: "TAV"}
			expect = []byte{
				0x01, 0x16, // len=22

				0x00,                               // true
				0x41, 0x82, 0x03, 0x54, 0x41, 0x56, // "TAV"

				0x01,                                                                               // false
				0x41, 0x82, 0x0b, 0x43, 0x4f, 0x4f, 0x4c, 0x20, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "COOL VRISKA"
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]binary", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]testBinary{
				413: {data: "JOHN", number: 1},
				612: {data: "VRISKA", number: 8},
			}
			expect = []byte{
				0x01, 0x1e, // len=30

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x09, // len=9
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"
				0x01, 0x01, // 1

				0x02, 0x02, 0x64, // 612:
				0x01, 0x0b, // len=11
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]text", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]testText{
				8:   {name: "VRISKA", value: 8, enabled: true},
				413: {name: "JOHN", enabled: false, value: 413},
				100: {name: "NEPETA", enabled: false, value: 100},
			}

			expect = []byte{
				0x01, 0x3b, // len=59

				// 8
				0x01, 0x08,
				// "8,true,VRISKA"
				0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,

				// 100
				0x01, 0x64,
				// "100,false,NEPETA"
				0x41, 0x82, 0x10, 0x31, 0x30, 0x30, 0x2c, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41,

				// 413
				0x02, 0x01, 0x9d,
				// "413,false,JOHN"
				0x41, 0x82, 0x0e, 0x34, 0x31, 0x33, 0x2c, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x4a, 0x4f, 0x48, 0x4e,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]struct", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]testStructManyFields{
				4: {Name: "KANAYA", Value: 8, Factor: 0.25, Enabled: true, hidden: make(chan int, 3), enabled: &sync.Mutex{}, inc: 48},
				2: {Name: "ROSE", Value: 12, Factor: 0.00390625, Enabled: false, hidden: make(chan int), enabled: nil, inc: 12},
			}

			expect = []byte{
				0x01, 0x68, // len=104

				0x01, 0x02, // 2:
				// "ROSE" struct:

				0x01, 0x2f, // struct len=47

				0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
				0x00, // false

				0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
				0x02, 0x3f, 0x70, // 0.00390625

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x04, 0x52, 0x4f, 0x53, 0x45, // "ROSE"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x0c, // 12

				0x01, 0x04, // 4:
				// "KANAYA" struct:

				0x01, 0x31, // struct len=49

				0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
				0x01, // true

				0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
				0x02, 0x3f, 0xd0, // 0.25

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4b, 0x41, 0x4e, 0x41, 0x59, 0x41, // "KANAYA"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x08, // 8
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]float64", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[int]float64{1: 0.25, 8: 8.5}
			expect = []byte{
				0x01, 0x0a, // len=10

				0x01, 0x01, // 1
				0x02, 0x3f, 0xd0, // 0.25

				0x01, 0x08, // 8
				0x02, 0x40, 0x21, // 8.5
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]bool", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]bool{
				413: true,
				612: true,
				100: false,
			}
			expect = []byte{
				0x01, 0x0b, // len=11

				0x01, 0x64, 0x00, // 100: false
				0x02, 0x01, 0x9d, 0x01, // 413: true
				0x02, 0x02, 0x64, 0x01, // 612: true
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]complex128", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]complex128{
				413: complex128(2.02499999999999991118215802999 + 1.0i),
				612: complex128(0.0 + 0.0i),
				100: complex128(8.0 + 8.0i),
			}
			expect = []byte{
				0x01, 0x21, // len=33

				0x01, 0x64, 0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20, // 100: 8.0+8.0i
				0x02, 0x01, 0x9d, 0x41, 0x80, 0x0c, 0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x02, 0x3f, 0xf0, // 413: 2.02499999999999991118215802999 + 1.0i
				0x02, 0x02, 0x64, 0x00, // 612: 0.0+0.0i
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int][3]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int][3]int{
				413: {4, 1, 3},
				612: {6, 1},
			}
			expect = []byte{
				0x01, 0x15, // len=21

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x06, // len=6
				0x01, 0x04, 0x01, 0x01, 0x01, 0x03, // {4, 1, 3}

				0x02, 0x02, 0x64, // 612:
				0x01, 0x05, // len=5
				0x01, 0x06, 0x01, 0x01, 0x00, // {6, 1, 0}
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int][]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int][]int{
				413: {4, 1, 3},
				612: {6, 1, 2},
			}
			expect = []byte{
				0x01, 0x16, // len=22

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x06, // len=6
				0x01, 0x04, 0x01, 0x01, 0x01, 0x03, // {4, 1, 3}

				0x02, 0x02, 0x64, // 612:
				0x01, 0x06, // len=6
				0x01, 0x06, 0x01, 0x01, 0x01, 0x02, // {6, 1, 2}
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("http.Header (map[string][]string)", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = http.Header{
				"X-Api-Key":    []string{"12345"},
				"Content-Type": []string{"application/json"},
			}
			expect = []byte{
				0x01, 0x3a, // map len=58

				0x41, 0x82, 0x0c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2d, 0x54, 0x79, 0x70, 0x65, // "Content-Type":
				0x01, 0x13, // slice len=19
				0x41, 0x82, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, // "application/json"

				0x41, 0x82, 0x09, 0x58, 0x2d, 0x41, 0x70, 0x69, 0x2d, 0x4b, 0x65, 0x79, // "X-Api-Key":
				0x01, 0x08, // slice len=8
				0x41, 0x82, 0x5, 0x31, 0x32, 0x33, 0x34, 0x35, // "12345"
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("meta map[int]map[int]string", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]map[int]string{
				413: {
					2: "JOHN",
					4: "ROSE",
				},
				612: {
					8: "VRISKA",
					4: "NEPETA",
				},
			}
			expect = []byte{
				0x01, 0x32, // len=50

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x12, // len=18
				0x01, 0x02, 0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // 2: "JOHN"
				0x01, 0x04, 0x41, 0x82, 0x04, 0x52, 0x4f, 0x53, 0x45, // 4: "ROSE"

				0x02, 0x02, 0x64, // 612:
				0x01, 0x16, // len=22
				0x01, 0x04, 0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // 4: "NEPETA"
				0x01, 0x08, 0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // 8: "VRISKA"
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})
}

func Test_Enc_Map_SelfIndirection(t *testing.T) {
	t.Run("nil *map[string]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *map[string]int
			expect = []byte{
				0xa0,
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("*map[string]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = map[string]int{"ONE": 1, "EIGHT": 8}
			input    = &inputVal
			expect   = []byte{
				0x01, 0x12, // len=18

				0x41, 0x82, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54, // "EIGHT"
				0x01, 0x08, // 8

				0x41, 0x82, 0x03, 0x4f, 0x4e, 0x45, // "ONE"
				0x01, 0x01, // 1
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**map[string]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = map[string]int{"ONE": 1, "EIGHT": 8}
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{
				0x01, 0x12, // len=18

				0x41, 0x82, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54, // "EIGHT"
				0x01, 0x08, // 8

				0x41, 0x82, 0x03, 0x4f, 0x4e, 0x45, // "ONE"
				0x01, 0x01, // 1
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**map[string]int, but nil map[string]int part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *map[string]int
			input  = &ptr
			expect = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})
}

func Test_Enc_Map_ValueIndirection(t *testing.T) {
	t.Run("map[string]*int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[string]*int{"ONE": ref(1), "EIGHT": ref(8)}
			expect = []byte{
				0x01, 0x12, // len=18

				0x41, 0x82, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54, // "EIGHT"
				0x01, 0x08, // 8

				0x41, 0x82, 0x03, 0x4f, 0x4e, 0x45, // "ONE"
				0x01, 0x01, // 1
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[string]*int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[string]*int{"ONE": ref(1), "EIGHT": nil}
			expect = []byte{
				0x01, 0x11, // len=17

				0x41, 0x82, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54, // "EIGHT"
				0xa0, // nil

				0x41, 0x82, 0x03, 0x4f, 0x4e, 0x45, // "ONE"
				0x01, 0x01, // 1
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[string]*int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[string]*int{"ONE": nil, "EIGHT": nil}
			expect = []byte{
				0x01, 0x10, // len=16

				0x41, 0x82, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54, // "EIGHT"
				0xa0, // nil

				0x41, 0x82, 0x03, 0x4f, 0x4e, 0x45, // "ONE"
				0xa0, // nil
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*uint64, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[int]*uint64{8: ref(uint64(10004138888888800612)), 1: ref(uint64(88888888))}
			expect = []byte{
				0x01, 0x12,

				0x01, 0x01,
				0x04, 0x05, 0x4c, 0x56, 0x38,

				0x01, 0x08,
				0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x64,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*uint64, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[int]*uint64{8: nil, 1: ref(uint64(88888888))}
			expect = []byte{
				0x01, 0x0a,

				0x01, 0x01,
				0x04, 0x05, 0x4c, 0x56, 0x38,

				0x01, 0x08,
				0xa0,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*uint64, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[int]*uint64{8: nil, 1: nil}
			expect = []byte{
				0x01, 0x06,

				0x01, 0x01,
				0xa0,

				0x01, 0x08,
				0xa0,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]*string, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[bool]*string{true: ref("COOL VRISKA"), false: ref("TAV")}
			expect = []byte{
				0x01, 0x16, // len=22

				0x00,                               // false
				0x41, 0x82, 0x03, 0x54, 0x41, 0x56, // "TAV"

				0x01,                                                                               // true
				0x41, 0x82, 0x0b, 0x43, 0x4f, 0x4f, 0x4c, 0x20, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "COOL VRISKA"
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]*string, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[bool]*string{true: ref("COOL VRISKA"), false: nil}
			expect = []byte{
				0x01, 0x11, // len=17

				0x00, // false
				0xa0, // nil

				0x01,                                                                               // true
				0x41, 0x82, 0x0b, 0x43, 0x4f, 0x4f, 0x4c, 0x20, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "COOL VRISKA"
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]*string, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[bool]*string{true: nil, false: nil}
			expect = []byte{
				0x01, 0x04,

				0x00,
				0xa0,

				0x01,
				0xa0,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*binary, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*testBinary{
				413: {data: "JOHN", number: 1},
				612: {data: "VRISKA", number: 8},
			}
			expect = []byte{
				0x01, 0x1e, // len=30

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x09, // len=9
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"
				0x01, 0x01, // 1

				0x02, 0x02, 0x64, // 612:
				0x01, 0x0b, // len=11
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*binary, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*testBinary{
				413: nil,
				612: {data: "VRISKA", number: 8},
			}
			expect = []byte{
				0x01, 0x14, // len=20

				0x02, 0x01, 0x9d, // 413:
				0xa0, // nil

				0x02, 0x02, 0x64, // 612:
				0x01, 0x0b, // len=11
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*binary, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*testBinary{
				413: nil,
				612: nil,
			}
			expect = []byte{
				0x01, 0x08, // len=8

				0x02, 0x01, 0x9d, // 413:
				0xa0,

				0x02, 0x02, 0x64, // 612:
				0xa0,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*text, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*testText{
				8:   {name: "VRISKA", value: 8, enabled: true},
				413: {name: "JOHN", enabled: false, value: 413},
				100: {name: "NEPETA", enabled: false, value: 100},
			}

			expect = []byte{
				0x01, 0x3b, // len=59

				// 8
				0x01, 0x08,
				// "8,true,VRISKA"
				0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,

				// 100
				0x01, 0x64,
				// "100,false,NEPETA"
				0x41, 0x82, 0x10, 0x31, 0x30, 0x30, 0x2c, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41,

				// 413
				0x02, 0x01, 0x9d,
				// "413,false,JOHN"
				0x41, 0x82, 0x0e, 0x34, 0x31, 0x33, 0x2c, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x4a, 0x4f, 0x48, 0x4e,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*text, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*testText{
				8:   {name: "VRISKA", value: 8, enabled: true},
				413: {name: "JOHN", enabled: false, value: 413},
				100: nil,
			}

			expect = []byte{
				0x01, 0x29, // len=41

				// 8
				0x01, 0x08,
				// "8,true,VRISKA"
				0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,

				// 100
				0x01, 0x64,
				// nil
				0xa0,

				// 413
				0x02, 0x01, 0x9d,
				// "413,false,JOHN"
				0x41, 0x82, 0x0e, 0x34, 0x31, 0x33, 0x2c, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x4a, 0x4f, 0x48, 0x4e,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*text, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*testText{
				8:   nil,
				413: nil,
				100: nil,
			}

			expect = []byte{
				0x01, 0x0a, // len=10

				// 8
				0x01, 0x08,
				0xa0,

				// 100
				0x01, 0x64,
				0xa0,

				// 413
				0x02, 0x01, 0x9d,
				0xa0,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*bool, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*bool{
				413: ref(true),
				612: ref(true),
				100: ref(false),
			}
			expect = []byte{
				0x01, 0x0b, // len=11

				0x01, 0x64, 0x00, // 100: false
				0x02, 0x01, 0x9d, 0x01, // 413: true
				0x02, 0x02, 0x64, 0x01, // 612: true
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*bool, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*bool{
				413: ref(true),
				612: nil,
				100: ref(false),
			}
			expect = []byte{
				0x01, 0x0b, // len=11

				0x01, 0x64, 0x00, // 100: false
				0x02, 0x01, 0x9d, 0x01, // 413: true
				0x02, 0x02, 0x64, 0xa0, // 612: nil
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*bool, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*bool{
				413: nil,
				612: nil,
				100: nil,
			}
			expect = []byte{
				0x01, 0x0b, // len=11

				0x01, 0x64, 0xa0, // 100: false
				0x02, 0x01, 0x9d, 0xa0, // 413: true
				0x02, 0x02, 0x64, 0xa0, // 612: true
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[]int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*[]int{
				413: {4, 1, 3},
				612: {6, 1, 2},
			}
			expect = []byte{
				0x01, 0x16, // len=22

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x06, // len=6
				0x01, 0x04, 0x01, 0x01, 0x01, 0x03, // {4, 1, 3}

				0x02, 0x02, 0x64, // 612:
				0x01, 0x06, // len=6
				0x01, 0x06, 0x01, 0x01, 0x01, 0x02, // {6, 1, 2}
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[]int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*[]int{
				413: nil,
				612: {6, 1, 2},
			}
			expect = []byte{
				0x01, 0x0f, // len=15

				0x02, 0x01, 0x9d, // 413:
				0xa0,

				0x02, 0x02, 0x64, // 612:
				0x01, 0x06, // len=6
				0x01, 0x06, 0x01, 0x01, 0x01, 0x02, // {6, 1, 2}
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[]int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*[]int{
				413: nil,
				612: nil,
			}
			expect = []byte{
				0x01, 0x08, // len=8

				0x02, 0x01, 0x9d, // 413:
				0xa0,

				0x02, 0x02, 0x64, // 612:
				0xa0,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[3]int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*[3]int{
				413: {4, 1, 3},
				612: {6, 1},
			}
			expect = []byte{
				0x01, 0x15, // len=21

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x06, // len=6
				0x01, 0x04, 0x01, 0x01, 0x01, 0x03, // {4, 1, 3}

				0x02, 0x02, 0x64, // 612:
				0x01, 0x05, // len=5
				0x01, 0x06, 0x01, 0x01, 0x00, // {6, 1, 0}
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[3]int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*[3]int{
				413: nil,
				612: {6, 1},
			}
			expect = []byte{
				0x01, 0x0e, // len=14

				0x02, 0x01, 0x9d, // 413:
				0xa0,

				0x02, 0x02, 0x64, // 612:
				0x01, 0x05, // len=5
				0x01, 0x06, 0x01, 0x01, 0x00, // {6, 1, 0}
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[3]int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*[3]int{
				413: nil,
				612: nil,
			}
			expect = []byte{
				0x01, 0x08, // len=8

				0x02, 0x01, 0x9d, // 413:
				0xa0,

				0x02, 0x02, 0x64, // 612:
				0xa0,
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*map[int]string, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*map[int]string{
				413: {
					2: "JOHN",
					4: "ROSE",
				},
				612: {
					8: "VRISKA",
					4: "NEPETA",
				},
			}
			expect = []byte{
				0x01, 0x32, // len=50

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x12, // len=18
				0x01, 0x02, 0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // 2: "JOHN"
				0x01, 0x04, 0x41, 0x82, 0x04, 0x52, 0x4f, 0x53, 0x45, // 4: "ROSE"

				0x02, 0x02, 0x64, // 612:
				0x01, 0x16, // len=22
				0x01, 0x04, 0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // 4: "NEPETA"
				0x01, 0x08, 0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // 8: "VRISKA"
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*map[int]string, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*map[int]string{
				413: {
					2: "JOHN",
					4: "ROSE",
				},
				612: nil,
			}
			expect = []byte{
				0x01, 0x1b, // len=27

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x12, // len=18
				0x01, 0x02, 0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // 2: "JOHN"
				0x01, 0x04, 0x41, 0x82, 0x04, 0x52, 0x4f, 0x53, 0x45, // 4: "ROSE"

				0x02, 0x02, 0x64, // 612:
				0xa0, // nil
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*map[int]string, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = map[int]*map[int]string{
				413: nil,
				612: nil,
			}
			expect = []byte{
				0x01, 0x08, // len=8

				0x02, 0x01, 0x9d, // 413:
				0xa0, // nil

				0x02, 0x02, 0x64, // 612:
				0xa0, // nil
			}
		)

		// execute
		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expect, actual)
	})
}

func Test_Dec_Map_NoIndirection(t *testing.T) {
	t.Run("nil map[string]int (implicit nil)", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x80,
			}
			expectConsumed = 1
		)

		// execute
		actual := map[string]int{"A": 1, "B": 2} // start with a value so we can check it is set to nil
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Nil(actual)
	})

	t.Run("nil map[string]int (explicit nil)", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0xa0,
			}
			expectConsumed = 1
		)

		// execute
		actual := map[string]int{"A": 1, "B": 2} // start with a value so we can check it is set to nil
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Nil(actual)
	})

	t.Run("map[string]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x10,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0x01, 0x08,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0x01, 0x01,
			}
			expect         = map[string]int{"ONE": 1, "EIGHT": 8}
			expectConsumed = 18
		)

		// execute
		var actual map[string]int
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]uint64", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x12,

				0x01, 0x01,
				0x04, 0x05, 0x4c, 0x56, 0x38,

				0x01, 0x08,
				0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x64,
			}
			expect         = map[int]uint64{8: 10004138888888800612, 1: 88888888}
			expectConsumed = 20
		)

		// execute
		var actual map[int]uint64
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]string", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x14,

				0x00,
				0x01, 0x03, 0x54, 0x41, 0x56,

				0x01,
				0x01, 0x0b, 0x43, 0x4f, 0x4f, 0x4c, 0x20, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,
			}
			expect         = map[bool]string{true: "COOL VRISKA", false: "TAV"}
			expectConsumed = 22
		)

		// execute
		var actual map[bool]string
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[float64]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0f, // len=15

				0x82, 0x3f, 0xf0, // -1.0
				0x01, 0x01, // 1

				0x02, 0x3f, 0xd0, // 0.25
				0x01, 0x01, // 1

				0x02, 0x40, 0x21, // 8.5
				0x01, 0x08, // 8
			}
			expect         = map[float64]int{0.25: 1, 8.5: 8, -1.0: 1}
			expectConsumed = 17
		)

		// execute
		var actual map[float64]int
		consumed, err := Dec(input, &actual)

		// asset
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]float64", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0f, // len=15

				0x01, 0x01, // 1
				0x02, 0x3f, 0xd0, // 0.25

				0x01, 0x02, // 2
				0x82, 0x3f, 0xf0, // -1.0

				0x01, 0x08, // 8
				0x02, 0x40, 0x21, // 8.5
			}
			expect         = map[int]float64{2: -1.0, 1: 0.25, 8: 8.5}
			expectConsumed = 17
		)

		// execute
		var actual map[int]float64
		consumed, err := Dec(input, &actual)

		// asset
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]complex128", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x21, // len=33

				0x01, 0x64, 0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20, // 100: 8.0+8.0i
				0x02, 0x01, 0x9d, 0x41, 0x80, 0x0c, 0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x02, 0x3f, 0xf0, // 413: 2.02499999999999991118215802999 + 1.0i
				0x02, 0x02, 0x64, 0x00, // 612: 0.0+0.0i
			}
			expect = map[int]complex128{
				413: complex128(2.02499999999999991118215802999 + 1.0i),
				612: complex128(0.0 + 0.0i),
				100: complex128(8.0 + 8.0i),
			}
			expectConsumed = 35
		)

		// execute
		var actual map[int]complex128
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]binary", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x1c, // len=28

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x08, // len=8
				0x01, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"
				0x01, 0x01, // 1

				0x02, 0x02, 0x64, // 612:
				0x01, 0x0a, // len=10
				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
			expect = map[int]testBinary{
				413: {data: "JOHN", number: 1},
				612: {data: "VRISKA", number: 8},
			}
			expectConsumed = 30
		)

		// execute
		var actual map[int]testBinary
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]text", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x3b, // len=59

				// 8
				0x01, 0x08,
				// "8,true,VRISKA"
				0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,

				// 100
				0x01, 0x64,
				// "100,false,NEPETA"
				0x41, 0x82, 0x10, 0x31, 0x30, 0x30, 0x2c, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41,

				// 413
				0x02, 0x01, 0x9d,
				// "413,false,JOHN"
				0x41, 0x82, 0x0e, 0x34, 0x31, 0x33, 0x2c, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x4a, 0x4f, 0x48, 0x4e,
			}
			expect = map[int]testText{
				8:   {name: "VRISKA", value: 8, enabled: true},
				413: {name: "JOHN", enabled: false, value: 413},
				100: {name: "NEPETA", enabled: false, value: 100},
			}
			expectConsumed = 61
		)

		// execute
		var actual map[int]testText
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]struct", func(t *testing.T) {
		// setup
		assert := assert.New(t)

		var (
			input = []byte{
				0x01, 0x68, // len=104

				0x01, 0x02, // 2:
				// "ROSE" struct:

				0x01, 0x2f, // struct len=47

				0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
				0x00, // false

				0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
				0x02, 0x3f, 0x70, // 0.00390625

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x04, 0x52, 0x4f, 0x53, 0x45, // "ROSE"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x0c, // 12

				0x01, 0x04, // 4:
				// "KANAYA" struct:

				0x01, 0x31, // struct len=49

				0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
				0x01, // true

				0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
				0x02, 0x3f, 0xd0, // 0.25

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4b, 0x41, 0x4e, 0x41, 0x59, 0x41, // "KANAYA"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x08, // 8
			}
			expect = map[int]testStructManyFields{
				4: {Name: "KANAYA", Value: 8, Factor: 0.25, Enabled: true, hidden: nil, enabled: nil, inc: 0},
				2: {Name: "ROSE", Value: 12, Factor: 0.00390625, Enabled: false, hidden: nil, enabled: nil, inc: 0},
			}
			expectConsumed = 106
		)

		// execute
		var actual map[int]testStructManyFields
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int][]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x16, // len=22

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x06, // len=6
				0x01, 0x04, 0x01, 0x01, 0x01, 0x03, // {4, 1, 3}

				0x02, 0x02, 0x64, // 612:
				0x01, 0x06, // len=6
				0x01, 0x06, 0x01, 0x01, 0x01, 0x02, // {6, 1, 2}
			}
			expect = map[int][]int{
				413: {4, 1, 3},
				612: {6, 1, 2},
			}
			expectConsumed = 24
		)

		// execute
		var actual map[int][]int
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int][3]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x15, // len=21

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x06, // len=6
				0x01, 0x04, 0x01, 0x01, 0x01, 0x03, // {4, 1, 3}

				0x02, 0x02, 0x64, // 612:
				0x01, 0x05, // len=5
				0x01, 0x06, 0x01, 0x01, 0x00, // {6, 1, 0}
			}
			expect = map[int][3]int{
				413: {4, 1, 3},
				612: {6, 1, 0},
			}
			expectConsumed = 23
		)

		// execute
		var actual map[int][3]int
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("http.Header (map[string][]string)", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x3a, // map len=58

				0x41, 0x82, 0x0c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2d, 0x54, 0x79, 0x70, 0x65, // "Content-Type":
				0x01, 0x13, // slice len=19
				0x41, 0x82, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, // "application/json"

				0x41, 0x82, 0x09, 0x58, 0x2d, 0x41, 0x70, 0x69, 0x2d, 0x4b, 0x65, 0x79, // "X-Api-Key":
				0x01, 0x08, // slice len=8
				0x41, 0x82, 0x5, 0x31, 0x32, 0x33, 0x34, 0x35, // "12345"
			}
			expect = http.Header{
				"X-Api-Key":    []string{"12345"},
				"Content-Type": []string{"application/json"},
			}
			expectConsumed = 60
		)

		// execute
		var actual http.Header
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("meta map[int]map[int]string", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x2e, // len=46

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x10, // len=16
				0x01, 0x02, 0x01, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // 2: "JOHN"
				0x01, 0x04, 0x01, 0x04, 0x52, 0x4f, 0x53, 0x45, // 4: "ROSE"

				0x02, 0x02, 0x64, // 612:
				0x01, 0x14, // len=20
				0x01, 0x04, 0x01, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // 4: "NEPETA"
				0x01, 0x08, 0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // 8: "VRISKA"
			}
			expect = map[int]map[int]string{
				413: {
					2: "JOHN",
					4: "ROSE",
				},
				612: {
					8: "VRISKA",
					4: "NEPETA",
				},
			}
			expectConsumed = 48
		)

		// execute
		var actual map[int]map[int]string
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})
}

func Test_Dec_Map_SelfIndirection(t *testing.T) {
	t.Run("*map[string]int (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0xa0,
			}
			expect         *map[string]int
			expectConsumed = 1
		)

		var actual *map[string]int = &map[string]int{"OTHER": 1}
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*map[string]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0x01, 0x10,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0x01, 0x08,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0x01, 0x01,
			}
			expectVal      = map[string]int{"ONE": 1, "EIGHT": 8}
			expect         = &expectVal
			expectConsumed = 18
		)

		var actual *map[string]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**map[string]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0x01, 0x10,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0x01, 0x08,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0x01, 0x01,
			}
			expectVal      = map[string]int{"ONE": 1, "EIGHT": 8}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 18
		)

		var actual **map[string]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**map[string]int, but nil map[string]int part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0xb0, 0x01, 0x01,
			}
			expectPtr      *map[string]int
			expect         = &expectPtr
			expectConsumed = 3
		)

		var actual **map[string]int = ref(&map[string]int{"OTHER": 1})
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*http.Header (*map[string][]string) (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0xa0,
			}
			expect         *http.Header
			expectConsumed = 1
		)

		var actual *http.Header = &http.Header{"X-Api-Key": []string{"blah"}}
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*http.Header (*map[string][]string)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0x01, 0x3a, // map len=58

				0x41, 0x82, 0x0c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2d, 0x54, 0x79, 0x70, 0x65, // "Content-Type":
				0x01, 0x13, // slice len=19
				0x41, 0x82, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, // "application/json"

				0x41, 0x82, 0x09, 0x58, 0x2d, 0x41, 0x70, 0x69, 0x2d, 0x4b, 0x65, 0x79, // "X-Api-Key":
				0x01, 0x08, // slice len=8
				0x41, 0x82, 0x5, 0x31, 0x32, 0x33, 0x34, 0x35, // "12345"
			}
			expectVal = http.Header{
				"X-Api-Key":    []string{"12345"},
				"Content-Type": []string{"application/json"},
			}
			expect         = &expectVal
			expectConsumed = 60
		)

		var actual *http.Header
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**http.Header (**map[string][]string)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0x01, 0x3a, // map len=58

				0x41, 0x82, 0x0c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2d, 0x54, 0x79, 0x70, 0x65, // "Content-Type":
				0x01, 0x13, // slice len=19
				0x41, 0x82, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, // "application/json"

				0x41, 0x82, 0x09, 0x58, 0x2d, 0x41, 0x70, 0x69, 0x2d, 0x4b, 0x65, 0x79, // "X-Api-Key":
				0x01, 0x08, // slice len=8
				0x41, 0x82, 0x5, 0x31, 0x32, 0x33, 0x34, 0x35, // "12345"
			}
			expectVal = http.Header{
				"X-Api-Key":    []string{"12345"},
				"Content-Type": []string{"application/json"},
			}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 60
		)

		var actual **http.Header
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**http.Header, but nil http.Header part (**map[string][]string)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0xb0, 0x01, 0x01,
			}
			expectPtr      *http.Header
			expect         = &expectPtr
			expectConsumed = 3
		)

		var actual **http.Header = ref(&http.Header{"X-Api-Key": []string{"blah"}})
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})
}

func Test_Dec_Map_ValueIndirection(t *testing.T) {

	t.Run("map[string]*int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x10,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0x01, 0x08,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0x01, 0x01,
			}
			expect         = map[string]*int{"ONE": ref(1), "EIGHT": ref(8)}
			expectConsumed = 18
		)

		// execute
		var actual map[string]*int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[string]*int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0f,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0xa0,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0x01, 0x01,
			}
			expect         = map[string]*int{"ONE": ref(1), "EIGHT": nil}
			expectConsumed = 17
		)

		// execute
		var actual map[string]*int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[string]*int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0e,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0xa0,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0xa0,
			}
			expect         = map[string]*int{"ONE": nil, "EIGHT": nil}
			expectConsumed = 16
		)

		// execute
		var actual map[string]*int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*uint64, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x12,

				0x01, 0x01,
				0x04, 0x05, 0x4c, 0x56, 0x38,

				0x01, 0x08,
				0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x64,
			}
			expect         = map[int]*uint64{8: ref(uint64(10004138888888800612)), 1: ref(uint64(88888888))}
			expectConsumed = 20
		)

		// execute
		var actual map[int]*uint64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*uint64, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0a,

				0x01, 0x01,
				0x04, 0x05, 0x4c, 0x56, 0x38,

				0x01, 0x08,
				0xa0,
			}
			expect         = map[int]*uint64{8: nil, 1: ref(uint64(88888888))}
			expectConsumed = 12
		)

		// execute
		var actual map[int]*uint64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*uint64, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x06,

				0x01, 0x01,
				0xa0,

				0x01, 0x08,
				0xa0,
			}
			expect         = map[int]*uint64{8: nil, 1: nil}
			expectConsumed = 8
		)

		// execute
		var actual map[int]*uint64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]*string, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x14,

				0x00,
				0x01, 0x03, 0x54, 0x41, 0x56,

				0x01,
				0x01, 0x0b, 0x43, 0x4f, 0x4f, 0x4c, 0x20, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,
			}
			expect         = map[bool]*string{true: ref("COOL VRISKA"), false: ref("TAV")}
			expectConsumed = 22
		)

		// execute
		var actual map[bool]*string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]*string, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x10,

				0x00,
				0xa0,

				0x01,
				0x01, 0x0b, 0x43, 0x4f, 0x4f, 0x4c, 0x20, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,
			}
			expect         = map[bool]*string{true: ref("COOL VRISKA"), false: nil}
			expectConsumed = 18
		)

		// execute
		var actual map[bool]*string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]*string, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x04,

				0x00,
				0xa0,

				0x01,
				0xa0,
			}
			expect         = map[bool]*string{true: nil, false: nil}
			expectConsumed = 6
		)

		// execute
		var actual map[bool]*string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*bool, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0b, // len=11

				0x01, 0x64, 0x00, // 100: false
				0x02, 0x01, 0x9d, 0x01, // 413: true
				0x02, 0x02, 0x64, 0x01, // 612: true
			}
			expect = map[int]*bool{
				413: ref(true),
				612: ref(true),
				100: ref(false),
			}
			expectConsumed = 13
		)

		// execute
		var actual map[int]*bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*bool, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0b, // len=11

				0x01, 0x64, 0x00, // 100: false
				0x02, 0x01, 0x9d, 0x01, // 413: true
				0x02, 0x02, 0x64, 0xa0, // 612: nil
			}
			expect = map[int]*bool{
				413: ref(true),
				612: nil,
				100: ref(false),
			}
			expectConsumed = 13
		)

		// execute
		var actual map[int]*bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*bool, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0b, // len=11

				0x01, 0x64, 0xa0, // 100: nil
				0x02, 0x01, 0x9d, 0xa0, // 413: nil
				0x02, 0x02, 0x64, 0xa0, // 612: nil
			}
			expect = map[int]*bool{
				413: nil,
				612: nil,
				100: nil,
			}
			expectConsumed = 13
		)

		// execute
		var actual map[int]*bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*binary, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x1c, // len=28

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x08, // len=8
				0x01, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // "JOHN"
				0x01, 0x01, // 1

				0x02, 0x02, 0x64, // 612:
				0x01, 0x0a, // len=10
				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
			expect = map[int]*testBinary{
				413: {data: "JOHN", number: 1},
				612: {data: "VRISKA", number: 8},
			}
			expectConsumed = 30
		)

		// execute
		var actual map[int]*testBinary
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*binary, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x13, // len=19

				0x02, 0x01, 0x9d, // 413:
				0xa0, // nil

				0x02, 0x02, 0x64, // 612:
				0x01, 0x0a, // len=10
				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
			expect = map[int]*testBinary{
				413: nil,
				612: {data: "VRISKA", number: 8},
			}
			expectConsumed = 21
		)

		// execute
		var actual map[int]*testBinary
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*binary, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x08, // len=08

				0x02, 0x01, 0x9d, // 413:
				0xa0, // nil

				0x02, 0x02, 0x64, // 612:
				0xa0, // nil
			}
			expect = map[int]*testBinary{
				413: nil,
				612: nil,
			}
			expectConsumed = 10
		)

		// execute
		var actual map[int]*testBinary
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*text, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x3b, // len=59

				// 8
				0x01, 0x08,
				// "8,true,VRISKA"
				0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,

				// 100
				0x01, 0x64,
				// "100,false,NEPETA"
				0x41, 0x82, 0x10, 0x31, 0x30, 0x30, 0x2c, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41,

				// 413
				0x02, 0x01, 0x9d,
				// "413,false,JOHN"
				0x41, 0x82, 0x0e, 0x34, 0x31, 0x33, 0x2c, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x4a, 0x4f, 0x48, 0x4e,
			}
			expect = map[int]*testText{
				8:   {name: "VRISKA", value: 8, enabled: true},
				413: {name: "JOHN", enabled: false, value: 413},
				100: {name: "NEPETA", enabled: false, value: 100},
			}
			expectConsumed = 61
		)

		// execute
		var actual map[int]*testText
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*text, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x29, // len=41

				// 8
				0x01, 0x08,
				// "8,true,VRISKA"
				0x41, 0x82, 0x0d, 0x38, 0x2c, 0x74, 0x72, 0x75, 0x65, 0x2c, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,

				// 100
				0x01, 0x64,
				// nil
				0xa0,

				// 413
				0x02, 0x01, 0x9d,
				// "413,false,JOHN"
				0x41, 0x82, 0x0e, 0x34, 0x31, 0x33, 0x2c, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x4a, 0x4f, 0x48, 0x4e,
			}
			expect = map[int]*testText{
				8:   {name: "VRISKA", value: 8, enabled: true},
				413: {name: "JOHN", enabled: false, value: 413},
				100: nil,
			}
			expectConsumed = 43
		)

		// execute
		var actual map[int]*testText
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*text, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0a, // len=10

				// 8
				0x01, 0x08,
				0xa0,

				// 100
				0x01, 0x64,
				0xa0,

				// 413
				0x02, 0x01, 0x9d,
				0xa0,
			}
			expect = map[int]*testText{
				8:   nil,
				413: nil,
				100: nil,
			}
			expectConsumed = 12
		)

		// execute
		var actual map[int]*testText
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[]int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x16, // len=22

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x06, // len=6
				0x01, 0x04, 0x01, 0x01, 0x01, 0x03, // {4, 1, 3}

				0x02, 0x02, 0x64, // 612:
				0x01, 0x06, // len=6
				0x01, 0x06, 0x01, 0x01, 0x01, 0x02, // {6, 1, 2}
			}
			expect = map[int]*[]int{
				413: {4, 1, 3},
				612: {6, 1, 2},
			}
			expectConsumed = 24
		)

		// execute
		var actual map[int]*[]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[]int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0f, // len=15

				0x02, 0x01, 0x9d, // 413:
				0xa0, // nil

				0x02, 0x02, 0x64, // 612:
				0x01, 0x06, // len=6
				0x01, 0x06, 0x01, 0x01, 0x01, 0x02, // {6, 1, 2}
			}
			expect = map[int]*[]int{
				413: nil,
				612: {6, 1, 2},
			}
			expectConsumed = 17
		)

		// execute
		var actual map[int]*[]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[]int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x08, // len=8

				0x02, 0x01, 0x9d, // 413:
				0xa0, // nil

				0x02, 0x02, 0x64, // 612:
				0xa0, // nil
			}
			expect = map[int]*[]int{
				413: nil,
				612: nil,
			}
			expectConsumed = 10
		)

		// execute
		var actual map[int]*[]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[3]int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x15, // len=21

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x06, // len=6
				0x01, 0x04, 0x01, 0x01, 0x01, 0x03, // {4, 1, 3}

				0x02, 0x02, 0x64, // 612:
				0x01, 0x05, // len=5
				0x01, 0x06, 0x01, 0x01, 0x00, // {6, 1, 0}
			}
			expect = map[int]*[3]int{
				413: {4, 1, 3},
				612: {6, 1},
			}
			expectConsumed = 23
		)

		// execute
		var actual map[int]*[3]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[3]int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0e, // len=14

				0x02, 0x01, 0x9d, // 413:
				0xa0, // nil

				0x02, 0x02, 0x64, // 612:
				0x01, 0x05, // len=5
				0x01, 0x06, 0x01, 0x01, 0x00, // {6, 1, 0}
			}
			expect = map[int]*[3]int{
				413: nil,
				612: {6, 1},
			}
			expectConsumed = 16
		)

		// execute
		var actual map[int]*[3]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*[3]int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x08, // len=8

				0x02, 0x01, 0x9d, // 413:
				0xa0, // nil

				0x02, 0x02, 0x64, // 612:
				0xa0, // nil
			}
			expect = map[int]*[3]int{
				413: nil,
				612: nil,
			}
			expectConsumed = 10
		)

		// execute
		var actual map[int]*[3]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*map[int]string, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x2e, // len=46

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x10, // len=16
				0x01, 0x02, 0x01, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // 2: "JOHN"
				0x01, 0x04, 0x01, 0x04, 0x52, 0x4f, 0x53, 0x45, // 4: "ROSE"

				0x02, 0x02, 0x64, // 612:
				0x01, 0x14, // len=20
				0x01, 0x04, 0x01, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // 4: "NEPETA"
				0x01, 0x08, 0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // 8: "VRISKA"
			}
			expect = map[int]*map[int]string{
				413: {
					2: "JOHN",
					4: "ROSE",
				},
				612: {
					8: "VRISKA",
					4: "NEPETA",
				},
			}
			expectConsumed = 48
		)

		// execute
		var actual map[int]*map[int]string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*map[int]string, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x19, // len=25

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x10, // len=16
				0x01, 0x02, 0x01, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // 2: "JOHN"
				0x01, 0x04, 0x01, 0x04, 0x52, 0x4f, 0x53, 0x45, // 4: "ROSE"

				0x02, 0x02, 0x64, // 612:
				0xa0, // nil
			}
			expect = map[int]*map[int]string{
				413: {
					2: "JOHN",
					4: "ROSE",
				},
				612: nil,
			}
			expectConsumed = 27
		)

		// execute
		var actual map[int]*map[int]string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("map[int]*map[int]string, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x08, // len=8

				0x02, 0x01, 0x9d, // 413:
				0xa0, // nil

				0x02, 0x02, 0x64, // 612:
				0xa0, // nil
			}
			expect = map[int]*map[int]string{
				413: nil,
				612: nil,
			}
			expectConsumed = 10
		)

		// execute
		var actual map[int]*map[int]string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})
}
