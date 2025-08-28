package utils

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestTimetask(t *testing.T) {
	// timetaskExample()
}

func timetaskExample() {
	for i := 0; i < 10; i++ {
		fmt.Println("NumGoroutine start:", runtime.NumGoroutine())
		task := NewTask(taskFuncTemp)
		task.Add(time.Now().Add(time.Second*10).Unix(), "", nil)
		task.Destroy()
		fmt.Println("NumGoroutine end:", runtime.NumGoroutine())
	}
}

func taskFuncTemp(class string, params []byte) {

}
