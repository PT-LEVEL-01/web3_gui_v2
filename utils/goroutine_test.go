package utils

import (
	"testing"
)

func TestPanic(t *testing.T) {
	defer PrintPanicStack(Log)
	panic("haha")
}
