package evmutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type hashT struct {
	input string
	want  string
}

var mydata = []hashT{
	{"alice.eth", "0x787192fc5378cc32aa956ddfdedbf26b24e8d78e40109add0eea2c1a012c3dec"},
	{"foo.eth", "0xde9b09fd7c5f901e23a3f19fecc54828e9c848539801e86591bd9801b019f84f"},
	{"eth", "0x93cdeb708b7545dc668eb9280176169d1c33cfd8ed6f04690a0bcc88a93fc4ae"},
}

func TestNameHash(t *testing.T) {
	for _, test := range mydata {
		assert.Equal(t, test.want, NameHash(test.input).String())
	}
}
