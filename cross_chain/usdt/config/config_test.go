package config

import (
	"fmt"
	"testing"
)

func TestCheckBlockConfirmationNumber(t *testing.T) {
	for i := range 10 {
		ok := CheckBlockConfirmationNumber(1, uint64(5+i), 10)
		fmt.Println(5+i, ok)
	}
}
