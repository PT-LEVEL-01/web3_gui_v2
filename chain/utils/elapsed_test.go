package utils

import (
	"testing"
	"time"
)

func TestElapsed(t *testing.T) {
	elapsed := NewElapsed(1)

	time.Sleep(time.Millisecond)
	elapsed.Step("记录1")

	time.Sleep(time.Millisecond)
	elapsed.Step("记录2")

	time.Sleep(time.Millisecond)
	elapsed.Step("记录3")

	elapsed.Print()
}
