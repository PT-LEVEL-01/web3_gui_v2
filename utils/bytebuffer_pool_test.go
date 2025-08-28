package utils

import (
	"fmt"
	"testing"
)

func TestBufferBytes(*testing.T) {
	// byteBuffPoolExample()
}

func byteBuffPoolExample() {
	bs1 := []byte("hello")
	buf := NewBufferByte(0)
	buf.Write(&bs1)
	fmt.Println(buf.Bytes())
	buf.Write(&bs1)
	fmt.Println(buf.Bytes())
	buf.Clean()
	buf.Write(&bs1)
	fmt.Println(buf.Bytes())
	buf.Clean()
	buf.SetLength(100)
	buf.Write(&bs1)
	fmt.Println(buf.Bytes())
}
