package utils

import (
	"fmt"
	"runtime"
	"testing"
)

func TestDiskSpace(t *testing.T) {
	// getDiskFreeSpace()
}

func getDiskFreeSpace() {
	diskPath := ""
	if runtime.GOOS == "linux" {
		diskPath = "/mnt/e"
	}
	if runtime.GOOS == "windows" {
		diskPath = "E:"
	}
	totalAll, _, free, err := GetDiskFreeSpace(diskPath)
	if err != nil {
		fmt.Println("", err.Error())
		return
	}
	fmt.Println("磁盘总容量:", totalAll/1024/1024, "M")
	fmt.Println("剩余可用容量:", free/1024/1024, "M")
}
