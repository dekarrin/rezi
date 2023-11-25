package rezi

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStructToEmbed struct {
	Value int
}

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

type testStructWithEmbedded struct {
	testStructToEmbed
	Name string
}

type testWithEmbeddedOverlap struct {
	testStructToEmbed
	Name  string
	Value float64
}

func Test_Enc_Struct(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		assert := assert.New(t)

		var (
			input  testStructEmpty
			expect = []byte{
				0x00,
			}
		)

		actual, err := Enc(input)
		if !assert.NoError(err) {
			return
		}

		assert.Equal(expect, actual)
	})
}
