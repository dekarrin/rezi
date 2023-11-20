package rezi

// Put tests of array cases for slices.go in this file.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Enc_Array_NoIndirection(t *testing.T) {
	// different types, can't rly be table driven easily

	t.Run("empty []int", func(t *testing.T) {
		// setup
		assert := assert.New(t)
		var (
			input  [3]int
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

	// t.Run("[]int", func(t *testing.T) {
	// 	// setup
	// 	assert := assert.New(t)
	// 	var (
	// 		input  = []int{1, 3, 4, 200, 281409}
	// 		expect = []byte{
	// 			0x01, 0x0c, 0x01, 0x01, 0x01, 0x03, 0x01, 0x04, 0x01, 0xc8,
	// 			0x03, 0x04, 0x4b, 0x41,
	// 		}
	// 	)

	// 	// execute
	// 	actual, err := Enc(input)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	// assert
	// 	assert.Equal(expect, actual)
	// })

	// t.Run("[]uint64", func(t *testing.T) {
	// 	// setup
	// 	assert := assert.New(t)
	// 	var (
	// 		input  = []uint64{10004138888888800612, 10004138888888800613}
	// 		expect = []byte{
	// 			0x01, 0x12, 0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55,
	// 			0x64, 0x08, 0x8a, 0xd5, 0xd7, 0x50, 0xb3, 0xe3, 0x55, 0x65,
	// 		}
	// 	)

	// 	// execute
	// 	actual, err := Enc(input)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	// assert
	// 	assert.Equal(expect, actual)
	// })

	// t.Run("[]float64", func(t *testing.T) {
	// 	// setup
	// 	assert := assert.New(t)
	// 	var (
	// 		input  = []float64{-2.02499999999999991118215802999, 256.01220703125, -1.0, math.Inf(0)}
	// 		expect = []byte{
	// 			0x01, 0x14, // len=20

	// 			0x88, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, // -2.02499999999999991118215802999
	// 			0x04, 0xc0, 0x70, 0x00, 0x32, // 256.01220703125
	// 			0x82, 0x3f, 0xf0, // -1.0
	// 			0x02, 0x7f, 0xf0, // +Inf
	// 		}
	// 	)

	// 	// execute
	// 	actual, err := Enc(input)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	// assert
	// 	assert.Equal(expect, actual)
	// })

	// t.Run("[]float32", func(t *testing.T) {
	// 	// setup
	// 	assert := assert.New(t)
	// 	var (
	// 		input  = []float32{8.38218975067138671875, 256.01220703125, -1.0, float32(math.Inf(-1))}
	// 		expect = []byte{
	// 			0x01, 0x11, // len=17

	// 			0x05, 0xc0, 0x20, 0xc3, 0xae, 0x60, // 8.38218975067138671875
	// 			0x04, 0xc0, 0x70, 0x00, 0x32, // 256.01220703125
	// 			0x82, 0x3f, 0xf0, // -1.0
	// 			0x82, 0x7f, 0xf0, // -Inf
	// 		}
	// 	)

	// 	// execute
	// 	actual, err := Enc(input)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	// assert
	// 	assert.Equal(expect, actual)
	// })

	// t.Run("[]string", func(t *testing.T) {
	// 	// setup
	// 	assert := assert.New(t)
	// 	var (
	// 		input  = []string{"VRISKA", "NEPETA", "TEREZI"}
	// 		expect = []byte{
	// 			0x01, 0x1b, // len=27

	// 			0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, // "VRISKA"
	// 			0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
	// 			0x41, 0x82, 0x06, 0x54, 0x45, 0x52, 0x45, 0x5a, 0x49, // "TEREZI"
	// 		}
	// 	)

	// 	// execute
	// 	actual, err := Enc(input)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	// assert
	// 	assert.Equal(expect, actual)
	// })

	// t.Run("[]bool", func(t *testing.T) {
	// 	// setup
	// 	assert := assert.New(t)
	// 	var (
	// 		input  = []bool{true, true, false, true}
	// 		expect = []byte{
	// 			0x01, 0x04,

	// 			0x01, 0x01, 0x00, 0x01,
	// 		}
	// 	)

	// 	// execute
	// 	actual, err := Enc(input)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	// assert
	// 	assert.Equal(expect, actual)
	// })

	// t.Run("[]complex128", func(t *testing.T) {
	// 	// setup
	// 	assert := assert.New(t)
	// 	var (
	// 		input = []complex128{
	// 			complex128(2.02499999999999991118215802999 + 1.0i),
	// 			complex128(0.0 + 0.0i),
	// 			complex128(8.0 + 8.0i),
	// 		}
	// 		expect = []byte{
	// 			0x01, 0x19, // len=25

	// 			0x41, 0x80, 0x0c, 0x08, 0x40, 0x00, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x02, 0x3f, 0xf0, // 2.02499999999999991118215802999 + 1.0i
	// 			0x00,                                                 // 0.0+0.0i
	// 			0x41, 0x80, 0x06, 0x02, 0x40, 0x20, 0x02, 0x40, 0x20, // 8.0+8.0i
	// 		}
	// 	)

	// 	// execute
	// 	actual, err := Enc(input)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	// assert
	// 	assert.Equal(expect, actual)
	// })

	// t.Run("[]binary", func(t *testing.T) {
	// 	// setup
	// 	assert := assert.New(t)
	// 	var (
	// 		input = []testBinary{
	// 			{data: "sup", number: 1},
	// 			{data: "VRISSY", number: 8},
	// 		}

	// 		expect = []byte{
	// 			0x01, 0x17, // len=23

	// 			0x01, 0x08, // len=8
	// 			0x41, 0x82, 0x03, 0x73, 0x75, 0x70, // "sup"
	// 			0x01, 0x01, // 1

	// 			0x01, 0x0b, // len=12
	// 			0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x53, 0x59, // "VRISSY"
	// 			0x01, 0x08, // 8
	// 		}
	// 	)

	// 	// execute
	// 	actual, err := Enc(input)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	// assert
	// 	assert.Equal(expect, actual)
	// })

	// t.Run("[]map[string]bool", func(t *testing.T) {
	// 	// setup
	// 	assert := assert.New(t)
	// 	var (
	// 		input = []map[string]bool{
	// 			{
	// 				"VRISKA":   true,
	// 				"ARANEA":   false,
	// 				"MINDFANG": true,
	// 			},
	// 			{
	// 				"NEPETA": true,
	// 			},
	// 			{
	// 				"JOHN": true,
	// 				"JADE": true,
	// 			},
	// 		}

	// 		expect = []byte{
	// 			0x01, 0x40, // len=64

	// 			0x01, 0x20, // len=32
	// 			0x41, 0x82, 0x06, 0x41, 0x52, 0x41, 0x4e, 0x45, 0x41, 0x00, // "ARANEA": false
	// 			0x41, 0x82, 0x08, 0x4d, 0x49, 0x4e, 0x44, 0x46, 0x41, 0x4e, 0x47, 0x01, // "MINDFANG": true
	// 			0x41, 0x82, 0x06, 0x56, 0x52, 0x49, 0x53, 0x4b, 0x41, 0x01, // "VRISKA": true

	// 			0x01, 0x0a, // len=10
	// 			0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, 0x01, // "NEPETA": true

	// 			0x01, 0x10, // len=16
	// 			0x41, 0x82, 0x04, 0x4a, 0x41, 0x44, 0x45, 0x01, // "JADE": true
	// 			0x41, 0x82, 0x04, 0x4a, 0x4f, 0x48, 0x4e, 0x01, // "JOHN": true
	// 		}
	// 	)

	// 	// execute
	// 	actual, err := Enc(input)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	// assert
	// 	assert.Equal(expect, actual)
	// })

	// t.Run("meta slice [][]int", func(t *testing.T) {
	// 	// setup
	// 	assert := assert.New(t)
	// 	var (
	// 		input = [][]int{
	// 			{1, 2, 3},
	// 			{8888},
	// 		}

	// 		expect = []byte{
	// 			0x01, 0x0d,

	// 			0x01, 0x06,
	// 			0x01, 0x01,
	// 			0x01, 0x02,
	// 			0x01, 0x03,

	// 			0x01, 0x03,
	// 			0x02, 0x22, 0xb8,
	// 		}
	// 	)

	// 	// execute
	// 	actual, err := Enc(input)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	// assert
	// 	assert.Equal(expect, actual)
	// })
}
