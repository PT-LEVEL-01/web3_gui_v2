package utils

import (
	"syscall"
	"unsafe"
)

/*
获取磁盘空间大小以及可用空间大小
@return    uint64    磁盘总大小
@return    uint64    磁盘可用空间大小
@return    uint64    磁盘可用空间大小
*/
func GetDiskFreeSpace(diskName string) (uint64, uint64, uint64, error) {
	// 磁盘
	// diskName := "E:"
	diskNameUtf16Ptr, _ := syscall.UTF16PtrFromString(diskName)
	// 一下参数类型需要跟API 的类型相符
	lpFreeBytesAvailable, lpTotalNumberOfBytes,
		lpTotalNumberOfFreeBytes := uint64(0), uint64(0), uint64(0)

	// 获取方法引用
	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		// panic("获取方法引用失败:")
		return 0, 0, 0, err
	}
	// 释放引用
	defer syscall.FreeLibrary(kernel32)

	getDisFreeSpaceEx, err := syscall.GetProcAddress(kernel32, "GetDiskFreeSpaceExW")
	if err != nil {
		// panic("失败1")
		return 0, 0, 0, err
	}

	// 根据参数个数使用对象SyscallN方法, 只需要4个参数
	r, _, errno := syscall.Syscall6(uintptr(getDisFreeSpaceEx), 4,
		uintptr(unsafe.Pointer(diskNameUtf16Ptr)), //
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)),
		0, 0)
	// 此处的errno不是error接口， 而是type Errno uintptr
	// MSDN GetDiskFreeSpaceEx function 文档说明：
	// Return value
	// 		If the function succeeds, the return value is nonzero.
	// 		If the function fails, the return value is zero (0). To get extended error information, call GetLastError.
	// 只要是0 就是错误
	// if r != 0 {
	// 	fmt.Printf("剩余空间 %d M.\n", lpFreeBytesAvailable)
	// 	fmt.Printf("用户可用总空间 %d G.\n", lpTotalNumberOfBytes)
	// 	fmt.Printf("剩余空间2 %d M.\n", lpTotalNumberOfFreeBytes)
	// } else {
	// 	fmt.Println("失败2")
	// 	panic(errno)
	// }

	// 只要是0 就是错误
	if r == 0 {
		return 0, 0, 0, errno
	}
	return lpTotalNumberOfBytes, lpFreeBytesAvailable, lpTotalNumberOfFreeBytes, nil
}
