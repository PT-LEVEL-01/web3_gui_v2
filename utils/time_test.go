package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeToken(*testing.T) {
	// timeTokenExample()
}

func timeTokenExample() {
	class1 := "1"
	class2 := "2"
	SetTimeToken(class1, time.Second)
	SetTimeToken(class1, time.Second)
	for i := 0; i < 1; i++ {
		time.Sleep(time.Second / 2)
		allow1 := GetTimeToken(class1, false)
		allow2 := GetTimeToken(class2, false)
		fmt.Println(allow1, allow2)
	}
}
