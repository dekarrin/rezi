package rezi

import (
	"sync"
	"testing"
)

func Test_Enc_Struct(t *testing.T) {
	type toEmbed struct {
		Value int
	}

	type testEmptyStruct struct{}

	type testOneMemberStruct struct {
		Value int
	}

	type testMultiMemberStruct struct {
		Value int
		Name  string
	}

	type testGoodStructWithUnexported struct {
		Value      int
		unexported float32
	}

	type testGoodStructWithUnexportedCase struct {
		Value int
		value float32
	}

	type testOnlyUnexported struct {
		value int
		name  string
	}

	type testManyFields struct {
		Name    string
		Factor  float64
		Value   int
		Enabled bool

		hidden  chan int
		inc     int
		enabled *sync.Mutex
	}

	type testWithEmbedded struct {
		toEmbed
		Name string
	}

	type testWithEmbeddedOverlap struct {
		toEmbed
		Name  string
		Value float64
	}
}
