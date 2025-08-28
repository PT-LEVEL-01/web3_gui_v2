package main

import (
	"fmt"
	"syscall"
)

func main() {
	example1()
}
func example1() {
	handler, err := syscall.LoadLibrary("peer.dll")
	fmt.Println("DLL文件加载完成", handler, err)
	fn, err := syscall.GetProcAddress(handler, "Start")
	fmt.Println("获得方法完成", fn, err)
	r1, r2, err := syscall.Syscall(fn, 0, 0, 0, 0)
	fmt.Println("执行结果", r1, r2, err)
}
