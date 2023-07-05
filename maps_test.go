package rezi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Enc_Map(t *testing.T) {
	// different types, can't easily be table-driven

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
		actual := Enc(input)

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[string]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[string]int{"ONE": 1, "EIGHT": 8}
			expect = []byte{
				0x01, 0x10,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0x01, 0x08,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0x01, 0x01,
			}
		)

		// execute
		actual := Enc(input)

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
		actual := Enc(input)

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]string", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[bool]string{true: "COOL VRISKA", false: "TAV"}
			expect = []byte{
				0x01, 0x14,

				0x00,
				0x01, 0x03, 0x54, 0x41, 0x56,

				0x01,
				0x01, 0x0b, 0x43, 0x4f, 0x4f, 0x4c, 0x20, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,
			}
		)

		// execute
		actual := Enc(input)

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
		)

		// execute
		actual := Enc(input)

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
		actual := Enc(input)

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
		actual := Enc(input)

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
		)

		// execute
		actual := Enc(input)

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("nil *map[string]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *map[string]int
			expect = []byte{
				0xa0,
			}
		)

		actual := Enc(input)

		assert.Equal(expect, actual)
	})

	t.Run("*map[string]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = map[string]int{"ONE": 1, "EIGHT": 8}
			input    = &inputVal
			expect   = []byte{
				0x01, 0x10,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0x01, 0x08,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0x01, 0x01,
			}
		)

		actual := Enc(input)

		assert.Equal(expect, actual)
	})

	t.Run("**map[string]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = map[string]int{"ONE": 1, "EIGHT": 8}
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{
				0x01, 0x10,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0x01, 0x08,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0x01, 0x01,
			}
		)

		actual := Enc(input)

		assert.Equal(expect, actual)
	})

	t.Run("**map[string]int, but nil map[string]int part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *map[string]int
			input  = &ptr
			expect = []byte{0xb0, 0x01, 0x01}
		)

		actual := Enc(input)

		assert.Equal(expect, actual)
	})

	t.Run("map[string]*int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[string]*int{"ONE": ref(1), "EIGHT": ref(8)}
			expect = []byte{
				0x01, 0x10,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0x01, 0x08,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0x01, 0x01,
			}
		)

		// execute
		actual := Enc(input)

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[string]*int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[string]*int{"ONE": ref(1), "EIGHT": nil}
			expect = []byte{
				0x01, 0x0f,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0xa0,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0x01, 0x01,
			}
		)

		// execute
		actual := Enc(input)

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[string]*int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[string]*int{"ONE": nil, "EIGHT": nil}
			expect = []byte{
				0x01, 0x0e,

				0x01, 0x05, 0x45, 0x49, 0x47, 0x48, 0x54,
				0xa0,

				0x01, 0x03, 0x4f, 0x4e, 0x45,
				0xa0,
			}
		)

		// execute
		actual := Enc(input)

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
		actual := Enc(input)

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
		actual := Enc(input)

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
		actual := Enc(input)

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]*string, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[bool]*string{true: ref("COOL VRISKA"), false: ref("TAV")}
			expect = []byte{
				0x01, 0x14,

				0x00,
				0x01, 0x03, 0x54, 0x41, 0x56,

				0x01,
				0x01, 0x0b, 0x43, 0x4f, 0x4f, 0x4c, 0x20, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,
			}
		)

		// execute
		actual := Enc(input)

		// assert
		assert.Equal(expect, actual)
	})

	t.Run("map[bool]*string, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = map[bool]*string{true: ref("COOL VRISKA"), false: nil}
			expect = []byte{
				0x01, 0x10,

				0x00,
				0xa0,

				0x01,
				0x01, 0x0b, 0x43, 0x4f, 0x4f, 0x4c, 0x20, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41,
			}
		)

		// execute
		actual := Enc(input)

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
		actual := Enc(input)

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
		)

		// execute
		actual := Enc(input)

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
				0x01, 0x13, // len=19

				0x02, 0x01, 0x9d, // 413:
				0xa0, // nil

				0x02, 0x02, 0x64, // 612:
				0x01, 0x0a, // len=10
				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x01, 0x08, // 8
			}
		)

		// execute
		actual := Enc(input)

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
		actual := Enc(input)

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
		actual := Enc(input)

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
		actual := Enc(input)

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
		actual := Enc(input)

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
		actual := Enc(input)

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
		actual := Enc(input)

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
		actual := Enc(input)

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
		)

		// execute
		actual := Enc(input)

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
				0x01, 0x19, // len=25

				0x02, 0x01, 0x9d, // 413:
				0x01, 0x10, // len=16
				0x01, 0x02, 0x01, 0x04, 0x4a, 0x4f, 0x48, 0x4e, // 2: "JOHN"
				0x01, 0x04, 0x01, 0x04, 0x52, 0x4f, 0x53, 0x45, // 4: "ROSE"

				0x02, 0x02, 0x64, // 612:
				0xa0, // nil
			}
		)

		// execute
		actual := Enc(input)

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
		actual := Enc(input)

		// assert
		assert.Equal(expect, actual)
	})
}

func Test_Dec_Map(t *testing.T) {
	// different types, can't easily be table-driven

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
}
