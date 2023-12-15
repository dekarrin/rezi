package rezi

// putting this in rezi_test and not rezi because we need to create a public
// type in it and we don't want to pollute the rezi package with it.

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStructEmpty struct{}

type testStructOneMember struct {
	Value int
}

type testStructMultiMember struct {
	Value int
	Name  string
}

type testStructWithUnexported struct {
	Value      int
	unexported float32
}

type testStructWithUnexportedCaseDistinguished struct {
	Value int
	value float32
}

type testStructOnlyUnexported struct {
	value int
	name  string
}

type testStructManyFields struct {
	Name    string
	Factor  float64
	Value   int
	Enabled bool

	hidden  chan int
	inc     int
	enabled *sync.Mutex
}

func Test_Enc_Struct(t *testing.T) {
	// we require an exported struct in order to test embedded struct fields.
	// we will declare it and other struct types it is embedded in here in this
	// function to avoid adding a new exported type to the rezi package.
	type TestStructToEmbed struct {
		Value int
	}
	type testStructWithEmbedded struct {
		TestStructToEmbed
		Name string
	}

	type testStructWithEmbeddedOverlap struct {
		TestStructToEmbed
		Name  string
		Value float64
	}

	// normal value test
	t.Run("empty struct", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = testStructEmpty{}
			expect = []byte{0x00}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, filled
	t.Run("*(empty struct)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = &testStructEmpty{}
			expect = []byte{0x00}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, nil
	t.Run("*(empty struct), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testStructEmpty
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, filled
	t.Run("**(empty struct)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr = &testStructEmpty{}
			input    = &inputPtr
			expect   = []byte{0x00}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, nil at first level
	t.Run("**(empty struct), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr *testStructEmpty
			input    = &inputPtr
			expect   = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// normal value test
	t.Run("one member", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = testStructOneMember{Value: 4}
			expect = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, filled
	t.Run("*(one member)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = &testStructOneMember{Value: 4}
			expect = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, nil
	t.Run("*(one member), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testStructOneMember
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, filled
	t.Run("**(one member)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr = &testStructOneMember{Value: 4}
			input    = &inputPtr
			expect   = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, nil at first level
	t.Run("**(one member), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr *testStructOneMember
			input    = &inputPtr
			expect   = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// normal value test
	t.Run("multi member", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = testStructMultiMember{Value: 4, Name: "NEPETA"}
			expect = []byte{
				0x01, 0x1a, // len=26

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, filled
	t.Run("*(multi member)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = &testStructMultiMember{Value: 4, Name: "NEPETA"}
			expect = []byte{
				0x01, 0x1a, // len=26

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, nil
	t.Run("*(multi member), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testStructMultiMember
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, filled
	t.Run("**(multi member)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr = &testStructMultiMember{Value: 4, Name: "NEPETA"}
			input    = &inputPtr
			expect   = []byte{
				0x01, 0x1a, // len=26

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, nil at first level
	t.Run("**(multi member), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr *testStructMultiMember
			input    = &inputPtr
			expect   = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// normal value test
	t.Run("with unexported", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = testStructWithUnexported{Value: 4, unexported: 9}
			expect = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, filled
	t.Run("*(with unexported)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = &testStructWithUnexported{Value: 4, unexported: 9}
			expect = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, nil
	t.Run("*(with unexported), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testStructWithUnexported
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, filled
	t.Run("**(with unexported)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr = &testStructWithUnexported{Value: 4, unexported: 9}
			input    = &inputPtr
			expect   = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, nil at first level
	t.Run("**(with unexported), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr *testStructWithUnexported
			input    = &inputPtr
			expect   = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// normal value test
	t.Run("with unexported case distinguished", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = testStructWithUnexportedCaseDistinguished{Value: 4, value: 3.2}
			expect = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, filled
	t.Run("*(with unexported case distinguished)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = &testStructWithUnexportedCaseDistinguished{Value: 4, value: 3.2}
			expect = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, nil
	t.Run("*(with unexported case distinguished), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testStructWithUnexportedCaseDistinguished
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, filled
	t.Run("**(with unexported case distinguished)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr = &testStructWithUnexportedCaseDistinguished{Value: 4, value: 3.2}
			input    = &inputPtr
			expect   = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, nil at first level
	t.Run("**(with unexported case distinguished), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr *testStructWithUnexportedCaseDistinguished
			input    = &inputPtr
			expect   = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// normal value test
	t.Run("only unexported", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = testStructOnlyUnexported{value: 3, name: "test"}
			expect = []byte{0x00}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, filled
	t.Run("*(only unexported)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  = &testStructOnlyUnexported{value: 3, name: "test"}
			expect = []byte{0x00}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, nil
	t.Run("*(only unexported), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testStructOnlyUnexported
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, filled
	t.Run("**(only unexported)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr = &testStructOnlyUnexported{value: 3, name: "test"}
			input    = &inputPtr
			expect   = []byte{0x00}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, nil at first level
	t.Run("**(only unexported), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr *testStructOnlyUnexported
			input    = &inputPtr
			expect   = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// normal value test
	t.Run("many fields", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = testStructManyFields{
				Value:   413,
				Name:    "Rose Lalonde",
				Enabled: true,
				Factor:  8.25,

				hidden:  make(chan int),
				inc:     17,
				enabled: &sync.Mutex{},
			}
			expect = []byte{
				0x01, 0x39, // len=57

				0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
				0x01, // true

				0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
				0x03, 0xc0, 0x20, 0x80, // 8.25

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x0c, 0x52, 0x6f, 0x73, 0x65, 0x20, 0x4c, 0x61, 0x6c, 0x6f, 0x6e, 0x64, 0x65, // "Rose Lalonde"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x02, 0x01, 0x9d, // 413
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, filled
	t.Run("*(many fields)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = &testStructManyFields{
				Value:   413,
				Name:    "Rose Lalonde",
				Enabled: true,
				Factor:  8.25,

				hidden:  make(chan int),
				inc:     17,
				enabled: &sync.Mutex{},
			}
			expect = []byte{
				0x01, 0x39, // len=57

				0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
				0x01, // true

				0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
				0x03, 0xc0, 0x20, 0x80, // 8.25

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x0c, 0x52, 0x6f, 0x73, 0x65, 0x20, 0x4c, 0x61, 0x6c, 0x6f, 0x6e, 0x64, 0x65, // "Rose Lalonde"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x02, 0x01, 0x9d, // 413
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, nil
	t.Run("*(many fields), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testStructManyFields
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, filled
	t.Run("**(many fields)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr = &testStructManyFields{
				Value:   413,
				Name:    "Rose Lalonde",
				Enabled: true,
				Factor:  8.25,

				hidden:  make(chan int),
				inc:     17,
				enabled: &sync.Mutex{},
			}
			input  = &inputPtr
			expect = []byte{
				0x01, 0x39, // len=57

				0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
				0x01, // true

				0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
				0x03, 0xc0, 0x20, 0x80, // 8.25

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x0c, 0x52, 0x6f, 0x73, 0x65, 0x20, 0x4c, 0x61, 0x6c, 0x6f, 0x6e, 0x64, 0x65, // "Rose Lalonde"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x02, 0x01, 0x9d, // 413
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, nil at first level
	t.Run("**(many fields), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr *testStructManyFields
			input    = &inputPtr
			expect   = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// normal value test
	t.Run("with embedded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = testStructWithEmbedded{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Name: "NEPETA",
			}
			expect = []byte{
				0x01, 0x30, // len=48

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, filled
	t.Run("*(with embedded)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = &testStructWithEmbedded{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Name: "NEPETA",
			}
			expect = []byte{
				0x01, 0x30, // len=48

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, nil
	t.Run("*(with embedded), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testStructWithEmbedded
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, filled
	t.Run("**(with embedded)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr = &testStructWithEmbedded{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Name: "NEPETA",
			}
			input  = &inputPtr
			expect = []byte{
				0x01, 0x30, // len=48

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, nil at first level
	t.Run("**(with embedded), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr *testStructWithEmbedded
			input    = &inputPtr
			expect   = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// normal value test
	t.Run("with embedded overlap", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = testStructWithEmbeddedOverlap{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Value: 8.25,
				Name:  "NEPETA",
			}
			expect = []byte{
				0x01, 0x3c, // len=60

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x03, 0xc0, 0x20, 0x80, // 8.25
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, filled
	t.Run("*(with embedded overlap)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input = &testStructWithEmbeddedOverlap{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Value: 8.25,
				Name:  "NEPETA",
			}
			expect = []byte{
				0x01, 0x3c, // len=60

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x03, 0xc0, 0x20, 0x80, // 8.25
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, nil
	t.Run("*(with embedded overlap), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  *testStructWithEmbeddedOverlap
			expect = []byte{0xa0}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, filled
	t.Run("**(with embedded overlap)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr = &testStructWithEmbeddedOverlap{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Value: 8.25,
				Name:  "NEPETA",
			}
			input  = &inputPtr
			expect = []byte{
				0x01, 0x3c, // len=60

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x03, 0xc0, 0x20, 0x80, // 8.25
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, nil at first level
	t.Run("**(with embedded overlap), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			inputPtr *testStructWithEmbeddedOverlap
			input    = &inputPtr
			expect   = []byte{0xb0, 0x01, 0x01}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})
}

func Test_Dec_Struct(t *testing.T) {
	// we require an exported struct in order to test embedded struct fields.
	// we will declare it and other struct types it is embedded in here in this
	// function to avoid adding a new exported type to the rezi package.
	type TestStructToEmbed struct {
		Value int
	}
	type testStructWithEmbedded struct {
		TestStructToEmbed
		Name string
	}

	type testStructWithEmbeddedOverlap struct {
		TestStructToEmbed
		Name  string
		Value float64
	}

	t.Run("no-member struct", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructEmpty
			input          = []byte{0x00}
			expect         = testStructEmpty{}
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run("no-member struct, no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructEmpty
			input          = []byte{0x00}
			expect         testStructEmpty
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, filled
	t.Run("*(no-member struct)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         *testStructEmpty
			input          = []byte{0x00}
			expectVal      = testStructEmpty{}
			expect         = &expectVal
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*(no-member struct), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         = &testStructEmpty{} // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *testStructEmpty
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, filled
	t.Run("**(no-member struct)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         **testStructEmpty
			input          = []byte{0x00}
			expectVal      = testStructEmpty{}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**(no-member struct), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &testStructEmpty{}
			actual           = &actualInitialPtr // initially set to enshore it's cleared
			input            = []byte{0xb0, 0x01, 0x01}
			expectPtr        *testStructEmpty
			expect           = &expectPtr
			expectConsumed   = 3
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// normal value test
	t.Run("one-member struct", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual testStructOneMember
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expect         = testStructOneMember{Value: 4}
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run("one-member struct, no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructOneMember
			input          = []byte{0x00}
			expect         testStructOneMember
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, filled
	t.Run("*(one-member struct)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual *testStructOneMember
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal      = testStructOneMember{Value: 4}
			expect         = &expectVal
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*(one-member struct), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         = &testStructOneMember{Value: 4} // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *testStructOneMember
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, filled
	t.Run("**(one-member struct)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual **testStructOneMember
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal      = testStructOneMember{Value: 4}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**(one-member struct), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &testStructOneMember{Value: 4}
			actual           = &actualInitialPtr // initially set to enshore it's cleared
			input            = []byte{0xb0, 0x01, 0x01}
			expectPtr        *testStructOneMember
			expect           = &expectPtr
			expectConsumed   = 3
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// normal value test
	t.Run("multi-member struct", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual testStructMultiMember
			input  = []byte{
				0x01, 0x1a, // len=26

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expect         = testStructMultiMember{Value: 4, Name: "NEPETA"}
			expectConsumed = 28
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run("multi-member struct, no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructMultiMember
			input          = []byte{0x00}
			expect         testStructMultiMember
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, filled
	t.Run("*(multi-member struct)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual *testStructMultiMember
			input  = []byte{
				0x01, 0x1a, // len=26

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal      = testStructMultiMember{Value: 4, Name: "NEPETA"}
			expect         = &expectVal
			expectConsumed = 28
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*(multi-member struct), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         = &testStructMultiMember{Value: 4, Name: "NEPETA"} // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *testStructMultiMember
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, filled
	t.Run("**(multi-member struct)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual **testStructMultiMember
			input  = []byte{
				0x01, 0x1a, // len=26

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal      = testStructMultiMember{Value: 4, Name: "NEPETA"}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 28
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**(multi-member struct), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &testStructMultiMember{Value: 4, Name: "NEPETA"}
			actual           = &actualInitialPtr // initially set to enshore it's cleared
			input            = []byte{0xb0, 0x01, 0x01}
			expectPtr        *testStructMultiMember
			expect           = &expectPtr
			expectConsumed   = 3
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// normal value test
	t.Run("with unexported", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual testStructWithUnexported
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expect         = testStructWithUnexported{Value: 4, unexported: 0}
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run("with unexported, no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructWithUnexported
			input          = []byte{0x00}
			expect         testStructWithUnexported
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, filled
	t.Run("*(with unexported)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual *testStructWithUnexported
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal      = testStructWithUnexported{Value: 4, unexported: 0}
			expect         = &expectVal
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*(with unexported), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         = &testStructWithUnexported{Value: 4, unexported: 0} // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *testStructWithUnexported
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, filled
	t.Run("**(with unexported)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual **testStructWithUnexported
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal      = testStructWithUnexported{Value: 4, unexported: 0}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**(with unexported), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &testStructWithUnexported{Value: 4, unexported: 0}
			actual           = &actualInitialPtr // initially set to enshore it's cleared
			input            = []byte{0xb0, 0x01, 0x01}
			expectPtr        *testStructWithUnexported
			expect           = &expectPtr
			expectConsumed   = 3
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// normal value test
	t.Run("with unexported values set", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual = testStructWithUnexported{unexported: 12}
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expect         = testStructWithUnexported{Value: 4, unexported: 12}
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run("with unexported values set, no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         = testStructWithUnexported{unexported: 12}
			input          = []byte{0x00}
			expect         = testStructWithUnexported{unexported: 12}
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, filled
	t.Run("*(with unexported values set)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual = &testStructWithUnexported{unexported: 12}
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal      = testStructWithUnexported{Value: 4, unexported: 12}
			expect         = &expectVal
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*(with unexported values set), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         = &testStructWithUnexported{Value: 4, unexported: 12} // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *testStructWithUnexported
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, filled
	t.Run("**(with unexported values set)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual = ref(&testStructWithUnexported{unexported: 12})
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal      = testStructWithUnexported{Value: 4, unexported: 12}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**(with unexported values set), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &testStructWithUnexported{Value: 4, unexported: 12}
			actual           = &actualInitialPtr // initially set to enshore it's cleared
			input            = []byte{0xb0, 0x01, 0x01}
			expectPtr        *testStructWithUnexported
			expect           = &expectPtr
			expectConsumed   = 3
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// normal value test
	t.Run("with unexported case distinguished", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual testStructWithUnexportedCaseDistinguished
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expect         = testStructWithUnexportedCaseDistinguished{Value: 4, value: 0}
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run("with unexported case distinguished, no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructWithUnexportedCaseDistinguished
			input          = []byte{0x00}
			expect         testStructWithUnexportedCaseDistinguished
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, filled
	t.Run("*(with unexported case distinguished)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual *testStructWithUnexportedCaseDistinguished
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal      = testStructWithUnexportedCaseDistinguished{Value: 4, value: 0}
			expect         = &expectVal
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*(with unexported case distinguished), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         = &testStructWithUnexportedCaseDistinguished{Value: 4, value: 0} // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *testStructWithUnexportedCaseDistinguished
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, filled
	t.Run("**(with unexported case distinguished)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual **testStructWithUnexportedCaseDistinguished
			input  = []byte{
				0x01, 0x0a, // len=10

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal      = testStructWithUnexportedCaseDistinguished{Value: 4, value: 0}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 12
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**(with unexported case distinguished), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &testStructWithUnexportedCaseDistinguished{Value: 4, value: 0}
			actual           = &actualInitialPtr // initially set to enshore it's cleared
			input            = []byte{0xb0, 0x01, 0x01}
			expectPtr        *testStructWithUnexportedCaseDistinguished
			expect           = &expectPtr
			expectConsumed   = 3
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// normal value test
	t.Run("only unexported", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructOnlyUnexported
			input          = []byte{0x00}
			expect         = testStructOnlyUnexported{value: 0, name: ""}
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run("only unexported, no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructOnlyUnexported
			input          = []byte{0x00}
			expect         testStructOnlyUnexported
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, filled
	t.Run("*(only unexported)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         *testStructOnlyUnexported
			input          = []byte{0x00}
			expectVal      = testStructOnlyUnexported{value: 0, name: ""}
			expect         = &expectVal
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*(only unexported), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         = &testStructOnlyUnexported{value: 0, name: ""} // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *testStructOnlyUnexported
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, filled
	t.Run("**(only unexported)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         **testStructOnlyUnexported
			input          = []byte{0x00}
			expectVal      = testStructOnlyUnexported{value: 0, name: ""}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**(only unexported), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &testStructOnlyUnexported{value: 0, name: ""}
			actual           = &actualInitialPtr // initially set to enshore it's cleared
			input            = []byte{0xb0, 0x01, 0x01}
			expectPtr        *testStructOnlyUnexported
			expect           = &expectPtr
			expectConsumed   = 3
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// normal value test
	t.Run("many fields", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual testStructManyFields
			input  = []byte{
				0x01, 0x39, // len=57

				0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
				0x01, // true

				0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
				0x03, 0xc0, 0x20, 0x80, // 8.25

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x0c, 0x52, 0x6f, 0x73, 0x65, 0x20, 0x4c, 0x61, 0x6c, 0x6f, 0x6e, 0x64, 0x65, // "Rose Lalonde"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x02, 0x01, 0x9d, // 413
			}
			expect = testStructManyFields{
				Value:   413,
				Name:    "Rose Lalonde",
				Enabled: true,
				Factor:  8.25,

				hidden:  nil,
				inc:     0,
				enabled: nil,
			}
			expectConsumed = 59
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run("many fields, no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructManyFields
			input          = []byte{0x00}
			expect         testStructManyFields
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, filled
	t.Run("*(many fields)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual *testStructManyFields
			input  = []byte{
				0x01, 0x39, // len=57

				0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
				0x01, // true

				0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
				0x03, 0xc0, 0x20, 0x80, // 8.25

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x0c, 0x52, 0x6f, 0x73, 0x65, 0x20, 0x4c, 0x61, 0x6c, 0x6f, 0x6e, 0x64, 0x65, // "Rose Lalonde"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x02, 0x01, 0x9d, // 413
			}
			expectVal = testStructManyFields{
				Value:   413,
				Name:    "Rose Lalonde",
				Enabled: true,
				Factor:  8.25,

				hidden:  nil,
				inc:     0,
				enabled: nil,
			}
			expect         = &expectVal
			expectConsumed = 59
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*(many fields), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual = &testStructManyFields{
				Value:   413,
				Name:    "Rose Lalonde",
				Enabled: true,
				Factor:  8.25,

				hidden:  nil,
				inc:     0,
				enabled: nil,
			} // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *testStructManyFields
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, filled
	t.Run("**(many fields)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual **testStructManyFields
			input  = []byte{
				0x01, 0x39, // len=57

				0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
				0x01, // true

				0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
				0x03, 0xc0, 0x20, 0x80, // 8.25

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x0c, 0x52, 0x6f, 0x73, 0x65, 0x20, 0x4c, 0x61, 0x6c, 0x6f, 0x6e, 0x64, 0x65, // "Rose Lalonde"

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x02, 0x01, 0x9d, // 413
			}
			expectVal = testStructManyFields{
				Value:   413,
				Name:    "Rose Lalonde",
				Enabled: true,
				Factor:  8.25,

				hidden:  nil,
				inc:     0,
				enabled: nil,
			}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 59
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**(many fields), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &testStructManyFields{
				Value:   413,
				Name:    "Rose Lalonde",
				Enabled: true,
				Factor:  8.25,

				hidden:  nil,
				inc:     0,
				enabled: nil,
			}
			actual         = &actualInitialPtr // initially set to enshore it's cleared
			input          = []byte{0xb0, 0x01, 0x01}
			expectPtr      *testStructManyFields
			expect         = &expectPtr
			expectConsumed = 3
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// normal value test
	t.Run("with embedded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual testStructWithEmbedded
			input  = []byte{
				0x01, 0x30, // len=48

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expect = testStructWithEmbedded{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Name: "NEPETA",
			}
			expectConsumed = 50
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run("with embedded, no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructWithEmbedded
			input          = []byte{0x00}
			expect         testStructWithEmbedded
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, filled
	t.Run("*(with embedded)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual *testStructWithEmbedded
			input  = []byte{
				0x01, 0x30, // len=48

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal = testStructWithEmbedded{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Name: "NEPETA",
			}
			expect         = &expectVal
			expectConsumed = 50
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*(with embedded), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual = &testStructWithEmbedded{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Name: "NEPETA",
			} // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *testStructWithEmbedded
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, filled
	t.Run("**(with embedded)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual **testStructWithEmbedded
			input  = []byte{
				0x01, 0x30, // len=48

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4
			}
			expectVal = testStructWithEmbedded{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Name: "NEPETA",
			}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 50
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**(with embedded), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &testStructWithEmbedded{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Name: "NEPETA",
			}
			actual         = &actualInitialPtr // initially set to enshore it's cleared
			input          = []byte{0xb0, 0x01, 0x01}
			expectPtr      *testStructWithEmbedded
			expect         = &expectPtr
			expectConsumed = 3
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// normal value test
	t.Run("with embedded overlap", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual testStructWithEmbeddedOverlap
			input  = []byte{
				0x01, 0x3c, // len=60

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x03, 0xc0, 0x20, 0x80, // 8.25
			}
			expect = testStructWithEmbeddedOverlap{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Value: 8.25,
				Name:  "NEPETA",
			}
			expectConsumed = 62
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run("with embedded overlap, no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         testStructWithEmbeddedOverlap
			input          = []byte{0x00}
			expect         testStructWithEmbeddedOverlap
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, filled
	t.Run("*(with embedded overlap)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual *testStructWithEmbeddedOverlap
			input  = []byte{
				0x01, 0x3c, // len=60

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x03, 0xc0, 0x20, 0x80, // 8.25
			}
			expectVal = testStructWithEmbeddedOverlap{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Value: 8.25,
				Name:  "NEPETA",
			}
			expect         = &expectVal
			expectConsumed = 62
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*(with embedded overlap), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual = &testStructWithEmbeddedOverlap{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Value: 8.25,
				Name:  "NEPETA",
			} // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *testStructWithEmbeddedOverlap
			expectConsumed = 1
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, filled
	t.Run("**(with embedded overlap)", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual **testStructWithEmbeddedOverlap
			input  = []byte{
				0x01, 0x3c, // len=60

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

				0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
				0x01, 0x0a, // len=10
				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x01, 0x04, // 4

				0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
				0x03, 0xc0, 0x20, 0x80, // 8.25
			}
			expectVal = testStructWithEmbeddedOverlap{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Value: 8.25,
				Name:  "NEPETA",
			}
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = 62
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**(with embedded overlap), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &testStructWithEmbeddedOverlap{
				TestStructToEmbed: TestStructToEmbed{
					Value: 4,
				},
				Value: 8.25,
				Name:  "NEPETA",
			}
			actual         = &actualInitialPtr // initially set to enshore it's cleared
			input          = []byte{0xb0, 0x01, 0x01}
			expectPtr      *testStructWithEmbeddedOverlap
			expect         = &expectPtr
			expectConsumed = 3
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	t.Run("missing values in encoded are left alone", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual = testStructMultiMember{Value: 8, Name: "JOHN"}
			input  = []byte{
				0x01, 0x10, // len=16

				0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
				0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
			}
			expect         = testStructMultiMember{Value: 8, Name: "NEPETA"}
			expectConsumed = 18
		)

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})
}
