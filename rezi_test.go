package rezi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Enc_String(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect []byte
	}{
		{
			name: "empty",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			actual := Enc(tc.input)
			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_Enc_Int(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}
func Test_Enc_Bool(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Enc_Binary(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Enc_Map(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Dec_Slice(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Dec_String(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Dec_Int(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}
func Test_Dec_Bool(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Dec_Binary(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Dec_Map(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}

func Test_Enc_Slice(t *testing.T) {
	testCases := []struct {
		name string
	}{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// assert := assert.New(t)
		})
	}

}
