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

	runEncTests(t, "empty struct", testStructEmpty{}, []byte{0x00})
	runEncTests(t, "one member", testStructOneMember{Value: 4}, []byte{
		0x01, 0x0a, // len=10

		0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
		0x01, 0x04, // 4
	})
	runEncTests(t, "multi member", testStructMultiMember{Value: 4, Name: "NEPETA"}, []byte{
		0x01, 0x1a, // len=26

		0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
		0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

		0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
		0x01, 0x04, // 4
	})
	runEncTests(t, "with unexported", testStructWithUnexported{Value: 4, unexported: 9}, []byte{
		0x01, 0x0a, // len=10

		0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
		0x01, 0x04, // 4
	})
	runEncTests(t, "with unexported case distinguished", testStructWithUnexportedCaseDistinguished{Value: 4, value: 3.2}, []byte{
		0x01, 0x0a, // len=10

		0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
		0x01, 0x04, // 4
	})
	runEncTests(t, "only unexported", testStructOnlyUnexported{value: 3, name: "test"}, []byte{0x00})
	runEncTests(t, "many fields", testStructManyFields{
		Value:   413,
		Name:    "Rose Lalonde",
		Enabled: true,
		Factor:  8.25,

		hidden:  make(chan int),
		inc:     17,
		enabled: &sync.Mutex{},
	}, []byte{
		0x01, 0x39, // len=57

		0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
		0x01, // true

		0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
		0x03, 0xc0, 0x20, 0x80, // 8.25

		0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
		0x41, 0x82, 0x0c, 0x52, 0x6f, 0x73, 0x65, 0x20, 0x4c, 0x61, 0x6c, 0x6f, 0x6e, 0x64, 0x65, // "Rose Lalonde"

		0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
		0x02, 0x01, 0x9d, // 413
	})
	runEncTests(t, "with embedded", testStructWithEmbedded{
		TestStructToEmbed: TestStructToEmbed{
			Value: 4,
		},
		Name: "NEPETA",
	}, []byte{
		0x01, 0x30, // len=48

		0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
		0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

		0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
		0x01, 0x0a, // len=10
		0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
		0x01, 0x04, // 4
	})
	runEncTests(t, "with embedded overlap", testStructWithEmbeddedOverlap{
		TestStructToEmbed: TestStructToEmbed{
			Value: 4,
		},
		Value: 8.25,
		Name:  "NEPETA",
	}, []byte{
		0x01, 0x3c, // len=60

		0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
		0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

		0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
		0x01, 0x0a, // len=10
		0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
		0x01, 0x04, // 4

		0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
		0x03, 0xc0, 0x20, 0x80, // 8.25
	})
}

