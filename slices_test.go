package rezi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Enc_Slice_NoIndirection(t *testing.T) {
	// different types, can't rly be table driven easily

	t.Run("nil []int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  []int
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

	t.Run("[]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []int{1, 3, 4, 200, 281409}
			expect = []byte{
				0x01, 0x0c, 0x01, 0x01, 0x01, 0x03, 0x01, 0x04, 0x01, 0xc8,
				0x03, 0x04, 0x4b, 0x41,
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

	t.Run("[]uint64", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []uint64{10004138888888800612, 10004138888888800613}
			expect = []byte{
				0x01, 0x12, 0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55,
				0x64, 0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x65,
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

	t.Run("[]string", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []string{"VRISKA", "NEPETA", "TEREZI"}
			expect = []byte{
				0x01, 0x1b, // len=27

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
				0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"
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

	t.Run("[]bool", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []bool{true, true, false, true}
			expect = []byte{
				0x01, 0x04,

				0x01, 0x01, 0x00, 0x01,
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

	t.Run("[]binary", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []testBinary{
				{data: "sup", number: 1},
				{data: "VRISSY", number: 8},
			}

			expect = []byte{
				0x01, 0x17, // len=23

				0x01, 0x08, // len=8
				0x41, 0x82, 0x03, 0x73, 0x75, 0x70, // "sup"
				0x01, 0x01, // 1

				0x01, 0x0b, // len=12
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x53, 0x59, // "VRISSY"
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

	t.Run("[]map[string]bool", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []map[string]bool{
				{
					"VRISKA":   true,
					"ARANEA":   false,
					"MINDFANG": true,
				},
				{
					"NEPETA": true,
				},
				{
					"JOHN": true,
					"JADE": true,
				},
			}

			expect = []byte{
				0x01, 0x40, // len=64

				0x01, 0x20, // len=32
				0x41, 0x82, 0x06, 0x41, 0x52, 0x41, 0x4e, 0x45, 0x41, 0x00, // "ARANEA": false
				0x41, 0x82, 0x08, 0x4d, 0x49, 0x4e, 0x44, 0x46, 0x41, 0x4e, 0x47, 0x01, // "MINDFANG": true
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, 0x01, // "VRISKA": true

				0x01, 0x0a, // len=10
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, 0x01, // "NEPETA": true

				0x01, 0x10, // len=16
				0x41, 0x82, 0x04, 0x4a, 0x41, 0x44, 0x45, 0x01, // "JADE": true
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, 0x01, // "JOHN": true
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

	t.Run("meta slice [][]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = [][]int{
				{1, 2, 3},
				{8888},
			}

			expect = []byte{
				0x01, 0x0d,

				0x01, 0x06,
				0x01, 0x01,
				0x01, 0x02,
				0x01, 0x03,

				0x01, 0x03,
				0x02, 0x22, 0xb8,
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

func Test_Enc_Slice_SelfIndirection(t *testing.T) {
	t.Run("*[]int (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *[]int
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

	t.Run("*[]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = []int{1, 2, 8, 8}
			input    = &inputVal
			expect   = []byte{
				0x01, 0x08, // len=8s

				0x01, 0x01, // 1
				0x01, 0x02, // 2
				0x01, 0x08, // 8
				0x01, 0x08, // 8
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**[]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputVal = []int{1, 2, 8, 8}
			inputPtr = &inputVal
			input    = &inputPtr
			expect   = []byte{
				0x01, 0x08, // len=8s

				0x01, 0x01, // 1
				0x01, 0x02, // 2
				0x01, 0x08, // 8
				0x01, 0x08, // 8
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	t.Run("**[]int, but nil []int part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			ptr    *[]int
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

func Test_Enc_Slice_ValueIndirection(t *testing.T) {
	t.Run("[]*int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*int{ref(1), ref(3), ref(4), ref(200), ref(281409)}
			expect = []byte{
				0x01, 0x0c,

				0x01, 0x01,
				0x01, 0x03,
				0x01, 0x04,
				0x01, 0xc8,
				0x03, 0x04, 0x4b, 0x41,
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

	t.Run("[]*int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*int{ref(1), ref(3), ref(4), ref(200), nil}
			expect = []byte{
				0x01, 0x09,

				0x01, 0x01,
				0x01, 0x03,
				0x01, 0x04,
				0x01, 0xc8,
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

	t.Run("[]*int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*int{nil, nil}
			expect = []byte{
				0x01, 0x02,

				0xa0,
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

	t.Run("[]*uint64, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*uint64{ref(uint64(10004138888888800612)), ref(uint64(10004138888888800613))}
			expect = []byte{
				0x01, 0x12,

				0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x64,
				0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x65,
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

	t.Run("[]*uint64, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*uint64{ref(uint64(10004138888888800612)), nil}
			expect = []byte{
				0x01, 0x0a,

				0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x64,
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

	t.Run("[]*uint64, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*uint64{nil, nil}
			expect = []byte{
				0x01, 0x02,

				0xa0,
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

	t.Run("[]*string, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*string{ref("VRISKA"), ref("NEPETA"), ref("TEREZI")}
			expect = []byte{
				0x01, 0x1b, // len=27

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4B, 0x41, // "VRISKA"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
				0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"
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

	t.Run("[]*string, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*string{ref("VRISKA"), nil, ref("TEREZI")}
			expect = []byte{
				0x01, 0x13, // len=19

				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4B, 0x41, // "VRISKA"
				0xa0,                                                 // nil
				0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"
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

	t.Run("[]*string, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*string{nil, nil}
			expect = []byte{
				0x01, 0x02,

				0xa0,
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

	t.Run("[]*bool, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*bool{ref(true), ref(true), ref(false), ref(true)}
			expect = []byte{
				0x01, 0x04,

				0x01, 0x01, 0x00, 0x01,
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

	t.Run("[]*bool, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*bool{ref(true), nil, ref(false), ref(true)}
			expect = []byte{
				0x01, 0x04,

				0x01, 0xa0, 0x00, 0x01,
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

	t.Run("[]*bool, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*bool{nil, nil, nil, nil}
			expect = []byte{
				0x01, 0x04,

				0xa0, 0xa0, 0xa0, 0xa0,
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

	t.Run("[]*binary, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []*testBinary{
				{data: "sup", number: 1},
				{data: "VRISSY", number: 8},
			}
			expect = []byte{
				0x01, 0x17, // len=23

				0x01, 0x08, // len=8
				0x41, 0x82, 0x03, 0x73, 0x75, 0x70, // "sup"
				0x01, 0x01, // 1

				0x01, 0x0b, // len=11
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x53, 0x59, // "VRISSY"
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

	t.Run("[]*binary, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []*testBinary{
				{data: "sup", number: 1},
				nil,
			}
			expect = []byte{
				0x01, 0x0b, // len=11

				0x01, 0x08, // len=8
				0x41, 0x82, 0x03, 0x73, 0x75, 0x70, // "sup"
				0x01, 0x01, // 1

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

	t.Run("[]*binary, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*testBinary{nil, nil}
			expect = []byte{
				0x01, 0x02,

				0xa0,
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

	t.Run("[]*map[string]bool, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []*map[string]bool{
				{
					"VRISKA":   true,
					"ARANEA":   false,
					"MINDFANG": true,
				},
				{
					"NEPETA": true,
				},
				{
					"JOHN": true,
					"JADE": true,
				},
			}
			expect = []byte{
				0x01, 0x40, // len=64

				0x01, 0x20, // len=32
				0x41, 0x82, 0x06, 0x41, 0x52, 0x41, 0x4e, 0x45, 0x41, 0x00, // "ARANEA": false
				0x41, 0x82, 0x08, 0x4d, 0x49, 0x4e, 0x44, 0x46, 0x41, 0x4e, 0x47, 0x01, // "MINDFANG": true
				0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, 0x01, // "VRISKA": true

				0x01, 0x0a, // len=10
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, 0x01, // "NEPETA": true

				0x01, 0x10, // len=16
				0x41, 0x82, 0x04, 0x4a, 0x41, 0x44, 0x45, 0x01, // "JADE": true
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, 0x01, // "JOHN": true
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

	t.Run("[]*map[string]bool, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []*map[string]bool{
				nil,
				{
					"NEPETA": true,
				},
				{
					"JOHN": true,
					"JADE": true,
				},
			}
			expect = []byte{
				0x01, 0x1f, // len=31

				0xa0, // nil

				0x01, 0x0a, // len=10
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, 0x01, // "NEPETA": true

				0x01, 0x10, // len=16
				0x41, 0x82, 0x04, 0x4a, 0x41, 0x44, 0x45, 0x01, // "JADE": true
				0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, 0x01, // "JOHN": true
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

	t.Run("[]*map[string]bool, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []*map[string]bool{
				nil, nil, nil,
			}
			expect = []byte{
				0x01, 0x03, // len=3

				0xa0, 0xa0, 0xa0,
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

	t.Run("[]*[]int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*[]int{{8, 8, 16, 24}, {1, 2, 3}, {10, 9, 8}}
			expect = []byte{
				0x01, 0x1a, // len=26

				0x01, 0x08, // len=8
				0x01, 0x08,
				0x01, 0x08,
				0x01, 0x10,
				0x01, 0x18,

				0x01, 0x06, // len=6
				0x01, 0x01,
				0x01, 0x02,
				0x01, 0x03,

				0x01, 0x06, // len=6
				0x01, 0x0a,
				0x01, 0x09,
				0x01, 0x08,
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

	t.Run("[]*[]int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*[]int{{8, 8, 16, 24}, nil, {10, 9, 8}}
			expect = []byte{
				0x01, 0x13, // len=19

				0x01, 0x08, // len=8
				0x01, 0x08,
				0x01, 0x08,
				0x01, 0x10,
				0x01, 0x18,

				0xa0, // nil

				0x01, 0x06, // len=6
				0x01, 0x0a,
				0x01, 0x09,
				0x01, 0x08,
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

	t.Run("[]*[]int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  = []*[]int{nil, nil, nil}
			expect = []byte{
				0x01, 0x03,

				0xa0,
				0xa0,
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
}

func Test_Dec_Slice_NoIndirection(t *testing.T) {
	t.Run("nil []int (implicit nil)", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x80,
			}
			expectConsumed = 1
		)

		// execute
		actual := []int{1, 2} // start with a value so we can check it is set to nil
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Nil(actual)
	})

	t.Run("nil []int (explicit nil)", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0xa0,
			}
			expectConsumed = 1
		)

		// execute
		actual := []int{1, 2} // start with a value so we can check it is set to nil
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Nil(actual)
	})

	t.Run("[]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0c, 0x01, 0x01, 0x01, 0x03, 0x01, 0x04, 0x01, 0xc8,
				0x03, 0x04, 0x4b, 0x41,
			}
			expect         = []int{1, 3, 4, 200, 281409}
			expectConsumed = 14
		)

		// execute
		var actual []int
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]uint64", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x12, 0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55,
				0x64, 0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x65,
			}
			expect         = []uint64{10004138888888800612, 10004138888888800613}
			expectConsumed = 20
		)

		// execute
		var actual []uint64
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]string", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x18, 0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4B, 0x41,
				0x01, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, 0x01, 0x06,
				0x54, 0x45, 0x52, 0x45, 0x5a, 0x49,
			}
			expect         = []string{"VRISKA", "NEPETA", "TEREZI"}
			expectConsumed = 26
		)

		// execute
		var actual []string
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]binary", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x15,

				0x01, 0x07,
				0x01, 0x03, 0x73, 0x75, 0x70,
				0x01, 0x01,

				0x01, 0x0a,
				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x53, 0x59,
				0x01, 0x08,
			}
			expect = []testBinary{
				{data: "sup", number: 1},
				{data: "VRISSY", number: 8},
			}
			expectConsumed = 23
		)

		// execute
		var actual []testBinary
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]map[string]bool", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x3a, // len=58

				0x01, 0x1d, // len=29
				0x01, 0x06, 0x41, 0x52, 0x41, 0x4e, 0x45, 0x41, 0x00, // "ARANEA": false
				0x01, 0x08, 0x4d, 0x49, 0x4e, 0x44, 0x46, 0x41, 0x4e, 0x47, 0x01, // "MINDFANG": true
				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, 0x01, // "VRISKA": true

				0x01, 0x09, // len=9
				0x01, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, 0x01, // "NEPETA": true

				0x01, 0x0e, // len=14
				0x01, 0x04, 0x4a, 0x41, 0x44, 0x45, 0x01, // "JADE": true
				0x01, 0x04, 0x4a, 0x4f, 0x48, 0x4e, 0x01, // "JOHN": true
			}
			expect = []map[string]bool{
				{
					"VRISKA":   true,
					"ARANEA":   false,
					"MINDFANG": true,
				},
				{
					"NEPETA": true,
				},
				{
					"JOHN": true,
					"JADE": true,
				},
			}
			expectConsumed = 60
		)

		// execute
		var actual []map[string]bool
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("meta slice [][]int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0d,

				0x01, 0x06,
				0x01, 0x01,
				0x01, 0x02,
				0x01, 0x03,

				0x01, 0x03,
				0x02, 0x22, 0xb8,
			}
			expect = [][]int{
				{1, 2, 3},
				{8888},
			}
			expectConsumed = 15
		)

		// execute
		var actual [][]int
		consumed, err := Dec(input, &actual)

		// assert
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})
}

func Test_Dec_Slice_SelfIndirection(t *testing.T) {
	t.Run("*[]int (nil)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0xa0,
			}
			expect         *[]int
			expectConsumed = 1
		)

		var actual *[]int = &[]int{1, 2}
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("*[]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0x01, 0x08, // len=8s

				0x01, 0x01, // 1
				0x01, 0x02, // 2
				0x01, 0x08, // 8
				0x01, 0x08, // 8
			}
			expectVal      = []int{1, 2, 8, 8}
			expect         = &expectVal
			expectConsumed = 10
		)

		var actual *[]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**[]int", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0x01, 0x08, // len=8s

				0x01, 0x01, // 1
				0x01, 0x02, // 2
				0x01, 0x08, // 8
				0x01, 0x08, // 8
			}
			expectVal      = []int{1, 2, 8, 8}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 10
		)

		var actual **[]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("**[]int, but nil []int part", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = []byte{
				0xb0, 0x01, 0x01,
			}
			expectPtr      *[]int
			expect         = &expectPtr
			expectConsumed = 3
		)

		var actual **[]int = ref(&[]int{1, 2, 3})
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})
}

func Test_Dec_Slice_ValueIndirection(t *testing.T) {

	t.Run("[]*int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0c,

				0x01, 0x01,
				0x01, 0x03,
				0x01, 0x04,
				0x01, 0xc8,
				0x03, 0x04, 0x4b, 0x41,
			}
			expect         = []*int{ref(1), ref(3), ref(4), ref(200), ref(281409)}
			expectConsumed = 14
		)

		// execute
		var actual []*int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0b,

				0x01, 0x01,
				0xa0,
				0x01, 0x04,
				0x01, 0xc8,
				0x03, 0x04, 0x4b, 0x41,
			}
			expect         = []*int{ref(1), nil, ref(4), ref(200), ref(281409)}
			expectConsumed = 13
		)

		// execute
		var actual []*int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x05,

				0xa0,
				0xa0,
				0xa0,
				0xa0,
				0xa0,
			}
			expect         = []*int{nil, nil, nil, nil, nil}
			expectConsumed = 7
		)

		// execute
		var actual []*int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*uint64, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x12,

				0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x64,
				0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x65,
			}
			expect         = []*uint64{ref(uint64(10004138888888800612)), ref(uint64(10004138888888800613))}
			expectConsumed = 20
		)

		// execute
		var actual []*uint64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*uint64, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0a,

				0xa0,
				0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x65,
			}
			expect         = []*uint64{nil, ref(uint64(10004138888888800613))}
			expectConsumed = 12
		)

		// execute
		var actual []*uint64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*uint64, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x02,

				0xa0,
				0xa0,
			}
			expect         = []*uint64{nil, nil}
			expectConsumed = 4
		)

		// execute
		var actual []*uint64
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*string, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x18,

				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4B, 0x41,
				0x01, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41,
				0x01, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49,
			}
			expect         = []*string{ref("VRISKA"), ref("NEPETA"), ref("TEREZI")}
			expectConsumed = 26
		)

		// execute
		var actual []*string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*string, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x11,

				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4B, 0x41,
				0x01, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41,
				0xa0,
			}
			expect         = []*string{ref("VRISKA"), ref("NEPETA"), nil}
			expectConsumed = 19
		)

		// execute
		var actual []*string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*string, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x03,

				0xa0,
				0xa0,
				0xa0,
			}
			expect         = []*string{nil, nil, nil}
			expectConsumed = 5
		)

		// execute
		var actual []*string
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*bool, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x04,

				0x01, 0x01, 0x00, 0x01,
			}
			expect         = []*bool{ref(true), ref(true), ref(false), ref(true)}
			expectConsumed = 6
		)

		// execute
		var actual []*bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*bool, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x04,

				0x01, 0xa0, 0x00, 0x01,
			}
			expect         = []*bool{ref(true), nil, ref(false), ref(true)}
			expectConsumed = 6
		)

		// execute
		var actual []*bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*bool, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x04,

				0xa0, 0xa0, 0xa0, 0xa0,
			}
			expect         = []*bool{nil, nil, nil, nil}
			expectConsumed = 6
		)

		// execute
		var actual []*bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*binary, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x15,

				0x01, 0x07,
				0x01, 0x03, 0x73, 0x75, 0x70,
				0x01, 0x01,

				0x01, 0x0a,
				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x53, 0x59,
				0x01, 0x08,
			}
			expect = []*testBinary{
				{data: "sup", number: 1},
				{data: "VRISSY", number: 8},
			}
			expectConsumed = 23
		)

		// execute
		var actual []*testBinary
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*binary, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x0a,

				0x01, 0x07,
				0x01, 0x03, 0x73, 0x75, 0x70,
				0x01, 0x01,

				0xa0,
			}
			expect = []*testBinary{
				{data: "sup", number: 1},
				nil,
			}
			expectConsumed = 12
		)

		// execute
		var actual []*testBinary
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*binary, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x02,

				0xa0,
				0xa0,
			}
			expect = []*testBinary{
				nil,
				nil,
			}
			expectConsumed = 4
		)

		// execute
		var actual []*testBinary
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*map[string]bool, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x3a, // len=58

				0x01, 0x1d, // len=29
				0x01, 0x06, 0x41, 0x52, 0x41, 0x4e, 0x45, 0x41, 0x00, // "ARANEA": false
				0x01, 0x08, 0x4d, 0x49, 0x4e, 0x44, 0x46, 0x41, 0x4e, 0x47, 0x01, // "MINDFANG": true
				0x01, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, 0x01, // "VRISKA": true

				0x01, 0x09, // len=9
				0x01, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, 0x01, // "NEPETA": true

				0x01, 0x0e, // len=14
				0x01, 0x04, 0x4a, 0x41, 0x44, 0x45, 0x01, // "JADE": true
				0x01, 0x04, 0x4a, 0x4f, 0x48, 0x4e, 0x01, // "JOHN": true
			}
			expect = []*map[string]bool{
				{
					"VRISKA":   true,
					"ARANEA":   false,
					"MINDFANG": true,
				},
				{
					"NEPETA": true,
				},
				{
					"JOHN": true,
					"JADE": true,
				},
			}
			expectConsumed = 60
		)

		// execute
		var actual []*map[string]bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*map[string]bool, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x1c, // len=28

				0xa0,

				0x01, 0x09, // len=9
				0x01, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, 0x01, // "NEPETA": true

				0x01, 0x0e, // len=14
				0x01, 0x04, 0x4a, 0x41, 0x44, 0x45, 0x01, // "JADE": true
				0x01, 0x04, 0x4a, 0x4f, 0x48, 0x4e, 0x01, // "JOHN": true
			}
			expect = []*map[string]bool{
				nil,
				{
					"NEPETA": true,
				},
				{
					"JOHN": true,
					"JADE": true,
				},
			}
			expectConsumed = 30
		)

		// execute
		var actual []*map[string]bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*map[string]bool, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x03, // len=3

				0xa0,
				0xa0,
				0xa0,
			}
			expect = []*map[string]bool{
				nil,
				nil,
				nil,
			}
			expectConsumed = 5
		)

		// execute
		var actual []*map[string]bool
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*[]int, all non-nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x1a, // len=26

				0x01, 0x08, // len=8
				0x01, 0x08,
				0x01, 0x08,
				0x01, 0x10,
				0x01, 0x18,

				0x01, 0x06, // len=6
				0x01, 0x01,
				0x01, 0x02,
				0x01, 0x03,

				0x01, 0x06, // len=6
				0x01, 0x0a,
				0x01, 0x09,
				0x01, 0x08,
			}
			expect         = []*[]int{{8, 8, 16, 24}, {1, 2, 3}, {10, 9, 8}}
			expectConsumed = 28
		)

		// execute
		var actual []*[]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*[]int, one nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x13, // len=19

				0x01, 0x08, // len=8
				0x01, 0x08,
				0x01, 0x08,
				0x01, 0x10,
				0x01, 0x18,

				0xa0,

				0x01, 0x06, // len=6
				0x01, 0x0a,
				0x01, 0x09,
				0x01, 0x08,
			}
			expect         = []*[]int{{8, 8, 16, 24}, nil, {10, 9, 8}}
			expectConsumed = 21
		)

		// execute
		var actual []*[]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})

	t.Run("[]*[]int, all nil", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input = []byte{
				0x01, 0x03, // len=19

				0xa0,
				0xa0,
				0xa0,
			}
			expect         = []*[]int{nil, nil, nil}
			expectConsumed = 5
		)

		// execute
		var actual []*[]int
		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		// assert
		assert.Equal(expectConsumed, consumed)
		assert.Equal(expect, actual)
	})
}