func runEncTests[E any](t *testing.T, name string, inputVal E, expect []byte) {
	// normal value test
	t.Run(name, func(t *testing.T) {
		assert := assert.New(t)

		input := inputVal

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, filled
	t.Run("*("+name+")", func(t *testing.T) {
		assert := assert.New(t)

		input := &inputVal

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// single pointer, nil
	t.Run("*("+name+"), nil", func(t *testing.T) {
		assert := assert.New(t)

		var input *E
		var nilExp = []byte{0xa0}

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(nilExp, actual)
	})

	// double pointer, filled
	t.Run("**("+name+")", func(t *testing.T) {
		assert := assert.New(t)

		inputPtr := &inputVal
		input := &inputPtr

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})

	// double pointer, nil at first level
	t.Run("**("+name+"), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var inputPtr *E
		var input = &inputPtr
		var nilFirstLevelExp = []byte{0xb0, 0x01, 0x01}

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(nilFirstLevelExp, actual)
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

	// runDecTests(t, "no-member struct", []byte{0x00}, testStructEmpty{}, 1, nil)

	// runDecTests(t, "one-member struct", []byte{
	// 	0x01, 0x0a, // len=10

	// 	0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
	// 	0x01, 0x04, // 4
	// }, testStructOneMember{Value: 4}, 12, nil)

	// runDecTests(t, "multi-member struct", []byte{
	// 	0x01, 0x1a, // len=26

	// 	0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
	// 	0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

	// 	0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
	// 	0x01, 0x04, // 4
	// }, testStructMultiMember{Value: 4, Name: "NEPETA"}, 28, nil)

	// runDecTests(t, "with unexported", []byte{
	// 	0x01, 0x0a, // len=10

	// 	0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
	// 	0x01, 0x04, // 4
	// }, testStructWithUnexported{Value: 4, unexported: 0}, 12, nil)

	runDecTests(t, "with unexported values set", []byte{
		0x01, 0x0a, // len=10

		0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
		0x01, 0x04, // 4
	}, testStructWithUnexported{Value: 4, unexported: 12}, 12, &testStructWithUnexported{unexported: 12})

	// runDecTests(t, "with unexported case distinguished", []byte{
	// 	0x01, 0x0a, // len=10

	// 	0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
	// 	0x01, 0x04, // 4
	// }, testStructWithUnexportedCaseDistinguished{Value: 4, value: 0}, 12, nil)

	// runDecTests(t, "only unexported", []byte{0x00}, testStructOnlyUnexported{value: 0, name: ""}, 1, nil)

	// runDecTests(t, "many fields", []byte{
	// 	0x01, 0x39, // len=57

	// 	0x41, 0x82, 0x07, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, // "Enabled"
	// 	0x01, // true

	// 	0x41, 0x82, 0x06, 0x46, 0x61, 0x63, 0x74, 0x6f, 0x72, // "Factor"
	// 	0x03, 0xc0, 0x20, 0x80, // 8.25

	// 	0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
	// 	0x41, 0x82, 0x0c, 0x52, 0x6f, 0x73, 0x65, 0x20, 0x4c, 0x61, 0x6c, 0x6f, 0x6e, 0x64, 0x65, // "Rose Lalonde"

	// 	0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
	// 	0x02, 0x01, 0x9d, // 413
	// }, testStructManyFields{
	// 	Value:   413,
	// 	Name:    "Rose Lalonde",
	// 	Enabled: true,
	// 	Factor:  8.25,

	// 	hidden:  nil,
	// 	inc:     0,
	// 	enabled: nil,
	// }, 59, nil)

	// runDecTests(t, "with embedded", []byte{
	// 	0x01, 0x30, // len=48

	// 	0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
	// 	0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

	// 	0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
	// 	0x01, 0x0a, // len=10
	// 	0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
	// 	0x01, 0x04, // 4
	// }, testStructWithEmbedded{
	// 	TestStructToEmbed: TestStructToEmbed{
	// 		Value: 4,
	// 	},
	// 	Name: "NEPETA",
	// }, 50, nil)

	// runDecTests(t, "with embedded overlap", []byte{
	// 	0x01, 0x3c, // len=60

	// 	0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
	// 	0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"

	// 	0x41, 0x82, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x6f, 0x45, 0x6d, 0x62, 0x65, 0x64, // "TestStructToEmbed"
	// 	0x01, 0x0a, // len=10
	// 	0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
	// 	0x01, 0x04, // 4

	// 	0x41, 0x82, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, // "Value"
	// 	0x03, 0xc0, 0x20, 0x80, // 8.25
	// }, testStructWithEmbeddedOverlap{
	// 	TestStructToEmbed: TestStructToEmbed{
	// 		Value: 4,
	// 	},
	// 	Value: 8.25,
	// 	Name:  "NEPETA",
	// }, 62, nil)

	// // specialized test we cannot abstract easily
	// t.Run("missing values in encoded are set to default", func(t *testing.T) {
	// 	assert := assert.New(t)

	// 	var (
	// 		actual = testStructMultiMember{Value: 8, Name: "JOHN"}
	// 		input  = []byte{
	// 			0x01, 0x10, // len=16

	// 			0x41, 0x82, 0x04, 0x4e, 0x61, 0x6d, 0x65, // "Name"
	// 			0x41, 0x82, 0x06, 0x4e, 0x45, 0x50, 0x45, 0x54, 0x41, // "NEPETA"
	// 		}
	// 		expect         = testStructMultiMember{Value: 0, Name: "NEPETA"}
	// 		expectConsumed = 18
	// 	)

	// 	consumed, err := Dec(input, &actual)
	// 	if !assert.NoError(err) {
	// 		return
	// 	}

	// 	assert.Equal(expect, actual, "value mismatch")
	// 	assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	// })
}

// expectConsumed used only in sub-tests where expect is the actual expected.
//
// if initVal is nil it will be set to an empty value
func runDecTests[E any](t *testing.T, name string, filledInput []byte, filledExpect E, filledExpectConsumed int, initVal *E) {
	// normal value test
	t.Run(name, func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         E
			input          = filledInput
			expect         = filledExpect
			expectConsumed = filledExpectConsumed
		)

		if initVal != nil {
			actual = *initVal
		}

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// 0-len struct
	t.Run(name+", no values encoded", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         = filledExpect // initially set to enshore it is cleared
			input          = []byte{0x00}
			expect         E
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
	t.Run("*("+name+")", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         *E
			input          = filledInput
			expectVal      = filledExpect
			expect         = &expectVal
			expectConsumed = filledExpectConsumed
		)

		if initVal != nil {
			actual = new(E)
			*actual = (*initVal)
		}

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// single pointer, nil
	t.Run("*("+name+"), nil", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         = &filledExpect // initially set to enshore it's cleared
			input          = []byte{0xa0}
			expect         *E
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
	t.Run("**("+name+")", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actual         **E
			input          = filledInput
			expectVal      = filledExpect
			expectPtr      = &expectVal
			expect         = &expectPtr
			expectConsumed = filledExpectConsumed
		)

		if initVal != nil {
			actual = new(*E)
			*actual = new(E)
			**actual = (*initVal)
		}

		consumed, err := Dec(input, &actual)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual, "value mismatch")
		assert.Equal(expectConsumed, consumed, "consumed bytes mismatch")
	})

	// double pointer, nil at first level
	t.Run("**("+name+"), nil at first level", func(t *testing.T) {
		assert := assert.New(t)

		var (
			actualInitialPtr = &filledExpect
			actual           = &actualInitialPtr // initially set to enshore it's cleared
			input            = []byte{0xb0, 0x01, 0x01}
			expectPtr        *E
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
}
